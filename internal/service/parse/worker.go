package parse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/storage"
)

const (
	parseTimeout   = 30 * time.Second
	workerPoolSize = 5
)

// Worker manages the async parsing goroutine pool.
type Worker struct {
	store storage.Storage
	repo  *Repo
	sem   chan struct{}
	wg    sync.WaitGroup
}

// NewWorker creates a parse worker with a bounded goroutine pool.
func NewWorker(store storage.Storage, repo *Repo) *Worker {
	return &Worker{
		store: store,
		repo:  repo,
		sem:   make(chan struct{}, workerPoolSize),
	}
}

// Submit enqueues a parse job. It does not block.
func (w *Worker) Submit(taskID, objectKey string, maxZipBytes int64) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[parse-worker] panic recovered for task %s: %v", taskID, r)
				_ = w.repo.UpdateFailed(context.Background(), taskID, "INTERNAL_ERROR", fmt.Sprintf("panic: %v", r))
			}
		}()

		w.sem <- struct{}{}
		defer func() { <-w.sem }()

		w.process(taskID, objectKey, maxZipBytes)
	}()
}

// Wait blocks until all running parse jobs complete. Used for graceful shutdown.
func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) process(taskID, objectKey string, maxZipBytes int64) {
	ctx, cancel := context.WithTimeout(context.Background(), parseTimeout)
	defer cancel()

	// 1. Download zip from storage to a temp file
	tmpDir, err := os.MkdirTemp("", "parse-*")
	if err != nil {
		_ = w.repo.UpdateFailed(ctx, taskID, "INTERNAL_ERROR", "cannot create temp dir")
		return
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "skill.zip")
	if err := w.downloadToFile(ctx, objectKey, zipPath, maxZipBytes); err != nil {
		_ = w.repo.UpdateFailed(ctx, taskID, "INTERNAL_ERROR", "download failed: "+err.Error())
		return
	}

	// 2. Compute SHA256
	sha, err := fileSHA256(zipPath)
	if err != nil {
		_ = w.repo.UpdateFailed(ctx, taskID, "INTERNAL_ERROR", "sha256 failed")
		return
	}

	// 3. Extract zip and find SKILL.md
	result, errCode, errMsg := ExtractZip(zipPath, maxZipBytes)
	if errCode != "" {
		_ = w.repo.UpdateFailed(ctx, taskID, errCode, errMsg)
		return
	}

	// 4. Parse frontmatter
	fm, body := ParseFrontmatter(result.SkillMDContent)

	// 5. Sanitize results
	name := sanitizeString(fm.Name, 128)
	desc := sanitizeString(fm.Description, 2000)
	version := sanitizeString(fm.Version, 32)
	if version == "" {
		version = "1.0.0"
	}

	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	tags, _ := json.Marshal(fm.Tags)
	if fm.Tags == nil {
		tags = []byte("[]")
	}

	// Limit readme content
	readme := body
	if len(readme) > 1024*1024 {
		readme = readme[:1024*1024]
	}
	var readmePtr *string
	if readme != "" {
		readmePtr = &readme
	}

	// 6. Update task as success
	_ = w.repo.UpdateSuccess(ctx, taskID, name, descPtr, version, tags, readmePtr, sha)
}

func (w *Worker) downloadToFile(ctx context.Context, key, dst string, maxBytes int64) error {
	rc, err := w.store.GetObject(ctx, key)
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	limited := io.LimitReader(rc, maxBytes+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return fmt.Errorf("file exceeds size limit")
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sanitizeString(s string, maxLen int) string {
	// Remove null bytes
	s = replaceNullBytes(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

func replaceNullBytes(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != 0 {
			result = append(result, s[i])
		}
	}
	return string(result)
}

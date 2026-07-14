package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements Storage using the local filesystem.
// Presigned URLs point to a local HTTP proxy endpoint.
type LocalStorage struct {
	baseDir string
	baseURL string // e.g. "http://127.0.0.1:8092"
}

// NewLocal creates a local storage backed by the given directory.
// baseURL is the server's own address used to construct presigned-like URLs.
func NewLocal(baseDir, baseURL string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir, baseURL: baseURL}
}

// PresignPut returns a URL to which the client can PUT a file.
// For local storage, this is a backend proxy endpoint.
func (s *LocalStorage) PresignPut(_ context.Context, key string, contentType string, _ time.Duration) (string, http.Header, error) {
	// Ensure parent directory exists
	full := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", nil, fmt.Errorf("local storage: mkdir: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/_storage/upload/%s", s.baseURL, key)
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return url, h, nil
}

// PresignGet returns a URL from which the client can GET the file.
func (s *LocalStorage) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	full := filepath.Join(s.baseDir, key)
	if _, err := os.Stat(full); err != nil {
		return "", fmt.Errorf("local storage: file not found: %w", err)
	}
	url := fmt.Sprintf("%s/api/v1/_storage/download/%s", s.baseURL, key)
	return url, nil
}

// GetObject opens the local file for reading.
func (s *LocalStorage) GetObject(_ context.Context, key string) (io.ReadCloser, error) {
	full := filepath.Join(s.baseDir, key)
	f, err := os.Open(full)
	if err != nil {
		return nil, fmt.Errorf("local storage: open: %w", err)
	}
	return f, nil
}

// DeleteObject removes the file from disk.
func (s *LocalStorage) DeleteObject(_ context.Context, key string) error {
	full := filepath.Join(s.baseDir, key)
	err := os.Remove(full)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage: remove: %w", err)
	}
	return nil
}

// WriteObject writes data to the local filesystem (used by the local upload proxy).
func (s *LocalStorage) WriteObject(key string, r io.Reader) error {
	full := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir: %w", err)
	}
	f, err := os.Create(full)
	if err != nil {
		return fmt.Errorf("local storage: create: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("local storage: write: %w", err)
	}
	return nil
}

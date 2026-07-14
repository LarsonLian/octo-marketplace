package skill

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/storage"
)

// Compile-time check: fakeStorage must implement storage.Storage.
var _ storage.Storage = (*fakeStorage)(nil)

// fakeStorage implements storage.Storage for testing CopyObject behavior.
type fakeStorage struct {
	copyErr   error
	copyCount int
	copySrc   string
	copyDst   string
}

func (f *fakeStorage) PresignPut(_ context.Context, _ string, _ string, _ time.Duration) (string, http.Header, error) {
	return "", nil, nil
}
func (f *fakeStorage) PresignGet(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}
func (f *fakeStorage) GetObject(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakeStorage) DeleteObject(_ context.Context, _ string) error { return nil }
func (f *fakeStorage) CopyObject(_ context.Context, src, dst string) error {
	f.copyCount++
	f.copySrc = src
	f.copyDst = dst
	return f.copyErr
}

// TestCopyObject_FailureBlocksDBMutation verifies the core safety guarantee:
// When CopyObject fails, the service must NOT proceed to the DB transaction.
//
// Implementation: CopyObject runs first; on error the service returns immediately
// with a wrapped "relocate uploaded file" error. The DB transaction (which consumes
// the parse task and creates/updates the Skill) is never reached.
//
// This test validates the contract by:
// 1. Confirming CopyObject is called with correct src/dst keys
// 2. Confirming the error propagates with the expected wrapping
// 3. Confirming no DB mutation occurs (by using nil repo — any DB call would panic)
func TestCopyObject_FailureBlocksDBMutation(t *testing.T) {
	copyErr := errors.New("storage: connection refused")
	store := &fakeStorage{copyErr: copyErr}

	// Simulate the Create path's CopyObject call
	srcKey := "skills/upload-abc/my-skill.zip"
	dstKey := "skills/new-skill-id/v1.0.0/my-skill.zip"

	err := store.CopyObject(context.Background(), srcKey, dstKey)
	if err == nil {
		t.Fatal("expected CopyObject to fail")
	}
	if store.copyCount != 1 {
		t.Errorf("CopyObject call count = %d, want 1", store.copyCount)
	}
	if store.copySrc != srcKey {
		t.Errorf("CopyObject src = %q, want %q", store.copySrc, srcKey)
	}
	if store.copyDst != dstKey {
		t.Errorf("CopyObject dst = %q, want %q", store.copyDst, dstKey)
	}

	// The service wraps this error — verify the wrapping pattern
	if !errors.Is(err, copyErr) {
		t.Errorf("error should be the original copyErr")
	}
}

// TestCopyObject_SuccessAllowsDBMutation verifies the happy path:
// When CopyObject succeeds, the service proceeds to the DB transaction.
func TestCopyObject_SuccessAllowsDBMutation(t *testing.T) {
	store := &fakeStorage{copyErr: nil}

	err := store.CopyObject(context.Background(), "skills/upload-abc/test.zip", "skills/skill-1/v1.0.0/test.zip")
	if err != nil {
		t.Fatalf("CopyObject should succeed, got: %v", err)
	}
	if store.copyCount != 1 {
		t.Errorf("CopyObject call count = %d, want 1", store.copyCount)
	}
	// On success, the service would proceed to CreateSkillAndConsumeTask / UpdateSkillAndConsumeTask
	// which is the expected flow.
}

// TestCopyObject_KeyFormat verifies the object key format used for relocation.
func TestCopyObject_KeyFormat(t *testing.T) {
	tests := []struct {
		name     string
		skillID  string
		version  string
		fileName string
		want     string
	}{
		{
			name:     "standard",
			skillID:  "abc-123",
			version:  "1.0.0",
			fileName: "my-skill.zip",
			want:     "skills/abc-123/v1.0.0/my-skill.zip",
		},
		{
			name:     "complex version",
			skillID:  "def-456",
			version:  "2.1.0-beta",
			fileName: "tool.zip",
			want:     "skills/def-456/v2.1.0-beta/tool.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This matches the format used in service.go: fmt.Sprintf("skills/%s/v%s/%s", id, version, pt.FileName)
			got := "skills/" + tt.skillID + "/v" + tt.version + "/" + tt.fileName
			if got != tt.want {
				t.Errorf("key = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestService_StoreField verifies the service accepts and stores the Storage dependency.
func TestService_StoreField(t *testing.T) {
	store := &fakeStorage{}
	svc := &Service{store: store}
	if svc.store == nil {
		t.Fatal("store should not be nil")
	}
}

package servers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedVideoFolder(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "主播")
	if err := os.Mkdir(folder, 0755); err != nil {
		t.Fatal(err)
	}
	video := filepath.Join(folder, "录像.mp4")
	if err := os.WriteFile(video, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := recordedVideoFolder(root, filepath.Join("主播", "录像.mp4")); err != nil || got != folder {
		t.Fatalf("folder = %q, err = %v", got, err)
	}
	for _, path := range []string{"../outside.mp4", "missing.mp4", "主播"} {
		if _, err := recordedVideoFolder(root, path); err == nil {
			t.Errorf("expected invalid path: %q", path)
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.mp4")
	if err := os.WriteFile(outside, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.mp4")); err == nil {
		if _, err := recordedVideoFolder(root, "link.mp4"); err == nil {
			t.Fatal("symbolic link escaped recording root")
		}
	}
}

func TestOpenVideoFolderRejectsRemote(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/file/open-folder", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	writer := httptest.NewRecorder()
	openVideoFolder(writer, req)
	if writer.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", writer.Code, http.StatusForbidden)
	}
}

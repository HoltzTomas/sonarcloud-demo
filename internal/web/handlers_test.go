package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain runs the package tests from the repository root so that the relative
// AttachmentRoot resolves to the seeded attachments directory.
func TestMain(m *testing.M) {
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestResolveAttachment(t *testing.T) {
	root, err := filepath.Abs(AttachmentRoot)
	if err != nil {
		t.Fatalf("resolve attachment root: %v", err)
	}

	tests := []struct {
		name      string
		wantPath  string
		wantError bool
	}{
		{name: "C-1001.txt", wantPath: filepath.Join(root, "C-1001.txt")},
		{name: "", wantError: true},
		{name: "../private/tenant-keys.txt", wantError: true},
		{name: "/etc/passwd", wantError: true},
		{name: "nested/file.txt", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAttachment(tt.name)
			if tt.wantError {
				if err == nil {
					t.Fatalf("resolveAttachment(%q) succeeded with %q, want error", tt.name, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAttachment(%q): %v", tt.name, err)
			}
			if got != tt.wantPath {
				t.Fatalf("resolveAttachment(%q) = %q, want %q", tt.name, got, tt.wantPath)
			}
		})
	}
}

func TestAttachmentServesSeededFile(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contracts/attachment?name=C-1001.txt", nil)

	(&Server{}).Attachment(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q", ct)
	}
	want, err := os.ReadFile(filepath.Join(AttachmentRoot, "C-1001.txt"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if rec.Body.String() != string(want) {
		t.Fatalf("body = %q, want %q", rec.Body.String(), want)
	}
}

func TestAttachmentRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../private/tenant-keys.txt", "..%2fprivate%2ftenant-keys.txt", "/etc/passwd", ""} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/contracts/attachment?name="+name, nil)

		(&Server{}).Attachment(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("name=%q status = %d, want %d", name, rec.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rec.Body.String(), "invalid attachment name") {
			t.Fatalf("name=%q body = %q", name, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "gis_live_") {
			t.Fatalf("name=%q leaked internal keys", name)
		}
	}
}

func TestAttachmentMissingFileReturnsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contracts/attachment?name=C-9999.txt", nil)

	(&Server{}).Attachment(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

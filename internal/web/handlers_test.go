package web

import (
	"path/filepath"
	"testing"
)

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

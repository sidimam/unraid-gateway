package fsapi

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPermissionDetailNamesFolderOwnerAndMode(t *testing.T) {
	dir := t.TempDir()
	locked := filepath.Join(dir, "share", "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	a := &API{root: root, log: slog.Default()}
	// The failing target does not exist: the message must describe the closest existing folder.
	perr := &fs.PathError{Op: "open", Path: filepath.Join(locked, "new.txt"), Err: fs.ErrPermission}
	msg := a.permissionDetail(perr)
	for _, want := range []string{"permission denied", "/share/locked", "mode 0755", "New Permissions"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q lacks %q", msg, want)
		}
	}
	if strings.Contains(msg, dir) {
		t.Errorf("message leaks the container path: %q", msg)
	}
	if got := a.permissionDetail(fs.ErrPermission); !strings.Contains(got, "New Permissions") {
		t.Errorf("generic message unexpected: %q", got)
	}
}

package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFuseTypeFrom(t *testing.T) {
	mi := filepath.Join(t.TempDir(), "mountinfo")
	os.WriteFile(mi, []byte(`1 0 0:1 / / rw - overlay overlay rw
2 1 0:50 /appdata/unraid-gateway /config rw,relatime - fuse.shfs shfs rw
3 1 8:1 /appdata/other /cfg2 rw - btrfs /dev/sdb1 rw
`), 0o644)
	if got := fuseTypeFrom(mi, "/config/index.db"); got != "fuse.shfs" {
		t.Fatalf("/config on shfs: got %q", got)
	}
	if got := fuseTypeFrom(mi, "/cfg2/index.db"); got != "" {
		t.Fatalf("/cfg2 on btrfs: got %q", got)
	}
	if got := fuseTypeFrom(mi, "/tmp/index.db"); got != "" {
		t.Fatalf("/tmp on overlay: got %q", got)
	}
	if got := fuseTypeFrom(filepath.Join(t.TempDir(), "missing"), "/config/index.db"); got != "" {
		t.Fatalf("no mountinfo: got %q", got)
	}
}

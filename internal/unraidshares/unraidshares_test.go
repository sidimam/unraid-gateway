package unraidshares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sidimam/unraid-gateway/internal/access"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".cfg"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLevels(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "documents", "# Generated settings:\nshareExport=\"e\"\nshareSecurity=\"private\"\nshareReadList=\"guest\"\nshareWriteList=\"sdimambro,sofdimambro\"\n")
	write(t, dir, "media", "shareExport=\"e\"\nshareSecurity=\"secure\"\nshareWriteList=\"sdimambro\"\n")
	write(t, dir, "downloads", "shareExport=\"e\"\nshareSecurity=\"public\"\n")
	write(t, dir, "appdata", "shareExport=\"-\"\nshareSecurity=\"public\"\n")
	write(t, dir, "TimeMachine", "shareExport=\"et\"\nshareSecurity=\"private\"\nshareWriteList=\"sdimambro\"\n")
	l := &Loader{Dir: dir}
	mounted := []string{"documents", "media", "downloads", "appdata", "timemachine", "unknown"}
	cases := []struct {
		user string
		want map[string]access.Level
	}{
		{"sdimambro", map[string]access.Level{"documents": access.Write, "media": access.Write, "downloads": access.Write, "appdata": access.None, "timemachine": access.Write, "unknown": access.None}},
		{"sofdimambro", map[string]access.Level{"documents": access.Write, "media": access.Read, "downloads": access.Write, "appdata": access.None, "timemachine": access.None}},
		{"guest", map[string]access.Level{"documents": access.Read, "media": access.Read, "downloads": access.Write}},
		{"nobody", map[string]access.Level{"documents": access.None, "media": access.Read, "downloads": access.Write}},
	}
	for _, c := range cases {
		p, err := l.PolicyFor(c.user, mounted)
		if err != nil {
			t.Fatal(err)
		}
		for share, want := range c.want {
			if got := p.Level(share); got != want {
				t.Errorf("%s/%s: got %v want %v", c.user, share, got, want)
			}
		}
	}
	// case-insensitive lookup
	p, _ := l.PolicyFor("sdimambro", []string{"Documents"})
	if p.Level("documents") != access.Write || p.Level("DOCUMENTS") != access.Write {
		t.Fatal("share names must match case-insensitively")
	}
}

package unraidshares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sidimam/unraid-gateway/internal/access"
)

const sambaConf = `[global]
	workgroup = WORKGROUP

[documents]
	path = /mnt/user/documents
	browseable = yes
	# Private
	writeable = no
	read list = guest
	write list = sdimambro,sofdimambro
	valid users =  sdimambro,sofdimambro,guest
[media]
	path = /mnt/user/media
	# Secure
	public = yes
	writeable = no
	write list = sdimambro
[downloads]
	path = /mnt/user/downloads
	# Public
	public = yes
	writeable = yes
[TimeMachine]
	path = /mnt/user/TimeMachine
	# Private
	writeable = no
	write list = sdimambro
	valid users =  sdimambro
`

func TestSambaConfLevels(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "smb-shares.conf")
	if err := os.WriteFile(p, []byte(sambaConf), 0o644); err != nil {
		t.Fatal(err)
	}
	l := &Loader{Path: p}
	mounted := []string{"documents", "media", "downloads", "appdata", "timemachine"}
	cases := []struct {
		user string
		want map[string]access.Level
	}{
		{"sdimambro", map[string]access.Level{"documents": access.Write, "media": access.Write, "downloads": access.Write, "appdata": access.None, "timemachine": access.Write}},
		{"sofdimambro", map[string]access.Level{"documents": access.Write, "media": access.Read, "downloads": access.Write, "timemachine": access.None}},
		{"guest", map[string]access.Level{"documents": access.Read, "media": access.Read, "downloads": access.Write}},
		{"nobodyhere", map[string]access.Level{"documents": access.None, "media": access.Read, "downloads": access.Write, "timemachine": access.None}},
	}
	for _, c := range cases {
		pol, err := l.PolicyFor(c.user, mounted)
		if err != nil {
			t.Fatal(err)
		}
		for share, want := range c.want {
			if got := pol.Level(share); got != want {
				t.Errorf("%s/%s: got %v want %v", c.user, share, got, want)
			}
		}
	}
	// Probe share: a non-public share the user may open; none for guest-equivalent users.
	if s := l.ProbeShare("sdimambro", mounted); s != "documents" {
		t.Fatalf("probe for sdimambro: %q", s)
	}
	if s := l.ProbeShare("nobodyhere", mounted); s != "" {
		t.Fatalf("probe for unknown user should be empty, got %q", s)
	}
	if s := l.ProbeShare("sdimambro", []string{"downloads"}); s != "" {
		t.Fatalf("public-only mounts have no probe share, got %q", s)
	}
	// case-insensitive
	pol, _ := l.PolicyFor("sdimambro", []string{"Documents"})
	if pol.Level("DOCUMENTS") != access.Write {
		t.Fatal("case-insensitive lookup")
	}
}

func TestCfgDirLevels(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name+".cfg"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("documents", "shareExport=\"e\"\nshareSecurity=\"private\"\nshareReadList=\"guest\"\nshareWriteList=\"sdimambro\"\n")
	write("media", "shareExport=\"e\"\nshareSecurity=\"secure\"\nshareWriteList=\"sdimambro\"\n")
	write("downloads", "shareExport=\"e\"\nshareSecurity=\"public\"\n")
	write("appdata", "shareExport=\"-\"\nshareSecurity=\"public\"\n")
	l := &Loader{Path: dir}
	pol, err := l.PolicyFor("guest", []string{"documents", "media", "downloads", "appdata"})
	if err != nil {
		t.Fatal(err)
	}
	if pol.Level("documents") != access.Read || pol.Level("media") != access.Read || pol.Level("downloads") != access.Write || pol.Level("appdata") != access.None {
		t.Fatalf("cfg dir levels: %v", pol)
	}
	pol, _ = l.PolicyFor("nobodyhere", []string{"documents"})
	if pol.Level("documents") != access.None {
		t.Fatal("private share must hide unknown users")
	}
}

package syncs

import (
	"path/filepath"
	"testing"
	"time"
)

func TestUpsertReportToggleAndPersist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "syncs.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	x, isNew, err := s.Upsert(Sync{ID: "a", DeviceID: "dev1", User: "simone", LocalLabel: "~/Desktop", RemotePath: "/documents/desktop", Enabled: true})
	if err != nil || !isNew || x.Mode != ModeUp || x.Kind != KindFolder {
		t.Fatalf("upsert: %v %v %+v", err, isNew, x)
	}
	// A report with files updates counters and keeps files newest first.
	files := []File{{Time: time.Now().Add(-time.Minute), Path: "/documents/desktop/a.txt", Direction: "up", Bytes: 10, Outcome: "ok"},
		{Time: time.Now(), Path: "/documents/desktop/b.txt", Direction: "up", Bytes: 20, Outcome: "ok"}}
	x, err = s.Report("a", Run{Files: 2, Bytes: 30, Status: "ok"}, files)
	if err != nil || x.Runs != 1 || x.TotalBytes != 30 || len(x.Files) != 2 || x.Files[0].Path != "/documents/desktop/b.txt" {
		t.Fatalf("report: %v %+v", err, x)
	}
	// The web UI pauses the pair: reports are refused and the app cannot re-enable it.
	if _, err := s.SetEnabled("a", false, "web ui"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Report("a", Run{Status: "ok"}, nil); err != ErrDisabled {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
	x, _, _ = s.Upsert(Sync{ID: "a", DeviceID: "dev1", LocalLabel: "~/Desktop", RemotePath: "/documents/desktop", Mode: ModeBoth, Enabled: true})
	if x.Enabled || x.Mode != ModeBoth || x.DisabledBy != "web ui" || x.Kind != KindFolder {
		t.Fatalf("app must not re-enable a pair paused from the web UI: %+v", x)
	}
	if _, err := s.SetEnabled("a", true, ""); err != nil {
		t.Fatal(err)
	}
	// Persisted.
	s2, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	y, ok := s2.Get("a")
	if !ok || !y.Enabled || y.Runs != 1 || len(y.Files) != 2 {
		t.Fatalf("reload: %v %+v", ok, y)
	}
	if n, _ := s2.RemoveForDevice("dev1"); n != 1 || len(s2.List()) != 0 {
		t.Fatalf("remove for device: %d", n)
	}
}

func TestKeepFiles(t *testing.T) {
	s, _ := Open("")
	s.Upsert(Sync{ID: "b", DeviceID: "d", Enabled: true})
	var files []File
	for i := 0; i < KeepFiles+20; i++ {
		files = append(files, File{Time: time.Now().Add(time.Duration(i) * time.Second), Path: "p", Outcome: "ok"})
	}
	x, err := s.Report("b", Run{Files: len(files), Status: "ok"}, files)
	if err != nil || len(x.Files) != KeepFiles {
		t.Fatalf("keep: %v %d", err, len(x.Files))
	}
}

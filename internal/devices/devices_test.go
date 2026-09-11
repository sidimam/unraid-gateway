package devices

import (
	"path/filepath"
	"testing"
)

func TestRegistry(t *testing.T) {
	p := filepath.Join(t.TempDir(), "devices.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	isNew, err := s.Register("d1", "Unraid Drive 1.3 (29) · iPhone · App", "simone", "Unraid Drive iPhone", "10.0.0.5")
	if err != nil || !isNew {
		t.Fatalf("first register: new=%v err=%v", isNew, err)
	}
	if isNew, _ = s.Register("d1", "", "", "", "10.0.0.6"); isNew {
		t.Fatal("second register must not be new")
	}
	if !s.Known("d1") || s.Known("nope") {
		t.Fatal("known")
	}
	if err := s.Touch("nope", ""); err != ErrNotRegistered {
		t.Fatal("touch unknown")
	}
	// Reload from disk.
	s2, err := Open(p)
	if err != nil || len(s2.List()) != 1 || s2.List()[0].Logins != 2 || s2.List()[0].LastIP != "10.0.0.6" {
		t.Fatalf("reload: %+v %v", s2.List(), err)
	}
	if _, err := s2.Remove("d1"); err != nil {
		t.Fatal(err)
	}
	if s2.Known("d1") {
		t.Fatal("still known after remove")
	}
	s2.Register("a", "", "", "", "")
	s2.Register("b", "", "", "", "")
	if n, _ := s2.RemoveAll(); n != 2 || len(s2.List()) != 0 {
		t.Fatal("remove all")
	}
}

// A reinstall registers a new id with the same device name and user: the previous entry is marked
// superseded (shown as "old"), a different user's identical phone is not, and an old installation
// that shows up again is not old any more.
func TestSupersede(t *testing.T) {
	s, _ := Open("")
	s.Register("old", "Unraid Drive 1.3 (29) · iPhone · iOS 26 · App", "simone", "k", "")
	s.Register("sofia", "Unraid Drive 1.3 (29) · iPhone · iOS 26 · App", "sofia", "k", "")
	isNew, _ := s.Register("new", "Unraid Drive 1.3 (30) · iPhone · iOS 26 · App", "simone", "k", "")
	if !isNew {
		t.Fatal("new id must be new")
	}
	o, _ := s.Get("old")
	if o.SupersededBy != "new" || o.SupersededAt.IsZero() {
		t.Fatalf("old not superseded: %+v", o)
	}
	if sf, _ := s.Get("sofia"); sf.SupersededBy != "" {
		t.Fatal("another user's device must not be superseded")
	}
	if n, _ := s.Get("new"); n.SupersededBy != "" {
		t.Fatal("the new device must not be superseded")
	}
	if err := s.Touch("old", "10.0.0.9"); err != nil {
		t.Fatal(err)
	}
	if o, _ := s.Get("old"); o.SupersededBy != "" {
		t.Fatal("an installation that talks again is not old")
	}
	if !sameDevice(&Device{Name: "Unraid Drive 1.2 (20) · Mac di Simone · macOS 26.6", User: ""}, &Device{Name: "Unraid Drive 1.3 (30) · Mac di Simone · macOS 26.6", User: ""}) {
		t.Fatal("version prefix must be ignored")
	}
}

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

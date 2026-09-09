package index

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newScanner(t *testing.T, root string) *Scanner {
	ix, err := Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ix.Close() })
	return &Scanner{Index: ix, Root: root, Skip: func(n string) bool { return n[0] == '.' }}
}

func TestReconcileCreateUpdateDelete(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "docs", "a.txt"), "aaa")
	write(t, filepath.Join(root, "docs", "sub", "b.txt"), "b")
	s := newScanner(t, root)
	s.full(context.Background())

	a, err := s.Index.ByPath("/docs/a.txt")
	if err != nil || a == nil || a.IsDir || a.Size != 3 {
		t.Fatalf("a.txt not indexed: %+v %v", a, err)
	}
	docs, _ := s.Index.ByPath("/docs")
	if a.ParentID != docs.ID || docs.ParentID != RootID {
		t.Fatalf("parent chain wrong: %+v %+v", a, docs)
	}
	latest, _ := s.Index.LatestSeq()

	// modify a file, delete the sub dir
	time.Sleep(10 * time.Millisecond)
	write(t, filepath.Join(root, "docs", "a.txt"), "aaaaaa")
	os.RemoveAll(filepath.Join(root, "docs", "sub"))
	if _, err := s.ScanDir("/docs"); err != nil {
		t.Fatal(err)
	}
	ch, newest, _, reset, err := s.Index.Changes(latest, 100)
	if err != nil || reset {
		t.Fatalf("changes: %v reset=%v", err, reset)
	}
	if newest <= latest || len(ch) < 3 {
		t.Fatalf("expected update + 2 deletes, got %d changes (%+v)", len(ch), ch)
	}
	kinds := map[string]int{}
	for _, c := range ch {
		kinds[c.Kind]++
	}
	if kinds["upsert"] != 1 || kinds["delete"] != 2 {
		t.Fatalf("kinds: %v", kinds)
	}
	a2, _ := s.Index.ByPath("/docs/a.txt")
	if a2.ID != a.ID || a2.Size != 6 {
		t.Fatalf("id must be stable across content changes: %+v vs %+v", a, a2)
	}
	if it, _ := s.Index.ByPath("/docs/sub/b.txt"); it != nil {
		t.Fatal("deleted item still indexed")
	}
}

func TestRenameKeepsID(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "docs", "old.txt"), "x")
	s := newScanner(t, root)
	s.full(context.Background())
	old, _ := s.Index.ByPath("/docs/old.txt")
	latest, _ := s.Index.LatestSeq()

	os.Rename(filepath.Join(root, "docs", "old.txt"), filepath.Join(root, "docs", "new.txt"))
	if _, err := s.ScanDir("/docs"); err != nil {
		t.Fatal(err)
	}
	nw, _ := s.Index.ByPath("/docs/new.txt")
	if nw == nil || nw.ID != old.ID {
		t.Fatalf("rename in place must keep the id: %+v -> %+v", old, nw)
	}
	ch, _, _, _, _ := s.Index.Changes(latest, 10)
	if len(ch) != 1 || ch[0].Kind != "move" || ch[0].OldPath != "/docs/old.txt" {
		t.Fatalf("expected one move change, got %+v", ch)
	}
}

func TestMoveAPIAndSubtree(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "docs", "dir", "f.txt"), "x")
	s := newScanner(t, root)
	s.full(context.Background())
	dir, _ := s.Index.ByPath("/docs/dir")
	f, _ := s.Index.ByPath("/docs/dir/f.txt")
	if err := s.Index.Move("/docs/dir", "/docs/renamed"); err != nil {
		t.Fatal(err)
	}
	d2, _ := s.Index.ByPath("/docs/renamed")
	f2, _ := s.Index.ByPath("/docs/renamed/f.txt")
	if d2 == nil || f2 == nil || d2.ID != dir.ID || f2.ID != f.ID {
		t.Fatalf("move must keep ids for the subtree: %+v %+v", d2, f2)
	}
	if err := s.Index.Delete("/docs/renamed"); err != nil {
		t.Fatal(err)
	}
	if it, _ := s.Index.ByPath("/docs/renamed/f.txt"); it != nil {
		t.Fatal("subtree not deleted")
	}
}

func TestChangesResetAndTruncate(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 5; i++ {
		write(t, filepath.Join(root, "docs", string(rune('a'+i))+".txt"), "x")
	}
	s := newScanner(t, root)
	s.full(context.Background())
	_, latest, _, reset, _ := s.Index.Changes(0, 10)
	if !reset || latest == 0 {
		t.Fatalf("since=0 must ask for a full enumeration: reset=%v latest=%d", reset, latest)
	}
	ch, newest, truncated, reset, _ := s.Index.Changes(1, 2)
	if reset || !truncated || len(ch) != 2 || newest != ch[1].Seq {
		t.Fatalf("pagination wrong: %d changes truncated=%v reset=%v newest=%d", len(ch), truncated, reset, newest)
	}
}

func TestDirScanPicksUpExternalChanges(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "docs", "readme.txt"), "x")
	write(t, filepath.Join(root, "media", "song.mp3"), "x")
	write(t, filepath.Join(root, "vm", "ubuntu", "disk.img"), "x")
	s := newScanner(t, root)
	s.full(context.Background())
	latest, _ := s.Index.LatestSeq()
	old, _ := s.Index.ByPath("/docs/readme.txt")

	time.Sleep(20 * time.Millisecond)
	os.Rename(filepath.Join(root, "docs", "readme.txt"), filepath.Join(root, "docs", "renamed.txt"))
	write(t, filepath.Join(root, "media", "new.mp3"), "y")
	os.RemoveAll(filepath.Join(root, "vm", "ubuntu"))
	s.dirs(context.Background())

	ch, _, _, reset, _ := s.Index.Changes(latest, 100)
	if reset {
		t.Fatal("unexpected reset")
	}
	got := map[string]string{}
	for _, c := range ch {
		got[c.Kind+" "+c.Path] = c.OldPath
	}
	if _, ok := got["move /docs/renamed.txt"]; !ok || got["move /docs/renamed.txt"] != "/docs/readme.txt" {
		t.Fatalf("rename not detected as move: %v", got)
	}
	if _, ok := got["upsert /media/new.mp3"]; !ok {
		t.Fatalf("new file not detected: %v", got)
	}
	if _, ok := got["delete /vm/ubuntu"]; !ok {
		t.Fatalf("deleted directory not detected: %v", got)
	}
	nw, _ := s.Index.ByPath("/docs/renamed.txt")
	if nw == nil || nw.ID != old.ID {
		t.Fatalf("renamed file must keep its id")
	}
}

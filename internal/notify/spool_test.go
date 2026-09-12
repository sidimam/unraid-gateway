package notify

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnraidSpool(t *testing.T) {
	dir := t.TempDir()
	n := New(Config{UnraidEnabled: true, UnraidDir: dir}, nil, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if !n.UnraidSpoolAvailable() {
		// unread does not exist yet: Send creates it
		if errs := n.Send(context.Background(), "", "New device", "iPhone\nsecond line", Warning); len(errs) != 0 {
			t.Fatalf("send: %v", errs)
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, "unread", "unraid-gateway_*.notify"))
	if len(files) != 1 {
		t.Fatalf("expected one .notify file, got %v", files)
	}
	b, _ := os.ReadFile(files[0])
	for _, want := range []string{"event=unraid-gateway", "subject=New device", "description=iPhone second line", "importance=warning", "timestamp="} {
		if !strings.Contains(string(b), want) {
			t.Errorf("missing %q in %q", want, b)
		}
	}
	if !n.UnraidSpoolAvailable() {
		t.Fatal("spool should be available after the first write")
	}
	if ch := n.Channels(); len(ch) != 1 || ch[0] != "unraid" {
		t.Fatalf("channels: %v", ch)
	}
}

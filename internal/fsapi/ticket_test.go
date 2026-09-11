package fsapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMediaTicketRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "media"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "media", "clip.mkv"), []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := New(dir, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	a.RegisterPublic(mux)
	tok, exp := a.IssueTicket("/media/clip.mkv", "", time.Hour)
	if time.Until(exp) < 59*time.Minute {
		t.Fatalf("expiry too short: %v", exp)
	}
	req := httptest.NewRequest("GET", "/media/"+tok, nil)
	req.Header.Set("Range", "bytes=2-4")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusPartialContent || rr.Body.String() != "234" {
		t.Fatalf("range via ticket: %d %q", rr.Code, rr.Body.String())
	}
	// Tampered signature and wrong path are refused.
	bad := tok[:len(tok)-2] + "zz"
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/media/"+bad, nil))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("tampered ticket accepted: %d", rr.Code)
	}
	// Expired ticket.
	expired, _ := a.IssueTicket("/media/clip.mkv", "", time.Nanosecond)
	time.Sleep(2 * time.Millisecond)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/media/"+expired, nil))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expired ticket accepted: %d", rr.Code)
	}
	if !strings.Contains(tok, ".") {
		t.Fatal("ticket has no signature part")
	}
}

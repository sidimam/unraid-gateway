package activity

import (
	"testing"
	"time"
)

func TestTrackerLifecycle(t *testing.T) {
	tr := New()
	w := Who{ClientID: "s:abc", User: "simone", IP: "10.0.0.2", Agent: "Unraid Drive 1.3 (28) · iPhone"}
	tr.Touch(w, "/media/x.mkv")
	x := tr.Begin(w, "stream", "/media/x.mkv", 1000, "bytes=0-")
	x.Add(250)
	snap := tr.Snapshot("")
	if len(snap.Clients) != 1 || snap.Clients[0].Active != 1 || snap.Clients[0].Requests != 1 {
		t.Fatalf("clients: %+v", snap.Clients)
	}
	if len(snap.Active) != 1 || snap.Active[0].Bytes != 250 || snap.Active[0].Kind != "stream" {
		t.Fatalf("active: %+v", snap.Active)
	}
	if got := tr.Snapshot("someoneelse"); len(got.Clients) != 0 || len(got.Active) != 0 {
		t.Fatalf("user filter leaked: %+v", got)
	}
	x.Add(750)
	x.End()
	x.End() // idempotent
	snap = tr.Snapshot("")
	if len(snap.Active) != 0 || len(snap.Recent) != 1 || snap.Recent[0].Bytes != 1000 || snap.Recent[0].Ended == nil {
		t.Fatalf("recent: %+v", snap.Recent)
	}
	// Idle clients vanish.
	tr.IdleAfter = time.Nanosecond
	time.Sleep(time.Millisecond)
	if got := tr.Snapshot(""); len(got.Clients) != 0 {
		t.Fatalf("idle client kept: %+v", got.Clients)
	}
	tr.Touch(w, "")
	tr.Forget("s:abc")
	if got := tr.Snapshot(""); len(got.Clients) != 0 {
		t.Fatal("forget failed")
	}
}

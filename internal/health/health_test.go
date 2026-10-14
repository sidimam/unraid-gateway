package health

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeQ struct {
	body string
	err  error
}

func (f fakeQ) Query(context.Context, string, string, map[string]any) (json.RawMessage, error) {
	return json.RawMessage(f.body), f.err
}

func TestAssessGreen(t *testing.T) {
	gw := Gateway{Version: "test", MountedShares: []string{"a"}, ConfigWritable: true, UserAuth: "off"}
	s := Assess(context.Background(), gw, fakeQ{body: `{"array":{"state":"STARTED","disks":[{"name":"disk1","status":"DISK_OK","temp":35}]},"notifications":{"overview":{"unread":{"total":0,"warning":0,"alert":0}}},"metrics":{"cpu":{"percentTotal":5},"memory":{"percentTotal":40}}}`}, "k")
	if s.Level != OK || len(s.Reasons) != 0 {
		t.Fatalf("expected ok, got %s %v", s.Level, s.Reasons)
	}
}

func TestAssessYellowAndRed(t *testing.T) {
	gw := Gateway{Version: "test", MountedShares: []string{"a"}, UnwritableShares: []string{"a"}, ConfigWritable: false, UserAuth: "optional"}
	s := Assess(context.Background(), gw, fakeQ{body: `{"array":{"state":"STARTED","disks":[{"name":"disk1","status":"DISK_OK","temp":55}]},"notifications":{"overview":{"unread":{"total":2,"warning":2,"alert":0}}}}`}, "k")
	for _, r := range s.Reasons {
		if r == "2 warning notification(s)" {
			t.Fatalf("Unraid notifications must not count: %v", s.Reasons)
		}
	}
	if s.Level != Warning || len(s.Reasons) != 3 {
		t.Fatalf("expected warning with the three gateway reasons, got %s %v", s.Level, s.Reasons)
	}
	// Unraid's own state never colours the dot (array stopped, hot disk): still the same two gateway warnings.
	s = Assess(context.Background(), gw, fakeQ{body: `{"array":{"state":"STOPPED","disks":[{"name":"disk1","status":"DISK_DSBL","temp":70}]}}`}, "k")
	if s.Level != Warning || len(s.Reasons) != 3 {
		t.Fatalf("Unraid state must not count, got %s %v", s.Level, s.Reasons)
	}
	s = Assess(context.Background(), Gateway{MountedShares: []string{"a"}, ConfigWritable: true, UserAuth: "off"}, fakeQ{body: `{"docker":{"containers":[{"names":["unraid-gateway"],"image":"ghcr.io/sidimam/unraid-gateway:latest","state":"EXITED"}]}}`}, "k")
	if s.Level != Warning || s.Reasons[0] != "unraid-gateway container exited" {
		t.Fatalf("expected container warning, got %s %v", s.Level, s.Reasons)
	}
	s = Assess(context.Background(), Gateway{MountedShares: []string{"a"}, ConfigWritable: true, UserAuth: "off"}, fakeQ{err: context.DeadlineExceeded}, "k")
	if s.Level != Error {
		t.Fatalf("unreachable unraid must be red, got %s", s.Level)
	}
}

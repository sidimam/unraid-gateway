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
	if s.Level != Warning || len(s.Reasons) < 3 {
		t.Fatalf("expected warning with reasons, got %s %v", s.Level, s.Reasons)
	}
	s = Assess(context.Background(), gw, fakeQ{body: `{"array":{"state":"STOPPED","disks":[]}}`}, "k")
	if s.Level != Error || s.Reasons[0] != "array stopped" {
		t.Fatalf("expected error array stopped, got %s %v", s.Level, s.Reasons)
	}
	s = Assess(context.Background(), Gateway{MountedShares: []string{"a"}, ConfigWritable: true, UserAuth: "off"}, fakeQ{err: context.DeadlineExceeded}, "k")
	if s.Level != Error {
		t.Fatalf("unreachable unraid must be red, got %s", s.Level)
	}
}

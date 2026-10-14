package metrics

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRecordSeriesAndPersist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "metrics.json")
	s := Open(p)
	s.Request("d1")
	s.Transfer("d1", "download", 1000)
	s.Transfer("d1", "upload", 500)
	s.Transfer("d1", "stream", 200)
	s.Log("d1", "login", "info", "signed in")
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	s2 := Open(p)
	tot := s2.Sum("d1", time.Now().Add(-time.Hour))
	if tot.Requests != 1 || tot.Down != 1200 || tot.Up != 500 || tot.Streams != 1 {
		t.Fatalf("totals wrong: %+v", tot)
	}
	series := s2.Series("d1", time.Now().Add(-24*time.Hour), time.Hour)
	if len(series) != 25 || series[len(series)-1].Down != 1200 {
		t.Fatalf("series wrong: len=%d last=%+v", len(series), series[len(series)-1])
	}
	days := s2.Series("d1", time.Now().Add(-7*24*time.Hour), 24*time.Hour)
	if len(days) < 7 || days[len(days)-1].Down != 1200 {
		t.Fatalf("daily series wrong: len=%d last=%+v", len(days), days[len(days)-1])
	}
	if ev := s2.Events("d1", 10); len(ev) != 1 || ev[0].Kind != "login" {
		t.Fatalf("events wrong: %+v", ev)
	}
	if len(s2.sortedHours("d1")) != 1 {
		t.Fatal("expected one hour bucket")
	}
	s2.Forget("d1")
	if s2.Sum("d1", time.Time{}).Requests != 0 {
		t.Fatal("forget failed")
	}
}

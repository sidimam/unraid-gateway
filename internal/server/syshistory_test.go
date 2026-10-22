package server

import (
	"testing"
	"time"
)

func TestSysHistoryAveragesPerStep(t *testing.T) {
	var h sysHistory
	base := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		h.add(sysSample{T: base.Add(time.Duration(i) * time.Minute), CPU: float64(i * 10), Mem: 50, MemUsed: 1, MemTotal: 2})
	}
	got := h.since(base.Add(-time.Hour), 5*time.Minute)
	if len(got) != 2 {
		t.Fatalf("want 2 buckets, got %d", len(got))
	}
	if got[0].CPU != 20 || got[1].CPU != 70 || got[0].Mem != 50 {
		t.Fatalf("averages wrong: %+v", got)
	}
	// Old samples are dropped.
	h.add(sysSample{T: base.Add(sysHistoryKeep + time.Hour)})
	if n := len(h.since(time.Time{}, time.Minute)); n != 1 {
		t.Fatalf("old samples kept: %d", n)
	}
}

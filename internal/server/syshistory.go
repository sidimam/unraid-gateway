package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// sysSample is one reading of Unraid's CPU and memory load (0.17): the web UI's Health section
// draws them over time. Samples live in memory only (a restart starts a new history).
type sysSample struct {
	T        time.Time `json:"t"`
	CPU      float64   `json:"cpu"`      // percent
	Mem      float64   `json:"mem"`      // percent
	MemUsed  int64     `json:"memUsed"`  // bytes
	MemTotal int64     `json:"memTotal"` // bytes
}

const sysHistoryKeep = 7 * 24 * time.Hour

type sysHistory struct {
	mu      sync.Mutex
	samples []sysSample
}

func (h *sysHistory) add(x sysSample) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.samples = append(h.samples, x)
	cut := x.T.Add(-sysHistoryKeep)
	i := 0
	for i < len(h.samples) && h.samples[i].T.Before(cut) {
		i++
	}
	h.samples = h.samples[i:]
}

// since returns the samples after t, averaged per step so that a week fits in a few hundred points.
func (h *sysHistory) since(t time.Time, step time.Duration) []sysSample {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := []sysSample{}
	var acc sysSample
	var n int
	var slot time.Time
	flush := func() {
		if n == 0 {
			return
		}
		out = append(out, sysSample{T: slot, CPU: acc.CPU / float64(n), Mem: acc.Mem / float64(n), MemUsed: acc.MemUsed / int64(n), MemTotal: acc.MemTotal / int64(n)})
		acc, n = sysSample{}, 0
	}
	for _, x := range h.samples {
		if x.T.Before(t) {
			continue
		}
		s := x.T.Truncate(step)
		if n > 0 && !s.Equal(slot) {
			flush()
		}
		slot = s
		acc.CPU += x.CPU
		acc.Mem += x.Mem
		acc.MemUsed += x.MemUsed
		acc.MemTotal += x.MemTotal
		n++
	}
	flush()
	return out
}

const sysSampleQuery = `{ metrics { cpu { percentTotal } memory { percentTotal used total } } }`

// runSystemSampler reads Unraid's load every minute with the key remembered for the web UI
// (WEBUI_API_KEY / /config/webui.key); without one, the Health charts stay empty.
func (s *Server) runSystemSampler(stop <-chan struct{}, every time.Duration) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	s.sampleSystem()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
			s.sampleSystem()
		}
	}
}

func (s *Server) sampleSystem() {
	key, _ := s.storedKey()
	if key == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	data, err := s.gql.Query(ctx, key, sysSampleQuery, nil)
	if err != nil {
		s.log.Debug("system sample failed", "err", err)
		return
	}
	var out struct {
		Metrics struct {
			CPU    struct{ PercentTotal float64 } `json:"cpu"`
			Memory struct {
				PercentTotal float64
				Used, Total  int64
			} `json:"memory"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return
	}
	mem := out.Metrics.Memory.PercentTotal
	if mem == 0 && out.Metrics.Memory.Total > 0 {
		mem = float64(out.Metrics.Memory.Used) * 100 / float64(out.Metrics.Memory.Total)
	}
	s.sys.add(sysSample{T: time.Now(), CPU: out.Metrics.CPU.PercentTotal, Mem: mem, MemUsed: out.Metrics.Memory.Used, MemTotal: out.Metrics.Memory.Total})
}

// GET /api/v1/system/history?range=24h|7d — CPU and memory load over time (0.17).
func (s *Server) handleSystemHistory(w http.ResponseWriter, r *http.Request) {
	since, step := time.Now().Add(-24*time.Hour), 5*time.Minute
	if r.URL.Query().Get("range") == "7d" {
		since, step = time.Now().Add(-7*24*time.Hour), time.Hour
	}
	writeJSON(w, http.StatusOK, map[string]any{"step": step.String(), "samples": s.sys.since(since, step), "sampling": s.storedKeyAvailable()})
}

func (s *Server) storedKeyAvailable() bool { k, _ := s.storedKey(); return k != "" }

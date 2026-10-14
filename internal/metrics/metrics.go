// Package metrics keeps, per registered device, what moved over time (requests, bytes up and
// down, streams) in hourly buckets for 90 days, and a short log of notable events (logins,
// failed transfers, revocations). It feeds the Devices detail page of the web UI (0.13) and
// GET /api/v1/devices/{id}/metrics|log. Persisted as JSON under /config, written lazily.
package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Bucket is one hour of one device.
type Bucket struct {
	Requests  int64 `json:"requests"`
	Up        int64 `json:"up"`   // bytes uploaded by the device (PUT/uploads)
	Down      int64 `json:"down"` // bytes downloaded or streamed to the device
	Streams   int64 `json:"streams"`
	Uploads   int64 `json:"uploads"`
	Downloads int64 `json:"downloads"`
}

// Event is one notable thing a device did or suffered.
type Event struct {
	Time   time.Time `json:"time"`
	Kind   string    `json:"kind"` // login, register, revoke, upload, download, stream, error
	Detail string    `json:"detail"`
	Level  string    `json:"level"` // info | warning | error
}

type device struct {
	Hours  map[int64]*Bucket `json:"hours"` // unix hour → bucket
	Events []Event           `json:"events"`
}

// Store is safe for concurrent use.
type Store struct {
	path    string
	mu      sync.Mutex
	devices map[string]*device
	dirty   bool
	// Retention of hourly buckets (default 90 days) and events per device (default 200).
	Keep       time.Duration
	KeepEvents int
}

// Open loads the file (missing = empty).
func Open(path string) *Store {
	s := &Store{path: path, devices: map[string]*device{}, Keep: 90 * 24 * time.Hour, KeepEvents: 200}
	if path != "" {
		if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &s.devices)
		}
	}
	for _, d := range s.devices {
		if d.Hours == nil {
			d.Hours = map[int64]*Bucket{}
		}
	}
	return s
}

func (s *Store) dev(id string) *device {
	d := s.devices[id]
	if d == nil {
		d = &device{Hours: map[int64]*Bucket{}}
		s.devices[id] = d
	}
	return d
}

func hourOf(t time.Time) int64 { return t.Unix() / 3600 * 3600 }

func (s *Store) bucket(id string, t time.Time) *Bucket {
	d := s.dev(id)
	h := hourOf(t)
	b := d.Hours[h]
	if b == nil {
		b = &Bucket{}
		d.Hours[h] = b
	}
	return b
}

// Request counts one authenticated request of a device.
func (s *Store) Request(id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	s.bucket(id, time.Now()).Requests++
	s.dirty = true
	s.mu.Unlock()
}

// Transfer records a finished download, stream or upload of n bytes.
func (s *Store) Transfer(id, kind string, n int64) {
	if id == "" {
		return
	}
	s.mu.Lock()
	b := s.bucket(id, time.Now())
	switch kind {
	case "upload":
		b.Up += n
		b.Uploads++
	case "stream":
		b.Down += n
		b.Streams++
	default:
		b.Down += n
		b.Downloads++
	}
	s.dirty = true
	s.mu.Unlock()
}

// Log appends an event for a device.
func (s *Store) Log(id, kind, level, detail string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	d := s.dev(id)
	d.Events = append(d.Events, Event{Time: time.Now(), Kind: kind, Level: level, Detail: detail})
	if len(d.Events) > s.KeepEvents {
		d.Events = d.Events[len(d.Events)-s.KeepEvents:]
	}
	s.dirty = true
	s.mu.Unlock()
}

// Point is one hour (or day) of the series returned to the UI.
type Point struct {
	Time      time.Time `json:"t"`
	Requests  int64     `json:"requests"`
	Up        int64     `json:"up"`
	Down      int64     `json:"down"`
	Streams   int64     `json:"streams"`
	Uploads   int64     `json:"uploads"`
	Downloads int64     `json:"downloads"`
}

// Totals sums a period.
type Totals struct {
	Requests  int64 `json:"requests"`
	Up        int64 `json:"up"`
	Down      int64 `json:"down"`
	Streams   int64 `json:"streams"`
	Uploads   int64 `json:"uploads"`
	Downloads int64 `json:"downloads"`
}

// Series returns the buckets since `since`, grouped per `step` (time.Hour or 24h), oldest first,
// with empty steps filled in so charts have a regular axis.
func (s *Store) Series(id string, since time.Time, step time.Duration) []Point {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.devices[id]
	start := since.Truncate(step)
	if step == 24*time.Hour {
		y, m, dd := since.Date()
		start = time.Date(y, m, dd, 0, 0, 0, 0, since.Location())
	}
	n := int(time.Since(start)/step) + 1
	out := make([]Point, n)
	for i := range out {
		out[i].Time = start.Add(time.Duration(i) * step)
	}
	if d == nil {
		return out
	}
	for h, b := range d.Hours {
		t := time.Unix(h, 0)
		if t.Before(start) {
			continue
		}
		i := int(t.Sub(start) / step)
		if step == 24*time.Hour {
			i = int(t.In(since.Location()).Sub(start) / step)
		}
		if i < 0 || i >= n {
			continue
		}
		p := &out[i]
		p.Requests += b.Requests
		p.Up += b.Up
		p.Down += b.Down
		p.Streams += b.Streams
		p.Uploads += b.Uploads
		p.Downloads += b.Downloads
	}
	return out
}

// Sum totals everything since `since`.
func (s *Store) Sum(id string, since time.Time) Totals {
	s.mu.Lock()
	defer s.mu.Unlock()
	var t Totals
	d := s.devices[id]
	if d == nil {
		return t
	}
	for h, b := range d.Hours {
		if time.Unix(h, 0).Before(since) {
			continue
		}
		t.Requests += b.Requests
		t.Up += b.Up
		t.Down += b.Down
		t.Streams += b.Streams
		t.Uploads += b.Uploads
		t.Downloads += b.Downloads
	}
	return t
}

// Events returns the newest events first (up to limit).
func (s *Store) Events(id string, limit int) []Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.devices[id]
	if d == nil {
		return []Event{}
	}
	ev := d.Events
	if limit > 0 && len(ev) > limit {
		ev = ev[len(ev)-limit:]
	}
	out := make([]Event, len(ev))
	for i, e := range ev {
		out[len(ev)-1-i] = e
	}
	return out
}

// LastActivity is the newest hour with any traffic for the device (zero if none).
func (s *Store) LastActivity(id string) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.devices[id]
	if d == nil {
		return time.Time{}
	}
	var last int64
	for h := range d.Hours {
		if h > last {
			last = h
		}
	}
	if last == 0 {
		return time.Time{}
	}
	return time.Unix(last, 0)
}

// Forget drops everything about a device (it was removed from the registry).
func (s *Store) Forget(id string) {
	s.mu.Lock()
	delete(s.devices, id)
	s.dirty = true
	s.mu.Unlock()
}

// Flush prunes old buckets and writes the file when something changed.
func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.dirty || s.path == "" {
		return nil
	}
	cut := hourOf(time.Now().Add(-s.Keep))
	for _, d := range s.devices {
		for h := range d.Hours {
			if h < cut {
				delete(d.Hours, h)
			}
		}
	}
	b, _ := json.Marshal(s.devices)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return err
	}
	s.dirty = false
	return nil
}

// Run flushes periodically until ctx is done, then once more.
func (s *Store) Run(stop <-chan struct{}, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_ = s.Flush()
		case <-stop:
			_ = s.Flush()
			return
		}
	}
}

// sortedHours is used by tests.
func (s *Store) sortedHours(id string) []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.devices[id]
	if d == nil {
		return nil
	}
	hs := make([]int64, 0, len(d.Hours))
	for h := range d.Hours {
		hs = append(hs, h)
	}
	sort.Slice(hs, func(i, j int) bool { return hs[i] < hs[j] })
	return hs
}

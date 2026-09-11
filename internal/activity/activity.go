// Package activity keeps an in-memory picture of who is using the gateway right
// now: connected clients (one per session or API-key+IP), the transfers in
// flight (downloads, streams, uploads) with the bytes moved so far, and the
// last completed ones. It feeds the "Activity" panel of the web UI and
// GET /api/v1/activity. Nothing is persisted.
package activity

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Who identifies the caller of a request.
type Who struct {
	ClientID string `json:"clientId"` // session id or key+ip
	User     string `json:"user,omitempty"`
	Key      string `json:"key,omitempty"`
	IP       string `json:"ip"`
	Agent    string `json:"agent"` // X-Unraid-Drive-Client if sent, else the User-Agent
}

// Client is a connected device/session.
type Client struct {
	Who
	FirstSeen  time.Time `json:"firstSeen"`
	LastSeen   time.Time `json:"lastSeen"`
	Requests   int64     `json:"requests"`
	LastPath   string    `json:"lastPath,omitempty"`
	LastPathAt time.Time `json:"lastPathAt,omitempty"`
	Active     int       `json:"active"`
}

// Transfer is a download, stream or upload.
type Transfer struct {
	ID int64 `json:"id"`
	Who
	Kind    string     `json:"kind"` // download | stream | upload
	Path    string     `json:"path"`
	Size    int64      `json:"size"`
	Bytes   int64      `json:"bytes"`
	Started time.Time  `json:"started"`
	Ended   *time.Time `json:"ended,omitempty"`
	Range   string     `json:"range,omitempty"`
	bytes   *atomic.Int64
	tracker *Tracker
}

// Add records bytes moved.
func (t *Transfer) Add(n int64) {
	if t != nil && t.bytes != nil {
		t.bytes.Add(n)
	}
}

// End marks the transfer finished and moves it to the recent list.
func (t *Transfer) End() {
	if t != nil && t.tracker != nil {
		t.tracker.end(t)
	}
}

// Tracker is safe for concurrent use.
type Tracker struct {
	mu      sync.Mutex
	clients map[string]*Client
	active  map[int64]*Transfer
	recent  []*Transfer
	seq     int64
	// IdleAfter drops clients not seen for this long (default 30 min).
	IdleAfter time.Duration
	// KeepRecent is how many finished transfers to keep (default 50).
	KeepRecent int
}

// New creates a tracker.
func New() *Tracker {
	return &Tracker{clients: map[string]*Client{}, active: map[int64]*Transfer{}, IdleAfter: 30 * time.Minute, KeepRecent: 50}
}

// Touch records a request by a client (called from the auth middleware).
func (t *Tracker) Touch(w Who, path string) {
	if w.ClientID == "" {
		return
	}
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()
	c, ok := t.clients[w.ClientID]
	if !ok {
		c = &Client{Who: w, FirstSeen: now}
		t.clients[w.ClientID] = c
	}
	c.LastSeen = now
	c.Requests++
	if w.User != "" {
		c.User = w.User
	}
	if w.Agent != "" {
		c.Agent = w.Agent
	}
	c.IP = w.IP
	if path != "" {
		c.LastPath, c.LastPathAt = path, now
	}
}

// Forget removes a client (logout).
func (t *Tracker) Forget(clientID string) {
	t.mu.Lock()
	delete(t.clients, clientID)
	t.mu.Unlock()
}

// Begin registers a transfer in flight.
func (t *Tracker) Begin(w Who, kind, path string, size int64, rng string) *Transfer {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.seq++
	tr := &Transfer{ID: t.seq, Who: w, Kind: kind, Path: path, Size: size, Started: time.Now(), Range: rng, bytes: new(atomic.Int64), tracker: t}
	t.active[tr.ID] = tr
	return tr
}

func (t *Tracker) end(tr *Transfer) {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.active[tr.ID]; !ok {
		return
	}
	delete(t.active, tr.ID)
	tr.Ended = &now
	tr.Bytes = tr.bytes.Load()
	t.recent = append([]*Transfer{tr}, t.recent...)
	if len(t.recent) > t.KeepRecent {
		t.recent = t.recent[:t.KeepRecent]
	}
}

// Snapshot is what the API returns.
type Snapshot struct {
	Now     time.Time  `json:"now"`
	Clients []Client   `json:"clients"`
	Active  []Transfer `json:"active"`
	Recent  []Transfer `json:"recent"`
}

// Snapshot returns the current state; with onlyUser set, only that user's rows.
func (t *Tracker) Snapshot(onlyUser string) Snapshot {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()
	activeBy := map[string]int{}
	for _, tr := range t.active {
		activeBy[tr.ClientID]++
	}
	out := Snapshot{Now: now}
	for id, c := range t.clients {
		if activeBy[id] == 0 && now.Sub(c.LastSeen) > t.IdleAfter {
			delete(t.clients, id)
			continue
		}
		if onlyUser != "" && c.User != onlyUser {
			continue
		}
		cc := *c
		cc.Active = activeBy[id]
		out.Clients = append(out.Clients, cc)
	}
	sort.Slice(out.Clients, func(i, j int) bool { return out.Clients[i].LastSeen.After(out.Clients[j].LastSeen) })
	for _, tr := range t.active {
		if onlyUser != "" && tr.User != onlyUser {
			continue
		}
		cp := *tr
		cp.Bytes = tr.bytes.Load()
		out.Active = append(out.Active, cp)
	}
	sort.Slice(out.Active, func(i, j int) bool { return out.Active[i].Started.Before(out.Active[j].Started) })
	for _, tr := range t.recent {
		if onlyUser != "" && tr.User != onlyUser {
			continue
		}
		out.Recent = append(out.Recent, *tr)
	}
	if out.Clients == nil {
		out.Clients = []Client{}
	}
	if out.Active == nil {
		out.Active = []Transfer{}
	}
	if out.Recent == nil {
		out.Recent = []Transfer{}
	}
	return out
}

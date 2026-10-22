// Package syncs keeps the registry of the folder syncs and backups the apps run against this
// gateway (0.14): which device syncs which local folder with which Unraid folder, in which
// direction, whether it is enabled, how the last runs went and which files moved recently.
// The apps own their pairs (they create and update them); the web UI can pause a pair
// (enabled=false), and the app honours that before every run. Stored as one JSON file next to
// devices.json.
package syncs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Kind of sync pair.
const (
	KindFolder    = "folder"    // a local folder ↔ an Unraid folder (Mac)
	KindPhotos    = "photos"    // the device's Photos library → an Unraid folder
	KindDownloads = "downloads" // the device's Downloads folder → an Unraid folder (iPhone/iPad/Vision Pro)
)

// Mode of a pair.
const (
	ModeUp   = "up"   // device → Unraid
	ModeDown = "down" // Unraid → device
	ModeBoth = "both" // two-way
)

// Run is the outcome of one execution, reported by the app.
type Run struct {
	At        time.Time `json:"at"`
	Duration  int64     `json:"durationMs"`
	Files     int       `json:"files"`
	Bytes     int64     `json:"bytes"`
	Conflicts int       `json:"conflicts"`
	Errors    int       `json:"errors"`
	Status    string    `json:"status"` // ok | partial | failed | skipped
	Error     string    `json:"error,omitempty"`
	Trigger   string    `json:"trigger,omitempty"` // manual | scheduled | change
}

// File is one file the pair moved (or failed to move), newest first in the lists.
type File struct {
	Time      time.Time `json:"time"`
	Path      string    `json:"path"`      // Unraid-side path
	Direction string    `json:"direction"` // up | down | delete | conflict
	Bytes     int64     `json:"bytes"`
	Outcome   string    `json:"outcome"` // ok | conflict | failed | skipped
	Detail    string    `json:"detail,omitempty"`
}

// Sync is one pair.
type Sync struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	User       string `json:"user,omitempty"`
	Kind       string `json:"kind"`
	LocalLabel string `json:"localLabel"`
	RemotePath string `json:"remotePath"`
	Mode       string `json:"mode"`
	// Conflict policy chosen in the app (0.16): newest | keepBoth | keepExisting | overwrite | ask.
	Conflicts string `json:"conflicts,omitempty"`
	// The app compares a quick content fingerprint before deciding: identical files are skipped (0.16).
	HashCheck bool      `json:"hashCheck"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// Set by the web UI when it pauses the pair; cleared when re-enabled. Shown to the user.
	DisabledBy string `json:"disabledBy,omitempty"`
	LastRun    *Run   `json:"lastRun,omitempty"`
	Runs       int64  `json:"runs"`
	TotalFiles int64  `json:"totalFiles"`
	TotalBytes int64  `json:"totalBytes"`
	// Recent files (≤ KeepFiles), newest first.
	Files []File `json:"files,omitempty"`
}

// KeepFiles is how many recent files are kept per pair.
const KeepFiles = 100

var ErrNotFound = errors.New("unknown sync")
var ErrDisabled = errors.New("sync disabled on the gateway")

// Store is safe for concurrent use.
type Store struct {
	mu    sync.Mutex
	path  string
	syncs map[string]*Sync
}

// Open loads the registry (an empty path keeps it in memory only).
func Open(path string) (*Store, error) {
	s := &Store{path: path, syncs: map[string]*Sync{}}
	if path == "" {
		return s, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	var list []*Sync
	if err := json.Unmarshal(b, &list); err != nil {
		return s, err
	}
	for _, x := range list {
		s.syncs[x.ID] = x
	}
	return s, nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	list := s.sortedLocked()
	b, err := json.MarshalIndent(list, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) sortedLocked() []*Sync {
	out := make([]*Sync, 0, len(s.syncs))
	for _, x := range s.syncs {
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DeviceID != out[j].DeviceID {
			return out[i].DeviceID < out[j].DeviceID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// Upsert creates or updates a pair owned by `in.DeviceID`. The app sends the definition; the
// gateway keeps what it owns (enabled flag set from the web UI, counters, history) when the pair
// already exists. Returns the stored pair and whether it was new.
func (s *Store) Upsert(in Sync) (Sync, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	cur, ok := s.syncs[in.ID]
	if !ok {
		in.CreatedAt, in.UpdatedAt = now, now
		if in.Mode == "" {
			in.Mode = ModeUp
		}
		if in.Kind == "" {
			in.Kind = KindFolder
		}
		in.Files = nil
		in.LastRun = nil
		in.Runs, in.TotalFiles, in.TotalBytes = 0, 0, 0
		c := in
		s.syncs[in.ID] = &c
		return c, true, s.save()
	}
	if in.LocalLabel != "" {
		cur.LocalLabel = in.LocalLabel
	}
	if in.RemotePath != "" {
		cur.RemotePath = in.RemotePath
	}
	if in.Mode != "" {
		cur.Mode = in.Mode
	}
	if in.Kind != "" {
		cur.Kind = in.Kind
	}
	if in.Conflicts != "" {
		cur.Conflicts = in.Conflicts
	}
	cur.HashCheck = in.HashCheck
	if in.User != "" {
		cur.User = in.User
	}
	// The app may re-enable a pair it disabled itself, never one the web UI paused.
	if cur.DisabledBy == "" {
		cur.Enabled = in.Enabled
	}
	cur.UpdatedAt = now
	return *cur, false, s.save()
}

// SetEnabled pauses or resumes a pair; `by` names who did it (web UI user / console), kept while
// the pair stays paused.
func (s *Store) SetEnabled(id string, enabled bool, by string) (Sync, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.syncs[id]
	if !ok {
		return Sync{}, ErrNotFound
	}
	cur.Enabled = enabled
	if enabled {
		cur.DisabledBy = ""
	} else {
		cur.DisabledBy = by
	}
	cur.UpdatedAt = time.Now()
	return *cur, s.save()
}

// Report records one run and its files (newest first kept, ≤ KeepFiles).
func (s *Store) Report(id string, run Run, files []File) (Sync, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.syncs[id]
	if !ok {
		return Sync{}, ErrNotFound
	}
	if !cur.Enabled {
		return *cur, ErrDisabled
	}
	if run.At.IsZero() {
		run.At = time.Now()
	}
	r := run
	cur.LastRun = &r
	cur.Runs++
	cur.TotalFiles += int64(run.Files)
	cur.TotalBytes += run.Bytes
	if len(files) > 0 {
		sort.Slice(files, func(i, j int) bool { return files[i].Time.After(files[j].Time) })
		cur.Files = append(files, cur.Files...)
		if len(cur.Files) > KeepFiles {
			cur.Files = cur.Files[:KeepFiles]
		}
	}
	cur.UpdatedAt = time.Now()
	return *cur, s.save()
}

// Get returns one pair.
func (s *Store) Get(id string) (Sync, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.syncs[id]
	if !ok {
		return Sync{}, false
	}
	return *x, true
}

// List returns every pair (sorted by device, then creation).
func (s *Store) List() []Sync {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Sync, 0, len(s.syncs))
	for _, x := range s.sortedLocked() {
		out = append(out, *x)
	}
	return out
}

// ForDevice returns the pairs of one device.
func (s *Store) ForDevice(deviceID string) []Sync {
	out := []Sync{}
	for _, x := range s.List() {
		if x.DeviceID == deviceID {
			out = append(out, x)
		}
	}
	return out
}

// Remove deletes one pair.
func (s *Store) Remove(id string) (Sync, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.syncs[id]
	if !ok {
		return Sync{}, ErrNotFound
	}
	delete(s.syncs, id)
	return *x, s.save()
}

// RemoveForDevice drops every pair of a device (the device was removed).
func (s *Store) RemoveForDevice(deviceID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, x := range s.syncs {
		if x.DeviceID == deviceID {
			delete(s.syncs, id)
			n++
		}
	}
	if n == 0 {
		return 0, nil
	}
	return n, s.save()
}

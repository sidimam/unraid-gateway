// Package devices keeps the registry of app installations allowed to use the
// gateway. A device registers once (when the user adds the server in Unraid
// Drive); removing it here revokes its sessions and forces the app to ask the
// user to sign in again. Persisted as JSON under /config.
package devices

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Device is one registered installation of the app (or a browser using the web UI).
type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`           // X-Unraid-Drive-Client description at registration
	User      string    `json:"user,omitempty"` // Unraid user used at registration, if any
	Key       string    `json:"key,omitempty"`  // API key name
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	LastIP    string    `json:"lastIp,omitempty"`
	Logins    int       `json:"logins"`
	// SupersededBy is set when a *new* device registered with the same name and user: the app was
	// reinstalled (new id) on the same hardware. The old entry stays listed as "old" with its last
	// seen date until the admin removes it; it is cleared again if that old installation shows up.
	SupersededBy string    `json:"supersededBy,omitempty"`
	SupersededAt time.Time `json:"supersededAt,omitempty"`
}

// Store is safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
	list map[string]*Device
}

// ErrNotRegistered is returned for unknown device ids.
var ErrNotRegistered = errors.New("device not registered")

// Open loads the registry (an empty file or a missing one is fine).
func Open(path string) (*Store, error) {
	s := &Store{path: path, list: map[string]*Device{}}
	if path == "" {
		return s, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	var arr []*Device
	if len(b) > 0 {
		if err := json.Unmarshal(b, &arr); err != nil {
			return nil, err
		}
	}
	for _, d := range arr {
		s.list[d.ID] = d
	}
	return s, nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	arr := s.sorted()
	b, _ := json.MarshalIndent(arr, "", "  ")
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) sorted() []*Device {
	arr := make([]*Device, 0, len(s.list))
	for _, d := range s.list {
		arr = append(arr, d)
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].LastSeen.After(arr[j].LastSeen) })
	return arr
}

// Register adds or refreshes a device. Returns true when it is new.
func (s *Store) Register(id, name, user, key, ip string) (isNew bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	d, ok := s.list[id]
	if !ok {
		d = &Device{ID: id, FirstSeen: now}
		s.list[id] = d
		isNew = true
	}
	if name != "" {
		d.Name = name
	}
	if user != "" {
		d.User = user
	}
	if key != "" {
		d.Key = key
	}
	d.LastSeen, d.LastIP = now, ip
	d.Logins++
	// A device that signs in again is not old, whatever happened in between.
	d.SupersededBy, d.SupersededAt = "", time.Time{}
	if isNew {
		// The same device name (model + user's device name) with the same user: a reinstall on the
		// same hardware, or the same person's new phone. Mark the older entries as superseded.
		for _, o := range s.list {
			if o.ID != id && o.SupersededBy == "" && sameDevice(o, d) {
				o.SupersededBy, o.SupersededAt = id, now
			}
		}
	}
	return isNew, s.save()
}

// sameDevice: identical description (case-insensitive, ignoring the app version/build prefix)
// and identical Unraid user.
func sameDevice(a, b *Device) bool {
	if a.User != b.User || a.Name == "" || b.Name == "" {
		return false
	}
	return strings.EqualFold(stripVersion(a.Name), stripVersion(b.Name))
}

// stripVersion drops the leading "Unraid Drive 1.3 (29) · " so that an update does not count as a
// different device.
func stripVersion(name string) string {
	if i := strings.Index(name, "·"); i >= 0 && strings.HasPrefix(name, "Unraid Drive") {
		return strings.TrimSpace(name[i+len("·"):])
	}
	return strings.TrimSpace(name)
}

// Touch updates last-seen data for a known device; unknown ids return ErrNotRegistered.
func (s *Store) Touch(id, ip string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.list[id]
	if !ok {
		return ErrNotRegistered
	}
	// An installation that still talks to the gateway is not old.
	revived := d.SupersededBy != ""
	d.SupersededBy, d.SupersededAt = "", time.Time{}
	// Save at most once a minute per device to spare the flash.
	if revived || time.Since(d.LastSeen) > time.Minute {
		d.LastSeen, d.LastIP = time.Now(), ip
		return s.save()
	}
	d.LastSeen, d.LastIP = time.Now(), ip
	return nil
}

// Known reports whether the id is registered.
func (s *Store) Known(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.list[id]
	return ok
}

// Get returns a copy of a device.
func (s *Store) Get(id string) (Device, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.list[id]
	if !ok {
		return Device{}, false
	}
	return *d, true
}

// List returns all devices, most recently seen first.
func (s *Store) List() []Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Device, 0, len(s.list))
	for _, d := range s.sorted() {
		out = append(out, *d)
	}
	return out
}

// Remove deletes one device.
func (s *Store) Remove(id string) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.list[id]
	if !ok {
		return Device{}, ErrNotRegistered
	}
	delete(s.list, id)
	return *d, s.save()
}

// RemoveAll deletes every device.
func (s *Store) RemoveAll() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.list)
	s.list = map[string]*Device{}
	return n, s.save()
}

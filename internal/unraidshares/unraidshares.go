// Package unraidshares derives, for a given Unraid user, the same access level
// the user has over SMB, from Unraid's own configuration.
//
// Two sources are supported (auto-detected from the configured path):
//
//   - /etc/samba/smb-shares.conf (a file, recommended): the share definitions
//     Unraid generates for Samba. World-readable, always current, and by
//     definition what Samba enforces: `public`, `writeable`, `read list`,
//     `write list`, `valid users`.
//   - /boot/config/shares/ (a directory): the per-share .cfg files
//     (shareExport, shareSecurity, shareReadList, shareWriteList). Only
//     readable by root on most systems, kept for completeness.
//
// Semantics (Unraid "SMB Security Settings"): Public → everyone read/write;
// Secure → everyone read, write list read/write; Private → write list
// read/write, read list read, everyone else nothing. Shares not exported via
// SMB grant nothing.
package unraidshares

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sidimam/unraid-gateway/internal/access"
)

// Share is the subset of a share definition the gateway cares about.
type Share struct {
	Name      string
	Exported  bool
	Public    bool // guests allowed (Public and Secure shares)
	Writeable bool // everyone may write (Public shares)
	ReadList  []string
	WriteList []string
	ValidList []string // Private shares: only these users may connect
}

// Level returns what user may do in this share.
func (s Share) Level(user string) access.Level {
	if !s.Exported {
		return access.None
	}
	in := func(list []string) bool {
		for _, u := range list {
			if strings.EqualFold(strings.TrimSpace(u), user) {
				return true
			}
		}
		return false
	}
	if len(s.ValidList) > 0 && !in(s.ValidList) {
		return access.None
	}
	if s.Writeable || in(s.WriteList) {
		return access.Write
	}
	if s.Public || in(s.ReadList) {
		return access.Read
	}
	return access.None
}

// Loader reads the configuration, caching by modification time.
type Loader struct {
	// Path is either the smb-shares.conf file or the shares .cfg directory.
	Path string

	mu     sync.Mutex
	loaded time.Time
	shares map[string]Share
}

// Shares returns all share definitions keyed by lower-case name, reloading when changed.
func (l *Loader) Shares() (map[string]Share, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	st, err := os.Stat(l.Path)
	if err != nil {
		return nil, err
	}
	if l.shares != nil && !st.ModTime().After(l.loaded) && time.Since(l.loaded) < time.Minute {
		return l.shares, nil
	}
	var out map[string]Share
	if st.IsDir() {
		out, err = loadCfgDir(l.Path)
	} else {
		out, err = loadSambaConf(l.Path)
	}
	if err != nil {
		return nil, err
	}
	l.shares, l.loaded = out, time.Now()
	return out, nil
}

// Lookup returns the definition of a share by (case-insensitive) name.
func (l *Loader) Lookup(name string) (Share, bool) {
	shares, err := l.Shares()
	if err != nil {
		return Share{}, false
	}
	s, ok := shares[strings.ToLower(name)]
	return s, ok
}

// PolicyFor builds the access policy of a user over the mounted shares.
// Mounted directories without a share definition grant nothing.
func (l *Loader) PolicyFor(user string, mounted []string) (access.Static, error) {
	shares, err := l.Shares()
	if err != nil {
		return nil, err
	}
	p := access.Static{}
	for _, m := range mounted {
		if s, ok := shares[strings.ToLower(m)]; ok {
			p[m] = s.Level(user)
		} else {
			p[m] = access.None
		}
	}
	return p, nil
}

// ProbeShare picks a share the user should be able to open but a guest could
// not: used to make sure the SMB credentials really belong to that user
// (Samba maps unknown users to guest). "" when no such share exists.
func (l *Loader) ProbeShare(user string, mounted []string) string {
	shares, err := l.Shares()
	if err != nil {
		return ""
	}
	for _, m := range mounted {
		if s, ok := shares[strings.ToLower(m)]; ok && !s.Public && s.Level(user) >= access.Read {
			return s.Name
		}
	}
	return ""
}

// ---- smb-shares.conf -------------------------------------------------------

func loadSambaConf(path string) (map[string]Share, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]Share{}
	var cur *Share
	flush := func() {
		if cur != nil {
			out[strings.ToLower(cur.Name)] = *cur
		}
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			name := strings.TrimSpace(line[1 : len(line)-1])
			if strings.EqualFold(name, "global") {
				cur = nil
				continue
			}
			cur = &Share{Name: name, Exported: true}
			continue
		}
		if cur == nil {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v)
		switch k {
		case "public", "guest ok":
			cur.Public = isYes(v)
		case "writeable", "writable", "write ok":
			cur.Writeable = isYes(v)
		case "read only":
			cur.Writeable = !isYes(v)
		case "read list":
			cur.ReadList = splitList(v)
		case "write list":
			cur.WriteList = splitList(v)
		case "valid users":
			cur.ValidList = splitList(v)
		}
	}
	flush()
	return out, sc.Err()
}

func isYes(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "yes", "true", "1":
		return true
	}
	return false
}

// ---- /boot/config/shares/*.cfg ---------------------------------------------

func loadCfgDir(dir string) (map[string]Share, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := map[string]Share{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cfg") {
			continue
		}
		s, err := parseCfg(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		s.Name = strings.TrimSuffix(e.Name(), ".cfg")
		out[strings.ToLower(s.Name)] = s
	}
	return out, nil
}

func parseCfg(path string) (Share, error) {
	f, err := os.Open(path)
	if err != nil {
		return Share{}, err
	}
	defer f.Close()
	s := Share{}
	security := "private"
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.Trim(strings.TrimSpace(v), `"`)
		switch k {
		case "shareExport":
			s.Exported = strings.Contains(v, "e")
		case "shareSecurity":
			security = strings.ToLower(v)
		case "shareReadList":
			s.ReadList = splitList(v)
		case "shareWriteList":
			s.WriteList = splitList(v)
		}
	}
	switch security {
	case "public":
		s.Public, s.Writeable = true, true
	case "secure":
		s.Public = true
	default: // private: only listed users may connect
		s.ValidList = append(append([]string{}, s.ReadList...), s.WriteList...)
	}
	return s, sc.Err()
}

func splitList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

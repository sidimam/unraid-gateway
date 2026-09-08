// Package unraidshares reads Unraid's per-share configuration files
// (/boot/config/shares/<share>.cfg) and derives, for a given user, the same
// access level the user has over SMB.
//
// Semantics (Unraid "SMB Security Settings"):
//   - shareExport must contain "e" (exported via SMB), otherwise nobody has access;
//   - Public: everyone read/write;
//   - Secure: everyone read, users in shareWriteList read/write;
//   - Private: users in shareWriteList read/write, users in shareReadList read, others nothing.
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

// Share is the subset of a .cfg file the gateway cares about.
type Share struct {
	Name      string
	Exported  bool
	Security  string // public | secure | private
	ReadList  []string
	WriteList []string
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
	switch strings.ToLower(s.Security) {
	case "public":
		return access.Write
	case "secure":
		if in(s.WriteList) {
			return access.Write
		}
		return access.Read
	default: // private
		if in(s.WriteList) {
			return access.Write
		}
		if in(s.ReadList) {
			return access.Read
		}
		return access.None
	}
}

// Loader reads the cfg directory, caching by directory mtime.
type Loader struct {
	Dir string

	mu     sync.Mutex
	loaded time.Time
	shares map[string]Share
}

// Shares returns all share configurations, reloading when files changed.
func (l *Loader) Shares() (map[string]Share, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	st, err := os.Stat(l.Dir)
	if err != nil {
		return nil, err
	}
	if l.shares != nil && !st.ModTime().After(l.loaded) && time.Since(l.loaded) < time.Minute {
		return l.shares, nil
	}
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil, err
	}
	out := map[string]Share{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cfg") {
			continue
		}
		s, err := parseFile(filepath.Join(l.Dir, e.Name()))
		if err != nil {
			continue
		}
		s.Name = strings.TrimSuffix(e.Name(), ".cfg")
		out[strings.ToLower(s.Name)] = s
	}
	l.shares, l.loaded = out, time.Now()
	return out, nil
}

// PolicyFor builds the access policy of a user over the mounted shares.
// Shares mounted in the container but without a cfg file are treated as private with no access.
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

func parseFile(path string) (Share, error) {
	f, err := os.Open(path)
	if err != nil {
		return Share{}, err
	}
	defer f.Close()
	s := Share{Security: "private"}
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
			s.Security = v
		case "shareReadList":
			s.ReadList = splitList(v)
		case "shareWriteList":
			s.WriteList = splitList(v)
		}
	}
	return s, sc.Err()
}

func splitList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

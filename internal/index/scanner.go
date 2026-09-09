package index

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Scanner keeps the index in step with the file system in the background.
//
//   - a full scan (every entry stat'ed) at start-up when the index is empty or
//     stale, then every FullInterval;
//   - a cheap directory scan every DirInterval: only directories are stat'ed and
//     a directory's children are re-listed when its mtime differs from the
//     index (that is what changes when entries are added, removed or renamed).
//
// Everything that goes through the API updates the index immediately, and a
// directory listed by a client is reconciled on the spot, so the scanner only
// has to catch changes made behind the gateway's back (SMB, other containers).
type Scanner struct {
	Index        *Index
	Root         string                 // host path of the data root
	Shares       func() []string        // current share names (mount points)
	Skip         func(name string) bool // entries to ignore (hidden, temp)
	DirInterval  time.Duration
	FullInterval time.Duration
	Log          *slog.Logger
}

const metaLastFull = "last_full_scan"

// Run blocks until ctx is cancelled.
func (s *Scanner) Run(ctx context.Context) {
	if s.DirInterval <= 0 {
		s.DirInterval = 5 * time.Minute
	}
	if s.FullInterval <= 0 {
		s.FullInterval = 6 * time.Hour
	}
	if s.Log == nil {
		s.Log = slog.Default()
	}
	last := s.lastFull()
	if time.Since(last) > s.FullInterval {
		s.full(ctx)
	} else {
		s.logger().Info("index: reusing catalogue", "items", s.count(), "lastFullScan", last.Format(time.RFC3339))
	}
	dirT := time.NewTicker(s.DirInterval)
	fullT := time.NewTicker(s.FullInterval)
	defer dirT.Stop()
	defer fullT.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-dirT.C:
			s.dirs(ctx)
		case <-fullT.C:
			s.full(ctx)
		}
	}
}

func (s *Scanner) logger() *slog.Logger {
	if s.Log == nil {
		return slog.Default()
	}
	return s.Log
}

func (s *Scanner) count() int64 { n, _ := s.Index.Count(); return n }

func (s *Scanner) lastFull() time.Time {
	v, _ := s.Index.GetMeta(metaLastFull)
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ScanDir reconciles one directory (gateway path) from disk. Used by the API
// when a client lists a directory, and by the scanner.
func (s *Scanner) ScanDir(rel string) (map[string]string, error) {
	rel = clean(rel)
	abs := filepath.Join(s.Root, filepath.FromSlash(rel))
	dirents, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	obs := make([]Observed, 0, len(dirents))
	for _, d := range dirents {
		name := d.Name()
		if s.Skip != nil && s.Skip(name) {
			continue
		}
		if rel == "/" && s.Shares != nil && !contains(s.Shares(), name) {
			continue
		}
		fi, err := d.Info()
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			if tfi, err := os.Stat(filepath.Join(abs, name)); err == nil {
				o := FromFileInfo(tfi)
				o.Name = name
				obs = append(obs, o)
			}
			continue
		}
		if !fi.IsDir() && !fi.Mode().IsRegular() {
			continue
		}
		obs = append(obs, FromFileInfo(fi))
	}
	return s.Index.Reconcile(rel, obs)
}

// full walks every share and reconciles every directory.
func (s *Scanner) full(ctx context.Context) {
	start := time.Now()
	n := 0
	if _, err := s.ScanDir("/"); err != nil {
		s.logger().Warn("index: cannot scan root", "err", err)
		return
	}
	for _, share := range s.shares() {
		filepath.WalkDir(filepath.Join(s.Root, share), func(p string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return fs.SkipAll
			}
			if err != nil || !d.IsDir() {
				return nil
			}
			if p != filepath.Join(s.Root, share) && s.Skip != nil && s.Skip(d.Name()) {
				return fs.SkipDir
			}
			rel := "/" + filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(p, s.Root), string(filepath.Separator)))
			if _, err := s.ScanDir(rel); err != nil {
				s.logger().Debug("index: scan failed", "dir", rel, "err", err)
			}
			n++
			return nil
		})
	}
	if ctx.Err() == nil {
		_ = s.Index.SetMeta(metaLastFull, time.Now().UTC().Format(time.RFC3339))
		_ = s.Index.Prune(30 * 24 * time.Hour)
		s.logger().Info("index: full scan done", "directories", n, "items", s.count(), "took", time.Since(start).Round(time.Millisecond))
	}
}

// dirs re-lists only the directories whose mtime changed since they were indexed
// (adding, removing or renaming an entry changes its directory's mtime).
func (s *Scanner) dirs(ctx context.Context) {
	start := time.Now()
	changed := 0
	var walk func(rel string)
	walk = func(rel string) {
		if ctx.Err() != nil {
			return
		}
		it, err := s.Index.ByPath(rel)
		if err != nil || it == nil {
			return
		}
		fi, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(rel)))
		if err != nil {
			return // vanished: the parent's re-listing removes it
		}
		if fi.ModTime().UnixNano() != it.MTime.UnixNano() {
			if _, err := s.ScanDir(rel); err == nil {
				changed++
			}
		}
		children, err := s.Index.Children(rel)
		if err != nil {
			return
		}
		for _, c := range children {
			if c.IsDir {
				walk(c.Path)
			}
		}
	}
	for _, share := range s.shares() {
		walk("/" + share)
	}
	// The root last: it refreshes the share entries (and notices mounted/unmounted shares).
	if _, err := s.ScanDir("/"); err != nil {
		return
	}
	if changed > 0 {
		s.logger().Info("index: directory scan", "changedDirs", changed, "took", time.Since(start).Round(time.Millisecond))
	}
}

func (s *Scanner) shares() []string {
	if s.Shares != nil {
		return s.Shares()
	}
	dirents, err := os.ReadDir(s.Root)
	if err != nil {
		return nil
	}
	var out []string
	for _, d := range dirents {
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			out = append(out, d.Name())
		}
	}
	sort.Strings(out)
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// Rel converts a host path under root to a gateway path.
func Rel(root, abs string) string {
	r, err := filepath.Rel(root, abs)
	if err != nil || r == "." {
		return "/"
	}
	return "/" + filepath.ToSlash(r)
}

var _ = path.Join

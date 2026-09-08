package fsapi

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ErrOutsideRoot is returned when a path escapes the data root.
var ErrOutsideRoot = errors.New("path outside data root")

// Root confines every filesystem operation to a directory.
type Root struct {
	abs string // absolute, symlink-resolved root
}

// NewRoot resolves and validates the data root.
func NewRoot(dir string) (*Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("data root is not a directory")
	}
	return &Root{abs: resolved}, nil
}

// Path returns the absolute root directory.
func (r *Root) Path() string { return r.abs }

// Clean normalises a client-supplied path to a canonical, rooted, slash form.
func Clean(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	c := path.Clean(p)
	return c
}

// Resolve maps a client path to an absolute filesystem path inside the root.
// It refuses traversal and symlinks that point outside the root. The target
// itself need not exist; the closest existing ancestor is checked instead.
func (r *Root) Resolve(clientPath string) (string, error) {
	rel := strings.TrimPrefix(Clean(clientPath), "/")
	if rel == "" {
		return r.abs, nil
	}
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." || seg == "" {
			return "", ErrOutsideRoot
		}
	}
	abs := filepath.Join(r.abs, filepath.FromSlash(rel))
	if !r.within(abs) {
		return "", ErrOutsideRoot
	}
	// Resolve symlinks on the deepest existing ancestor (or the path itself).
	probe := abs
	var tail []string
	for {
		resolved, err := filepath.EvalSymlinks(probe)
		if err == nil {
			if !r.within(resolved) {
				return "", ErrOutsideRoot
			}
			if len(tail) == 0 {
				return resolved, nil
			}
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			if !r.within(resolved) {
				return "", ErrOutsideRoot
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent, base := filepath.Split(probe)
		parent = strings.TrimRight(parent, string(filepath.Separator))
		if parent == "" || len(parent) < len(r.abs) {
			return "", ErrOutsideRoot
		}
		tail = append(tail, base)
		probe = parent
	}
}

// Rel converts an absolute path back into a client path.
func (r *Root) Rel(abs string) string {
	rel, err := filepath.Rel(r.abs, abs)
	if err != nil || rel == "." {
		return "/"
	}
	return "/" + filepath.ToSlash(rel)
}

func (r *Root) within(abs string) bool {
	return abs == r.abs || strings.HasPrefix(abs, r.abs+string(filepath.Separator))
}

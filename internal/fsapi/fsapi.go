// Package fsapi implements the file API used by the iOS File Provider
// extension: listing, streaming download/upload, mutations and a change feed.
package fsapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Options tunes the file API.
type Options struct {
	ReadOnly         bool
	MaxJSONBody      int64
	ChangesWalkLimit int
	ChangesDeadline  time.Duration
	UploadTTL        time.Duration
}

// API serves the /fs endpoints.
type API struct {
	root    *Root
	opts    Options
	log     *slog.Logger
	uploads *uploadStore
}

// New creates the file API rooted at dir.
func New(dir string, opts Options, log *slog.Logger) (*API, error) {
	root, err := NewRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("data root: %w", err)
	}
	if opts.MaxJSONBody <= 0 {
		opts.MaxJSONBody = 1 << 20
	}
	if opts.ChangesWalkLimit <= 0 {
		opts.ChangesWalkLimit = 250000
	}
	if opts.UploadTTL <= 0 {
		opts.UploadTTL = 24 * time.Hour
	}
	if opts.ChangesDeadline <= 0 {
		opts.ChangesDeadline = 20 * time.Second
	}
	a := &API{root: root, opts: opts, log: log}
	a.uploads = newUploadStore(a, opts.UploadTTL)
	return a, nil
}

// Entry is the JSON representation of a file or directory.
type Entry struct {
	Name  string    `json:"name"`
	Path  string    `json:"path"`
	Type  string    `json:"type"` // "file" | "dir"
	Size  int64     `json:"size"`
	MTime time.Time `json:"mtime"`
	ETag  string    `json:"etag"`
	Mode  string    `json:"mode,omitempty"`
}

func (a *API) entry(abs string, info fs.FileInfo) Entry {
	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}
	return Entry{
		Name:  info.Name(),
		Path:  a.root.Rel(abs),
		Type:  typ,
		Size:  info.Size(),
		MTime: info.ModTime().UTC(),
		ETag:  etag(info),
		Mode:  info.Mode().Perm().String(),
	}
}

func etag(info fs.FileInfo) string {
	return fmt.Sprintf("\"%x-%x\"", info.Size(), info.ModTime().UnixNano())
}

// Register mounts the handlers on mux under prefix (e.g. "/api/v1/fs").
func (a *API) Register(mux *http.ServeMux, prefix string) {
	p := strings.TrimRight(prefix, "/")
	mux.HandleFunc("GET "+p+"/list", a.handleList)
	mux.HandleFunc("GET "+p+"/stat", a.handleStat)
	mux.HandleFunc("GET "+p+"/content", a.handleDownload)
	mux.HandleFunc("HEAD "+p+"/content", a.handleDownload)
	mux.HandleFunc("PUT "+p+"/content", a.mutating(a.handlePut))
	mux.HandleFunc("POST "+p+"/mkdir", a.mutating(a.handleMkdir))
	mux.HandleFunc("POST "+p+"/move", a.mutating(a.handleMove))
	mux.HandleFunc("POST "+p+"/copy", a.mutating(a.handleCopy))
	mux.HandleFunc("POST "+p+"/delete", a.mutating(a.handleDelete))
	mux.HandleFunc("GET "+p+"/changes", a.handleChanges)
	mux.HandleFunc("POST "+p+"/uploads", a.mutating(a.uploads.handleCreate))
	mux.HandleFunc("HEAD "+p+"/uploads/{id}", a.uploads.handleStatus)
	mux.HandleFunc("GET "+p+"/uploads/{id}", a.uploads.handleStatus)
	mux.HandleFunc("PATCH "+p+"/uploads/{id}", a.mutating(a.uploads.handleAppend))
	mux.HandleFunc("POST "+p+"/uploads/{id}/commit", a.mutating(a.uploads.handleCommit))
	mux.HandleFunc("DELETE "+p+"/uploads/{id}", a.uploads.handleAbort)
}

func (a *API) mutating(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.opts.ReadOnly {
			writeErr(w, http.StatusForbidden, "gateway is read-only")
			return
		}
		h(w, r)
	}
}

// ---- helpers ---------------------------------------------------------------

// isShareRoot reports whether abs is a direct child of the data root, i.e. a
// mounted share. Shares are mount points: they can be listed and written
// into, but never renamed, moved, replaced or deleted through the API.
func (a *API) isShareRoot(abs string) bool {
	return filepath.Dir(abs) == a.root.Path() && abs != a.root.Path()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *API) fsErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrOutsideRoot):
		writeErr(w, http.StatusBadRequest, "invalid path")
	case errors.Is(err, os.ErrNotExist):
		writeErr(w, http.StatusNotFound, "not found")
	case errors.Is(err, os.ErrExist):
		writeErr(w, http.StatusConflict, "already exists")
	case errors.Is(err, os.ErrPermission):
		writeErr(w, http.StatusForbidden, "permission denied")
	case errors.Is(err, syscall.EROFS):
		writeErr(w, http.StatusForbidden, "share is mounted read-only")
	case errors.Is(err, syscall.ENOSPC):
		writeErr(w, http.StatusInsufficientStorage, "no space left on share")
	default:
		a.log.Error("fs error", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
	}
}

func (a *API) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, a.opts.MaxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

func queryPath(r *http.Request) string {
	p := r.URL.Query().Get("path")
	if p == "" {
		p = "/"
	}
	return p
}

// ---- read handlers ---------------------------------------------------------

type listResponse struct {
	Path    string    `json:"path"`
	ETag    string    `json:"etag"`
	MTime   time.Time `json:"mtime"`
	Entries []Entry   `json:"entries"`
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if !info.IsDir() {
		writeErr(w, http.StatusBadRequest, "not a directory")
		return
	}
	dirents, err := os.ReadDir(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	showHidden := r.URL.Query().Get("hidden") == "1"
	out := make([]Entry, 0, len(dirents))
	for _, d := range dirents {
		name := d.Name()
		if isInternal(name) || (!showHidden && strings.HasPrefix(name, ".")) {
			continue
		}
		fi, err := d.Info()
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			// Present symlinks as their target when it stays inside the root.
			target, err := a.root.Resolve(a.root.Rel(filepath.Join(abs, name)))
			if err != nil {
				continue
			}
			if tfi, err := os.Stat(target); err == nil {
				fi = renamed{tfi, name}
			} else {
				continue
			}
		}
		if !fi.IsDir() && !fi.Mode().IsRegular() {
			continue
		}
		out = append(out, a.entry(filepath.Join(abs, name), fi))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type == "dir"
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	w.Header().Set("ETag", etag(info))
	writeJSON(w, http.StatusOK, listResponse{Path: a.root.Rel(abs), ETag: etag(info), MTime: info.ModTime().UTC(), Entries: out})
}

// renamed lets a symlink keep its own name while reporting the target's stat.
type renamed struct {
	fs.FileInfo
	name string
}

func (r renamed) Name() string { return r.name }

func isInternal(name string) bool {
	return strings.Contains(name, partSuffix)
}

func (a *API) handleStat(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	w.Header().Set("ETag", etag(info))
	writeJSON(w, http.StatusOK, a.entry(abs, info))
}

func (a *API) handleDownload(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	f, err := os.Open(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if info.IsDir() {
		writeErr(w, http.StatusBadRequest, "is a directory")
		return
	}
	w.Header().Set("ETag", etag(info))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+escapeFilename(info.Name()))
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func escapeFilename(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '-' || c == '_' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// ---- write handlers --------------------------------------------------------

// handlePut streams a whole file to a temporary sibling and renames it into
// place so readers never observe a partial file.
func (a *API) handlePut(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if abs == a.root.Path() {
		writeErr(w, http.StatusBadRequest, "cannot write to root")
		return
	}
	existing, statErr := os.Stat(abs)
	if statErr == nil {
		if existing.IsDir() {
			writeErr(w, http.StatusConflict, "target is a directory")
			return
		}
		if r.URL.Query().Get("overwrite") == "false" {
			writeErr(w, http.StatusConflict, "already exists")
			return
		}
		if m := r.Header.Get("If-Match"); m != "" && m != etag(existing) {
			writeErr(w, http.StatusPreconditionFailed, "etag mismatch")
			return
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		a.fsErr(w, statErr)
		return
	} else if r.Header.Get("If-Match") != "" {
		writeErr(w, http.StatusPreconditionFailed, "target does not exist")
		return
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o775); err != nil {
		a.fsErr(w, err)
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(abs), "."+filepath.Base(abs)+partSuffix)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	tmpName := tmp.Name()
	cleanup := func() { tmp.Close(); os.Remove(tmpName) }
	if _, err := io.Copy(tmp, r.Body); err != nil {
		cleanup()
		a.log.Warn("upload aborted", "path", a.root.Rel(abs), "err", err)
		writeErr(w, http.StatusBadRequest, "upload interrupted")
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		a.fsErr(w, err)
		return
	}
	if mt := r.Header.Get("X-Mtime"); mt != "" {
		if t, err := time.Parse(time.RFC3339Nano, mt); err == nil {
			_ = os.Chtimes(tmpName, t, t)
		}
	}
	_ = os.Chmod(tmpName, 0o664)
	if err := os.Rename(tmpName, abs); err != nil {
		os.Remove(tmpName)
		a.fsErr(w, err)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	status := http.StatusCreated
	if statErr == nil {
		status = http.StatusOK
	}
	w.Header().Set("ETag", etag(info))
	writeJSON(w, status, a.entry(abs, info))
}

type pathRequest struct {
	Path      string `json:"path"`
	Parents   bool   `json:"parents,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
}

func (a *API) handleMkdir(w http.ResponseWriter, r *http.Request) {
	var req pathRequest
	if !a.decode(w, r, &req) {
		return
	}
	abs, err := a.root.Resolve(req.Path)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if abs == a.root.Path() {
		writeErr(w, http.StatusConflict, "already exists")
		return
	}
	if a.isShareRoot(abs) {
		writeErr(w, http.StatusForbidden, "the root only contains mounted shares; create folders inside a share")
		return
	}
	if req.Parents {
		err = os.MkdirAll(abs, 0o775)
	} else {
		err = os.Mkdir(abs, 0o775)
	}
	if err != nil {
		a.fsErr(w, err)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a.entry(abs, info))
}

type moveRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Overwrite bool   `json:"overwrite,omitempty"`
}

func (a *API) resolvePair(w http.ResponseWriter, req moveRequest) (string, string, bool) {
	from, err := a.root.Resolve(req.From)
	if err != nil {
		a.fsErr(w, err)
		return "", "", false
	}
	to, err := a.root.Resolve(req.To)
	if err != nil {
		a.fsErr(w, err)
		return "", "", false
	}
	if from == a.root.Path() || to == a.root.Path() {
		writeErr(w, http.StatusBadRequest, "root cannot be moved or replaced")
		return "", "", false
	}
	if a.isShareRoot(from) || a.isShareRoot(to) {
		writeErr(w, http.StatusForbidden, "shares are mount points and cannot be moved, renamed or replaced")
		return "", "", false
	}
	if to == from || strings.HasPrefix(to, from+string(filepath.Separator)) {
		writeErr(w, http.StatusBadRequest, "destination is inside source")
		return "", "", false
	}
	if _, err := os.Lstat(from); err != nil {
		a.fsErr(w, err)
		return "", "", false
	}
	if _, err := os.Lstat(to); err == nil {
		if !req.Overwrite {
			writeErr(w, http.StatusConflict, "destination already exists")
			return "", "", false
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		a.fsErr(w, err)
		return "", "", false
	}
	return from, to, true
}

func (a *API) handleMove(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if !a.decode(w, r, &req) {
		return
	}
	from, to, ok := a.resolvePair(w, req)
	if !ok {
		return
	}
	if req.Overwrite {
		if st, err := os.Lstat(to); err == nil && st.IsDir() {
			if err := os.RemoveAll(to); err != nil {
				a.fsErr(w, err)
				return
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o775); err != nil {
		a.fsErr(w, err)
		return
	}
	if err := os.Rename(from, to); err != nil {
		// Cross-device (e.g. share -> different pool): fall back to copy+delete.
		if cerr := copyTree(from, to); cerr != nil {
			a.fsErr(w, err)
			return
		}
		if rerr := os.RemoveAll(from); rerr != nil {
			a.fsErr(w, rerr)
			return
		}
	}
	info, err := os.Stat(to)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a.entry(to, info))
}

func (a *API) handleCopy(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if !a.decode(w, r, &req) {
		return
	}
	from, to, ok := a.resolvePair(w, req)
	if !ok {
		return
	}
	if req.Overwrite {
		if err := os.RemoveAll(to); err != nil {
			a.fsErr(w, err)
			return
		}
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o775); err != nil {
		a.fsErr(w, err)
		return
	}
	if err := copyTree(from, to); err != nil {
		a.fsErr(w, err)
		return
	}
	info, err := os.Stat(to)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a.entry(to, info))
}

func copyTree(src, dst string) error {
	st, err := os.Lstat(src)
	if err != nil {
		return err
	}
	switch {
	case st.IsDir():
		if err := os.MkdirAll(dst, st.Mode().Perm()|0o700); err != nil {
			return err
		}
		ents, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range ents {
			if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return os.Chtimes(dst, st.ModTime(), st.ModTime())
	case st.Mode().IsRegular():
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm()|0o600)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
		return os.Chtimes(dst, st.ModTime(), st.ModTime())
	default:
		return nil // skip symlinks and special files
	}
}

func (a *API) handleDelete(w http.ResponseWriter, r *http.Request) {
	var req pathRequest
	if !a.decode(w, r, &req) {
		return
	}
	abs, err := a.root.Resolve(req.Path)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if abs == a.root.Path() {
		writeErr(w, http.StatusBadRequest, "cannot delete root")
		return
	}
	if a.isShareRoot(abs) {
		writeErr(w, http.StatusForbidden, "shares are mount points and cannot be deleted")
		return
	}
	st, err := os.Lstat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if st.IsDir() && !req.Recursive {
		err = os.Remove(abs) // fails if non-empty
		if err != nil {
			var pe *os.PathError
			if errors.As(err, &pe) {
				writeErr(w, http.StatusConflict, "directory not empty (use recursive)")
				return
			}
		}
	} else {
		err = os.RemoveAll(abs)
	}
	if err != nil {
		a.fsErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- change feed -----------------------------------------------------------

type changesResponse struct {
	Path      string   `json:"path"`
	Since     int64    `json:"since"`
	Cursor    int64    `json:"cursor"`
	Dirs      []string `json:"dirs"`
	Files     []Entry  `json:"files"`
	Truncated bool     `json:"truncated"`
	Next      string   `json:"next,omitempty"`
	Scanned   int      `json:"scanned"`
}

// handleChanges walks the subtree and reports directories whose mtime changed
// (entries added/removed/renamed inside them) and files modified after the
// cursor. A client re-enumerates the listed directories and refreshes the
// listed files. since=0 returns everything.
//
// Large trees on shfs are slow to walk, so a scan is bounded by
// ChangesWalkLimit entries and ChangesDeadline. When it stops early the
// response has truncated=true and next=<last path scanned>; the client repeats
// the call with after=<next> (and the same since) to continue exactly where
// the walk stopped, then stores cursor only after a non-truncated page.
func (a *API) handleChanges(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	var since int64
	if s := r.URL.Query().Get("since"); s != "" {
		since, err = strconv.ParseInt(s, 10, 64)
		if err != nil || since < 0 {
			writeErr(w, http.StatusBadRequest, "since must be unix nanoseconds")
			return
		}
	}
	limit := a.opts.ChangesWalkLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n < limit {
			limit = n
		}
	}
	after := ""
	if q := r.URL.Query().Get("after"); q != "" {
		after = Clean(q)
	}
	// Cursor is taken before the walk so anything modified during the walk is
	// reported again next time rather than lost.
	cursor := time.Now().UnixNano() - int64(2*time.Second)
	if c := r.URL.Query().Get("cursor"); c != "" {
		// Continuation pages keep the cursor of the first page.
		if n, err := strconv.ParseInt(c, 10, 64); err == nil && n > 0 && n < cursor {
			cursor = n
		}
	}
	resp := changesResponse{Path: a.root.Rel(abs), Since: since, Cursor: cursor, Dirs: []string{}, Files: []Entry{}}
	deadline := time.Now().Add(a.opts.ChangesDeadline)
	last := ""
	walkErr := filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entries are skipped
		}
		rel := a.root.Rel(p)
		name := d.Name()
		if p != abs && (strings.HasPrefix(name, ".") || isInternal(name)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		// Resume support: skip everything at or before `after` in walk order.
		if after != "" && walkCompare(rel, after) <= 0 {
			if d.IsDir() && (rel == after || isAncestor(rel, after)) {
				// The resume point itself or one of its ancestors: already
				// reported, but its children still have to be walked.
				return nil
			}
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		resp.Scanned++
		if resp.Scanned > limit || time.Now().After(deadline) {
			resp.Truncated = true
			resp.Next = last
			return fs.SkipAll
		}
		last = rel
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().UnixNano() <= since {
			return nil
		}
		if d.IsDir() {
			resp.Dirs = append(resp.Dirs, rel)
		} else if info.Mode().IsRegular() {
			resp.Files = append(resp.Files, a.entry(p, info))
		}
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
		a.fsErr(w, walkErr)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// walkCompare orders two client paths the way filepath.WalkDir visits them:
// component by component, with a directory before its own children.
func walkCompare(a, b string) int {
	as := strings.Split(strings.Trim(a, "/"), "/")
	bs := strings.Split(strings.Trim(b, "/"), "/")
	if a == "/" {
		as = nil
	}
	if b == "/" {
		bs = nil
	}
	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] != bs[i] {
			if as[i] < bs[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(as) < len(bs):
		return -1
	case len(as) > len(bs):
		return 1
	}
	return 0
}

// isAncestor reports whether dir is "/" or a strict path prefix of p.
func isAncestor(dir, p string) bool {
	if dir == "/" {
		return p != "/"
	}
	return strings.HasPrefix(p, dir+"/")
}

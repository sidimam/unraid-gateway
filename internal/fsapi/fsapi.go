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

	"github.com/sidimam/unraid-gateway/internal/access"
	"github.com/sidimam/unraid-gateway/internal/index"
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
	root *Root
	opts Options
	log  *slog.Logger
	// Signs media tickets (see ticket.go); random per process.
	ticketSecret []byte
	uploads      *uploadStore
	idx          *index.Index
	scanner      *index.Scanner
}

// UseIndex attaches the persistent item index: listings then carry stable ids,
// writes are journaled and /fs/changes can answer from the journal.
func (a *API) UseIndex(ix *index.Index, sc *index.Scanner) {
	a.idx = ix
	a.scanner = sc
}

// Index returns the attached index (nil when disabled).
func (a *API) Index() *index.Index { return a.idx }

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
	a := &API{root: root, opts: opts, log: log, ticketSecret: newTicketSecret()}
	a.uploads = newUploadStore(a, opts.UploadTTL)
	return a, nil
}

// Entry is the JSON representation of a file or directory.
type Entry struct {
	ID       string    `json:"id,omitempty"`
	ParentID string    `json:"parentId,omitempty"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Type     string    `json:"type"` // "file" | "dir"
	Size     int64     `json:"size"`
	MTime    time.Time `json:"mtime"`
	ETag     string    `json:"etag"`
	Mode     string    `json:"mode,omitempty"`
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

// withID fills ID/ParentID from the index, indexing the entry on the fly when needed.
func (a *API) withID(e Entry, info fs.FileInfo) Entry {
	if a.idx == nil {
		return e
	}
	it, err := a.idx.ByPath(e.Path)
	if err == nil && it == nil {
		if _, err := a.idx.Upsert(e.Path, index.FromFileInfo(info)); err == nil {
			it, _ = a.idx.ByPath(e.Path)
		}
	}
	if it != nil {
		e.ID, e.ParentID = it.ID, it.ParentID
	}
	return e
}

// indexed records a path just written through the API.
func (a *API) indexed(abs string, info fs.FileInfo) {
	if a.idx == nil {
		return
	}
	if _, err := a.idx.Upsert(a.root.Rel(abs), index.FromFileInfo(info)); err != nil {
		a.log.Warn("index update failed", "path", a.root.Rel(abs), "err", err)
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
	mux.HandleFunc("GET "+p+"/item", a.handleItem)
	mux.HandleFunc("POST "+p+"/ticket", a.handleTicket)
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

// errNoShare is returned for paths whose first component is not a mounted share.
var errNoShare = errors.New("share not found")

// checkShare verifies that the share (first path component) of abs exists as
// a directory, so writes can never create new top-level entries in the data
// root by accident (e.g. a PUT to /typo/file.txt).
func (a *API) checkShare(abs string) error {
	rel, err := filepath.Rel(a.root.Path(), abs)
	if err != nil || rel == "." {
		return errNoShare
	}
	first := strings.SplitN(rel, string(filepath.Separator), 2)[0]
	if !a.isShare(first) {
		return errNoShare
	}
	return nil
}

// shareOf returns the share name (first component under the data root) of abs, or "" for the root.
func (a *API) shareOf(abs string) string {
	rel, err := filepath.Rel(a.root.Path(), abs)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return ""
	}
	return strings.SplitN(rel, string(filepath.Separator), 2)[0]
}

// allowed checks the request's share policy for abs and writes 403/404 when it fails.
// Shares the principal may not even read are reported as not found, like SMB does.
func (a *API) allowed(w http.ResponseWriter, r *http.Request, abs string, need access.Level) bool {
	pol := access.FromContext(r.Context())
	if !pol.Restricted() {
		return true
	}
	share := a.shareOf(abs)
	if share == "" {
		if need > access.Read {
			writeErr(w, http.StatusForbidden, "cannot write at root level")
			return false
		}
		return true
	}
	have := pol.Level(share)
	switch {
	case have == access.None:
		writeErr(w, http.StatusNotFound, "not found")
		return false
	case have < need:
		writeErr(w, http.StatusForbidden, "this share is read-only for your user")
		return false
	}
	return true
}

// MountedShares lists the shares: the top-level directories of the data root
// that are real mount points. Directories left behind in the root volume after
// a mapping was removed are ignored, so a share disappears from the API as
// soon as its mount is gone. When mount information is unavailable (tests,
// non-Linux), every top-level directory counts.
func (a *API) MountedShares() []string {
	ents, err := os.ReadDir(a.root.Path())
	if err != nil {
		return nil
	}
	mounts := a.mountPoints()
	var out []string
	for _, e := range ents {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if mounts != nil && !mounts[filepath.Join(a.root.Path(), e.Name())] {
			continue
		}
		out = append(out, e.Name())
	}
	return out
}

// isShare reports whether name is a share according to MountedShares.
func (a *API) isShare(name string) bool {
	for _, s := range a.MountedShares() {
		if s == name {
			return true
		}
	}
	return false
}

// mountPoints returns the mount points directly under the data root from
// /proc/self/mountinfo, or nil when there are none / it cannot be read.
func (a *API) mountPoints() map[string]bool {
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return nil
	}
	out := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) < 5 {
			continue
		}
		mp := strings.ReplaceAll(f[4], "\\040", " ")
		if filepath.Dir(mp) == a.root.Path() {
			out[mp] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

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
	case errors.Is(err, errNoShare):
		writeErr(w, http.StatusNotFound, "share not found")
	case errors.Is(err, os.ErrNotExist):
		writeErr(w, http.StatusNotFound, "not found")
	case errors.Is(err, os.ErrExist):
		writeErr(w, http.StatusConflict, "already exists")
	case errors.Is(err, os.ErrPermission):
		msg := a.permissionDetail(err)
		a.log.Warn("permission denied", "detail", msg)
		writeErr(w, http.StatusForbidden, msg)
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
	if !a.allowed(w, r, abs, access.Read) {
		return
	}
	pol := access.FromContext(r.Context())
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
		// At the root only real shares are listed, and only those the user may read.
		if abs == a.root.Path() {
			if !a.isShare(name) {
				continue
			}
			if pol.Restricted() && pol.Level(name) == access.None {
				continue
			}
		}
		out = append(out, a.entry(filepath.Join(abs, name), fi))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type == "dir"
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	if a.scanner != nil {
		// On-demand reconcile: the listing the client sees is what the index knows.
		if ids, err := a.scanner.ScanDir(a.root.Rel(abs)); err == nil {
			parentID := index.RootID
			if abs != a.root.Path() {
				if it, _ := a.idx.ByPath(a.root.Rel(abs)); it != nil {
					parentID = it.ID
				}
			}
			for i := range out {
				out[i].ID = ids[out[i].Name]
				out[i].ParentID = parentID
			}
		} else {
			a.log.Warn("index reconcile failed", "dir", a.root.Rel(abs), "err", err)
		}
	}
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
	if !a.allowed(w, r, abs, access.Read) {
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	w.Header().Set("ETag", etag(info))
	writeJSON(w, http.StatusOK, a.withID(a.entry(abs, info), info))
}

// handleItem resolves an item id to its current entry (GET /item?id=...), so a
// client that only remembers ids (after a reinstall, say) can find its way back.
func (a *API) handleItem(w http.ResponseWriter, r *http.Request) {
	if a.idx == nil {
		writeErr(w, http.StatusNotFound, "item index disabled")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}
	it, err := a.idx.ByID(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if it == nil {
		writeErr(w, http.StatusNotFound, "unknown item")
		return
	}
	abs, err := a.root.Resolve(it.Path)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if !a.allowed(w, r, abs, access.Read) {
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		_ = a.idx.Delete(it.Path)
		a.fsErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a.withID(a.entry(abs, info), info))
}

func (a *API) handleDownload(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if !a.allowed(w, r, abs, access.Read) {
		return
	}
	a.serveFile(w, r, abs, r.URL.Query().Get("download") == "1")
}

// serveFile streams a regular file with ETag and Range support (access already checked).
func (a *API) serveFile(w http.ResponseWriter, r *http.Request, abs string, attachment bool) {
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
	if attachment {
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
	if err := a.checkShare(abs); err != nil {
		a.fsErr(w, err)
		return
	}
	if !a.allowed(w, r, abs, access.Write) {
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
	if err := os.MkdirAll(filepath.Dir(abs), DirMode); err != nil {
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
	_ = os.Chmod(tmpName, FileMode)
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
	a.indexed(abs, info)
	w.Header().Set("ETag", etag(info))
	writeJSON(w, status, a.withID(a.entry(abs, info), info))
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
	if err := a.checkShare(abs); err != nil {
		a.fsErr(w, err)
		return
	}
	if !a.allowed(w, r, abs, access.Write) {
		return
	}
	if req.Parents {
		err = os.MkdirAll(abs, DirMode)
	} else {
		err = os.Mkdir(abs, DirMode)
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
	a.indexed(abs, info)
	writeJSON(w, http.StatusCreated, a.withID(a.entry(abs, info), info))
}

type moveRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Overwrite bool   `json:"overwrite,omitempty"`
}

func (a *API) resolvePair(w http.ResponseWriter, r *http.Request, req moveRequest, sourceNeeds access.Level) (string, string, bool) {
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
	if err := a.checkShare(to); err != nil {
		a.fsErr(w, err)
		return "", "", false
	}
	if to == from || strings.HasPrefix(to, from+string(filepath.Separator)) {
		writeErr(w, http.StatusBadRequest, "destination is inside source")
		return "", "", false
	}
	if !a.allowed(w, r, from, sourceNeeds) || !a.allowed(w, r, to, access.Write) {
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
	from, to, ok := a.resolvePair(w, r, req, access.Write)
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
	if err := os.MkdirAll(filepath.Dir(to), DirMode); err != nil {
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
	if a.idx != nil {
		if err := a.idx.Move(a.root.Rel(from), a.root.Rel(to)); err != nil {
			a.log.Warn("index move failed", "err", err)
		}
		a.indexed(to, info)
	}
	writeJSON(w, http.StatusOK, a.withID(a.entry(to, info), info))
}

func (a *API) handleCopy(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if !a.decode(w, r, &req) {
		return
	}
	from, to, ok := a.resolvePair(w, r, req, access.Read)
	if !ok {
		return
	}
	if req.Overwrite {
		if err := os.RemoveAll(to); err != nil {
			a.fsErr(w, err)
			return
		}
	}
	if err := os.MkdirAll(filepath.Dir(to), DirMode); err != nil {
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
	a.indexed(to, info)
	if a.scanner != nil && info.IsDir() {
		// Index the copied subtree so the change feed announces it.
		filepath.WalkDir(to, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				_, _ = a.scanner.ScanDir(a.root.Rel(p))
			}
			return nil
		})
	}
	writeJSON(w, http.StatusCreated, a.withID(a.entry(to, info), info))
}

func copyTree(src, dst string) error {
	st, err := os.Lstat(src)
	if err != nil {
		return err
	}
	switch {
	case st.IsDir():
		if err := os.MkdirAll(dst, st.Mode().Perm()|DirMode); err != nil {
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
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm()|FileMode)
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
	if !a.allowed(w, r, abs, access.Write) {
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
	if a.idx != nil {
		if err := a.idx.Delete(a.root.Rel(abs)); err != nil {
			a.log.Warn("index delete failed", "err", err)
		}
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
// journalChange is one entry of the id-based change feed (index mode).
type journalChange struct {
	Seq     int64  `json:"seq"`
	Kind    string `json:"kind"` // upsert | delete | move
	ID      string `json:"id"`
	Path    string `json:"path"`
	OldPath string `json:"oldPath,omitempty"`
	Entry   *Entry `json:"entry,omitempty"`
}

type journalResponse struct {
	Seq       int64           `json:"seq"`   // pass back as seq= next time
	Reset     bool            `json:"reset"` // true: forget everything and enumerate again, then continue from seq
	Changes   []journalChange `json:"changes"`
	Truncated bool            `json:"truncated"` // more available: call again with seq=<Seq>
}

// handleJournal answers /fs/changes?seq=N from the index journal: every item
// created, modified, moved or deleted after sequence N, with stable ids. It is
// instant regardless of the tree size and needs no walk.
func (a *API) handleJournal(w http.ResponseWriter, r *http.Request, since int64) {
	pol := access.FromContext(r.Context())
	limit := 1000
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 5000 {
			limit = n
		}
	}
	changes, latest, truncated, reset, err := a.idx.Changes(since, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := journalResponse{Seq: latest, Reset: reset, Changes: []journalChange{}, Truncated: truncated}
	// Compaction: an item deleted later in this page needs no earlier upsert/move.
	deleted := map[string]bool{}
	for _, c := range changes {
		if c.Kind == "delete" {
			deleted[c.ID] = true
		}
	}
	for _, c := range changes {
		if c.Kind != "delete" && deleted[c.ID] {
			continue
		}
		if c.ID == index.RootID || c.Path == "/" {
			continue // bootstrap marker
		}
		share := strings.SplitN(strings.TrimPrefix(c.Path, "/"), "/", 2)[0]
		if pol.Restricted() && pol.Level(share) == access.None {
			continue // shares this user may not see never appear in their feed
		}
		if strings.Contains(c.Path, "/.") || isInternal(c.Path) {
			continue
		}
		jc := journalChange{Seq: c.Seq, Kind: c.Kind, ID: c.ID, Path: c.Path, OldPath: c.OldPath}
		if c.Item != nil {
			typ := "file"
			if c.Item.IsDir {
				typ = "dir"
			}
			jc.Entry = &Entry{ID: c.Item.ID, ParentID: c.Item.ParentID, Name: c.Item.Name, Path: c.Item.Path, Type: typ,
				Size: c.Item.Size, MTime: c.Item.MTime, ETag: fmt.Sprintf("\"%x-%x\"", c.Item.Size, c.Item.MTime.UnixNano())}
		} else if c.Kind != "delete" {
			// The item vanished after the change was journaled: report it as gone.
			jc.Kind = "delete"
		}
		resp.Changes = append(resp.Changes, jc)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) handleChanges(w http.ResponseWriter, r *http.Request) {
	abs, err := a.root.Resolve(queryPath(r))
	if err != nil {
		a.fsErr(w, err)
		return
	}
	if !a.allowed(w, r, abs, access.Read) {
		return
	}
	if a.idx != nil {
		if q := r.URL.Query().Get("seq"); q != "" {
			n, err := strconv.ParseInt(q, 10, 64)
			if err != nil || n < 0 {
				writeErr(w, http.StatusBadRequest, "seq must be a non-negative integer")
				return
			}
			a.handleJournal(w, r, n)
			return
		}
	}
	pol := access.FromContext(r.Context())
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
		// Skip whole shares the user may not read.
		if pol.Restricted() && d.IsDir() && a.isShareRoot(p) && pol.Level(name) == access.None {
			return fs.SkipDir
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

// permissionDetail explains an EACCES in terms the Unraid admin can act on:
// which folder the gateway could not write, who owns it and with which mode,
// and what to do about it. Files created via SSH, rsync or other containers
// often end up 0755/0644 and owned by another user, which locks out the
// gateway's nobody:users account.
func (a *API) permissionDetail(err error) string {
	var pe *fs.PathError
	if !errors.As(err, &pe) || pe.Path == "" {
		return "permission denied: the gateway account " + gatewayIdentity() + " may not write here. On Unraid run Tools › New Permissions on the share."
	}
	// Describe the closest existing path (the target or its parent folder).
	target := pe.Path
	st, statErr := os.Lstat(target)
	for statErr != nil && target != "/" && target != "." {
		target = filepath.Dir(target)
		st, statErr = os.Lstat(target)
	}
	if statErr != nil {
		return "permission denied: the gateway account " + gatewayIdentity() + " may not write here. On Unraid run Tools › New Permissions on the share."
	}
	shown := a.root.Rel(target)
	if !a.root.within(target) {
		shown = target
	}
	return fmt.Sprintf("permission denied: the gateway runs as %s and may not write in %q (%s, mode %04o). On Unraid run Tools › New Permissions on this share, or chmod -R ugo+rwX the folder.",
		gatewayIdentity(), shown, ownerOf(st), st.Mode().Perm())
}

// WarnUnwritableShares logs, at start-up, every mounted share the gateway
// account cannot write to although the mount is read-write: typically a
// folder created over SSH or by another container with a restrictive owner
// or mode. Read-only mounts are legitimate and skipped.
func (a *API) WarnUnwritableShares() {
	if a.opts.ReadOnly {
		return
	}
	for _, share := range a.MountedShares() {
		dir := filepath.Join(a.root.Path(), share)
		f, err := os.CreateTemp(dir, ".gw-write-test-*")
		if err == nil {
			f.Close()
			os.Remove(f.Name())
			continue
		}
		if errors.Is(err, syscall.EROFS) {
			continue
		}
		if errors.Is(err, os.ErrPermission) {
			a.log.Warn("share not writable by the gateway account: run Tools › New Permissions on it in Unraid", "share", share, "detail", a.permissionDetail(err))
			continue
		}
		a.log.Warn("share write test failed", "share", share, "err", err)
	}
}

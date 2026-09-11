package fsapi

import (
	"github.com/sidimam/unraid-gateway/internal/access"

	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// partSuffix marks temporary files created by the gateway; they are hidden
// from listings and never reported by the change feed.
const partSuffix = ".gwpart"

// uploadStore implements resumable uploads: a client creates a session,
// appends chunks at an explicit offset, then commits. The temporary file lives
// next to the destination so the final rename is atomic on the same
// filesystem.
type uploadStore struct {
	api *API
	ttl time.Duration
	mu  sync.Mutex
	m   map[string]*upload
}

type upload struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Offset    int64     `json:"offset"`
	Size      int64     `json:"size,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
	tmp       string
	dst       string
	busy      bool
}

func newUploadStore(a *API, ttl time.Duration) *uploadStore {
	s := &uploadStore{api: a, ttl: ttl, m: map[string]*upload{}}
	go s.janitor()
	return s
}

type createUploadRequest struct {
	Path      string `json:"path"`
	Size      int64  `json:"size,omitempty"`
	Overwrite bool   `json:"overwrite,omitempty"`
}

func (s *uploadStore) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createUploadRequest
	if !s.api.decode(w, r, &req) {
		return
	}
	dst, err := s.api.root.Resolve(req.Path)
	if err != nil {
		s.api.fsErr(w, err)
		return
	}
	if dst == s.api.root.Path() {
		writeErr(w, http.StatusBadRequest, "cannot write to root")
		return
	}
	if err := s.api.checkShare(dst); err != nil {
		s.api.fsErr(w, err)
		return
	}
	if !s.api.allowed(w, r, dst, access.Write) {
		return
	}
	if st, err := os.Stat(dst); err == nil {
		if st.IsDir() {
			writeErr(w, http.StatusConflict, "target is a directory")
			return
		}
		if !req.Overwrite {
			writeErr(w, http.StatusConflict, "already exists")
			return
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		s.api.fsErr(w, err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(dst), DirMode); err != nil {
		s.api.fsErr(w, err)
		return
	}
	f, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+partSuffix)
	if err != nil {
		s.api.fsErr(w, err)
		return
	}
	f.Close()
	idb := make([]byte, 16)
	if _, err := rand.Read(idb); err != nil {
		os.Remove(f.Name())
		s.api.fsErr(w, err)
		return
	}
	u := &upload{
		ID:        hex.EncodeToString(idb),
		Path:      s.api.root.Rel(dst),
		Size:      req.Size,
		ExpiresAt: time.Now().Add(s.ttl),
		tmp:       f.Name(),
		dst:       dst,
	}
	s.mu.Lock()
	s.m[u.ID] = u
	s.mu.Unlock()
	w.Header().Set("Upload-Offset", "0")
	writeJSON(w, http.StatusCreated, u)
}

func (s *uploadStore) get(id string) (*upload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.m[id]
	if !ok || time.Now().After(u.ExpiresAt) {
		return nil, false
	}
	return u, true
}

// acquire marks an upload busy so concurrent appends cannot interleave.
func (s *uploadStore) acquire(id string) (*upload, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.m[id]
	if !ok || time.Now().After(u.ExpiresAt) {
		return nil, http.StatusNotFound
	}
	if u.busy {
		return nil, http.StatusConflict
	}
	u.busy = true
	return u, 0
}

func (s *uploadStore) release(u *upload) {
	s.mu.Lock()
	u.busy = false
	s.mu.Unlock()
}

func (s *uploadStore) drop(u *upload) {
	s.mu.Lock()
	delete(s.m, u.ID)
	s.mu.Unlock()
	os.Remove(u.tmp)
}

func (s *uploadStore) handleStatus(w http.ResponseWriter, r *http.Request) {
	u, ok := s.get(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "upload not found or expired")
		return
	}
	s.mu.Lock()
	off := u.Offset
	s.mu.Unlock()
	w.Header().Set("Upload-Offset", strconv.FormatInt(off, 10))
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// handleAppend appends the request body at Upload-Offset, which must match the
// server-side offset. On mismatch the client should HEAD and resume.
func (s *uploadStore) handleAppend(w http.ResponseWriter, r *http.Request) {
	u, code := s.acquire(r.PathValue("id"))
	if code != 0 {
		if code == http.StatusConflict {
			writeErr(w, code, "upload busy")
		} else {
			writeErr(w, code, "upload not found or expired")
		}
		return
	}
	defer s.release(u)
	off, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Upload-Offset header required")
		return
	}
	if off != u.Offset {
		w.Header().Set("Upload-Offset", strconv.FormatInt(u.Offset, 10))
		writeErr(w, http.StatusConflict, "offset mismatch")
		return
	}
	f, err := os.OpenFile(u.tmp, os.O_WRONLY, 0)
	if err != nil {
		s.api.fsErr(w, err)
		return
	}
	if _, err := f.Seek(u.Offset, io.SeekStart); err != nil {
		f.Close()
		s.api.fsErr(w, err)
		return
	}
	n, copyErr := io.Copy(f, r.Body)
	closeErr := f.Close()
	s.mu.Lock()
	u.Offset += n
	u.ExpiresAt = time.Now().Add(s.ttl)
	newOff := u.Offset
	s.mu.Unlock()
	w.Header().Set("Upload-Offset", strconv.FormatInt(newOff, 10))
	if copyErr != nil || closeErr != nil {
		// Partial data is kept; the client resumes from Upload-Offset.
		writeErr(w, http.StatusBadRequest, "chunk interrupted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *uploadStore) handleCommit(w http.ResponseWriter, r *http.Request) {
	u, code := s.acquire(r.PathValue("id"))
	if code != 0 {
		writeErr(w, code, "upload not found, expired or busy")
		return
	}
	if u.Size > 0 && u.Offset != u.Size {
		s.release(u)
		w.Header().Set("Upload-Offset", strconv.FormatInt(u.Offset, 10))
		writeErr(w, http.StatusBadRequest, "upload incomplete")
		return
	}
	if mt := r.Header.Get("X-Mtime"); mt != "" {
		if t, err := time.Parse(time.RFC3339Nano, mt); err == nil {
			_ = os.Chtimes(u.tmp, t, t)
		}
	}
	_ = os.Chmod(u.tmp, FileMode)
	if err := os.Rename(u.tmp, u.dst); err != nil {
		s.release(u)
		s.api.fsErr(w, err)
		return
	}
	s.mu.Lock()
	delete(s.m, u.ID)
	s.mu.Unlock()
	info, err := os.Stat(u.dst)
	if err != nil {
		s.api.fsErr(w, err)
		return
	}
	s.api.indexed(u.dst, info)
	w.Header().Set("ETag", etag(info))
	writeJSON(w, http.StatusCreated, s.api.withID(s.api.entry(u.dst, info), info))
}

func (s *uploadStore) handleAbort(w http.ResponseWriter, r *http.Request) {
	u, ok := s.get(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "upload not found or expired")
		return
	}
	s.drop(u)
	w.WriteHeader(http.StatusNoContent)
}

func (s *uploadStore) janitor() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		var expired []*upload
		s.mu.Lock()
		for id, u := range s.m {
			if now.After(u.ExpiresAt) && !u.busy {
				delete(s.m, id)
				expired = append(expired, u)
			}
		}
		s.mu.Unlock()
		for _, u := range expired {
			os.Remove(u.tmp)
		}
	}
}

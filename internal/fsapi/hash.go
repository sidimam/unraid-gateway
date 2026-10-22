package fsapi

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/sidimam/unraid-gateway/internal/access"
)

// QuickHashWindow is how much of the head and of the tail of a file the quick fingerprint reads.
// Files up to twice this size are hashed whole. Unraid Drive 1.5.1 computes the same fingerprint
// locally to tell "same file, different dates" from a real conflict without transferring anything.
const QuickHashWindow = 1 << 20

type hashResponse struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	MTime string `json:"mtime"`
	ETag  string `json:"etag"`
	Algo  string `json:"algo"`
	Quick bool   `json:"quick"`
	Hash  string `json:"hash"`
}

// handleHash answers GET /hash?path=…&algo=md5|sha256&quick=1 with a content fingerprint
// (0.16). quick=1 (default): hash of the first and last QuickHashWindow bytes plus the decimal
// size — cheap on a 4 GB video, good enough to recognise the same file uploaded twice. quick=0:
// the whole file.
func (a *API) handleHash(w http.ResponseWriter, r *http.Request) {
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
	if info.IsDir() {
		writeErr(w, http.StatusBadRequest, "hash: path is a directory")
		return
	}
	algo := r.URL.Query().Get("algo")
	if algo == "" {
		algo = "md5"
	}
	quick := r.URL.Query().Get("quick") != "0"
	f, err := os.Open(abs)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	defer f.Close()
	sum, err := Fingerprint(f, info.Size(), algo, quick)
	if err != nil {
		if err == errBadAlgo {
			writeErr(w, http.StatusBadRequest, "algo must be md5 or sha256")
			return
		}
		a.fsErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, hashResponse{Path: a.root.Rel(abs), Size: info.Size(), MTime: info.ModTime().UTC().Format(timeLayout), ETag: etag(info), Algo: algo, Quick: quick, Hash: sum})
}

const timeLayout = "2006-01-02T15:04:05.000000000Z07:00"

var errBadAlgo = os.ErrInvalid

// Fingerprint hashes r (of the given size) with md5 or sha256. Quick mode reads only the head and
// tail windows and mixes in the size, so two files with the same fingerprint are the same file for
// sync purposes (not for security).
func Fingerprint(r io.ReadSeeker, size int64, algo string, quick bool) (string, error) {
	var h hash.Hash
	switch algo {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	default:
		return "", errBadAlgo
	}
	if !quick || size <= 2*QuickHashWindow {
		if _, err := io.Copy(h, r); err != nil {
			return "", err
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	}
	if _, err := io.CopyN(h, r, QuickHashWindow); err != nil {
		return "", err
	}
	if _, err := r.Seek(size-QuickHashWindow, io.SeekStart); err != nil {
		return "", err
	}
	if _, err := io.CopyN(h, r, QuickHashWindow); err != nil {
		return "", err
	}
	h.Write([]byte(strconv.FormatInt(size, 10)))
	return hex.EncodeToString(h.Sum(nil)), nil
}

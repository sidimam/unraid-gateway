package fsapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sidimam/unraid-gateway/internal/access"
)

// Media tickets let players that cannot send HTTP headers (libmpv/mpv, VLC, AVPlayer on some
// systems, a browser <video> tag) stream one file. The client asks for a ticket with its normal
// credentials; the ticket is an HMAC-signed, self-contained token bound to that file and to an
// expiry, redeemable at the public `GET /media/<ticket>` URL with full Range support. The signing
// secret is random per process, so tickets die with a restart. Nothing secret is ever in the URL:
// leaking a ticket grants read access to that single file until it expires.

const (
	ticketDefaultTTL = 8 * time.Hour
	ticketMaxTTL     = 24 * time.Hour
)

type ticketPayload struct {
	Path string `json:"p"`
	Exp  int64  `json:"e"`
	User string `json:"u,omitempty"`
}

func newTicketSecret() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("ticket secret: " + err.Error())
	}
	return b
}

func (a *API) signTicket(payload []byte) string {
	m := hmac.New(sha256.New, a.ticketSecret)
	m.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// IssueTicket returns a ticket for a client path.
func (a *API) IssueTicket(clientPath, user string, ttl time.Duration) (string, time.Time) {
	if ttl <= 0 {
		ttl = ticketDefaultTTL
	}
	if ttl > ticketMaxTTL {
		ttl = ticketMaxTTL
	}
	exp := time.Now().Add(ttl).Truncate(time.Second)
	payload, _ := json.Marshal(ticketPayload{Path: Clean(clientPath), Exp: exp.Unix(), User: user})
	return a.signTicket(payload), exp
}

var errBadTicket = errors.New("invalid or expired media ticket")

// verifyTicket returns the client path a ticket was issued for.
func (a *API) verifyTicket(tok string) (string, error) {
	parts := strings.SplitN(tok, ".", 2)
	if len(parts) != 2 {
		return "", errBadTicket
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errBadTicket
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errBadTicket
	}
	m := hmac.New(sha256.New, a.ticketSecret)
	m.Write(payload)
	if !hmac.Equal(sig, m.Sum(nil)) {
		return "", errBadTicket
	}
	var p ticketPayload
	if err := json.Unmarshal(payload, &p); err != nil || p.Path == "" {
		return "", errBadTicket
	}
	if time.Now().Unix() >= p.Exp {
		return "", errBadTicket
	}
	return p.Path, nil
}

// handleTicket: POST /api/v1/fs/ticket {"path": "/share/movie.mkv", "ttl": "2h"} → ticket + URL.
// Requires read access to the file with the caller's normal credentials.
func (a *API) handleTicket(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		TTL  string `json:"ttl"`
	}
	if !a.decode(w, r, &body) {
		return
	}
	abs, err := a.root.Resolve(body.Path)
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
		writeErr(w, http.StatusBadRequest, "is a directory")
		return
	}
	var ttl time.Duration
	if body.TTL != "" {
		if ttl, err = time.ParseDuration(body.TTL); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid ttl")
			return
		}
	}
	tok, exp := a.IssueTicket(body.Path, "", ttl)
	writeJSON(w, http.StatusOK, map[string]any{
		"ticket":    tok,
		"url":       "/media/" + tok,
		"path":      Clean(body.Path),
		"expiresAt": exp.UTC().Format(time.RFC3339),
	})
}

// HandleMedia: GET|HEAD /media/{ticket} — streams the file the ticket was issued for.
func (a *API) HandleMedia(w http.ResponseWriter, r *http.Request) {
	clientPath, err := a.verifyTicket(r.PathValue("ticket"))
	if err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	abs, err := a.root.Resolve(clientPath)
	if err != nil {
		a.fsErr(w, err)
		return
	}
	a.serveFile(w, r, abs, false, "stream")
}

// RegisterPublic mounts the endpoints that carry their own authentication.
func (a *API) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("GET /media/{ticket}", a.HandleMedia)
	mux.HandleFunc("HEAD /media/{ticket}", a.HandleMedia)
}

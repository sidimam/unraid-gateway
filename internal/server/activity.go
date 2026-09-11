package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/sidimam/unraid-gateway/internal/activity"
	"github.com/sidimam/unraid-gateway/internal/auth"
	"github.com/sidimam/unraid-gateway/internal/fsapi"
)

// shortHash identifies a session or key without exposing it.
func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:4])
}

// whoFor describes the caller: which session/device, which Unraid user, from where.
// The apps send "X-Unraid-Drive-Client: Unraid Drive 1.3 (28) · iPhone · iOS 26 · File Provider";
// other clients are shown by their User-Agent.
func whoFor(r *http.Request, p auth.Principal, ip string) activity.Who {
	agent := strings.TrimSpace(r.Header.Get("X-Unraid-Drive-Client"))
	if agent == "" {
		agent = r.UserAgent()
	}
	id := "t:" + ip // header-less media tickets
	if tok := bearer(r); tok != "" {
		id = "s:" + shortHash(tok)
	} else if key := strings.TrimSpace(r.Header.Get("x-api-key")); key != "" {
		id = "k:" + shortHash(key) + "@" + ip
	}
	return activity.Who{ClientID: id, User: p.User, Key: p.Identity.Name, IP: ip, Agent: agent}
}

// activityHook adapts the tracker to the file API.
type activityHook struct {
	t          *activity.Tracker
	trustProxy bool
}

func (h *activityHook) Begin(r *http.Request, kind, path string, size int64) fsapi.TransferHandle {
	return h.t.Begin(whoFor(r, principal(r), auth.ClientIP(r, h.trustProxy)), kind, path, size, r.Header.Get("Range"))
}

// handleActivity: GET /api/v1/activity. API-key-only sessions and ADMIN keys see everyone;
// a session opened with an Unraid user sees only that user's devices and transfers.
func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	only := ""
	if p.User != "" && !hasRole(p.Identity.Roles, "ADMIN") {
		only = p.User
	}
	snap := s.activity.Snapshot(only)
	writeJSON(w, http.StatusOK, map[string]any{
		"now":       snap.Now,
		"scope":     map[bool]string{true: "user", false: "all"}[only != ""],
		"clients":   snap.Clients,
		"active":    snap.Active,
		"recent":    snap.Recent,
		"idleAfter": s.activity.IdleAfter.String(),
	})
}

func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, want) {
			return true
		}
	}
	return false
}

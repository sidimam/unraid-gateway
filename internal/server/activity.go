package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

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
	if r.URL.Query().Get("format") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(renderActivity(snap)))
		return
	}
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

// loopbackOr serves the activity to the container's own console (loopback, no proxy headers
// trusted here) and hands everything else to the authenticated handler.
func (s *Server) loopbackOr(authed http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
				s.handleActivity(w, r)
				return
			}
		}
		authed.ServeHTTP(w, r)
	})
}

// renderActivity is the plain-text version used by `gw activity`.
func renderActivity(snap activity.Snapshot) string {
	var b strings.Builder
	ago := func(t time.Time) string {
		d := snap.Now.Sub(t).Round(time.Second)
		switch {
		case d < 5*time.Second:
			return "now"
		case d < time.Minute:
			return fmt.Sprintf("%ds ago", int(d.Seconds()))
		case d < time.Hour:
			return fmt.Sprintf("%dm ago", int(d.Minutes()))
		default:
			return fmt.Sprintf("%dh%02dm ago", int(d.Hours()), int(d.Minutes())%60)
		}
	}
	size := func(n int64) string {
		f := float64(n)
		for _, u := range []string{"B", "KB", "MB", "GB", "TB"} {
			if f < 1024 || u == "TB" {
				if u == "B" {
					return fmt.Sprintf("%d %s", n, u)
				}
				return fmt.Sprintf("%.1f %s", f, u)
			}
			f /= 1024
		}
		return ""
	}
	who := func(w activity.Who) string {
		if w.User != "" {
			return w.User
		}
		if w.Key != "" {
			return "key:" + w.Key
		}
		return "-"
	}
	fmt.Fprintf(&b, "Connected devices (%d)\n", len(snap.Clients))
	if len(snap.Clients) == 0 {
		b.WriteString("  nobody right now (sessions expire from this list 30 min after their last request)\n")
	}
	for _, c := range snap.Clients {
		mark := " "
		if c.Active > 0 {
			mark = "▶"
		}
		fmt.Fprintf(&b, "%s %-12s %-16s since %-10s seen %-8s %5d req  %s\n", mark, who(c.Who), c.IP, ago(c.FirstSeen), ago(c.LastSeen), c.Requests, c.Agent)
		if c.LastPath != "" {
			fmt.Fprintf(&b, "  %-12s last file: %s (%s)\n", "", c.LastPath, ago(c.LastPathAt))
		}
	}
	fmt.Fprintf(&b, "\nIn progress (%d)\n", len(snap.Active))
	if len(snap.Active) == 0 {
		b.WriteString("  nothing playing or transferring\n")
	}
	for _, t := range snap.Active {
		pct := ""
		if t.Size > 0 {
			pct = fmt.Sprintf(" %d%%", t.Bytes*100/t.Size)
		}
		fmt.Fprintf(&b, "  %-8s %-12s %s  %s/%s%s  %s  %s\n", t.Kind, who(t.Who), t.Path, size(t.Bytes), size(t.Size), pct, snap.Now.Sub(t.Started).Round(time.Second), t.Agent)
	}
	fmt.Fprintf(&b, "\nRecent (%d)\n", len(snap.Recent))
	for i, t := range snap.Recent {
		if i >= 15 {
			break
		}
		ended := ""
		if t.Ended != nil {
			ended = ago(*t.Ended)
		}
		fmt.Fprintf(&b, "  %-8s %-12s %s  %s  %s\n", t.Kind, who(t.Who), t.Path, size(t.Bytes), ended)
	}
	return b.String()
}

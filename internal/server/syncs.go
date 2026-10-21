package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sidimam/unraid-gateway/internal/auth"
	"github.com/sidimam/unraid-gateway/internal/devices"
	"github.com/sidimam/unraid-gateway/internal/metrics"
	"github.com/sidimam/unraid-gateway/internal/syncs"
)

// ---- syncs (0.14): the folder syncs and backups the apps run ---------------------------------

// canSeeDevice: admins see everything; a user its own devices; a device itself.
func (s *Server) canSeeDevice(p auth.Principal, d devices.Device) bool {
	return isAdmin(p) || (p.User != "" && d.User == p.User) || (p.DeviceID != "" && d.ID == p.DeviceID)
}

func (s *Server) canSeeSync(p auth.Principal, x syncs.Sync) bool {
	if isAdmin(p) || (p.DeviceID != "" && x.DeviceID == p.DeviceID) {
		return true
	}
	if p.User != "" && x.User == p.User {
		return true
	}
	if d, ok := s.devices.Get(x.DeviceID); ok {
		return s.canSeeDevice(p, d)
	}
	return false
}

// GET /api/v1/syncs[?device=ID|user=NAME|all=1] — by default the caller's device pairs (an app),
// or everything the caller may see when it has no device (web UI / console).
func (s *Server) handleSyncsList(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	q := r.URL.Query()
	out := []syncs.Sync{}
	for _, x := range s.syncs.List() {
		if !s.canSeeSync(p, x) {
			continue
		}
		if d := q.Get("device"); d != "" && x.DeviceID != d {
			continue
		}
		if u := q.Get("user"); u != "" && s.userOfSync(x) != u {
			continue
		}
		if q.Get("device") == "" && q.Get("user") == "" && q.Get("all") == "" && p.DeviceID != "" && x.DeviceID != p.DeviceID {
			continue
		}
		out = append(out, x)
	}
	if r.URL.Query().Get("format") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if len(out) == 0 {
			fmt.Fprintln(w, "  (no syncs defined by any app yet)")
			return
		}
		for _, x := range out {
			dev := x.DeviceID
			if d, ok := s.devices.Get(x.DeviceID); ok {
				dev = d.Name
			}
			state := "enabled"
			if !x.Enabled {
				state = "PAUSED"
				if x.DisabledBy != "" {
					state += " by " + x.DisabledBy
				}
			}
			last := "never run"
			if x.LastRun != nil {
				last = fmt.Sprintf("last %s %s: %d files, %s, %d conflicts, %d errors", x.LastRun.At.Format("2006-01-02 15:04"), x.LastRun.Status, x.LastRun.Files, humanBytes(x.LastRun.Bytes), x.LastRun.Conflicts, x.LastRun.Errors)
			}
			fmt.Fprintf(w, "  %s  %-9s %-5s %s ↔ %s  [%s]  user=%s  device=%s\n    %s · %d runs · %d files · %s total\n", x.ID, x.Kind, x.Mode, x.LocalLabel, x.RemotePath, state, x.User, dev, last, x.Runs, x.TotalFiles, humanBytes(x.TotalBytes))
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"syncs": out, "thisDevice": p.DeviceID})
}

// PUT /api/v1/syncs/{id} — an app creates or updates one of its pairs.
func (s *Server) handleSyncPut(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	id := r.PathValue("id")
	var in syncs.Sync
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	in.ID = id
	if in.DeviceID == "" {
		in.DeviceID = p.DeviceID
	}
	if in.DeviceID == "" {
		writeErr(w, http.StatusBadRequest, "a sync belongs to a registered device: sign in with a device id")
		return
	}
	if !isAdmin(p) && in.DeviceID != p.DeviceID {
		writeErr(w, http.StatusForbidden, "you may only define syncs for your own device")
		return
	}
	if in.User == "" {
		in.User = p.User
	}
	switch in.Mode {
	case "", syncs.ModeUp, syncs.ModeDown, syncs.ModeBoth:
	default:
		writeErr(w, http.StatusBadRequest, "mode must be up, down or both")
		return
	}
	switch in.Kind {
	case "", syncs.KindFolder, syncs.KindPhotos, syncs.KindDownloads:
	default:
		writeErr(w, http.StatusBadRequest, "kind must be folder, photos or downloads")
		return
	}
	if !strings.HasPrefix(in.RemotePath, "/") {
		writeErr(w, http.StatusBadRequest, "remotePath must start with /")
		return
	}
	x, isNew, err := s.syncs.Upsert(in)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if isNew {
		s.log.Info("sync created", "sync", id, "device", x.DeviceID, "kind", x.Kind, "local", x.LocalLabel, "remote", x.RemotePath, "mode", x.Mode)
		s.metrics.Log(x.DeviceID, "sync", "info", fmt.Sprintf("sync added: %s ↔ %s (%s, %s)", x.LocalLabel, x.RemotePath, x.Kind, x.Mode))
	}
	writeJSON(w, http.StatusOK, x)
}

// PATCH /api/v1/syncs/{id} {"enabled": bool} — web UI / console pause or resume; the owning app
// may do it too.
func (s *Server) handleSyncPatch(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	id := r.PathValue("id")
	x, ok := s.syncs.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "unknown sync")
		return
	}
	if !s.canSeeSync(p, x) {
		writeErr(w, http.StatusForbidden, "not your sync")
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.Enabled == nil {
		writeErr(w, http.StatusBadRequest, "body must be {\"enabled\": true|false}")
		return
	}
	by := who(p)
	if p.DeviceID == x.DeviceID && p.DeviceID != "" {
		by = "" // the app itself: not a web UI pause
	}
	y, err := s.syncs.SetEnabled(id, *req.Enabled, by)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	state := "resumed"
	if !*req.Enabled {
		state = "paused"
	}
	s.log.Info("sync "+state, "sync", id, "device", y.DeviceID, "by", who(p))
	s.metrics.Log(y.DeviceID, "sync", "info", fmt.Sprintf("sync %s by %s: %s ↔ %s", state, who(p), y.LocalLabel, y.RemotePath))
	writeJSON(w, http.StatusOK, y)
}

// DELETE /api/v1/syncs/{id}
func (s *Server) handleSyncDelete(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	id := r.PathValue("id")
	x, ok := s.syncs.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "unknown sync")
		return
	}
	if !s.canSeeSync(p, x) {
		writeErr(w, http.StatusForbidden, "not your sync")
		return
	}
	if _, err := s.syncs.Remove(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.log.Info("sync removed", "sync", id, "device", x.DeviceID, "by", who(p))
	s.metrics.Log(x.DeviceID, "sync", "info", fmt.Sprintf("sync removed by %s: %s ↔ %s", who(p), x.LocalLabel, x.RemotePath))
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/syncs/{id}/runs {run: {...}, files: [...]} — the app reports one execution.
// 409 sync_disabled when the pair was paused on the gateway: the app must stop running it.
func (s *Server) handleSyncReport(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	id := r.PathValue("id")
	x, ok := s.syncs.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "unknown sync")
		return
	}
	if !isAdmin(p) && x.DeviceID != p.DeviceID {
		writeErr(w, http.StatusForbidden, "only the owning device reports runs")
		return
	}
	var req struct {
		Run   syncs.Run    `json:"run"`
		Files []syncs.File `json:"files"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.Run.Status == "" {
		req.Run.Status = "ok"
	}
	y, err := s.syncs.Report(id, req.Run, req.Files)
	if errors.Is(err, syncs.ErrDisabled) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "this sync was paused on the gateway", "code": "sync_disabled", "disabledBy": y.DisabledBy})
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	level := "info"
	if req.Run.Status == "failed" {
		level = "error"
	} else if req.Run.Status == "partial" || req.Run.Errors > 0 || req.Run.Conflicts > 0 {
		level = "warning"
	}
	detail := fmt.Sprintf("sync %s: %s ↔ %s — %d files, %s, %d conflicts, %d errors", req.Run.Status, y.LocalLabel, y.RemotePath, req.Run.Files, humanBytes(req.Run.Bytes), req.Run.Conflicts, req.Run.Errors)
	if req.Run.Error != "" {
		detail += ": " + req.Run.Error
	}
	s.metrics.Log(y.DeviceID, "sync", level, detail)
	writeJSON(w, http.StatusOK, y)
}

func humanBytes(n int64) string {
	const u = "KMGT"
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	f, i := float64(n), -1
	for f >= 1024 && i < len(u)-1 {
		f /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %cB", f, u[i])
}

// ---- users (0.14): everything about one Unraid user — devices, syncs, activity ---------------

type userSummary struct {
	User     string         `json:"user"`
	Devices  int            `json:"devices"`
	Syncs    int            `json:"syncs"`
	Paused   int            `json:"pausedSyncs"`
	LastSeen time.Time      `json:"lastSeen"`
	Today    metrics.Totals `json:"today"`
	Warnings int            `json:"warnings"` // sync runs with conflicts/errors + stale devices
}

// userDevices returns the registered devices of one user (API-key-only sessions are grouped
// under their key name, shown as "key <name>").
func (s *Server) userDevices(name string) []devices.Device {
	out := []devices.Device{}
	for _, d := range s.devices.List() {
		if userOf(d) == name {
			out = append(out, d)
		}
	}
	return out
}

func userOf(d devices.Device) string {
	if d.User != "" {
		return d.User
	}
	if d.Key != "" {
		return "key " + d.Key
	}
	return "—"
}

// userOfSync names the user a sync belongs to, the same way the users view groups devices
// (API-key-only devices under "key <name>").
func (s *Server) userOfSync(x syncs.Sync) string {
	if d, ok := s.devices.Get(x.DeviceID); ok {
		return userOf(d)
	}
	if x.User != "" {
		return x.User
	}
	return "—"
}

func (s *Server) canSeeUser(p auth.Principal, name string) bool {
	if isAdmin(p) {
		return true
	}
	return p.User != "" && p.User == name
}

// GET /api/v1/users — one row per user with devices, syncs, last seen and today's totals.
func (s *Server) handleUsersList(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	byUser := map[string]*userSummary{}
	now := time.Now()
	y, m, dd := now.Date()
	today := time.Date(y, m, dd, 0, 0, 0, 0, now.Location())
	for _, d := range s.devices.List() {
		u := userOf(d)
		if !s.canSeeUser(p, u) {
			continue
		}
		row := byUser[u]
		if row == nil {
			row = &userSummary{User: u}
			byUser[u] = row
		}
		row.Devices++
		if d.LastSeen.After(row.LastSeen) {
			row.LastSeen = d.LastSeen
		}
		if !d.LastSeen.IsZero() && now.Sub(d.LastSeen) > 30*24*time.Hour && d.SupersededBy == "" {
			row.Warnings++
		}
		t := s.metrics.Sum(d.ID, today)
		row.Today.Requests += t.Requests
		row.Today.Up += t.Up
		row.Today.Down += t.Down
		row.Today.Streams += t.Streams
		row.Today.Uploads += t.Uploads
		row.Today.Downloads += t.Downloads
	}
	for _, x := range s.syncs.List() {
		d, ok := s.devices.Get(x.DeviceID)
		u := x.User
		if ok {
			u = userOf(d)
		}
		row := byUser[u]
		if row == nil {
			continue
		}
		row.Syncs++
		if !x.Enabled {
			row.Paused++
		}
		if x.LastRun != nil && (x.LastRun.Status == "failed" || x.LastRun.Errors > 0 || x.LastRun.Conflicts > 0) {
			row.Warnings++
		}
	}
	out := make([]userSummary, 0, len(byUser))
	for _, v := range byUser {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].User < out[j].User })
	if r.URL.Query().Get("format") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if len(out) == 0 {
			fmt.Fprintln(w, "  (no registered device yet)")
			return
		}
		for _, u := range out {
			fmt.Fprintf(w, "  %-20s %d devices  %d syncs (%d paused)  last seen %s  today ↓%s ↑%s  warnings=%d\n", u.User, u.Devices, u.Syncs, u.Paused, u.LastSeen.Format("2006-01-02 15:04"), humanBytes(u.Today.Down), humanBytes(u.Today.Up), u.Warnings)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

// GET /api/v1/users/{name} — devices, syncs (with their recent files) and totals of one user.
func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	name := r.PathValue("name")
	if !s.canSeeUser(p, name) {
		writeErr(w, http.StatusForbidden, "you may only see yourself")
		return
	}
	devs := s.userDevices(name)
	ids := map[string]bool{}
	for _, d := range devs {
		ids[d.ID] = true
	}
	sy := []syncs.Sync{}
	for _, x := range s.syncs.List() {
		if ids[x.DeviceID] || (x.User == name && !ids[x.DeviceID]) {
			sy = append(sy, x)
		}
	}
	now := time.Now()
	y, m, dd := now.Date()
	today := time.Date(y, m, dd, 0, 0, 0, 0, now.Location())
	if r.URL.Query().Get("format") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "  %s — %d devices, %d syncs; today ↓%s ↑%s, 7 days ↓%s ↑%s, 30 days ↓%s ↑%s\n", name, len(devs), len(sy),
			humanBytes(s.sumMany(devs, today).Down), humanBytes(s.sumMany(devs, today).Up),
			humanBytes(s.sumMany(devs, now.Add(-7*24*time.Hour)).Down), humanBytes(s.sumMany(devs, now.Add(-7*24*time.Hour)).Up),
			humanBytes(s.sumMany(devs, now.Add(-30*24*time.Hour)).Down), humanBytes(s.sumMany(devs, now.Add(-30*24*time.Hour)).Up))
		fmt.Fprintln(w, "  devices:")
		for _, d := range devs {
			old := ""
			if d.SupersededBy != "" {
				old = "  OLD"
			}
			fmt.Fprintf(w, "    %s  %s  last seen %s  logins=%d%s\n", d.ID, d.Name, d.LastSeen.Format("2006-01-02 15:04"), d.Logins, old)
		}
		fmt.Fprintln(w, "  syncs:")
		if len(sy) == 0 {
			fmt.Fprintln(w, "    (none)")
		}
		for _, x := range sy {
			state := "enabled"
			if !x.Enabled {
				state = "PAUSED"
			}
			last := "never run"
			if x.LastRun != nil {
				last = fmt.Sprintf("last %s %s (%d files, %d errors)", x.LastRun.At.Format("2006-01-02 15:04"), x.LastRun.Status, x.LastRun.Files, x.LastRun.Errors)
			}
			fmt.Fprintf(w, "    %s  %-9s %-5s %s ↔ %s  [%s]  %s\n", x.ID, x.Kind, x.Mode, x.LocalLabel, x.RemotePath, state, last)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":    name,
		"devices": devs,
		"syncs":   sy,
		"today":   s.sumMany(devs, today),
		"week":    s.sumMany(devs, now.Add(-7*24*time.Hour)),
		"month":   s.sumMany(devs, now.Add(-30*24*time.Hour)),
	})
}

func (s *Server) sumMany(devs []devices.Device, since time.Time) metrics.Totals {
	var t metrics.Totals
	for _, d := range devs {
		x := s.metrics.Sum(d.ID, since)
		t.Requests += x.Requests
		t.Up += x.Up
		t.Down += x.Down
		t.Streams += x.Streams
		t.Uploads += x.Uploads
		t.Downloads += x.Downloads
	}
	return t
}

// GET /api/v1/users/{name}/metrics?range=24h|7d|30d|90d — the user's devices summed over time.
func (s *Server) handleUserMetrics(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	name := r.PathValue("name")
	if !s.canSeeUser(p, name) {
		writeErr(w, http.StatusForbidden, "you may only see yourself")
		return
	}
	since, step := rangeOf(r.URL.Query().Get("range"))
	var series []metrics.Point
	devs := s.userDevices(name)
	for _, d := range devs {
		pts := s.metrics.Series(d.ID, since, step)
		if series == nil {
			series = pts
			continue
		}
		for i := range pts {
			if i >= len(series) {
				series = append(series, pts[i])
				continue
			}
			series[i].Requests += pts[i].Requests
			series[i].Up += pts[i].Up
			series[i].Down += pts[i].Down
			series[i].Streams += pts[i].Streams
			series[i].Uploads += pts[i].Uploads
			series[i].Downloads += pts[i].Downloads
		}
	}
	if series == nil {
		series = []metrics.Point{}
	}
	now := time.Now()
	y, m, dd := now.Date()
	today := time.Date(y, m, dd, 0, 0, 0, 0, now.Location())
	writeJSON(w, http.StatusOK, map[string]any{
		"series": series, "step": step.String(),
		"today": s.sumMany(devs, today), "week": s.sumMany(devs, now.Add(-7*24*time.Hour)), "month": s.sumMany(devs, now.Add(-30*24*time.Hour)),
	})
}

func rangeOf(q string) (time.Time, time.Duration) {
	now := time.Now()
	switch q {
	case "24h":
		return now.Add(-24 * time.Hour), time.Hour
	case "7d":
		return now.Add(-7 * 24 * time.Hour), 24 * time.Hour
	case "90d":
		return now.Add(-90 * 24 * time.Hour), 24 * time.Hour
	default:
		return now.Add(-30 * 24 * time.Hour), 24 * time.Hour
	}
}

type userEvent struct {
	metrics.Event
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
}

// GET /api/v1/users/{name}/log?limit=N — the events of every device of the user, merged, newest first.
func (s *Server) handleUserLog(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	name := r.PathValue("name")
	if !s.canSeeUser(p, name) {
		writeErr(w, http.StatusForbidden, "you may only see yourself")
		return
	}
	limit := 200
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 1000 {
		limit = v
	}
	out := []userEvent{}
	for _, d := range s.userDevices(name) {
		for _, e := range s.metrics.Events(d.ID, limit) {
			out = append(out, userEvent{Event: e, DeviceID: d.ID, DeviceName: d.Name})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	if len(out) > limit {
		out = out[:limit]
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": out})
}

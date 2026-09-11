package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sidimam/unraid-gateway/internal/auth"
	"github.com/sidimam/unraid-gateway/internal/notify"
)

// isAdmin: API-key-only sessions and ADMIN keys manage the gateway; user sessions only see themselves.
func isAdmin(p auth.Principal) bool { return p.User == "" || hasRole(p.Identity.Roles, "ADMIN") }

// GET /api/v1/devices
func (s *Server) handleDevicesList(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	all := s.devices.List()
	out := all[:0:0]
	for _, d := range all {
		if isAdmin(p) || d.User == p.User || d.ID == p.DeviceID {
			out = append(out, d)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": out, "registration": s.cfg.DeviceRegistration, "thisDevice": p.DeviceID})
}

// DELETE /api/v1/devices/{id} — revokes one device: its sessions die, the app must register again.
func (s *Server) handleDeviceRemove(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	id := r.PathValue("id")
	d, ok := s.devices.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "unknown device")
		return
	}
	if !isAdmin(p) && d.User != p.User && d.ID != p.DeviceID {
		writeErr(w, http.StatusForbidden, "you may only remove your own devices")
		return
	}
	if _, err := s.devices.Remove(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	n := s.sessions.LogoutDevice(id)
	s.log.Info("device removed", "device", id, "name", d.Name, "sessions", n, "by", who(p), "ip", auth.ClientIP(r, s.cfg.TrustProxy))
	go s.notifier.Send(context.Background(), p.APIKey, "Device removed from unraid-gateway", fmt.Sprintf("%s\nUser: %s\nRemoved by: %s", d.Name, d.User, who(p)), notify.Info)
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/devices — removes every device (admins only).
func (s *Server) handleDevicesRemoveAll(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	if !isAdmin(p) {
		writeErr(w, http.StatusForbidden, "only an API-key or ADMIN session may remove all devices")
		return
	}
	n, err := s.devices.RemoveAll()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	k := s.sessions.LogoutDevice("")
	s.log.Info("all devices removed", "devices", n, "sessions", k, "by", who(p))
	go s.notifier.Send(context.Background(), p.APIKey, "All devices removed from unraid-gateway", fmt.Sprintf("%d devices, %d sessions closed. Every app will ask to sign in again.", n, k), notify.Info)
	writeJSON(w, http.StatusOK, map[string]any{"removed": n, "sessionsClosed": k})
}

func who(p auth.Principal) string {
	if p.User != "" {
		return p.User
	}
	if p.Identity.Name != "" {
		return "key " + p.Identity.Name
	}
	return "console"
}

// GET /api/v1/notify/channels
func (s *Server) handleNotifyChannels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"channels": s.notifier.Channels()})
}

// POST /api/v1/notify/test — sends a test message through every configured channel.
func (s *Server) handleNotifyTest(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	errs := s.notifier.Send(r.Context(), p.APIKey, "unraid-gateway test notification", fmt.Sprintf("Sent by %s at %s. If you read this, notifications work.", who(p), time.Now().Format("2006-01-02 15:04:05")), notify.Info)
	out := map[string]any{"channels": s.notifier.Channels(), "errors": []string{}}
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		out["errors"] = msgs
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- Unraid API keys (the key must have the ADMIN role, Unraid enforces it) -------------------

const keysQuery = `{ apiKeys { id name description roles createdAt } }`

// GET /api/v1/keys
func (s *Server) handleKeysList(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	data, err := s.gql.Query(r.Context(), p.APIKey, keysQuery, nil)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "Unraid refused to list the API keys (an ADMIN key is required): "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// POST /api/v1/keys {"name": "...", "roles": ["VIEWER"]} → the new key, shown once.
func (s *Server) handleKeyCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeErr(w, http.StatusBadRequest, "expected {\"name\": \"...\", \"roles\": [\"VIEWER\"]}")
		return
	}
	if len(req.Roles) == 0 {
		req.Roles = []string{"VIEWER"}
	}
	data, err := s.createKey(r.Context(), principal(r).APIKey, req.Name, req.Roles)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "Unraid refused to create the key (an ADMIN key is required): "+err.Error())
		return
	}
	s.log.Info("api key created via gateway", "name", req.Name, "roles", strings.Join(req.Roles, ","), "by", who(principal(r)))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (s *Server) createKey(ctx context.Context, adminKey, name string, roles []string) (json.RawMessage, error) {
	const q = `mutation($input: CreateApiKeyInput!) { apiKey { create(input: $input) { id key name roles createdAt } } }`
	return s.gql.Query(ctx, adminKey, q, map[string]any{"input": map[string]any{"name": name, "roles": roles, "description": "created by unraid-gateway"}})
}

// DELETE /api/v1/keys/{id}
func (s *Server) handleKeyDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	const q = `mutation($input: DeleteApiKeyInput!) { apiKey { delete(input: $input) } }`
	if _, err := s.gql.Query(r.Context(), principal(r).APIKey, q, map[string]any{"input": map[string]any{"ids": []string{id}}}); err != nil {
		writeErr(w, http.StatusBadGateway, "Unraid refused to delete the key: "+err.Error())
		return
	}
	s.log.Info("api key deleted via gateway", "id", id, "by", who(principal(r)))
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/keys/rotate {"name": "...", "remember": true, "deleteOldId": "..."} — creates a
// replacement key (VIEWER, or the roles given), optionally stores it as the web UI key and deletes
// the old one. The new key is returned once: copy it into the apps.
func (s *Server) handleKeyRotate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Roles       []string `json:"roles"`
		Remember    bool     `json:"remember"`
		DeleteOldID string   `json:"deleteOldId"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	_ = json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.Name) == "" {
		req.Name = "Unraid Drive " + time.Now().Format("2006 01 02 1504")
	}
	if len(req.Roles) == 0 {
		req.Roles = []string{"VIEWER"}
	}
	p := principal(r)
	data, err := s.createKey(r.Context(), p.APIKey, req.Name, req.Roles)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "Unraid refused to create the key (an ADMIN key is required): "+err.Error())
		return
	}
	var created struct {
		APIKey struct {
			Create struct {
				ID  string `json:"id"`
				Key string `json:"key"`
			} `json:"create"`
		} `json:"apiKey"`
	}
	_ = json.Unmarshal(data, &created)
	newKey := created.APIKey.Create.Key
	out := map[string]any{"id": created.APIKey.Create.ID, "key": newKey, "name": req.Name, "roles": req.Roles}
	if req.Remember && newKey != "" && s.cfg.WebUIKeyFile != "" {
		if err := writeSecret(s.cfg.WebUIKeyFile, newKey); err != nil {
			out["rememberError"] = err.Error()
		} else {
			out["remembered"] = true
		}
	}
	if req.DeleteOldID != "" {
		const q = `mutation($input: DeleteApiKeyInput!) { apiKey { delete(input: $input) } }`
		if _, err := s.gql.Query(r.Context(), p.APIKey, q, map[string]any{"input": map[string]any{"ids": []string{req.DeleteOldID}}}); err != nil {
			out["deleteError"] = err.Error()
		} else {
			out["deletedOld"] = true
		}
	}
	s.log.Info("api key rotated via gateway", "name", req.Name, "deletedOld", req.DeleteOldID != "", "by", who(p))
	writeJSON(w, http.StatusOK, out)
}

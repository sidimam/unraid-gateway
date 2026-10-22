package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sidimam/unraid-gateway/internal/notify"
)

// Notification channels can be configured from the web UI (0.18). The container variables are the
// defaults; what the admin saves here is kept in /config/notify.json (0600) and wins over them.

const masked = "••••••••"

type notifySettings struct {
	UnraidEnabled bool   `json:"unraidEnabled"`
	UnraidAPIKey  string `json:"unraidApiKey,omitempty"`
	SMTPHost      string `json:"smtpHost"`
	SMTPPort      string `json:"smtpPort"`
	SMTPUser      string `json:"smtpUser"`
	SMTPPassword  string `json:"smtpPassword,omitempty"`
	SMTPFrom      string `json:"smtpFrom"`
	SMTPTo        string `json:"smtpTo"`
	SMTPTLS       string `json:"smtpTls"`
	TelegramToken string `json:"telegramToken,omitempty"`
	TelegramChat  string `json:"telegramChatId"`
}

func (x notifySettings) toConfig() notify.Config {
	return notify.Config{UnraidEnabled: x.UnraidEnabled, UnraidAPIKey: x.UnraidAPIKey,
		SMTPHost: x.SMTPHost, SMTPPort: x.SMTPPort, SMTPUser: x.SMTPUser, SMTPPassword: x.SMTPPassword, SMTPFrom: x.SMTPFrom, SMTPTo: x.SMTPTo, SMTPTLS: x.SMTPTLS,
		TelegramToken: x.TelegramToken, TelegramChatID: x.TelegramChat}
}

func settingsOf(c notify.Config) notifySettings {
	return notifySettings{UnraidEnabled: c.UnraidEnabled, UnraidAPIKey: c.UnraidAPIKey,
		SMTPHost: c.SMTPHost, SMTPPort: c.SMTPPort, SMTPUser: c.SMTPUser, SMTPPassword: c.SMTPPassword, SMTPFrom: c.SMTPFrom, SMTPTo: c.SMTPTo, SMTPTLS: c.SMTPTLS,
		TelegramToken: c.TelegramToken, TelegramChat: c.TelegramChatID}
}

func (s *Server) notifyFile() string {
	if s.cfg.DevicesFile == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(s.cfg.DevicesFile), "notify.json")
}

// loadNotifySettings applies /config/notify.json over the environment defaults at start-up.
func (s *Server) loadNotifySettings() {
	path := s.notifyFile()
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var x notifySettings
	if err := json.Unmarshal(b, &x); err != nil {
		s.log.Warn("notify.json unreadable, keeping the container variables", "err", err)
		return
	}
	s.notifier.SetConfig(x.toConfig())
	s.log.Info("notification channels loaded from notify.json", "channels", strings.Join(s.notifier.Channels(), ","))
}

// GET /api/v1/notify/settings (admin) — secrets are masked, `has*` says whether one is set.
func (s *Server) handleNotifySettingsGet(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(principal(r)) {
		writeErr(w, http.StatusForbidden, "admin only")
		return
	}
	c := s.notifier.Config()
	x := settingsOf(c)
	out := map[string]any{"unraidEnabled": x.UnraidEnabled, "hasUnraidApiKey": x.UnraidAPIKey != "", "unraidSpool": s.notifier.UnraidSpoolAvailable(),
		"smtpHost": x.SMTPHost, "smtpPort": x.SMTPPort, "smtpUser": x.SMTPUser, "hasSmtpPassword": x.SMTPPassword != "", "smtpFrom": x.SMTPFrom, "smtpTo": x.SMTPTo, "smtpTls": x.SMTPTLS,
		"hasTelegramToken": x.TelegramToken != "", "telegramChatId": x.TelegramChat,
		"source": "env", "channels": s.notifier.Channels()}
	if p := s.notifyFile(); p != "" {
		if _, err := os.Stat(p); err == nil {
			out["source"] = "file"
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// PUT /api/v1/notify/settings (admin) — a secret sent empty or masked keeps its current value.
func (s *Server) handleNotifySettingsPut(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(principal(r)) {
		writeErr(w, http.StatusForbidden, "admin only")
		return
	}
	var in notifySettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	cur := settingsOf(s.notifier.Config())
	keep := func(newv, old string) string {
		if newv == "" || newv == masked {
			return old
		}
		return strings.TrimSpace(newv)
	}
	in.UnraidAPIKey = keep(in.UnraidAPIKey, cur.UnraidAPIKey)
	in.SMTPPassword = keep(in.SMTPPassword, cur.SMTPPassword)
	in.TelegramToken = keep(in.TelegramToken, cur.TelegramToken)
	switch in.SMTPTLS {
	case "", "starttls", "tls", "none":
	default:
		writeErr(w, http.StatusBadRequest, "smtpTls must be starttls, tls or none")
		return
	}
	for _, f := range []*string{&in.SMTPHost, &in.SMTPPort, &in.SMTPUser, &in.SMTPFrom, &in.SMTPTo, &in.TelegramChat} {
		*f = strings.TrimSpace(*f)
	}
	path := s.notifyFile()
	if path == "" {
		writeErr(w, http.StatusConflict, "no /config: set DEVICES_FILE (the settings would not survive a restart)")
		return
	}
	b, _ := json.MarshalIndent(in, "", "  ")
	if err := writeSecret(path, string(b)); err != nil {
		writeErr(w, http.StatusInternalServerError, "cannot write notify.json: "+err.Error())
		return
	}
	s.notifier.SetConfig(in.toConfig())
	s.metrics.Log(principal(r).DeviceID, "admin", "info", "notification channels changed: "+strings.Join(s.notifier.Channels(), ", "))
	s.handleNotifySettingsGet(w, r)
}

// DELETE /api/v1/notify/settings (admin) — back to the container variables.
func (s *Server) handleNotifySettingsDelete(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(principal(r)) {
		writeErr(w, http.StatusForbidden, "admin only")
		return
	}
	if p := s.notifyFile(); p != "" {
		_ = os.Remove(p)
	}
	s.notifier.SetConfig(s.envNotifyConfig())
	s.handleNotifySettingsGet(w, r)
}

func (s *Server) envNotifyConfig() notify.Config {
	cfg := s.cfg
	return notify.Config{UnraidEnabled: cfg.NotifyUnraid, UnraidAPIKey: cfg.NotifyUnraidAPIKey, UnraidDir: cfg.NotifyUnraidDir,
		SMTPHost: cfg.SMTPHost, SMTPPort: cfg.SMTPPort, SMTPUser: cfg.SMTPUser, SMTPPassword: cfg.SMTPPassword, SMTPFrom: cfg.SMTPFrom, SMTPTo: cfg.SMTPTo, SMTPTLS: cfg.SMTPTLS,
		TelegramToken: cfg.TelegramToken, TelegramChatID: cfg.TelegramChatID}
}

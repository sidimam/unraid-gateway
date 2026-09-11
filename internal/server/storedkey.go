package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sidimam/unraid-gateway/internal/auth"
)

// The web UI can log in with a key kept on the gateway, so the admin does not paste it at every
// visit: either the WEBUI_API_KEY variable, or a file written by "Remember this key on the gateway"
// (WEBUI_API_KEY_FILE, default /config/webui.key, mode 0600). Anyone who can open the web UI can
// then use that key, so this is meant for a LAN-only or Access-protected gateway; the Unraid user
// and password, when USER_AUTH is on, are still asked.

// storedKey returns the key and where it comes from ("env", "file" or "").
func (s *Server) storedKey() (string, string) {
	if k := strings.TrimSpace(s.cfg.WebUIKey); k != "" {
		return k, "env"
	}
	if s.cfg.WebUIKeyFile != "" {
		if b, err := os.ReadFile(s.cfg.WebUIKeyFile); err == nil {
			if k := strings.TrimSpace(string(b)); k != "" {
				return k, "file"
			}
		}
	}
	return "", ""
}

// GET /api/v1/auth/stored-key → {"available": bool, "source": "env"|"file"|"", "canForget": bool}
func (s *Server) handleStoredKeyInfo(w http.ResponseWriter, _ *http.Request) {
	_, src := s.storedKey()
	writeJSON(w, http.StatusOK, map[string]any{"available": src != "", "source": src, "canForget": src == "file", "canRemember": s.cfg.WebUIKeyFile != "" && strings.TrimSpace(s.cfg.WebUIKey) == ""})
}

// POST /api/v1/auth/remember — stores the current session's API key on the gateway.
func (s *Server) handleRemember(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	if p.APIKey == "" {
		writeErr(w, http.StatusBadRequest, "this session has no API key to remember")
		return
	}
	if s.cfg.WebUIKeyFile == "" {
		writeErr(w, http.StatusBadRequest, "WEBUI_API_KEY_FILE is empty: remembering is disabled")
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.cfg.WebUIKeyFile), 0o700); err != nil {
		writeErr(w, http.StatusInternalServerError, "cannot create the key folder: "+err.Error())
		return
	}
	if err := writeSecret(s.cfg.WebUIKeyFile, p.APIKey); err != nil {
		writeErr(w, http.StatusInternalServerError, "cannot write the key file: "+err.Error()+" (mount /config read-write)")
		return
	}
	s.log.Info("web UI key remembered", "file", s.cfg.WebUIKeyFile, "key", p.Identity.Name, "ip", auth.ClientIP(r, s.cfg.TrustProxy))
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/auth/remember — forgets the stored key file (the env variable cannot be removed here).
func (s *Server) handleForget(w http.ResponseWriter, r *http.Request) {
	if s.cfg.WebUIKeyFile != "" {
		_ = os.Remove(s.cfg.WebUIKeyFile)
	}
	s.log.Info("web UI key forgotten", "file", s.cfg.WebUIKeyFile, "ip", auth.ClientIP(r, s.cfg.TrustProxy))
	w.WriteHeader(http.StatusNoContent)
}

// writeSecret stores a secret in a 0600 file (folder created if needed).
func writeSecret(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value+"\n"), 0o600)
}

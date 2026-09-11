// Package server wires configuration, authentication, the file API and the
// GraphQL proxy into one http.Handler.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sidimam/unraid-gateway/internal/index"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sidimam/unraid-gateway/internal/access"
	"github.com/sidimam/unraid-gateway/internal/activity"
	"github.com/sidimam/unraid-gateway/internal/auth"
	"github.com/sidimam/unraid-gateway/internal/config"
	"github.com/sidimam/unraid-gateway/internal/devices"
	"github.com/sidimam/unraid-gateway/internal/fsapi"
	"github.com/sidimam/unraid-gateway/internal/notify"
	"github.com/sidimam/unraid-gateway/internal/proxy"
	"github.com/sidimam/unraid-gateway/internal/smbauth"
	"github.com/sidimam/unraid-gateway/internal/unraidshares"
	"github.com/sidimam/unraid-gateway/internal/webui"
)

// Version is set at build time via -ldflags.
var Version = "dev"

type ctxKey struct{}

// Server is the composed HTTP handler.
type Server struct {
	cfg      config.Config
	log      *slog.Logger
	sessions *auth.Store
	files    *fsapi.API
	gql      *proxy.GraphQL
	smb      *smbauth.Authenticator
	shares   *unraidshares.Loader
	handler  http.Handler
	index    *index.Index
	scanner  *index.Scanner
	activity *activity.Tracker
	devices  *devices.Store
	notifier *notify.Notifier
}

// StartIndexer runs the background scanner until ctx is cancelled (no-op without index).
func (s *Server) StartIndexer(ctx context.Context) {
	if s.scanner == nil {
		return
	}
	go s.scanner.Run(ctx)
}

// Close releases the index database.
func (s *Server) Close() error {
	if s.index != nil {
		return s.index.Close()
	}
	return nil
}

func writable(dir string) bool {
	f, err := os.CreateTemp(dir, ".gw-write-test-*")
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(f.Name())
	return true
}

// New builds the server.
func New(cfg config.Config, log *slog.Logger) (*Server, error) {
	validator := auth.NewUnraidValidator(cfg.UnraidURL, cfg.ValidateQuery, cfg.UnraidInsecureTLS)
	tracker := activity.New()
	files, err := fsapi.New(cfg.DataRoot, fsapi.Options{
		Activity:         &activityHook{t: tracker, trustProxy: cfg.TrustProxy},
		ReadOnly:         cfg.ReadOnly,
		MaxJSONBody:      cfg.MaxJSONBody,
		ChangesWalkLimit: cfg.ChangesWalkLimit,
		ChangesDeadline:  cfg.ChangesDeadline,
		UploadTTL:        cfg.UploadTTL,
	}, log)
	if err != nil {
		return nil, err
	}
	files.WarnUnwritableShares()
	var ix *index.Index
	var scanner *index.Scanner
	if cfg.IndexDB != "" && strings.ToLower(cfg.IndexDB) != "off" {
		dbPath := cfg.IndexDB
		if dbPath != ":memory:" {
			if err := os.MkdirAll(filepath.Dir(dbPath), fsapi.DirMode); err != nil || !writable(filepath.Dir(dbPath)) {
				fallback := filepath.Join(os.TempDir(), "unraid-gateway-index.db")
				log.Warn("index: INDEX_DB location not writable, using a temporary database (mount /config to keep it across restarts)", "wanted", dbPath, "using", fallback)
				dbPath = fallback
			}
		}
		ix, err = index.Open(dbPath)
		if err != nil {
			log.Warn("index: disabled, cannot open database", "path", dbPath, "err", err)
			ix = nil
		} else {
			sc := &index.Scanner{Index: ix, Root: cfg.DataRoot, Shares: files.MountedShares,
				Skip:        func(name string) bool { return strings.HasPrefix(name, ".") || strings.Contains(name, ".gwpart") },
				DirInterval: cfg.IndexDirScan, FullInterval: cfg.IndexFullScan, Log: log}
			files.UseIndex(ix, sc)
			scanner = sc
			log.Info("index: enabled", "db", dbPath, "dirScan", cfg.IndexDirScan, "fullScan", cfg.IndexFullScan)
		}
	}
	devStore, err := devices.Open(cfg.DevicesFile)
	if err != nil {
		log.Warn("devices: registry unreadable, starting empty", "file", cfg.DevicesFile, "err", err)
		devStore, _ = devices.Open("")
	}
	gqlProxy := proxy.New(cfg.UnraidURL, cfg.UnraidInsecureTLS)
	s := &Server{
		cfg:     cfg,
		devices: devStore,
		notifier: notify.New(notify.Config{
			UnraidEnabled: cfg.NotifyUnraid, UnraidAPIKey: cfg.NotifyUnraidAPIKey,
			SMTPHost: cfg.SMTPHost, SMTPPort: cfg.SMTPPort, SMTPUser: cfg.SMTPUser, SMTPPassword: cfg.SMTPPassword, SMTPFrom: cfg.SMTPFrom, SMTPTo: cfg.SMTPTo, SMTPTLS: cfg.SMTPTLS,
			TelegramToken: cfg.TelegramToken, TelegramChatID: cfg.TelegramChatID,
		}, gqlProxy, log),
		index:    ix,
		scanner:  scanner,
		log:      log,
		sessions: auth.NewStore(validator, cfg.SessionTTL, cfg.MaxLoginAttempts, cfg.LoginLockout),
		files:    files,
		gql:      gqlProxy,
		smb:      &smbauth.Authenticator{Addr: cfg.SMBAddr, Timeout: 10 * time.Second},
		shares:   &unraidshares.Loader{Path: cfg.SharesConfig},
		activity: tracker,
	}
	if cfg.UserAuth != "off" {
		if _, err := s.shares.Shares(); err != nil {
			log.Warn("USER_AUTH is enabled but the Unraid share configuration is not readable: user logins will fail until /etc/samba/smb-shares.conf is mounted", "path", cfg.SharesConfig, "err", err)
		}
	}
	s.handler = s.routes()
	return s, nil
}

// SharesSummary lists the mounted shares for the startup banner.
func (s *Server) SharesSummary() string {
	shares := s.files.MountedShares()
	if len(shares) == 0 {
		return "none mounted!"
	}
	return strings.Join(shares, ", ")
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.handler.ServeHTTP(w, r) }

func (s *Server) routes() http.Handler {
	public := http.NewServeMux()
	public.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": Version})
	})
	public.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	public.HandleFunc("GET /api/v1/auth/stored-key", s.handleStoredKeyInfo)

	private := http.NewServeMux()
	private.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	private.HandleFunc("POST /api/v1/auth/remember", s.handleRemember)
	private.HandleFunc("DELETE /api/v1/auth/remember", s.handleForget)
	private.HandleFunc("GET /api/v1/auth/session", s.handleSession)
	private.HandleFunc("GET /api/v1/info", s.handleInfo)
	private.HandleFunc("GET /api/v1/activity", s.handleActivity)
	private.HandleFunc("GET /api/v1/devices", s.handleDevicesList)
	private.HandleFunc("DELETE /api/v1/devices/{id}", s.handleDeviceRemove)
	private.HandleFunc("DELETE /api/v1/devices", s.handleDevicesRemoveAll)
	private.HandleFunc("POST /api/v1/notify/test", s.handleNotifyTest)
	private.HandleFunc("GET /api/v1/notify/channels", s.handleNotifyChannels)
	private.HandleFunc("GET /api/v1/keys", s.handleKeysList)
	private.HandleFunc("POST /api/v1/keys", s.handleKeyCreate)
	private.HandleFunc("DELETE /api/v1/keys/{id}", s.handleKeyDelete)
	private.HandleFunc("POST /api/v1/keys/rotate", s.handleKeyRotate)
	// The container console (`gw activity`) reads it on loopback without a token; everyone else
	// goes through the normal authentication. More specific pattern, so it wins over /api/v1/.
	public.Handle("GET /api/v1/activity", s.loopbackOr(s.authenticate(private)))
	public.Handle("GET /api/v1/devices", s.loopbackOr(s.authenticate(private)))
	public.Handle("DELETE /api/v1/devices", s.loopbackOr(s.authenticate(private)))
	public.Handle("DELETE /api/v1/devices/{id}", s.loopbackOr(s.authenticate(private)))
	public.Handle("POST /api/v1/notify/test", s.loopbackOr(s.authenticate(private)))
	public.Handle("GET /api/v1/notify/channels", s.loopbackOr(s.authenticate(private)))
	private.HandleFunc("POST /api/v1/graphql", s.handleGraphQL)
	s.files.Register(private, "/api/v1/fs")

	// Media tickets carry their own signature: no bearer token, players can stream by URL.
	s.files.RegisterPublic(public)
	public.Handle("/api/v1/", s.authenticate(private))
	webui.Register(public)
	return s.recoverer(s.securityHeaders(s.accessLog(public)))
}

// ---- auth ------------------------------------------------------------------

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey         string `json:"apiKey"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		UseStoredKey   bool   `json:"useStoredKey"`
		DeviceID       string `json:"deviceId"`
		DeviceName     string `json:"deviceName"`
		RegisterDevice bool   `json:"registerDevice"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "expected JSON {\"apiKey\": \"...\", \"username\": \"...\", \"password\": \"...\"}")
		return
	}
	ip := auth.ClientIP(r, s.cfg.TrustProxy)
	if req.UseStoredKey {
		key, _ := s.storedKey()
		if key == "" {
			writeErr(w, http.StatusBadRequest, "no API key is stored on this gateway (set WEBUI_API_KEY or tick \"Remember this key\" once)")
			return
		}
		req.APIKey = key
	}
	req.Username = strings.TrimSpace(req.Username)
	if s.cfg.UserAuth == "required" && req.Username == "" {
		writeErr(w, http.StatusUnauthorized, "this gateway requires an Unraid username and password in addition to the API key")
		return
	}
	if s.cfg.UserAuth == "off" {
		req.Username, req.Password = "", ""
	}
	// 1. The API key, validated by Unraid (also enforces the per-IP lockout).
	sess, err := s.sessions.Login(r.Context(), ip, req.APIKey)
	if err != nil {
		s.authError(w, ip, err)
		return
	}
	// 2. Optional Unraid user, validated by Samba; the session then carries the
	//    user's share permissions read from Unraid's share configuration.
	if req.Username != "" {
		mounted := s.files.MountedShares()
		policy, err := s.shares.PolicyFor(req.Username, mounted)
		if err != nil {
			s.sessions.Logout(sess.Token)
			s.log.Error("shares config unreadable", "err", err, "path", s.cfg.SharesConfig)
			writeErr(w, http.StatusBadGateway, "the Unraid share configuration is not readable by the gateway (SHARES_CONFIG)")
			return
		}
		// Samba maps unknown users to guest, so also open a share only the real user may open.
		probe := s.shares.ProbeShare(req.Username, mounted)
		if err := s.smb.Check(r.Context(), req.Username, req.Password, probe); err != nil {
			s.sessions.Logout(sess.Token)
			if errors.Is(err, smbauth.ErrInvalidCredentials) || errors.Is(err, smbauth.ErrGuest) {
				s.sessions.Fail(ip)
				s.log.Warn("login failed: Unraid rejected the username/password", "user", req.Username, "ip", ip, "guest", errors.Is(err, smbauth.ErrGuest))
				writeErr(w, http.StatusUnauthorized, "invalid unraid username or password")
				return
			}
			s.log.Error("smb auth error", "err", err, "addr", s.cfg.SMBAddr)
			writeErr(w, http.StatusBadGateway, "cannot reach the Unraid SMB service to verify the user")
			return
		}
		visible := 0
		for _, m := range mounted {
			if policy.Level(m) != access.None {
				visible++
			}
		}
		if visible == 0 {
			s.sessions.Logout(sess.Token)
			s.sessions.Fail(ip)
			s.log.Warn("login refused: user has no access to any mounted share", "user", req.Username, "ip", ip)
			writeErr(w, http.StatusUnauthorized, "this Unraid user has no access to any share mounted in the gateway")
			return
		}
		sess.User, sess.Policy = req.Username, policy
	}
	// 3. Device registration: apps identify their installation; a device the admin removed must be
	//    registered again explicitly (the app asks the user), which also raises a notification.
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if s.cfg.DeviceRegistration != "off" && req.DeviceID != "" {
		name := strings.TrimSpace(req.DeviceName)
		if name == "" {
			name = strings.TrimSpace(r.Header.Get("X-Unraid-Drive-Client"))
		}
		if name == "" {
			name = r.UserAgent()
		}
		if !s.devices.Known(req.DeviceID) && !req.RegisterDevice {
			s.sessions.Logout(sess.Token)
			s.log.Warn("login refused: device not registered", "device", req.DeviceID, "name", name, "user", sess.User, "ip", ip)
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "this device is not registered on the gateway: sign in again to register it", "code": "device_not_registered"})
			return
		}
		isNew, err := s.devices.Register(req.DeviceID, name, sess.User, sess.Identity.Name, ip)
		if err != nil {
			s.log.Warn("devices: cannot save the registry", "err", err, "file", s.cfg.DevicesFile)
		}
		sess.DeviceID = req.DeviceID
		if isNew {
			s.log.Info("device registered", "device", req.DeviceID, "name", name, "user", sess.User, "key", sess.Identity.Name, "ip", ip)
			who := sess.User
			if who == "" {
				who = "key " + sess.Identity.Name
			}
			body := fmt.Sprintf("%s\nUser: %s\nFrom: %s\nWhen: %s\n\nNot you? Remove it in the gateway web UI (port 8484) → Devices, or with `gw devices rm`.", name, who, ip, time.Now().Format("2006-01-02 15:04"))
			go s.notifier.Send(context.Background(), sess.APIKey, "New device registered on unraid-gateway", body, notify.Warning)
		}
	}
	if sess.User != "" {
		s.log.Info("login ok (user)", "user", sess.User, "key", sess.Identity.Name, "ip", ip, "shares", sharesLine(sess, s.files.MountedShares()))
	} else {
		s.log.Info("login ok (api key only)", "key", sess.Identity.Name, "ip", ip)
	}
	writeJSON(w, http.StatusOK, s.sessionResponse(sess))
}

// sharesLine renders "documents=rw media=ro" for the log.
func sharesLine(sess *auth.Session, mounted []string) string {
	parts := []string{}
	for _, m := range mounted {
		if l := sess.Policy.Level(m); l != access.None {
			parts = append(parts, m+"="+l.String())
		}
	}
	return strings.Join(parts, " ")
}

// sessionResponse describes a session to the client, including the effective share access.
func (s *Server) sessionResponse(sess *auth.Session) map[string]any {
	out := map[string]any{
		"token":     sess.Token,
		"expiresAt": sess.ExpiresAt.UTC(),
		"identity":  sess.Identity,
		"readOnly":  s.cfg.ReadOnly,
		"version":   Version,
		"userAuth":  s.cfg.UserAuth,
	}
	if sess.DeviceID != "" {
		out["deviceId"] = sess.DeviceID
	}
	if sess.User != "" {
		out["user"] = sess.User
		shares := map[string]string{}
		for _, name := range s.files.MountedShares() {
			if l := sess.Policy.Level(name); l != access.None {
				shares[name] = l.String()
			}
		}
		out["shares"] = shares
	}
	return out
}

func (s *Server) authError(w http.ResponseWriter, ip string, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidKey):
		s.log.Warn("login failed: Unraid rejected the API key", "ip", ip)
		writeErr(w, http.StatusUnauthorized, "invalid api key")
	case errors.Is(err, auth.ErrLocked):
		s.log.Warn("login refused: too many failed attempts from this IP", "ip", ip)
		w.Header().Set("Retry-After", "900")
		writeErr(w, http.StatusTooManyRequests, "too many failed attempts")
	default:
		s.log.Error("unraid validation error", "err", err)
		writeErr(w, http.StatusBadGateway, "cannot reach unraid api")
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if tok := bearer(r); tok != "" {
		s.sessions.Logout(tok)
		s.activity.Forget("s:" + shortHash(tok))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	out := map[string]any{"identity": p.Identity, "readOnly": s.cfg.ReadOnly, "userAuth": s.cfg.UserAuth}
	if p.User != "" {
		out["user"] = p.User
		shares := map[string]string{}
		for _, name := range s.files.MountedShares() {
			if l := p.Policy.Level(name); l != access.None {
				shares[name] = l.String()
			}
		}
		out["shares"] = shares
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":  Version,
		"readOnly": s.cfg.ReadOnly,
		"userAuth": s.cfg.UserAuth,
		"identity": principal(r).Identity,
		"user":     principal(r).User,
		"features": []string{"fs.list", "fs.content", "fs.range", "fs.uploads", "fs.changes", "graphql", "users"},
	})
}

func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	s.gql.Forward(w, r, principal(r).APIKey, s.cfg.MaxJSONBody)
}

// authenticate accepts either a session bearer token or a raw x-api-key.
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := auth.ClientIP(r, s.cfg.TrustProxy)
		var p auth.Principal
		if tok := bearer(r); tok != "" {
			sess, ok := s.sessions.Lookup(tok)
			if !ok {
				writeErr(w, http.StatusUnauthorized, "invalid or expired session")
				return
			}
			if sess.DeviceID != "" && s.cfg.DeviceRegistration != "off" && !s.devices.Known(sess.DeviceID) {
				s.sessions.Logout(tok)
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "this device was removed from the gateway: sign in again to register it", "code": "device_revoked"})
				return
			}
			if sess.DeviceID != "" {
				_ = s.devices.Touch(sess.DeviceID, ip)
			}
			p = auth.Principal{APIKey: sess.APIKey, Identity: sess.Identity, User: sess.User, Policy: sess.Policy, DeviceID: sess.DeviceID}
		} else if key := strings.TrimSpace(r.Header.Get("x-api-key")); key != "" {
			if s.cfg.UserAuth == "required" {
				writeErr(w, http.StatusUnauthorized, "this gateway requires a login with Unraid username and password")
				return
			}
			id, err := s.sessions.AuthenticateKey(r.Context(), ip, key)
			if err != nil {
				s.authError(w, ip, err)
				return
			}
			p = auth.Principal{APIKey: key, Identity: id}
		} else {
			w.Header().Set("WWW-Authenticate", `Bearer realm="unraid-gateway"`)
			writeErr(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if rec, ok := w.(principalRecorder); ok {
			rec.setPrincipal(p)
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, p)
		if p.Policy != nil {
			ctx = access.WithPolicy(ctx, p.Policy)
		}
		s.activity.Touch(whoFor(r, p, ip), r.URL.Query().Get("path"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func principal(r *http.Request) auth.Principal {
	p, _ := r.Context().Value(ctxKey{}).(auth.Principal)
	return p
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// ---- middleware ------------------------------------------------------------

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status    int
	principal *auth.Principal
}

// principalRecorder lets the authenticate middleware report who made the request to the access log.
type principalRecorder interface{ setPrincipal(auth.Principal) }

func (w *statusWriter) setPrincipal(p auth.Principal) { w.principal = &p }

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	if !s.cfg.LogRequests {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if r.URL.Path == "/healthz" && !s.cfg.LogHealthchecks {
			return
		}
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"ms", time.Since(start).Milliseconds(),
			"ip", auth.ClientIP(r, s.cfg.TrustProxy),
		}
		// Which file/folder a file-API call touched, and as which user.
		if p := r.URL.Query().Get("path"); p != "" {
			attrs = append(attrs, "file", p)
		}
		if pr := sw.principal; pr != nil {
			if pr.User != "" {
				attrs = append(attrs, "user", pr.User)
			} else if pr.Identity.Name != "" {
				attrs = append(attrs, "user", "key:"+pr.Identity.Name)
			}
		}
		s.log.Info("request", attrs...)
	})
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("panic", "err", rec, "path", r.URL.Path)
				writeErr(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

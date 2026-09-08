// Package server wires configuration, authentication, the file API and the
// GraphQL proxy into one http.Handler.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sidimam/unraid-gateway/internal/auth"
	"github.com/sidimam/unraid-gateway/internal/config"
	"github.com/sidimam/unraid-gateway/internal/fsapi"
	"github.com/sidimam/unraid-gateway/internal/proxy"
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
	handler  http.Handler
}

// New builds the server.
func New(cfg config.Config, log *slog.Logger) (*Server, error) {
	validator := auth.NewUnraidValidator(cfg.UnraidURL, cfg.ValidateQuery, cfg.UnraidInsecureTLS)
	files, err := fsapi.New(cfg.DataRoot, fsapi.Options{
		ReadOnly:         cfg.ReadOnly,
		MaxJSONBody:      cfg.MaxJSONBody,
		ChangesWalkLimit: cfg.ChangesWalkLimit,
		ChangesDeadline:  cfg.ChangesDeadline,
		UploadTTL:        cfg.UploadTTL,
	}, log)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:      cfg,
		log:      log,
		sessions: auth.NewStore(validator, cfg.SessionTTL, cfg.MaxLoginAttempts, cfg.LoginLockout),
		files:    files,
		gql:      proxy.New(cfg.UnraidURL, cfg.UnraidInsecureTLS),
	}
	s.handler = s.routes()
	return s, nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.handler.ServeHTTP(w, r) }

func (s *Server) routes() http.Handler {
	public := http.NewServeMux()
	public.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": Version})
	})
	public.HandleFunc("POST /api/v1/auth/login", s.handleLogin)

	private := http.NewServeMux()
	private.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	private.HandleFunc("GET /api/v1/auth/session", s.handleSession)
	private.HandleFunc("GET /api/v1/info", s.handleInfo)
	private.HandleFunc("POST /api/v1/graphql", s.handleGraphQL)
	s.files.Register(private, "/api/v1/fs")

	public.Handle("/api/v1/", s.authenticate(private))
	public.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeErr(w, http.StatusNotFound, "not found")
	})
	return s.recoverer(s.securityHeaders(s.accessLog(public)))
}

// ---- auth ------------------------------------------------------------------

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"apiKey"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "expected JSON {\"apiKey\": \"...\"}")
		return
	}
	ip := auth.ClientIP(r, s.cfg.TrustProxy)
	sess, err := s.sessions.Login(r.Context(), ip, req.APIKey)
	if err != nil {
		s.authError(w, ip, err)
		return
	}
	s.log.Info("login", "ip", ip, "user", sess.Identity.Name)
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     sess.Token,
		"expiresAt": sess.ExpiresAt.UTC(),
		"identity":  sess.Identity,
		"readOnly":  s.cfg.ReadOnly,
		"version":   Version,
	})
}

func (s *Server) authError(w http.ResponseWriter, ip string, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidKey):
		s.log.Warn("login failed", "ip", ip)
		writeErr(w, http.StatusUnauthorized, "invalid api key")
	case errors.Is(err, auth.ErrLocked):
		s.log.Warn("login locked", "ip", ip)
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
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	writeJSON(w, http.StatusOK, map[string]any{"identity": p.Identity, "readOnly": s.cfg.ReadOnly})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":  Version,
		"readOnly": s.cfg.ReadOnly,
		"identity": principal(r).Identity,
		"features": []string{"fs.list", "fs.content", "fs.range", "fs.uploads", "fs.changes", "graphql"},
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
			p = auth.Principal{APIKey: sess.APIKey, Identity: sess.Identity}
		} else if key := strings.TrimSpace(r.Header.Get("x-api-key")); key != "" {
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, p)))
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
	status int
}

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
		s.log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"ms", time.Since(start).Milliseconds(),
			"ip", auth.ClientIP(r, s.cfg.TrustProxy),
		)
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

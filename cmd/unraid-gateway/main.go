// Command unraid-gateway exposes an Unraid server to mobile clients through a
// single authenticated HTTPS port: a file API for the shares plus a proxy to
// the Unraid GraphQL API, both authenticated with an Unraid API key.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sidimam/unraid-gateway/internal/config"
	"github.com/sidimam/unraid-gateway/internal/logfmt"
	"github.com/sidimam/unraid-gateway/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(logfmt.New("text", slog.LevelInfo)).Error("configuration error", "err", err)
		os.Exit(2)
	}
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	log := slog.New(logfmt.New(cfg.LogFormat, level))
	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("startup error", "err", err)
		os.Exit(1)
	}
	indexCtx, stopIndexer := context.WithCancel(context.Background())
	srv.StartIndexer(indexCtx)
	httpSrv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
		// No WriteTimeout/ReadTimeout: uploads and downloads may be long.
	}
	if cfg.LogFormat != "json" {
		logfmt.Banner(os.Stdout, server.Version, map[string]string{
			"Listening on":        cfg.ListenAddr + map[bool]string{true: " (TLS)", false: " (HTTP, put TLS in front)"}[cfg.TLSCert != ""],
			"Unraid API":          cfg.UnraidURL,
			"Shares (data root)":  cfg.DataRoot + "  →  " + srv.SharesSummary(),
			"User authentication": cfg.UserAuth + "  (SMB " + cfg.SMBAddr + ", share config " + cfg.SharesConfig + ")",
			"Read-only":           map[bool]string{true: "yes", false: "no"}[cfg.ReadOnly],
			"Trust proxy headers": map[bool]string{true: "yes", false: "no"}[cfg.TrustProxy],
			"Sessions / lockout":  cfg.SessionTTL.String() + " / " + fmt.Sprintf("%d attempts, %s", cfg.MaxLoginAttempts, cfg.LoginLockout),
			"Item index":          map[bool]string{true: cfg.IndexDB + "  (dir scan " + cfg.IndexDirScan.String() + ", full scan " + cfg.IndexFullScan.String() + ")", false: "off"}[cfg.IndexDB != "" && cfg.IndexDB != "off"],
		}, []string{"Listening on", "Unraid API", "Shares (data root)", "User authentication", "Read-only", "Trust proxy headers", "Sessions / lockout", "Item index"})
	}
	go func() {
		log.Info("unraid-gateway listening",
			"addr", cfg.ListenAddr, "unraid", cfg.UnraidURL, "data", cfg.DataRoot,
			"readOnly", cfg.ReadOnly, "tls", cfg.TLSCert != "", "version", server.Version)
		var err error
		if cfg.TLSCert != "" {
			err = httpSrv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			err = httpSrv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")
	stopIndexer()
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}

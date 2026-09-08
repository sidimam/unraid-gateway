// Command unraid-gateway exposes an Unraid server to mobile clients through a
// single authenticated HTTPS port: a file API for the shares plus a proxy to
// the Unraid GraphQL API, both authenticated with an Unraid API key.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sidimam/unraid-gateway/internal/config"
	"github.com/sidimam/unraid-gateway/internal/server"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration error", "err", err)
		os.Exit(2)
	}
	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("startup error", "err", err)
		os.Exit(1)
	}
	httpSrv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
		// No WriteTimeout/ReadTimeout: uploads and downloads may be long.
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}

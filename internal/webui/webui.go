// Package webui serves the small embedded browser UI: login with an Unraid
// API key, browse/upload/download files, try the GraphQL proxy. It is a
// convenience for testing and administration; the iOS app talks to the JSON
// API directly.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var static embed.FS

// Register mounts the UI at "/" and its assets under "/ui/".
func Register(mux *http.ServeMux) {
	sub, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}
	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err)
	}
	icon, err := fs.ReadFile(sub, "icon.png")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(sub))
	// index.html is written directly: http.FileServer would 301 it to "./".
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; connect-src 'self'; object-src 'none'; frame-ancestors 'none'")
		h.Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
	mux.Handle("GET /ui/", http.StripPrefix("/ui/", files))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(icon)
	})
}

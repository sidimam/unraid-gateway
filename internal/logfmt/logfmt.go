// Package logfmt provides a human-readable slog handler for the container
// logs shown in the Unraid Docker tab, plus the JSON handler for machines.
package logfmt

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// New returns a handler for the requested format ("text" or "json").
func New(format string, level slog.Level) slog.Handler {
	if strings.EqualFold(format, "json") {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	return &textHandler{w: os.Stdout, level: level, mu: &sync.Mutex{}}
}

// textHandler prints one aligned, readable line per record:
//
//	20:15:03 INFO   login ok                 user=sdimambro key="unraid gateway" ip=203.0.113.7
//	20:15:04 INFO   PUT  /api/v1/fs/content  → 201 in 42ms  path=/documents/a.pdf user=sdimambro ip=…
type textHandler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	mu    *sync.Mutex
}

func (h *textHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &textHandler{w: h.w, level: h.level, mu: h.mu, attrs: append(append([]slog.Attr{}, h.attrs...), attrs...)}
}

func (h *textHandler) WithGroup(string) slog.Handler { return h }

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := map[string]string{}
	var order []string
	add := func(a slog.Attr) {
		if a.Key == "" {
			return
		}
		if _, seen := attrs[a.Key]; !seen {
			order = append(order, a.Key)
		}
		attrs[a.Key] = a.Value.String()
	}
	for _, a := range h.attrs {
		add(a)
	}
	r.Attrs(func(a slog.Attr) bool { add(a); return true })

	var b strings.Builder
	b.WriteString(r.Time.Format("2006-01-02 15:04:05"))
	b.WriteByte(' ')
	b.WriteString(fmt.Sprintf("%-5s ", levelName(r.Level)))

	if r.Message == "request" {
		// Access line: METHOD PATH → STATUS in Nms, then the interesting attributes.
		method, path, status, ms := attrs["method"], attrs["path"], attrs["status"], attrs["ms"]
		b.WriteString(fmt.Sprintf("%-6s %-28s → %s in %sms", method, path, status, ms))
		for _, k := range []string{"file", "user", "ip"} {
			if v, ok := attrs[k]; ok && v != "" {
				b.WriteString(fmt.Sprintf("  %s=%s", k, quote(v)))
			}
		}
	} else {
		b.WriteString(fmt.Sprintf("%-32s", r.Message))
		sort.SliceStable(order, func(i, j int) bool { return rank(order[i]) < rank(order[j]) })
		for _, k := range order {
			if v := attrs[k]; v != "" {
				b.WriteString(fmt.Sprintf(" %s=%s", k, quote(v)))
			}
		}
	}
	b.WriteByte('\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

func levelName(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERROR"
	case l >= slog.LevelWarn:
		return "WARN"
	case l >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

// rank keeps the most useful attributes first.
func rank(k string) int {
	switch k {
	case "user":
		return 0
	case "key":
		return 1
	case "ip":
		return 2
	case "err":
		return 9
	default:
		return 5
	}
}

func quote(v string) string {
	if strings.ContainsAny(v, " \t\"") {
		return fmt.Sprintf("%q", v)
	}
	return v
}

// Banner prints a readable startup summary.
func Banner(w io.Writer, version string, lines map[string]string, order []string) {
	fmt.Fprintf(w, "\n  unraid-gateway %s\n  %s\n", version, strings.Repeat("─", 60))
	for _, k := range order {
		fmt.Fprintf(w, "  %-22s %s\n", k, lines[k])
	}
	fmt.Fprintf(w, "  %s\n  Open the Console of this container for an interactive walkthrough (type: gw).\n\n", strings.Repeat("─", 60))
	_ = time.Now
}

package server

import (
	"sync"
	"time"
	"unicode/utf8"
)

// utf8Header repairs a header value that a client sent as Latin-1 bytes instead of UTF-8
// ("Unraid Drive 1.5 (45) · iPhone" arrives with a lone 0xB7 for "·", which json.Marshal would turn
// into U+FFFD). Valid UTF-8 is returned unchanged; otherwise every byte is read as one Latin-1 rune.
func utf8Header(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	r := make([]rune, 0, len(s))
	for i := 0; i < len(s); i++ {
		r = append(r, rune(s[i]))
	}
	return string(r)
}

// loginCoalescer keeps one "signed in" event per device per window: the File Provider extension
// re-authenticates at almost every request, which used to flood the device log (0.15).
type loginCoalescer struct {
	mu   sync.Mutex
	last map[string]time.Time
}

// Allow reports whether a sign-in by device id should be logged now and records it if so.
func (c *loginCoalescer) Allow(id string, window time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.last == nil {
		c.last = map[string]time.Time{}
	}
	now := time.Now()
	if t, ok := c.last[id]; ok && now.Sub(t) < window {
		return false
	}
	c.last[id] = now
	return true
}

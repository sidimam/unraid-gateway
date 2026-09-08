// Package auth validates Unraid API keys against the Unraid GraphQL API and
// manages short-lived gateway sessions.
package auth

import (
	"github.com/sidimam/unraid-gateway/internal/access"

	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ErrInvalidKey is returned when Unraid rejects the API key.
var ErrInvalidKey = errors.New("invalid api key")

// ErrLocked is returned when an IP exceeded the login attempts budget.
var ErrLocked = errors.New("too many failed attempts, try again later")

// Identity describes the Unraid principal behind an API key.
type Identity struct {
	Name  string   `json:"name,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// Principal is what handlers receive after authentication.
type Principal struct {
	APIKey   string
	Identity Identity
	// User is the Unraid user name when the session was opened with user credentials.
	User string
	// Policy is the per-share access of the session (nil = unrestricted).
	Policy access.Policy
}

// Validator checks API keys.
type Validator interface {
	Validate(ctx context.Context, apiKey string) (Identity, error)
}

// UnraidValidator validates keys by running a GraphQL query on Unraid.
type UnraidValidator struct {
	URL    string
	Query  string
	Client *http.Client
}

// NewUnraidValidator builds a validator for the given Unraid base URL.
func NewUnraidValidator(baseURL, query string, insecure bool) *UnraidValidator {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in for self-signed Unraid certs
	}
	return &UnraidValidator{
		URL:    strings.TrimRight(baseURL, "/") + "/graphql",
		Query:  query,
		Client: &http.Client{Transport: tr, Timeout: 15 * time.Second},
	}
}

var unauthorizedRe = regexp.MustCompile(`(?i)unauth|forbidden|api.?key|not.?allowed|invalid.?token`)

type gqlResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code string `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
}

// Validate implements Validator.
func (v *UnraidValidator) Validate(ctx context.Context, apiKey string) (Identity, error) {
	body, _ := json.Marshal(map[string]string{"query": v.Query})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.URL, bytes.NewReader(body))
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, err := v.Client.Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("unraid unreachable: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Identity{}, err
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return Identity{}, ErrInvalidKey
	case resp.StatusCode/100 != 2:
		return Identity{}, fmt.Errorf("unraid returned HTTP %d", resp.StatusCode)
	}
	var gr gqlResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return Identity{}, fmt.Errorf("unraid returned non-JSON body")
	}
	for _, e := range gr.Errors {
		code := strings.ToUpper(e.Extensions.Code)
		if code == "UNAUTHENTICATED" || code == "FORBIDDEN" || unauthorizedRe.MatchString(e.Message) {
			return Identity{}, ErrInvalidKey
		}
	}
	if len(gr.Errors) > 0 && len(gr.Data) == 0 {
		return Identity{}, fmt.Errorf("validation query failed: %s (check UNRAID_VALIDATE_QUERY)", gr.Errors[0].Message)
	}
	id := Identity{}
	if me, ok := gr.Data["me"]; ok {
		var m struct {
			Name  string   `json:"name"`
			Roles []string `json:"roles"`
		}
		if json.Unmarshal(me, &m) == nil {
			id.Name, id.Roles = m.Name, m.Roles
		}
	}
	return id, nil
}

// Session is an authenticated gateway session.
type Session struct {
	Token     string
	APIKey    string
	Identity  Identity
	User      string
	Policy    access.Policy
	ExpiresAt time.Time
}

// Store keeps sessions, an API-key validation cache and the login limiter.
type Store struct {
	validator Validator
	ttl       time.Duration
	limiter   *limiter

	mu       sync.Mutex
	sessions map[string]*Session
	keyCache map[string]cachedKey
}

type cachedKey struct {
	identity Identity
	expires  time.Time
}

// NewStore creates a session store.
func NewStore(v Validator, ttl time.Duration, maxAttempts int, lockout time.Duration) *Store {
	s := &Store{
		validator: v,
		ttl:       ttl,
		limiter:   newLimiter(maxAttempts, lockout),
		sessions:  map[string]*Session{},
		keyCache:  map[string]cachedKey{},
	}
	go s.janitor()
	return s
}

// Login validates the API key and returns a new session.
func (s *Store) Login(ctx context.Context, ip, apiKey string) (*Session, error) {
	if s.limiter.locked(ip) {
		return nil, ErrLocked
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		s.limiter.fail(ip)
		return nil, ErrInvalidKey
	}
	id, err := s.validator.Validate(ctx, apiKey)
	if err != nil {
		if errors.Is(err, ErrInvalidKey) {
			s.limiter.fail(ip)
		}
		return nil, err
	}
	s.limiter.reset(ip)
	tok, err := randomToken()
	if err != nil {
		return nil, err
	}
	sess := &Session{Token: tok, APIKey: apiKey, Identity: id, ExpiresAt: time.Now().Add(s.ttl)}
	s.mu.Lock()
	s.sessions[tok] = sess
	s.mu.Unlock()
	return sess, nil
}

// Fail records a failed attempt for ip (used for user/password failures).
func (s *Store) Fail(ip string) { s.limiter.fail(ip) }

// Locked reports whether ip is currently locked out.
func (s *Store) Locked(ip string) bool { return s.limiter.locked(ip) }

// Logout removes a session.
func (s *Store) Logout(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// Lookup returns the session for a token, if valid.
func (s *Store) Lookup(token string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[token]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, false
	}
	return sess, true
}

// AuthenticateKey validates a raw API key, using a short cache so that clients
// sending x-api-key on every request do not hammer Unraid.
func (s *Store) AuthenticateKey(ctx context.Context, ip, apiKey string) (Identity, error) {
	h := hashKey(apiKey)
	s.mu.Lock()
	if c, ok := s.keyCache[h]; ok && time.Now().Before(c.expires) {
		s.mu.Unlock()
		return c.identity, nil
	}
	s.mu.Unlock()
	if s.limiter.locked(ip) {
		return Identity{}, ErrLocked
	}
	id, err := s.validator.Validate(ctx, apiKey)
	if err != nil {
		if errors.Is(err, ErrInvalidKey) {
			s.limiter.fail(ip)
		}
		return Identity{}, err
	}
	s.limiter.reset(ip)
	s.mu.Lock()
	s.keyCache[h] = cachedKey{identity: id, expires: time.Now().Add(5 * time.Minute)}
	s.mu.Unlock()
	return id, nil
}

func (s *Store) janitor() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.sessions {
			if now.After(v.ExpiresAt) {
				delete(s.sessions, k)
			}
		}
		for k, v := range s.keyCache {
			if now.After(v.expires) {
				delete(s.keyCache, k)
			}
		}
		s.mu.Unlock()
		s.limiter.sweep()
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashKey(k string) string {
	sum := sha256.Sum256([]byte(k))
	return hex.EncodeToString(sum[:])
}

// limiter tracks failed logins per client IP.
type limiter struct {
	max     int
	lockout time.Duration
	mu      sync.Mutex
	entries map[string]*attempt
}

type attempt struct {
	failures    int
	lockedUntil time.Time
	last        time.Time
}

func newLimiter(max int, lockout time.Duration) *limiter {
	if max <= 0 {
		max = 5
	}
	return &limiter{max: max, lockout: lockout, entries: map[string]*attempt{}}
}

func (l *limiter) locked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.entries[ip]
	return ok && time.Now().Before(a.lockedUntil)
}

func (l *limiter) fail(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.entries[ip]
	if !ok {
		a = &attempt{}
		l.entries[ip] = a
	}
	a.failures++
	a.last = time.Now()
	if a.failures >= l.max {
		a.lockedUntil = time.Now().Add(l.lockout)
		a.failures = 0
	}
}

func (l *limiter) reset(ip string) {
	l.mu.Lock()
	delete(l.entries, ip)
	l.mu.Unlock()
}

func (l *limiter) sweep() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for ip, a := range l.entries {
		if now.After(a.lockedUntil) && now.Sub(a.last) > l.lockout {
			delete(l.entries, ip)
		}
	}
}

// ClientIP extracts the client address, optionally honouring X-Forwarded-For.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i >= 0 {
				xff = xff[:i]
			}
			return strings.TrimSpace(xff)
		}
		if rip := r.Header.Get("X-Real-IP"); rip != "" {
			return strings.TrimSpace(rip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

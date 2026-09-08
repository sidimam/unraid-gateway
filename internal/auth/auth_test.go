package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func fakeUnraid(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" || r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Header.Get("x-api-key") {
		case "good":
			_, _ = w.Write([]byte(`{"data":{"me":{"id":"1","name":"root","roles":["admin"]}}}`))
		case "gqlerror":
			_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Unauthorized","extensions":{"code":"UNAUTHENTICATED"}}]}`))
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	}))
}

func TestValidatorAndStore(t *testing.T) {
	srv := fakeUnraid(t)
	defer srv.Close()
	v := NewUnraidValidator(srv.URL, "query { me { id name roles } }", false)
	id, err := v.Validate(context.Background(), "good")
	if err != nil || id.Name != "root" || len(id.Roles) != 1 {
		t.Fatalf("valid key: %+v %v", id, err)
	}
	for _, k := range []string{"bad", "gqlerror"} {
		if _, err := v.Validate(context.Background(), k); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("%s: expected ErrInvalidKey, got %v", k, err)
		}
	}

	store := NewStore(v, time.Hour, 2, time.Minute)
	sess, err := store.Login(context.Background(), "1.2.3.4", "good")
	if err != nil || sess.Token == "" {
		t.Fatalf("login: %v", err)
	}
	if got, ok := store.Lookup(sess.Token); !ok || got.APIKey != "good" {
		t.Fatal("lookup failed")
	}
	store.Logout(sess.Token)
	if _, ok := store.Lookup(sess.Token); ok {
		t.Fatal("session should be gone after logout")
	}
	// Two failures lock the IP; even a good key is then refused.
	for i := 0; i < 2; i++ {
		if _, err := store.Login(context.Background(), "9.9.9.9", "bad"); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	if _, err := store.Login(context.Background(), "9.9.9.9", "good"); !errors.Is(err, ErrLocked) {
		t.Fatalf("expected lockout, got %v", err)
	}
	if _, err := store.Login(context.Background(), "5.5.5.5", "good"); err != nil {
		t.Fatalf("other ip should not be locked: %v", err)
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.2")
	if ip := ClientIP(r, false); ip != "10.0.0.1" {
		t.Fatalf("untrusted proxy: %s", ip)
	}
	if ip := ClientIP(r, true); ip != "203.0.113.7" {
		t.Fatalf("trusted proxy: %s", ip)
	}
}

package server

import (
	"testing"
	"time"
)

func TestUTF8Header(t *testing.T) {
	ok := "Unraid Drive 1.5 (45) · iPhone"
	if got := utf8Header(ok); got != ok {
		t.Fatalf("valid UTF-8 changed: %q", got)
	}
	latin1 := "Unraid Drive 1.5 (45) \xb7 iPhone"
	if got := utf8Header(latin1); got != ok {
		t.Fatalf("Latin-1 not repaired: %q", got)
	}
}

func TestLoginCoalescer(t *testing.T) {
	var c loginCoalescer
	if !c.Allow("a", time.Minute) {
		t.Fatal("first sign-in must be logged")
	}
	if c.Allow("a", time.Minute) {
		t.Fatal("second sign-in inside the window must be skipped")
	}
	if !c.Allow("b", time.Minute) {
		t.Fatal("another device is independent")
	}
	c.last["a"] = time.Now().Add(-2 * time.Minute)
	if !c.Allow("a", time.Minute) {
		t.Fatal("after the window the sign-in is logged again")
	}
}

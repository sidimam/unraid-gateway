// Package access defines per-share permission levels and carries the policy of
// the authenticated principal through the request context.
package access

import "context"

// Level is what a principal may do inside a share.
type Level int

const (
	None Level = iota
	Read
	Write
)

func (l Level) String() string {
	switch l {
	case Read:
		return "ro"
	case Write:
		return "rw"
	default:
		return "none"
	}
}

// Policy answers "what may this principal do in share X".
type Policy interface {
	// Level returns the access level for a share name (the first path component under the data root).
	Level(share string) Level
	// Restricted is false for principals with full, mount-defined access (API key only).
	Restricted() bool
}

// Unrestricted grants everything the container mounts allow.
type Unrestricted struct{}

func (Unrestricted) Level(string) Level { return Write }
func (Unrestricted) Restricted() bool   { return false }

// Static is a fixed share → level map (case-insensitive share names).
type Static map[string]Level

func (s Static) Level(share string) Level {
	for k, v := range s {
		if equalFold(k, share) {
			return v
		}
	}
	return None
}
func (s Static) Restricted() bool { return true }

type ctxKey struct{}

// WithPolicy attaches a policy to the context.
func WithPolicy(ctx context.Context, p Policy) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// FromContext returns the request policy, or Unrestricted when none was set.
func FromContext(ctx context.Context) Policy {
	if p, ok := ctx.Value(ctxKey{}).(Policy); ok && p != nil {
		return p
	}
	return Unrestricted{}
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

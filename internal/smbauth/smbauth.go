// Package smbauth validates Unraid user credentials by opening an SMB2
// session against the Unraid server, exactly like a file browser would.
// Samba (not the gateway) is the authority on users and passwords.
package smbauth

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

// ErrInvalidCredentials is returned when Samba rejects the user/password.
var ErrInvalidCredentials = errors.New("invalid unraid username or password")

// Authenticator checks credentials against an SMB server (host:port).
type Authenticator struct {
	Addr    string
	Timeout time.Duration
}

// Check opens and closes an SMB session with the given credentials.
func (a *Authenticator) Check(ctx context.Context, user, password string) error {
	if strings.TrimSpace(user) == "" || password == "" {
		return ErrInvalidCredentials
	}
	timeout := a.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", a.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	d := &smb2.Dialer{Initiator: &smb2.NTLMInitiator{User: user, Password: password}}
	sess, err := d.DialContext(ctx, conn)
	if err != nil {
		// go-smb2 reports authentication failures as STATUS_LOGON_FAILURE responses.
		if strings.Contains(strings.ToUpper(err.Error()), "LOGON") || strings.Contains(strings.ToLower(err.Error()), "auth") {
			return ErrInvalidCredentials
		}
		return err
	}
	_ = sess.Logoff()
	return nil
}

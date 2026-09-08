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

// ErrGuest is returned when the credentials were accepted only as a guest
// (Samba "map to guest = bad user"), i.e. the user does not really exist.
var ErrGuest = errors.New("credentials accepted as guest only")

// Check opens an SMB session with the given credentials. When probeShare is
// not empty it additionally connects to that share: a share the real user may
// open but a guest may not, so a guest-mapped session is detected.
func (a *Authenticator) Check(ctx context.Context, user, password, probeShare string) error {
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
	defer func() { _ = sess.Logoff() }()
	if probeShare != "" {
		fs, err := sess.Mount(probeShare)
		if err != nil {
			up := strings.ToUpper(err.Error())
			if strings.Contains(up, "ACCESS_DENIED") || strings.Contains(up, "ACCESS DENIED") {
				return ErrGuest
			}
			return err
		}
		_ = fs.Umount()
	}
	return nil
}

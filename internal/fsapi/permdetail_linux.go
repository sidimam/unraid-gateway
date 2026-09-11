//go:build linux

package fsapi

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// ownerOf describes who owns a path, for permission-denied diagnostics.
func ownerOf(st os.FileInfo) string {
	sys, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return "unknown owner"
	}
	name := strconv.Itoa(int(sys.Uid))
	if u, err := user.LookupId(name); err == nil && u.Username != "" {
		name = u.Username + " (uid " + name + ")"
	} else {
		name = "uid " + name
	}
	return fmt.Sprintf("owner %s, gid %d", name, sys.Gid)
}

// gatewayIdentity names the account the gateway runs as.
func gatewayIdentity() string {
	uid := os.Getuid()
	if u, err := user.LookupId(strconv.Itoa(uid)); err == nil && u.Username != "" {
		return fmt.Sprintf("%s (uid %d)", u.Username, uid)
	}
	return fmt.Sprintf("uid %d", uid)
}

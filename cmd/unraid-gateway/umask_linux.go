//go:build linux

package main

import "syscall"

// Unraid's convention is that everything on the shares is created as
// nobody:users with mode 0777 (folders) / 0666 (files), so every Unraid user,
// SMB session and container can read and modify it (this is what
// Tools › New Permissions enforces). The gateway asks for those modes when it
// creates files and folders; a non-zero umask inherited from the container
// runtime would silently reduce them to 0755/0644 and lock other users out.
func init() { syscall.Umask(0) }

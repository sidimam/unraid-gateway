//go:build !linux

package fsapi

import (
	"fmt"
	"os"
)

func ownerOf(_ os.FileInfo) string { return "unknown owner" }

func gatewayIdentity() string { return fmt.Sprintf("uid %d", os.Getuid()) }

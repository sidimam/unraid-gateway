package fsapi

import "os"

// DirMode and FileMode are the modes the gateway asks for when it creates
// folders and files: Unraid's convention (what Tools › New Permissions sets),
// so that anything created through the app stays writable for every Unraid
// user over SMB and for other containers. main() sets umask 0 so these are
// applied as-is.
const (
	DirMode  os.FileMode = 0o777
	FileMode os.FileMode = 0o666
)

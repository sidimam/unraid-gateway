package server

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// fuseType returns the file system type of the mount holding path when it is a
// FUSE file system (Unraid's /mnt/user is "fuse.shfs"), "" otherwise or when
// /proc/self/mountinfo cannot be read (not Linux).
func fuseType(path string) string {
	return fuseTypeFrom("/proc/self/mountinfo", path)
}

func fuseTypeFrom(mountinfo, path string) string {
	f, err := os.Open(mountinfo)
	if err != nil {
		return ""
	}
	defer f.Close()
	dir := filepath.Clean(filepath.Dir(path))
	best, bestType := "", ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// id parent major:minor root mountpoint options [optional...] - fstype source superopts
		line := sc.Text()
		sep := strings.Index(line, " - ")
		if sep < 0 {
			continue
		}
		left, right := strings.Fields(line[:sep]), strings.Fields(line[sep+3:])
		if len(left) < 5 || len(right) < 1 {
			continue
		}
		mp := strings.ReplaceAll(left[4], `\040`, " ")
		if (dir == mp || strings.HasPrefix(dir, strings.TrimSuffix(mp, "/")+"/") || mp == "/") && len(mp) > len(best) {
			best, bestType = mp, right[0]
		}
	}
	if strings.HasPrefix(bestType, "fuse") {
		return bestType
	}
	return ""
}

//go:build linux || darwin || freebsd || openbsd || netbsd

package filesystem

import (
	"os"
	"os/user"
	"strconv"
	"sync"
	"syscall"
)

var (
	userCache  sync.Map
	groupCache sync.Map
)

// populateMetadata extracts system-level file ownership and allocation attributes.
func populateMetadata(fi os.FileInfo, entry *Entry) {
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return
	}

	entry.Blocks = int64(stat.Blocks)
	entry.Inode = uint64(stat.Ino)

	uid := strconv.Itoa(int(stat.Uid))
	if cached, ok := userCache.Load(uid); ok {
		entry.Owner = cached.(string)
	} else {
		name := uid
		if u, err := user.LookupId(uid); err == nil && u.Username != "" {
			name = u.Username
		}
		userCache.Store(uid, name)
		entry.Owner = name
	}

	gid := strconv.Itoa(int(stat.Gid))
	if cached, ok := groupCache.Load(gid); ok {
		entry.Group = cached.(string)
	} else {
		gname := gid
		if g, err := user.LookupGroupId(gid); err == nil && g.Name != "" {
			gname = g.Name
		}
		groupCache.Store(gid, gname)
		entry.Group = gname
	}
}

// formatPermissionsString formats standard 10-character Unix permissions (-rwxr-xr-x).
func formatPermissionsString(mode os.FileMode) string {
	var buf [10]byte

	switch {
	case mode.IsDir():
		buf[0] = 'd'
	case mode&os.ModeSymlink != 0:
		buf[0] = 'l'
	case mode&os.ModeNamedPipe != 0:
		buf[0] = 'p'
	case mode&os.ModeSocket != 0:
		buf[0] = 's'
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			buf[0] = 'c'
		} else {
			buf[0] = 'b'
		}
	default:
		buf[0] = '-'
	}

	const rwx = "rwxrwxrwx"
	for i := 0; i < 9; i++ {
		if mode&(1<<uint(8-i)) != 0 {
			buf[i+1] = rwx[i]
		} else {
			buf[i+1] = '-'
		}
	}

	// Handle setuid, setgid, sticky bit
	if mode&os.ModeSetuid != 0 {
		if buf[3] == 'x' {
			buf[3] = 's'
		} else {
			buf[3] = 'S'
		}
	}
	if mode&os.ModeSetgid != 0 {
		if buf[6] == 'x' {
			buf[6] = 's'
		} else {
			buf[6] = 'S'
		}
	}
	if mode&os.ModeSticky != 0 {
		if buf[9] == 'x' {
			buf[9] = 't'
		} else {
			buf[9] = 'T'
		}
	}

	return string(buf[:])
}

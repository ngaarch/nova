//go:build darwin

package filesystem

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

func populateDetailedPlatformStat(fi os.FileInfo, stat *DetailedStat) {
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok || s == nil {
		return
	}

	stat.OwnerUID = s.Uid
	stat.GroupGID = s.Gid
	stat.Device = uint64(s.Dev)
	stat.Inode = uint64(s.Ino)
	stat.HardLinks = uint64(s.Nlink)
	stat.IOBlockSize = int64(s.Blksize)
	stat.Blocks = int64(s.Blocks)

	stat.AccessTime = time.Unix(s.Atimespec.Sec, s.Atimespec.Nsec)
	stat.ModifyTime = time.Unix(s.Mtimespec.Sec, s.Mtimespec.Nsec)
	stat.ChangeTime = time.Unix(s.Ctimespec.Sec, s.Ctimespec.Nsec)
	stat.BirthTime = time.Unix(s.Birthtimespec.Sec, s.Birthtimespec.Nsec)

	uid := strconv.Itoa(int(s.Uid))
	if cached, ok := userCache.Load(uid); ok {
		stat.OwnerUser = cached.(string)
	} else {
		name := uid
		if u, err := user.LookupId(uid); err == nil && u.Username != "" {
			name = u.Username
		}
		userCache.Store(uid, name)
		stat.OwnerUser = name
	}

	gid := strconv.Itoa(int(s.Gid))
	if cached, ok := groupCache.Load(gid); ok {
		stat.GroupUser = cached.(string)
	} else {
		gname := gid
		if g, err := user.LookupGroupId(gid); err == nil && g.Name != "" {
			gname = g.Name
		}
		groupCache.Store(gid, gname)
		stat.GroupUser = gname
	}
}

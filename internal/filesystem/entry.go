package filesystem

import (
	"os"
	"time"

	"nova/internal/theme"
)

// Entry encapsulates detailed filesystem metadata for a file or directory.
type Entry struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	Size        int64            `json:"size"`
	Mode        os.FileMode      `json:"mode"`
	ModTime     time.Time        `json:"mod_time"`
	IsDir       bool             `json:"is_dir"`
	IsSymlink   bool             `json:"is_symlink"`
	IsBroken    bool             `json:"is_broken"`
	LinkTarget  string           `json:"link_target,omitempty"`
	Owner       string           `json:"owner,omitempty"`
	Group       string           `json:"group,omitempty"`
	Permissions string           `json:"permissions"`
	Blocks      int64            `json:"blocks,omitempty"`
	Inode       uint64           `json:"inode,omitempty"`
	EntityType  theme.EntityType `json:"entity_type"`
	GitStatus   string           `json:"git_status,omitempty"`
}

// ClassifyEntityType determines the theme.EntityType based on entry attributes and file permissions.
func ClassifyEntityType(mode os.FileMode, isSymlink, isBroken bool) theme.EntityType {
	if isBroken {
		return theme.TypeBrokenSymlink
	}
	if isSymlink {
		return theme.TypeSymlink
	}
	if mode.IsDir() {
		return theme.TypeDirectory
	}
	if mode&os.ModeNamedPipe != 0 {
		return theme.TypePipe
	}
	if mode&os.ModeSocket != 0 {
		return theme.TypeSocket
	}
	if mode&os.ModeDevice != 0 {
		return theme.TypeDevice
	}
	if mode&0o111 != 0 {
		return theme.TypeExecutable
	}
	return theme.TypeRegular
}

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nova/internal/git"
	"nova/internal/theme"
)

// DetailedStat contains complete low-level metadata for stat inspection.
type DetailedStat struct {
	Path        string           `json:"path"`
	Name        string           `json:"name"`
	Size        int64            `json:"size"`
	Mode        os.FileMode      `json:"mode"`
	ModeOctal   string           `json:"mode_octal"`
	ModeString  string           `json:"mode_string"`
	EntityType  theme.EntityType `json:"entity_type"`
	OwnerUID    uint32           `json:"owner_uid"`
	OwnerUser   string           `json:"owner_user"`
	GroupGID    uint32           `json:"group_gid"`
	GroupUser   string           `json:"group_user"`
	Device      uint64           `json:"device"`
	Inode       uint64           `json:"inode"`
	HardLinks   uint64           `json:"hard_links"`
	IOBlockSize int64            `json:"io_block_size"`
	Blocks      int64            `json:"blocks"`
	AccessTime  time.Time        `json:"access_time"`
	ModifyTime  time.Time        `json:"modify_time"`
	ChangeTime  time.Time        `json:"change_time"`
	BirthTime   time.Time        `json:"birth_time,omitempty"`
	IsSymlink   bool             `json:"is_symlink"`
	IsBroken    bool             `json:"is_broken"`
	LinkTarget  string           `json:"link_target,omitempty"`
	GitBranch   string           `json:"git_branch,omitempty"`
	GitStatus   string           `json:"git_status,omitempty"`
}

// GetDetailedStat returns rich filesystem status metadata for path.
func GetDetailedStat(path string) (*DetailedStat, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}

	mode := fi.Mode()
	isSymlink := mode&os.ModeSymlink != 0
	isBroken := false
	linkTarget := ""

	if isSymlink {
		if target, err := os.Readlink(path); err == nil {
			linkTarget = target
			if _, sErr := os.Stat(path); sErr != nil {
				isBroken = true
			}
		} else {
			isBroken = true
		}
	}

	stat := &DetailedStat{
		Path:        path,
		Name:        filepath.Base(path),
		Size:        fi.Size(),
		Mode:        mode,
		ModeOctal:   fmt.Sprintf("%04o", mode.Perm()),
		ModeString:  formatPermissionsString(mode),
		EntityType:  ClassifyEntityType(mode, isSymlink, isBroken),
		ModifyTime:  fi.ModTime(),
		AccessTime:  fi.ModTime(),
		ChangeTime:  fi.ModTime(),
		IsSymlink:   isSymlink,
		IsBroken:    isBroken,
		LinkTarget:  linkTarget,
		HardLinks:   1,
		IOBlockSize: 4096,
	}

	populateDetailedPlatformStat(fi, stat)

	// Populate Git repository metadata if within repository
	dir := path
	if !fi.IsDir() {
		dir = filepath.Dir(path)
	}
	if repo, err := git.GetRepoStatus(dir, 50*time.Millisecond); err == nil && repo != nil {
		stat.GitBranch = repo.Branch
		st := repo.GetStatus(path)
		if st != git.StatusClean {
			stat.GitStatus = string(st)
		}
	}

	return stat, nil
}

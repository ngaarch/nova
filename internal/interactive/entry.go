package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nova/internal/theme"
)

// Entry represents a filesystem entity in the interactive browser.
type Entry struct {
	Name          string
	Path          string
	Target        string
	IsDir         bool
	IsSymlink     bool
	IsBroken      bool
	IsExec        bool
	Size          int64
	Mode          os.FileMode
	ModTime       time.Time
	Extension     string
	Icon          string
}

// ReadDir reads the directory at path and returns a sorted slice of Entries.
func ReadDir(dirPath string, showHidden bool, unicodeSupported bool) ([]Entry, error) {
	cleanPath := filepath.Clean(dirPath)
	dirEntries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", cleanPath, err)
	}

	var entries []Entry
	for _, de := range dirEntries {
		name := de.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(cleanPath, name)
		fi, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		isSymlink := fi.Mode()&os.ModeSymlink != 0
		var isBroken bool
		var target string
		var isDir bool
		var size int64
		var isExec bool

		if isSymlink {
			t, rErr := os.Readlink(fullPath)
			if rErr == nil {
				target = t
			}
			// Resolve target metadata
			targetFi, statErr := os.Stat(fullPath)
			if statErr != nil {
				isBroken = true
			} else {
				isDir = targetFi.IsDir()
				size = targetFi.Size()
				isExec = targetFi.Mode()&0111 != 0
			}
		} else {
			isDir = fi.IsDir()
			size = fi.Size()
			isExec = fi.Mode()&0111 != 0
		}

		ext := strings.ToLower(filepath.Ext(name))
		entType := theme.TypeRegular
		if isBroken {
			entType = theme.TypeBrokenSymlink
		} else if isSymlink {
			entType = theme.TypeSymlink
		} else if isDir {
			entType = theme.TypeDirectory
		} else if isExec {
			entType = theme.TypeExecutable
		}
		icon := theme.LookupIcon(name, entType, unicodeSupported)

		entries = append(entries, Entry{
			Name:       name,
			Path:       fullPath,
			Target:     target,
			IsDir:      isDir,
			IsSymlink:  isSymlink,
			IsBroken:   isBroken,
			IsExec:     isExec,
			Size:       size,
			Mode:       fi.Mode(),
			ModTime:    fi.ModTime(),
			Extension:  ext,
			Icon:       icon,
		})
	}

	// Sort directories first, then alphabetically case-insensitive
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, nil
}

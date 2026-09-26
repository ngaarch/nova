package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadDir reads the directory entries at targetPath, extracting metadata and classifying entity types.
// If targetPath is a regular file, it returns a slice containing the entry for that file.
func ReadDir(targetPath string, showHidden bool, needDetails bool) ([]Entry, error) {
	fi, err := os.Lstat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", targetPath, err)
	}

	// Single file target
	if !fi.IsDir() {
		entry := buildEntry(filepath.Dir(targetPath), fi.Name(), fi, needDetails)
		return []Entry{entry}, nil
	}

	dirEntries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", targetPath, err)
	}

	var entries []Entry
	for _, de := range dirEntries {
		name := de.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(targetPath, name)
		fi, err := os.Lstat(fullPath)
		if err != nil {
			// File may have been removed or permission denied; continue with partial entry
			mode := de.Type()
			entries = append(entries, Entry{
				Name:        name,
				Path:        fullPath,
				Mode:        mode,
				IsDir:       de.IsDir(),
				Permissions: mode.String(),
				EntityType:  ClassifyEntityType(mode, mode&os.ModeSymlink != 0, false),
			})
			continue
		}

		entries = append(entries, buildEntry(targetPath, name, fi, needDetails))
	}

	return entries, nil
}

// buildEntry constructs an Entry from an os.FileInfo.
func buildEntry(dir, name string, fi os.FileInfo, needDetails bool) Entry {
	fullPath := filepath.Join(dir, name)
	mode := fi.Mode()
	isSymlink := mode&os.ModeSymlink != 0
	isBroken := false
	linkTarget := ""

	if isSymlink {
		if target, err := os.Readlink(fullPath); err == nil {
			linkTarget = target
			// Test if target exists
			if _, err := os.Stat(fullPath); err != nil {
				isBroken = true
			}
		} else {
			isBroken = true
		}
	}

	entry := Entry{
		Name:        name,
		Path:        fullPath,
		Size:        fi.Size(),
		Mode:        mode,
		ModTime:     fi.ModTime(),
		IsDir:       fi.IsDir(),
		IsSymlink:   isSymlink,
		IsBroken:    isBroken,
		LinkTarget:  linkTarget,
		Permissions: formatPermissionsString(mode),
		EntityType:  ClassifyEntityType(mode, isSymlink, isBroken),
	}

	if needDetails {
		populateMetadata(fi, &entry)
	}

	return entry
}

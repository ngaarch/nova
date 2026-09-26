package ls

import (
	"path/filepath"
	"sort"
	"strings"

	"nova/internal/filesystem"
)

// SortEntries sorts a slice of filesystem.Entry according to the requested criteria.
func SortEntries(entries []filesystem.Entry, sortBy string, reverse bool, dirsFirst bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]

		// 1. Dirs-first sorting
		if dirsFirst && a.IsDir != b.IsDir {
			return a.IsDir
		}

		// 2. Main sort key
		var less bool
		switch sortBy {
		case "size":
			if a.Size != b.Size {
				less = a.Size < b.Size
			} else {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		case "time":
			if !a.ModTime.Equal(b.ModTime) {
				less = a.ModTime.After(b.ModTime) // newer files first by default
			} else {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		case "ext":
			extA := strings.ToLower(filepath.Ext(a.Name))
			extB := strings.ToLower(filepath.Ext(b.Name))
			if extA != extB {
				less = extA < extB
			} else {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		default: // "name"
			nameA := strings.ToLower(a.Name)
			nameB := strings.ToLower(b.Name)
			if nameA != nameB {
				less = nameA < nameB
			} else {
				less = a.Name < b.Name
			}
		}

		if reverse {
			return !less
		}
		return less
	})
}

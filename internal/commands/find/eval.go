package find

import (
	"os"
	"path/filepath"
	"strings"

	"nova/internal/filesystem"
	"nova/internal/theme"
)

// Match evaluates whether a filesystem entry satisfies all predicates.
func Match(entry *filesystem.Entry, relPath string, depth int, p *Predicates) bool {
	if depth < p.MinDepth {
		return false
	}
	if p.MaxDepth >= 0 && depth > p.MaxDepth {
		return false
	}

	// Name glob match (-name)
	if p.NamePattern != "" {
		matched, err := filepath.Match(p.NamePattern, entry.Name)
		if err != nil || !matched {
			return false
		}
	}

	// Case-insensitive name glob match (-iname)
	if p.INamePattern != "" {
		matched, err := filepath.Match(p.INamePattern, strings.ToLower(entry.Name))
		if err != nil || !matched {
			return false
		}
	}

	// Path glob match (-path)
	if p.PathPattern != "" {
		matched, err := filepath.Match(p.PathPattern, relPath)
		if err != nil || !matched {
			return false
		}
	}

	// Entity type match (-type)
	if p.EntityType != "" {
		switch p.EntityType {
		case "f":
			if entry.IsDir || entry.IsSymlink {
				return false
			}
		case "d":
			if !entry.IsDir || entry.IsSymlink {
				return false
			}
		case "l":
			if !entry.IsSymlink {
				return false
			}
		case "s":
			if entry.EntityType != theme.TypeSocket {
				return false
			}
		case "p":
			if entry.EntityType != theme.TypePipe {
				return false
			}
		}
	}

	// Size predicates (-size)
	if p.HasMinSize && entry.Size <= p.MinSize {
		return false
	}
	if p.HasMaxSize && entry.Size >= p.MaxSize {
		return false
	}
	if p.HasExactSize && entry.Size != p.ExactSize {
		return false
	}

	// Modification time predicates (-mtime)
	if p.HasModAfter && !entry.ModTime.After(p.ModifiedAfter) {
		return false
	}
	if p.HasModBefore && !entry.ModTime.Before(p.ModifiedBefore) {
		return false
	}

	// Empty predicate (-empty)
	if p.EmptyOnly {
		if !entry.IsDir {
			if entry.Size != 0 {
				return false
			}
		} else {
			// Check if directory is empty
			f, err := os.Open(entry.Path)
			if err != nil {
				return false
			}
			names, _ := f.Readdirnames(1)
			f.Close()
			if len(names) > 0 {
				return false
			}
		}
	}

	return true
}

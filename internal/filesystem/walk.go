package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WalkOptions controls directory traversal behavior.
type WalkOptions struct {
	MaxDepth       int                                  // -1 for unlimited, 0 for root only
	IncludeHidden  bool                                 // whether to include dotfiles
	FollowSymlinks bool                                 // whether to follow directory symlinks
	DetectCycles   bool                                 // track visited directories to prevent infinite loops
	DirsOnly       bool                                 // only visit directories
	SortBy         string                               // "name", "size", "time", "ext", or "none"
	Reverse        bool                                 // reverse sort order
	DirsFirst      bool                                 // group directories before files
	NeedDetails    bool                                 // extract full metadata (ownership, blocks, inode)
	Prune          func(path string, entry *Entry) bool // return true to skip descending into directory
	OnError        func(path string, err error) error   // error callback, return nil to continue
}

// DefaultWalkOptions returns standard safe traversal options.
func DefaultWalkOptions() WalkOptions {
	return WalkOptions{
		MaxDepth:       -1,
		IncludeHidden:  false,
		FollowSymlinks: false,
		DetectCycles:   true,
		DirsOnly:       false,
		SortBy:         "name",
		Reverse:        false,
		DirsFirst:      true,
		NeedDetails:    false,
		Prune:          nil,
		OnError:        nil,
	}
}

// TreeNode represents a node in a hierarchical directory tree with aggregated metrics.
type TreeNode struct {
	Entry     Entry       `json:"entry"`
	Depth     int         `json:"depth"`
	Children  []*TreeNode `json:"children,omitempty"`
	FileCount int         `json:"file_count"`
	DirCount  int         `json:"dir_count"`
	TotalSize int64       `json:"total_size"`
	IsCycle   bool        `json:"is_cycle,omitempty"`
}

// Walk traverses directory hierarchy starting at root, invoking fn for each entry.
func Walk(root string, opts WalkOptions, fn func(path string, entry *Entry, depth int) error) error {
	root = filepath.Clean(root)
	fi, err := os.Lstat(root)
	if err != nil {
		if opts.OnError != nil {
			return opts.OnError(root, err)
		}
		return fmt.Errorf("stat %q: %w", root, err)
	}

	rootEntry := buildEntry(filepath.Dir(root), filepath.Base(root), fi, opts.NeedDetails)
	if opts.DirsOnly && !rootEntry.IsDir {
		return nil
	}

	if err := fn(root, &rootEntry, 0); err != nil {
		return err
	}

	if !rootEntry.IsDir {
		return nil
	}

	visited := make(map[string]bool)
	absRoot, _ := filepath.Abs(root)
	visited[absRoot] = true

	return walkRecursive(root, 1, opts, visited, fn)
}

func walkRecursive(dirPath string, depth int, opts WalkOptions, visited map[string]bool, fn func(path string, entry *Entry, depth int) error) error {
	if opts.MaxDepth >= 0 && depth > opts.MaxDepth {
		return nil
	}

	entries, err := ReadDir(dirPath, opts.IncludeHidden, opts.NeedDetails)
	if err != nil {
		if opts.OnError != nil {
			return opts.OnError(dirPath, err)
		}
		return nil // Continue traversal gracefully on unreadable dirs
	}

	sortEntries(entries, opts.SortBy, opts.Reverse, opts.DirsFirst)

	for i := range entries {
		entry := &entries[i]

		if opts.DirsOnly && !entry.IsDir {
			continue
		}

		if err := fn(entry.Path, entry, depth); err != nil {
			return err
		}

		if entry.IsDir {
			// Check prune
			if opts.Prune != nil && opts.Prune(entry.Path, entry) {
				continue
			}

			// Cycle detection
			targetPath := entry.Path
			if entry.IsSymlink {
				if !opts.FollowSymlinks {
					continue
				}
				resolved, err := filepath.EvalSymlinks(entry.Path)
				if err != nil {
					continue
				}
				targetPath = resolved
			}

			absPath, _ := filepath.Abs(targetPath)
			if opts.DetectCycles && visited[absPath] {
				continue // Avoid infinite recursion
			}

			visited[absPath] = true
			err := walkRecursive(entry.Path, depth+1, opts, visited, fn)
			delete(visited, absPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BuildTree constructs a hierarchical TreeNode rooted at targetPath with aggregated counts and sizes.
func BuildTree(root string, opts WalkOptions) (*TreeNode, error) {
	root = filepath.Clean(root)
	fi, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", root, err)
	}

	rootEntry := buildEntry(filepath.Dir(root), filepath.Base(root), fi, opts.NeedDetails)
	node := &TreeNode{
		Entry:     rootEntry,
		Depth:     0,
		FileCount: 0,
		DirCount:  0,
		TotalSize: rootEntry.Size,
	}

	if !rootEntry.IsDir {
		node.FileCount = 1
		return node, nil
	}

	node.DirCount = 1
	visited := make(map[string]bool)
	absRoot, _ := filepath.Abs(root)
	visited[absRoot] = true

	err = buildTreeRecursive(node, root, 1, opts, visited)
	return node, err
}

func buildTreeRecursive(parent *TreeNode, dirPath string, depth int, opts WalkOptions, visited map[string]bool) error {
	if opts.MaxDepth >= 0 && depth > opts.MaxDepth {
		return nil
	}

	entries, err := ReadDir(dirPath, opts.IncludeHidden, opts.NeedDetails)
	if err != nil {
		if opts.OnError != nil {
			return opts.OnError(dirPath, err)
		}
		return nil
	}

	sortEntries(entries, opts.SortBy, opts.Reverse, opts.DirsFirst)

	for i := range entries {
		entry := entries[i]
		if opts.DirsOnly && !entry.IsDir {
			continue
		}

		child := &TreeNode{
			Entry:     entry,
			Depth:     depth,
			TotalSize: entry.Size,
		}

		if entry.IsDir {
			child.DirCount = 1
			if opts.Prune == nil || !opts.Prune(entry.Path, &entry) {
				targetPath := entry.Path
				if entry.IsSymlink {
					if opts.FollowSymlinks {
						if resolved, err := filepath.EvalSymlinks(entry.Path); err == nil {
							targetPath = resolved
						}
					} else {
						// Don't descend into symlink dir if not following symlinks
						parent.Children = append(parent.Children, child)
						parent.DirCount += child.DirCount
						parent.TotalSize += child.TotalSize
						continue
					}
				}

				absPath, _ := filepath.Abs(targetPath)
				if opts.DetectCycles && visited[absPath] {
					child.IsCycle = true
				} else {
					visited[absPath] = true
					_ = buildTreeRecursive(child, entry.Path, depth+1, opts, visited)
					delete(visited, absPath)
				}
			}
			parent.DirCount += child.DirCount
			parent.FileCount += child.FileCount
			parent.TotalSize += child.TotalSize
		} else {
			child.FileCount = 1
			parent.FileCount++
			parent.TotalSize += child.TotalSize
		}

		parent.Children = append(parent.Children, child)
	}

	return nil
}

func sortEntries(entries []Entry, sortBy string, reverse bool, dirsFirst bool) {
	if sortBy == "none" {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if dirsFirst && entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		var less bool
		switch sortBy {
		case "size":
			less = entries[i].Size < entries[j].Size
		case "time":
			less = entries[i].ModTime.Before(entries[j].ModTime)
		case "ext":
			extI := strings.ToLower(filepath.Ext(entries[i].Name))
			extJ := strings.ToLower(filepath.Ext(entries[j].Name))
			if extI != extJ {
				less = extI < extJ
			} else {
				less = strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			}
		default: // "name"
			less = strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}

		if reverse {
			return !less
		}
		return less
	})
}

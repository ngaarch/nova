package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Common operational errors.
var (
	ErrRootProtected    = errors.New("refusing to remove root directory")
	ErrDestInsideSource = errors.New("cannot copy directory into itself")
	ErrAlreadyExists    = errors.New("destination already exists")
)

// CopyOptions controls file and directory copying.
type CopyOptions struct {
	PreserveMetadata    bool
	DereferenceSymlinks bool
	Recursive           bool
	NoClobber           bool
	Force               bool
	DryRun              bool
	Verbose             bool
	Confirm             func(target string) (bool, error)
	OnProgress          func(copiedBytes, totalBytes int64, currentFile string)
}

// MoveOptions controls file and directory movement.
type MoveOptions struct {
	NoClobber bool
	Force     bool
	DryRun    bool
	Verbose   bool
	Confirm   func(target string) (bool, error)
}

// RemoveOptions controls file and directory deletion.
type RemoveOptions struct {
	Recursive   bool
	Force       bool
	DryRun      bool
	Verbose     bool
	ProtectRoot bool
	Confirm     func(target string) (bool, error)
}

// MkdirOptions controls directory creation.
type MkdirOptions struct {
	Parents bool
	Mode    os.FileMode
	Verbose bool
	DryRun  bool
}

// CopyFile copies a single file or symlink from src to dst.
func CopyFile(src, dst string, opts CopyOptions) error {
	srcFi, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", src, err)
	}

	// Handle symlink
	if srcFi.Mode()&os.ModeSymlink != 0 && !opts.DereferenceSymlinks {
		target, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("readlink %q: %w", src, err)
		}

		if _, err := os.Lstat(dst); err == nil {
			if opts.NoClobber {
				return nil
			}
			if opts.Confirm != nil {
				ok, cErr := opts.Confirm(dst)
				if cErr != nil || !ok {
					return cErr
				}
			}
			if !opts.DryRun {
				_ = os.Remove(dst)
			}
		}

		if opts.DryRun {
			return nil
		}
		return os.Symlink(target, dst)
	}

	if srcFi.IsDir() {
		return fmt.Errorf("source %q is a directory (use -r / --recursive)", src)
	}

	// Check if destination exists
	if _, err := os.Lstat(dst); err == nil {
		if opts.NoClobber {
			return nil
		}
		if opts.Confirm != nil {
			ok, cErr := opts.Confirm(dst)
			if cErr != nil || !ok {
				return cErr
			}
		}
	}

	if opts.DryRun {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcFi.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer out.Close()

	buf := make([]byte, 64*1024)
	totalSize := srcFi.Size()
	var copied int64

	for {
		nr, rErr := in.Read(buf)
		if nr > 0 {
			nw, wErr := out.Write(buf[:nr])
			if nw > 0 {
				copied += int64(nw)
				if opts.OnProgress != nil {
					opts.OnProgress(copied, totalSize, src)
				}
			}
			if wErr != nil {
				return fmt.Errorf("write %q: %w", dst, wErr)
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return fmt.Errorf("read %q: %w", src, rErr)
		}
	}

	// Preserve metadata
	if opts.PreserveMetadata {
		_ = os.Chmod(dst, srcFi.Mode().Perm())
		_ = os.Chtimes(dst, time.Now(), srcFi.ModTime())
	}

	return nil
}

// CopyDir recursively copies directory src to dst.
func CopyDir(src, dst string, opts CopyOptions) error {
	srcFi, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", src, err)
	}
	if !srcFi.IsDir() {
		return CopyFile(src, dst, opts)
	}

	// Prevent copying a directory into itself
	absSrc, _ := filepath.Abs(src)
	absDst, _ := filepath.Abs(dst)
	if strings.HasPrefix(absDst, absSrc+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s inside %s", ErrDestInsideSource, dst, src)
	}

	if opts.DryRun {
		// Verify entries can be read
		_, err := os.ReadDir(src)
		return err
	}

	if err := os.MkdirAll(dst, srcFi.Mode().Perm()); err != nil {
		return fmt.Errorf("mkdir %q: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", src, err)
	}

	for _, entry := range entries {
		subSrc := filepath.Join(src, entry.Name())
		subDst := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(subSrc, subDst, opts); err != nil {
				return err
			}
		} else {
			if err := CopyFile(subSrc, subDst, opts); err != nil {
				return err
			}
		}
	}

	if opts.PreserveMetadata {
		_ = os.Chmod(dst, srcFi.Mode().Perm())
		_ = os.Chtimes(dst, time.Now(), srcFi.ModTime())
	}

	return nil
}

// Move moves a file or directory from src to dst.
func Move(src, dst string, opts MoveOptions) error {
	srcFi, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("cannot stat %q: %w", src, err)
	}

	// Guard against moving a directory inside itself
	if srcFi.IsDir() {
		absSrc, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		absDst, err := filepath.Abs(dst)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absSrc, absDst)
		if err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
			return ErrDestInsideSource
		}
	}

	if _, err := os.Lstat(dst); err == nil {
		if opts.NoClobber {
			return nil
		}
		if opts.Confirm != nil {
			ok, cErr := opts.Confirm(dst)
			if cErr != nil || !ok {
				return cErr
			}
		}
	}

	if opts.DryRun {
		return nil
	}

	// First try atomic rename
	err = os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Check if cross-device error
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) && errors.Is(linkErr.Err, syscall.EXDEV) {
		// Fallback across device boundary: Copy then Remove
		copyOpts := CopyOptions{
			PreserveMetadata:    true,
			DereferenceSymlinks: false,
			Recursive:           true,
			NoClobber:           opts.NoClobber,
			Force:               opts.Force,
		}
		if cErr := CopyDir(src, dst, copyOpts); cErr != nil {
			return fmt.Errorf("cross-device copy failed: %w", cErr)
		}
		rmOpts := RemoveOptions{
			Recursive:   true,
			Force:       true,
			ProtectRoot: true,
		}
		return Remove(src, rmOpts)
	}

	return fmt.Errorf("rename %q -> %q: %w", src, dst, err)
}

// ProtectRoot returns true if the specified path resolves to root, working dir, parent, or empty path.
func ProtectRoot(path string) bool {
	clean := filepath.Clean(path)
	if clean == "/" || clean == "." || clean == ".." || clean == "" {
		return true
	}
	abs, err := filepath.Abs(clean)
	if err == nil && abs == "/" {
		return true
	}
	return false
}

// Remove deletes a file, symlink, or directory.
func Remove(path string, opts RemoveOptions) error {
	cleanPath := filepath.Clean(path)

	// Root protection
	if opts.ProtectRoot {
		if ProtectRoot(cleanPath) {
			return fmt.Errorf("%w: %q", ErrRootProtected, path)
		}
	}

	fi, err := os.Lstat(cleanPath)
	if err != nil {
		if opts.Force && os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot remove %q: %w", path, err)
	}

	// Symlinks are removed as symlinks (never follow into target directory)
	if fi.Mode()&os.ModeSymlink != 0 {
		if opts.Confirm != nil {
			ok, cErr := opts.Confirm(cleanPath)
			if cErr != nil || !ok {
				return cErr
			}
		}
		if opts.DryRun {
			return nil
		}
		return os.Remove(cleanPath)
	}

	if fi.IsDir() {
		if !opts.Recursive {
			return fmt.Errorf("cannot remove %q: Is a directory (use -r / -R to remove recursively)", path)
		}
		if opts.Confirm != nil {
			ok, cErr := opts.Confirm(cleanPath)
			if cErr != nil || !ok {
				return cErr
			}
		}
		if opts.DryRun {
			return nil
		}
		return os.RemoveAll(cleanPath)
	}

	if opts.Confirm != nil {
		ok, cErr := opts.Confirm(cleanPath)
		if cErr != nil || !ok {
			return cErr
		}
	}

	if opts.DryRun {
		return nil
	}
	return os.Remove(cleanPath)
}

// MakeDir creates a directory at path.
func MakeDir(path string, opts MkdirOptions) error {
	mode := opts.Mode
	if mode == 0 {
		mode = 0755
	}

	if opts.DryRun {
		return nil
	}

	if opts.Parents {
		return os.MkdirAll(path, mode)
	}
	return os.Mkdir(path, mode)
}

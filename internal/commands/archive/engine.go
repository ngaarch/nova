package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ArchiveEntry represents metadata of an archived item.
type ArchiveEntry struct {
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	CompressedSize   int64     `json:"compressed_size"`
	IsDir            bool      `json:"is_dir"`
	ModTime          time.Time `json:"mod_time"`
	Mode             string    `json:"mode"`
	CompressionRatio float64   `json:"compression_ratio"`
}

// PackZip creates a ZIP archive from the specified target files and directories.
func PackZip(zipPath string, targets []string) (int, int64, error) {
	out, err := os.Create(zipPath)
	if err != nil {
		return 0, 0, err
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	count := 0
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if err := addFileToZip(w, target, filepath.Base(target)); err == nil {
				count++
			}
			continue
		}

		baseDir := filepath.Dir(target)
		err = filepath.Walk(target, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			rel, err := filepath.Rel(baseDir, path)
			if err != nil {
				rel = path
			}
			if fi.IsDir() {
				rel += "/"
			}
			header, err := zip.FileInfoHeader(fi)
			if err != nil {
				return nil
			}
			header.Name = filepath.ToSlash(rel)
			header.Method = zip.Deflate

			writer, err := w.CreateHeader(header)
			if err != nil {
				return nil
			}

			if !fi.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return nil
				}
				defer file.Close()
				_, _ = io.Copy(writer, file)
				count++
			}
			return nil
		})
		if err != nil {
			return count, 0, err
		}
	}

	if err := w.Close(); err != nil {
		return count, 0, err
	}

	fi, err := os.Stat(zipPath)
	if err != nil {
		return count, 0, err
	}

	return count, fi.Size(), nil
}

func addFileToZip(w *zip.Writer, filePath, zipName string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(zipName)
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, f)
	return err
}

// UnpackZip extracts all items from a ZIP archive with zip-slip path validation.
func UnpackZip(zipPath, destDir string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, err
	}

	count := 0
	for _, f := range r.File {
		// Zip slip attack guard
		cleaned := filepath.Clean(f.Name)
		if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
			continue
		}
		targetPath := filepath.Join(destDir, cleaned)
		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) && targetPath != destDir {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			continue
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err == nil {
			count++
		}
	}

	return count, nil
}

// ListZip inspects and returns metadata for all entries in a ZIP file.
func ListZip(zipPath string) ([]ArchiveEntry, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var entries []ArchiveEntry
	for _, f := range r.File {
		var ratio float64
		if f.UncompressedSize64 > 0 {
			ratio = (1.0 - float64(f.CompressedSize64)/float64(f.UncompressedSize64)) * 100.0
		}
		entries = append(entries, ArchiveEntry{
			Name:             f.Name,
			Size:             int64(f.UncompressedSize64),
			CompressedSize:   int64(f.CompressedSize64),
			IsDir:            f.FileInfo().IsDir(),
			ModTime:          f.Modified,
			Mode:             f.Mode().String(),
			CompressionRatio: ratio,
		})
	}

	return entries, nil
}

// PackTarGz creates a .tar.gz archive.
func PackTarGz(tarPath string, targets []string) (int, int64, error) {
	out, err := os.Create(tarPath)
	if err != nil {
		return 0, 0, err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	count := 0
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if err := addFileToTar(tw, target, filepath.Base(target)); err == nil {
				count++
			}
			continue
		}

		baseDir := filepath.Dir(target)
		_ = filepath.Walk(target, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			rel, _ := filepath.Rel(baseDir, path)
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return nil
			}
			header.Name = filepath.ToSlash(rel)
			if fi.IsDir() {
				header.Name += "/"
			}

			if err := tw.WriteHeader(header); err != nil {
				return nil
			}

			if !fi.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return nil
				}
				defer file.Close()
				_, _ = io.Copy(tw, file)
				count++
			}
			return nil
		})
	}

	_ = tw.Close()
	_ = gw.Close()

	fi, err := os.Stat(tarPath)
	if err != nil {
		return count, 0, err
	}

	return count, fi.Size(), nil
}

func addFileToTar(tw *tar.Writer, filePath, tarName string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(tarName)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, f)
	return err
}

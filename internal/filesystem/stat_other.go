//go:build !linux && !darwin

package filesystem

import "os"

func populateDetailedPlatformStat(fi os.FileInfo, stat *DetailedStat) {
	// Fallback platform stat
}

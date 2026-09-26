//go:build windows

package filesystem

import (
	"os"
)

func populateMetadata(fi os.FileInfo, entry *Entry) {
	entry.Owner = "user"
	entry.Group = "group"
}

func formatPermissionsString(mode os.FileMode) string {
	return mode.String()
}

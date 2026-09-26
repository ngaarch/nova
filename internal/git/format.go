package git

import (
	"fmt"

	"nova/internal/terminal"
	"nova/internal/theme"
)

// FormatStatusBadge returns a colored string badge for a FileStatus.
func FormatStatusBadge(status FileStatus, th *theme.Theme, profile terminal.ColorProfile) string {
	if status == StatusClean {
		return " "
	}

	role := theme.RoleMuted
	switch status {
	case StatusModified:
		role = theme.RoleWarning
	case StatusAdded:
		role = theme.RoleSuccess
	case StatusUntracked:
		role = theme.RoleAccent
	case StatusDeleted, StatusConflict:
		role = theme.RoleError
	case StatusRenamed:
		role = theme.RoleInfo
	case StatusIgnored:
		role = theme.RoleMuted
	}

	return th.Format(role, string(status), profile)
}

// FormatBranch formats the Git branch with an icon or prefix.
func FormatBranch(branch string, isDetached bool, unicode bool, th *theme.Theme, profile terminal.ColorProfile) string {
	if branch == "" {
		return ""
	}

	prefix := "git:("
	suffix := ")"
	if unicode {
		prefix = "⎇ "
		suffix = ""
	}

	if isDetached {
		prefix += "detached:"
	}

	text := fmt.Sprintf("%s%s%s", prefix, branch, suffix)
	return th.Format(theme.RoleAccent, text, profile)
}

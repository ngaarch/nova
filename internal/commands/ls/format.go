package ls

import (
	"fmt"
	"strings"
	"time"

	"nova/internal/filesystem"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// FormatSize converts a byte count into a human-readable string (B, KB, MB, GB, TB) or raw number string.
func FormatSize(size int64, human bool) string {
	if !human {
		return fmt.Sprintf("%d", size)
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	val := float64(size) / float64(div)
	if val < 10 {
		return fmt.Sprintf("%.1f %s", val, units[exp])
	}
	return fmt.Sprintf("%.0f %s", val, units[exp])
}

// FormatTime formats timestamps conforming to modern Unix listing standards.
func FormatTime(t time.Time) string {
	now := time.Now()
	sixMonths := 180 * 24 * time.Hour

	if now.Sub(t) > sixMonths || t.After(now) {
		return t.Format("Jan 02  2006")
	}
	return t.Format("Jan 02 15:04")
}

// FormatPermissions formats a 10-char permission string with thematic colors for read, write, exec.
func FormatPermissions(perm string, th *theme.Theme, prof terminal.ColorProfile) string {
	if prof == terminal.ColorNone || len(perm) < 10 {
		return perm
	}

	var b strings.Builder
	// Entity type char (d, l, -, etc.)
	switch perm[0] {
	case 'd':
		b.WriteString(th.Format(theme.RoleDirectory, "d", prof))
	case 'l':
		b.WriteString(th.Format(theme.RoleSymlink, "l", prof))
	default:
		b.WriteByte(perm[0])
	}

	for i := 1; i < 10; i++ {
		ch := perm[i]
		switch ch {
		case 'r':
			b.WriteString(th.Format(theme.RolePermRead, "r", prof))
		case 'w':
			b.WriteString(th.Format(theme.RolePermWrite, "w", prof))
		case 'x', 's', 't':
			b.WriteString(th.Format(theme.RolePermExec, string(ch), prof))
		default:
			b.WriteString(th.Format(theme.RoleMuted, "-", prof))
		}
	}
	return b.String()
}

// FormatName formats the entry filename with appropriate theme role and prepended icon.
func FormatName(entry filesystem.Entry, icon string, th *theme.Theme, prof terminal.ColorProfile) string {
	var role theme.Role

	switch entry.EntityType {
	case theme.TypeDirectory:
		role = theme.RoleDirectory
	case theme.TypeSymlink:
		role = theme.RoleSymlink
	case theme.TypeBrokenSymlink:
		role = theme.RoleBrokenSymlink
	case theme.TypeExecutable:
		role = theme.RoleExecutable
	case theme.TypePipe:
		role = theme.RolePipe
	case theme.TypeSocket:
		role = theme.RoleSocket
	case theme.TypeDevice:
		role = theme.RoleDevice
	default:
		role = theme.RoleRegularFile
	}

	name := entry.Name
	if strings.HasPrefix(name, ".") && role == theme.RoleRegularFile {
		role = theme.RoleHidden
	}

	styledName := th.Format(role, name, prof)
	if icon != "" {
		styledName = icon + styledName
	}

	if entry.IsSymlink {
		arrow := th.Format(theme.RoleMuted, " -> ", prof)
		targetRole := theme.RoleSymlink
		if entry.IsBroken {
			targetRole = theme.RoleBrokenSymlink
		}
		targetStyled := th.Format(targetRole, entry.LinkTarget, prof)
		styledName += arrow + targetStyled
	}

	return styledName
}

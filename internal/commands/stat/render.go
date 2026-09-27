package stat

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/git"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// RenderHuman prints a rich structured metadata card for the file.
func RenderHuman(ctx *command.Context, st *filesystem.DetailedStat) {
	th := ctx.Theme
	profile := ctx.Caps.ColorProfile

	bTopL, bTopR, bBotL, bBotR := "╭", "╮", "╰", "╯"
	bHor, bVert, bSepL, bSepR := "─", "│", "├", "┤"
	if !ctx.Caps.UnicodeSupported {
		bTopL, bTopR, bBotL, bBotR = "+", "+", "+", "+"
		bHor, bVert, bSepL, bSepR = "-", "|", "+", "+"
	}

	cardW := 68
	if ctx.Caps.Width > 20 && ctx.Caps.Width < cardW {
		cardW = ctx.Caps.Width - 2
	}
	bFmt := func(s string) string {
		return th.Format(theme.RoleMuted, s, profile)
	}

	// Header line
	icon := ""
	if profile != terminal.ColorNone {
		icon = theme.LookupIcon(st.Name, st.EntityType, ctx.Caps.UnicodeSupported)
		if icon != "" {
			icon += " "
		}
	}

	titleRole := theme.RoleRegularFile
	typeDesc := "regular file"
	if st.Mode.IsDir() {
		titleRole = theme.RoleDirectory
		typeDesc = "directory"
	} else if st.IsBroken {
		titleRole = theme.RoleBrokenSymlink
		typeDesc = "broken symbolic link"
	} else if st.IsSymlink {
		titleRole = theme.RoleSymlink
		typeDesc = "symbolic link"
	} else if st.EntityType == theme.TypeExecutable {
		titleRole = theme.RoleExecutable
		typeDesc = "executable"
	} else if st.EntityType == theme.TypeSocket {
		typeDesc = "socket"
	} else if st.EntityType == theme.TypePipe {
		typeDesc = "named pipe (FIFO)"
	}

	labelFile := th.Format(theme.RoleAccent, "File: ", profile)
	nameStyled := th.Format(titleRole, st.Path, profile)
	ctx.Printer.Println(bFmt(bTopL+bHor+"[ ") + labelFile + icon + nameStyled + bFmt(" ]"+strings.Repeat(bHor, 8)+bTopR))

	vPre := bFmt(bVert) + "  "

	if st.IsSymlink {
		labelTarget := th.Format(theme.RoleAccent, "Target: ", profile)
		targetStr := st.LinkTarget
		if st.IsBroken {
			targetStr += " [broken link]"
		}
		ctx.Printer.Println(vPre + labelTarget + th.Format(theme.RoleSymlink, targetStr, profile))
	}

	// Size and allocation block
	sizeHuman := renderer.FormatSize(st.Size, true)
	lineSize := fmt.Sprintf("Size: %-15s Blocks: %-10d IO Block: %-6d %s",
		fmt.Sprintf("%d (%s)", st.Size, sizeHuman), st.Blocks, st.IOBlockSize, typeDesc)
	ctx.Printer.Println(vPre + th.Format(theme.RoleMuted, lineSize, profile))

	// Device and Inodes
	lineDev := fmt.Sprintf("Device: %-15d Inode: %-11d Links: %-5d",
		st.Device, st.Inode, st.HardLinks)
	ctx.Printer.Println(vPre + th.Format(theme.RoleMuted, lineDev, profile))

	// Access and Ownership
	ownerStr := fmt.Sprintf("(%d/%s)", st.OwnerUID, st.OwnerUser)
	groupStr := fmt.Sprintf("(%d/%s)", st.GroupGID, st.GroupUser)
	lineAccess := fmt.Sprintf("Access: (%s/%s)   Uid: %-16s Gid: %s",
		th.Format(theme.RoleAccent, st.ModeOctal, profile), st.ModeString, ownerStr, groupStr)
	ctx.Printer.Println(vPre + lineAccess)

	ctx.Printer.Println(bFmt(bSepL + strings.Repeat(bHor, cardW) + bSepR))

	// Timestamps
	timeFmt := "2006-01-02 15:04:05.000000000 -0700"
	relMod := renderer.FormatRelativeTime(st.ModifyTime)
	ctx.Printer.Println(vPre + fmt.Sprintf("Access: %s", st.AccessTime.Format(timeFmt)))
	ctx.Printer.Println(vPre + fmt.Sprintf("Modify: %s (%s)", st.ModifyTime.Format(timeFmt), th.Format(theme.RoleSuccess, relMod, profile)))
	ctx.Printer.Println(vPre + fmt.Sprintf("Change: %s", st.ChangeTime.Format(timeFmt)))
	if !st.BirthTime.IsZero() {
		ctx.Printer.Println(vPre + fmt.Sprintf(" Birth: %s", st.BirthTime.Format(timeFmt)))
	}

	// Git status
	if st.GitBranch != "" {
		ctx.Printer.Println(bFmt(bSepL + strings.Repeat(bHor, cardW) + bSepR))
		branchFmt := git.FormatBranch(st.GitBranch, false, ctx.Caps.UnicodeSupported, th, profile)
		gitStatusDesc := "clean"
		if st.GitStatus != "" {
			badge := git.FormatStatusBadge(git.FileStatus(st.GitStatus), th, profile)
			gitStatusDesc = fmt.Sprintf("%s (%s)", badge, st.GitStatus)
		}
		labelGit := th.Format(theme.RoleAccent, "Git: ", profile)
		ctx.Printer.Println(vPre + fmt.Sprintf("%s%s   Status: %s", labelGit, branchFmt, gitStatusDesc))
	}

	ctx.Printer.Println(bFmt(bBotL + strings.Repeat(bHor, cardW) + bBotR))
}

// RenderPlain prints tab-separated key-value pairs suitable for scripts.
func RenderPlain(ctx *command.Context, st *filesystem.DetailedStat) {
	timeFmt := "2006-01-02 15:04:05"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("File:\t%s\n", st.Path))
	b.WriteString(fmt.Sprintf("Size:\t%d\n", st.Size))
	b.WriteString(fmt.Sprintf("Blocks:\t%d\n", st.Blocks))
	b.WriteString(fmt.Sprintf("IO Block:\t%d\n", st.IOBlockSize))
	b.WriteString(fmt.Sprintf("Device:\t%d\n", st.Device))
	b.WriteString(fmt.Sprintf("Inode:\t%d\n", st.Inode))
	b.WriteString(fmt.Sprintf("Links:\t%d\n", st.HardLinks))
	b.WriteString(fmt.Sprintf("Access:\t%s (%s)\n", st.ModeOctal, st.ModeString))
	b.WriteString(fmt.Sprintf("Uid:\t%d (%s)\n", st.OwnerUID, st.OwnerUser))
	b.WriteString(fmt.Sprintf("Gid:\t%d (%s)\n", st.GroupGID, st.GroupUser))
	b.WriteString(fmt.Sprintf("AccessTime:\t%s\n", st.AccessTime.Format(timeFmt)))
	b.WriteString(fmt.Sprintf("ModifyTime:\t%s\n", st.ModifyTime.Format(timeFmt)))
	b.WriteString(fmt.Sprintf("ChangeTime:\t%s\n", st.ChangeTime.Format(timeFmt)))
	if st.IsSymlink {
		b.WriteString(fmt.Sprintf("LinkTarget:\t%s\n", st.LinkTarget))
	}
	if st.GitBranch != "" {
		b.WriteString(fmt.Sprintf("GitBranch:\t%s\n", st.GitBranch))
		stDesc := st.GitStatus
		if stDesc == "" {
			stDesc = "clean"
		}
		b.WriteString(fmt.Sprintf("GitStatus:\t%s\n", stDesc))
	}
	ctx.Printer.Print(b.String())
}

// RenderJSON serializes DetailedStat to structured JSON.
func RenderJSON(ctx *command.Context, st *filesystem.DetailedStat) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

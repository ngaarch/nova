package theme

// Role represents a semantic categorization for syntax highlighting and file display.
type Role string

const (
	RoleDefault       Role = "default"
	RoleDirectory     Role = "directory"
	RoleRegularFile   Role = "regular_file"
	RoleExecutable    Role = "executable"
	RoleSymlink       Role = "symlink"
	RoleBrokenSymlink Role = "broken_symlink"
	RolePipe          Role = "pipe"
	RoleSocket        Role = "socket"
	RoleDevice        Role = "device"
	RoleArchive       Role = "archive"
	RoleCode          Role = "code"
	RoleDocument      Role = "document"
	RoleImage         Role = "image"
	RoleAudio         Role = "audio"
	RoleVideo         Role = "video"
	RoleHidden        Role = "hidden"

	// Status & notification roles
	RoleSuccess   Role = "success"
	RoleWarning   Role = "warning"
	RoleError     Role = "error"
	RoleInfo      Role = "info"
	RoleMuted     Role = "muted"
	RoleSelection Role = "selection"
	RoleAccent    Role = "accent"

	// Metadata attributes
	RoleSize      Role = "size"
	RoleDate      Role = "date"
	RoleUser      Role = "user"
	RoleGroup     Role = "group"
	RolePermRead  Role = "perm_read"
	RolePermWrite Role = "perm_write"
	RolePermExec  Role = "perm_exec"
)

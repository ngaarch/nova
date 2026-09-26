package theme

import (
	"path/filepath"
	"strings"
)

// IconPair defines an icon with a rich Unicode glyph and an ASCII fallback.
type IconPair struct {
	Unicode string
	ASCII   string
}

// Icon entries for special filesystem entities
var (
	IconDirectory     = IconPair{Unicode: "📁 ", ASCII: "/ "}
	IconSymlink       = IconPair{Unicode: "🔗 ", ASCII: "@ "}
	IconBrokenSymlink = IconPair{Unicode: "❌ ", ASCII: "! "}
	IconPipe          = IconPair{Unicode: "🚰 ", ASCII: "| "}
	IconSocket        = IconPair{Unicode: "🔌 ", ASCII: "= "}
	IconDevice        = IconPair{Unicode: "💻 ", ASCII: "% "}
	IconExecutable    = IconPair{Unicode: "⚡ ", ASCII: "* "}
	IconDefaultFile   = IconPair{Unicode: "📄 ", ASCII: "  "}
)

// Extension-specific icons
var extIcons = map[string]IconPair{
	// Code
	".go":   {Unicode: "🐹 ", ASCII: "  "},
	".py":   {Unicode: "🐍 ", ASCII: "  "},
	".rs":   {Unicode: "🦀 ", ASCII: "  "},
	".js":   {Unicode: "📜 ", ASCII: "  "},
	".jsx":  {Unicode: "📜 ", ASCII: "  "},
	".ts":   {Unicode: "📘 ", ASCII: "  "},
	".tsx":  {Unicode: "📘 ", ASCII: "  "},
	".c":    {Unicode: "⚙️ ", ASCII: "  "},
	".h":    {Unicode: "⚙️ ", ASCII: "  "},
	".cpp":  {Unicode: "⚙️ ", ASCII: "  "},
	".hpp":  {Unicode: "⚙️ ", ASCII: "  "},
	".sh":   {Unicode: "🐚 ", ASCII: "  "},
	".bash": {Unicode: "🐚 ", ASCII: "  "},
	".zsh":  {Unicode: "🐚 ", ASCII: "  "},
	".html": {Unicode: "🌐 ", ASCII: "  "},
	".css":  {Unicode: "🎨 ", ASCII: "  "},
	".java": {Unicode: "☕ ", ASCII: "  "},
	".rb":   {Unicode: "💎 ", ASCII: "  "},
	".php":  {Unicode: "🐘 ", ASCII: "  "},

	// Data & Config
	".json": {Unicode: "🔧 ", ASCII: "  "},
	".yaml": {Unicode: "🔧 ", ASCII: "  "},
	".yml":  {Unicode: "🔧 ", ASCII: "  "},
	".toml": {Unicode: "🔧 ", ASCII: "  "},
	".xml":  {Unicode: "🔧 ", ASCII: "  "},
	".sql":  {Unicode: "🗄️ ", ASCII: "  "},

	// Archives
	".zip": {Unicode: "📦 ", ASCII: "  "},
	".tar": {Unicode: "📦 ", ASCII: "  "},
	".gz":  {Unicode: "📦 ", ASCII: "  "},
	".bz2": {Unicode: "📦 ", ASCII: "  "},
	".xz":  {Unicode: "📦 ", ASCII: "  "},
	".7z":  {Unicode: "📦 ", ASCII: "  "},
	".rar": {Unicode: "📦 ", ASCII: "  "},

	// Documents
	".md":   {Unicode: "📝 ", ASCII: "  "},
	".txt":  {Unicode: "📄 ", ASCII: "  "},
	".pdf":  {Unicode: "📕 ", ASCII: "  "},
	".doc":  {Unicode: "📘 ", ASCII: "  "},
	".docx": {Unicode: "📘 ", ASCII: "  "},

	// Media
	".png":  {Unicode: "🖼️ ", ASCII: "  "},
	".jpg":  {Unicode: "🖼️ ", ASCII: "  "},
	".jpeg": {Unicode: "🖼️ ", ASCII: "  "},
	".gif":  {Unicode: "🖼️ ", ASCII: "  "},
	".webp": {Unicode: "🖼️ ", ASCII: "  "},
	".svg":  {Unicode: "📐 ", ASCII: "  "},

	".mp3":  {Unicode: "🎵 ", ASCII: "  "},
	".wav":  {Unicode: "🎵 ", ASCII: "  "},
	".flac": {Unicode: "🎵 ", ASCII: "  "},
	".ogg":  {Unicode: "🎵 ", ASCII: "  "},

	".mp4":  {Unicode: "🎬 ", ASCII: "  "},
	".mkv":  {Unicode: "🎬 ", ASCII: "  "},
	".webm": {Unicode: "🎬 ", ASCII: "  "},
	".mov":  {Unicode: "🎬 ", ASCII: "  "},
}

// EntityType describes the filesystem nature of an item.
type EntityType int

const (
	TypeRegular EntityType = iota
	TypeDirectory
	TypeSymlink
	TypeBrokenSymlink
	TypePipe
	TypeSocket
	TypeDevice
	TypeExecutable
)

// LookupIcon returns the appropriate icon string according to the entity metadata and Unicode capability.
func LookupIcon(name string, entityType EntityType, unicodeEnabled bool) string {
	var pair IconPair

	switch entityType {
	case TypeDirectory:
		pair = IconDirectory
	case TypeSymlink:
		pair = IconSymlink
	case TypeBrokenSymlink:
		pair = IconBrokenSymlink
	case TypePipe:
		pair = IconPipe
	case TypeSocket:
		pair = IconSocket
	case TypeDevice:
		pair = IconDevice
	case TypeExecutable:
		pair = IconExecutable
	default:
		ext := strings.ToLower(filepath.Ext(name))
		if icon, ok := extIcons[ext]; ok {
			pair = icon
		} else {
			pair = IconDefaultFile
		}
	}

	if unicodeEnabled {
		return pair.Unicode
	}
	return pair.ASCII
}

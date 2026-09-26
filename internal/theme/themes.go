package theme

import (
	"strings"

	"nova/internal/terminal"
)

// Theme encapsulates a mapped collection of visual styles for semantic roles.
type Theme struct {
	Name   string
	Styles map[Role]Style
}

// Style returns the Style assigned to a role, or an empty default style if unassigned.
func (t *Theme) Style(role Role) Style {
	if t != nil && t.Styles != nil {
		if s, ok := t.Styles[role]; ok {
			return s
		}
	}
	return Style{}
}

// Format formats the provided text with the theme style corresponding to the role.
func (t *Theme) Format(role Role, text string, profile terminal.ColorProfile) string {
	return t.Style(role).Format(text, profile)
}

func rgbPtr(c RGB) *RGB {
	return &c
}

// Built-in theme definitions
var (
	ThemeDefault = &Theme{
		Name: "default",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x61AFEF)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xABB2BF))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x98C379)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x56B6C2)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xE06C75)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xD19A66))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xC678DD)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xE5C07B))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xE06C75))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x61AFEF))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xE5C07B))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xC678DD))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x56B6C2))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x98C379))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x5C6370))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x98C379)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xE5C07B)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xE06C75)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x61AFEF))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x5C6370))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xFFFFFF)), Bg: rgbPtr(HexRGB(0x3E4451)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0x61AFEF)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x98C379))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x56B6C2))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xE5C07B))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xD19A66))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xE5C07B))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xE06C75))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x98C379)), Bold: true},
		},
	}

	ThemeMinimal = &Theme{
		Name: "minimal",
		Styles: map[Role]Style{
			RoleDirectory:     {Bold: true},
			RoleExecutable:    {Bold: true},
			RoleSymlink:       {Italic: true},
			RoleBrokenSymlink: {Underline: true},
			RoleHidden:        {Dim: true},
			RoleMuted:         {Dim: true},
			RoleError:         {Bold: true, Underline: true},
			RoleWarning:       {Bold: true},
			RoleSuccess:       {Bold: true},
			RoleSelection:     {Underline: true, Bold: true},
		},
	}

	ThemeMono = &Theme{
		Name:   "mono",
		Styles: map[Role]Style{},
	}

	ThemeNord = &Theme{
		Name: "nord",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x88C0D0)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xD8DEE9))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0xA3BE8C)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x81A1C1)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xBF616A)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xD08770))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xB48EAD)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xEBCB8B))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xD08770))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x88C0D0))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xE5E9F0))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xB48EAD))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x81A1C1))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0xA3BE8C))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x4C566A))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0xA3BE8C)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xEBCB8B)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xBF616A)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x88C0D0))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x4C566A))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xECEFF4)), Bg: rgbPtr(HexRGB(0x434C5E)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0x88C0D0)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0xA3BE8C))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x81A1C1))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xEBCB8B))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xD08770))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xEBCB8B))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xBF616A))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0xA3BE8C)), Bold: true},
		},
	}

	ThemeDracula = &Theme{
		Name: "dracula",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0xBD93F9)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xF8F8F2))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x50FA7B)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x8BE9FD)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xFF5555)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFFB86C))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xFF79C6)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xF1FA8C))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xFF79C6))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x8BE9FD))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xF8F8F2))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xBD93F9))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0xFFB86C))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x50FA7B))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x6272A4))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x50FA7B)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xF1FA8C)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xFF5555)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x8BE9FD))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x6272A4))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xF8F8F2)), Bg: rgbPtr(HexRGB(0x44475A)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0xBD93F9)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x50FA7B))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x8BE9FD))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xF1FA8C))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFFB86C))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xF1FA8C))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xFF5555))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x50FA7B)), Bold: true},
		},
	}

	ThemeNeon = &Theme{
		Name: "neon",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x00FFFF)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xFFFFFF))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x00E5FF)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xFF0033)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFF9900))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xFF007F)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xFFF700))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xFF007F))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x00FFFF))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xFFF700))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0x9D00FF))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x00FFFF))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x39FF14))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x555555))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xFFF700)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xFF0033)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x00FFFF))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x555555))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0x000000)), Bg: rgbPtr(HexRGB(0x00FFFF)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0x00FFFF)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x39FF14))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x00FFFF))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xFFF700))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFF9900))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xFFF700))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xFF0033))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
		},
	}
)

// Get returns the Theme with the given name, falling back to ThemeDefault if not found.
func Get(name string) *Theme {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "minimal":
		return ThemeMinimal
	case "mono":
		return ThemeMono
	case "nord":
		return ThemeNord
	case "dracula":
		return ThemeDracula
	case "neon":
		return ThemeNeon
	default:
		return ThemeDefault
	}
}

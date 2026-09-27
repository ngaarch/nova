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

	ThemeCyberpunk = &Theme{
		Name: "cyberpunk",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x00F0FF)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xFAF0F6))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0xFF007F)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xFF0033)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFFE600))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0x9D00FF)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xFFE600))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xFF007F))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x00F0FF))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xFFE600))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0x9D00FF))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x00F0FF))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x39FF14))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x606080))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xFFE600)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xFF0033)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x00F0FF))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x606080))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0x000000)), Bg: rgbPtr(HexRGB(0x00F0FF)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0xFF007F)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x39FF14))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x00F0FF))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xFFE600))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFF007F))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xFFE600))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xFF0033))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x39FF14)), Bold: true},
		},
	}

	ThemeSynthwave = &Theme{
		Name: "synthwave",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0xFF71CE)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xFEE8FF))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x01CDFE)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x05FFA1)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xFF2A85)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFFB961))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xB967FF)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xFFFB96))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xFF71CE))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x01CDFE))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xFFFB96))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xB967FF))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x05FFA1))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x01CDFE))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x61486F))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x05FFA1)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xFFB961)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xFF2A85)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x01CDFE))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x61486F))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0x261447)), Bg: rgbPtr(HexRGB(0xFF71CE)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0xFF71CE)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x05FFA1))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x01CDFE))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xFFFB96))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xB967FF))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xFFFB96))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xFF2A85))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x01CDFE)), Bold: true},
		},
	}

	ThemeTokyoNight = &Theme{
		Name: "tokyo-night",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x7AA2F7)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xC0CAF5))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0x9ECE6A)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x7DCFFF)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xF7768E)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFF9E64))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xBB9AF7)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xE0AF68))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xF7768E))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x7AA2F7))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xE0AF68))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xBB9AF7))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x7DCFFF))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0x9ECE6A))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x565F89))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0x9ECE6A)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xE0AF68)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xF7768E)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x7DCFFF))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x565F89))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xC0CAF5)), Bg: rgbPtr(HexRGB(0x283457)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0x7AA2F7)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0x9ECE6A))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x7DCFFF))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xE0AF68))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFF9E64))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xE0AF68))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xF7768E))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0x9ECE6A)), Bold: true},
		},
	}

	ThemeCatppuccin = &Theme{
		Name: "catppuccin",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x89B4FA)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xCDD6F4))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0xA6E3A1)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x94E2D5)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xF38BA8)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFAB387))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xCBA6F7)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xF9E2AF))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xF38BA8))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x89B4FA))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xF9E2AF))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xCBA6F7))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x94E2D5))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0xA6E3A1))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x6C7086))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0xA6E3A1)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xF9E2AF)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xF38BA8)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x89DCEB))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x6C7086))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xCDD6F4)), Bg: rgbPtr(HexRGB(0x45475A)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0xCBA6F7)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0xA6E3A1))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x89DCEB))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xF9E2AF))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFAB387))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xF9E2AF))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xF38BA8))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0xA6E3A1)), Bold: true},
		},
	}

	ThemeGruvbox = &Theme{
		Name: "gruvbox",
		Styles: map[Role]Style{
			RoleDirectory:     {Fg: rgbPtr(HexRGB(0x83A598)), Bold: true},
			RoleRegularFile:   {Fg: rgbPtr(HexRGB(0xEBDBB2))},
			RoleExecutable:    {Fg: rgbPtr(HexRGB(0xB8BB26)), Bold: true},
			RoleSymlink:       {Fg: rgbPtr(HexRGB(0x8EC07C)), Italic: true},
			RoleBrokenSymlink: {Fg: rgbPtr(HexRGB(0xFB4934)), Bold: true, Underline: true},
			RolePipe:          {Fg: rgbPtr(HexRGB(0xFE8019))},
			RoleSocket:        {Fg: rgbPtr(HexRGB(0xD3869B)), Bold: true},
			RoleDevice:        {Fg: rgbPtr(HexRGB(0xFABD2F))},
			RoleArchive:       {Fg: rgbPtr(HexRGB(0xFB4934))},
			RoleCode:          {Fg: rgbPtr(HexRGB(0x83A598))},
			RoleDocument:      {Fg: rgbPtr(HexRGB(0xFABD2F))},
			RoleImage:         {Fg: rgbPtr(HexRGB(0xD3869B))},
			RoleAudio:         {Fg: rgbPtr(HexRGB(0x8EC07C))},
			RoleVideo:         {Fg: rgbPtr(HexRGB(0xB8BB26))},
			RoleHidden:        {Fg: rgbPtr(HexRGB(0x928374))},

			RoleSuccess:   {Fg: rgbPtr(HexRGB(0xB8BB26)), Bold: true},
			RoleWarning:   {Fg: rgbPtr(HexRGB(0xFABD2F)), Bold: true},
			RoleError:     {Fg: rgbPtr(HexRGB(0xFB4934)), Bold: true},
			RoleInfo:      {Fg: rgbPtr(HexRGB(0x83A598))},
			RoleMuted:     {Fg: rgbPtr(HexRGB(0x928374))},
			RoleSelection: {Fg: rgbPtr(HexRGB(0xEBDBB2)), Bg: rgbPtr(HexRGB(0x504945)), Bold: true},
			RoleAccent:    {Fg: rgbPtr(HexRGB(0xFE8019)), Bold: true},

			RoleSize:      {Fg: rgbPtr(HexRGB(0xB8BB26))},
			RoleDate:      {Fg: rgbPtr(HexRGB(0x8EC07C))},
			RoleUser:      {Fg: rgbPtr(HexRGB(0xFABD2F))},
			RoleGroup:     {Fg: rgbPtr(HexRGB(0xFE8019))},
			RolePermRead:  {Fg: rgbPtr(HexRGB(0xFABD2F))},
			RolePermWrite: {Fg: rgbPtr(HexRGB(0xFB4934))},
			RolePermExec:  {Fg: rgbPtr(HexRGB(0xB8BB26)), Bold: true},
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
	case "cyberpunk":
		return ThemeCyberpunk
	case "synthwave":
		return ThemeSynthwave
	case "tokyonight", "tokyo-night":
		return ThemeTokyoNight
	case "catppuccin", "catppuccin-mocha", "mocha":
		return ThemeCatppuccin
	case "gruvbox", "gruvbox-dark":
		return ThemeGruvbox
	default:
		return ThemeDefault
	}
}

// ListThemes returns the list of all available theme names.
func ListThemes() []string {
	return []string{
		"default",
		"dracula",
		"nord",
		"neon",
		"cyberpunk",
		"synthwave",
		"tokyo-night",
		"catppuccin",
		"gruvbox",
		"minimal",
		"mono",
	}
}

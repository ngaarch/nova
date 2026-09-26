package interactive

import (
	"strings"
)

// FilterEntries returns a subset of entries matching the given query string.
// If query is empty, all entries are returned.
func FilterEntries(entries []Entry, query string) []Entry {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return entries
	}

	qLower := strings.ToLower(trimmed)
	var matched []Entry
	for _, e := range entries {
		nameLower := strings.ToLower(e.Name)
		if strings.Contains(nameLower, qLower) || fuzzyMatch(nameLower, qLower) {
			matched = append(matched, e)
		}
	}
	return matched
}

// fuzzyMatch checks if pattern runes appear in order in s.
func fuzzyMatch(s, pattern string) bool {
	if len(pattern) == 0 {
		return true
	}
	if len(pattern) > len(s) {
		return false
	}

	pRunes := []rune(pattern)
	sRunes := []rune(s)

	pIdx := 0
	for _, r := range sRunes {
		if r == pRunes[pIdx] {
			pIdx++
			if pIdx == len(pRunes) {
				return true
			}
		}
	}
	return false
}

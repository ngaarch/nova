package diff

import (
	"strings"
	"time"
)

// LineType defines whether a diff line is context, deletion, or insertion.
type LineType int

const (
	LineContext LineType = iota
	LineDelete
	LineInsert
)

// DiffLine represents a single line inside a diff hunk.
type DiffLine struct {
	Type    LineType `json:"type"`
	OldNum  int      `json:"old_num,omitempty"`
	NewNum  int      `json:"new_num,omitempty"`
	Content string   `json:"content"`
}

// Hunk represents a contiguous block of differences surrounded by context lines.
type Hunk struct {
	OldStart int        `json:"old_start"`
	OldCount int        `json:"old_count"`
	NewStart int        `json:"new_start"`
	NewCount int        `json:"new_count"`
	Lines    []DiffLine `json:"lines"`
}

// DiffResult holds the complete evaluation of difference between two files.
type DiffResult struct {
	File1     string    `json:"file1"`
	File2     string    `json:"file2"`
	Time1     time.Time `json:"time1"`
	Time2     time.Time `json:"time2"`
	Identical bool      `json:"identical"`
	IsBinary  bool      `json:"is_binary"`
	Additions int       `json:"additions"`
	Deletions int       `json:"deletions"`
	Hunks     []Hunk    `json:"hunks"`
}

// ComputeDiff calculates the unified diff between lines1 and lines2.
func ComputeDiff(lines1, lines2 []string, opts Options) DiffResult {
	norm1 := make([]string, len(lines1))
	norm2 := make([]string, len(lines2))
	for i, l := range lines1 {
		norm1[i] = normalizeLine(l, opts)
	}
	for i, l := range lines2 {
		norm2[i] = normalizeLine(l, opts)
	}

	m := len(norm1)
	n := len(norm2)

	// Compute LCS table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if norm1[i-1] == norm2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to reconstruct edits
	type edit struct {
		typ    LineType
		oldIdx int
		newIdx int
	}
	var edits []edit

	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && norm1[i-1] == norm2[j-1] {
			edits = append(edits, edit{typ: LineContext, oldIdx: i - 1, newIdx: j - 1})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			edits = append(edits, edit{typ: LineInsert, oldIdx: -1, newIdx: j - 1})
			j--
		} else if i > 0 && (j == 0 || dp[i][j-1] < dp[i-1][j]) {
			edits = append(edits, edit{typ: LineDelete, oldIdx: i - 1, newIdx: -1})
			i--
		}
	}

	// Reverse edits to chronological order
	for k := 0; k < len(edits)/2; k++ {
		opp := len(edits) - 1 - k
		edits[k], edits[opp] = edits[opp], edits[k]
	}

	// Check if identical
	hasDiff := false
	for _, ed := range edits {
		if ed.typ != LineContext {
			hasDiff = true
			break
		}
	}

	if !hasDiff {
		return DiffResult{
			Identical: true,
		}
	}

	// Group edits into hunks
	contextLen := opts.ContextLines
	if contextLen < 0 {
		contextLen = 3
	}

	var hunks []Hunk
	totalAdd := 0
	totalDel := 0

	var curHunk *Hunk
	var pendingContext []DiffLine

	oldLineNum := 1
	newLineNum := 1

	for _, ed := range edits {
		switch ed.typ {
		case LineContext:
			dl := DiffLine{
				Type:    LineContext,
				OldNum:  oldLineNum,
				NewNum:  newLineNum,
				Content: lines1[ed.oldIdx],
			}
			oldLineNum++
			newLineNum++

			if curHunk == nil {
				pendingContext = append(pendingContext, dl)
				if len(pendingContext) > contextLen {
					pendingContext = pendingContext[1:]
				}
			} else {
				curHunk.Lines = append(curHunk.Lines, dl)
				// If consecutive context lines exceed 2 * contextLen, close the hunk
				tailContext := 0
				for k := len(curHunk.Lines) - 1; k >= 0; k-- {
					if curHunk.Lines[k].Type == LineContext {
						tailContext++
					} else {
						break
					}
				}
				if tailContext > contextLen {
					// Trim excess context from end
					excess := tailContext - contextLen
					curHunk.Lines = curHunk.Lines[:len(curHunk.Lines)-excess]
					finalizeHunk(curHunk)
					hunks = append(hunks, *curHunk)
					curHunk = nil
					pendingContext = []DiffLine{dl}
				}
			}

		case LineDelete:
			totalDel++
			dl := DiffLine{
				Type:    LineDelete,
				OldNum:  oldLineNum,
				Content: lines1[ed.oldIdx],
			}
			oldLineNum++

			if curHunk == nil {
				curHunk = &Hunk{
					Lines: append([]DiffLine{}, pendingContext...),
				}
				pendingContext = nil
			}
			curHunk.Lines = append(curHunk.Lines, dl)

		case LineInsert:
			totalAdd++
			dl := DiffLine{
				Type:    LineInsert,
				NewNum:  newLineNum,
				Content: lines2[ed.newIdx],
			}
			newLineNum++

			if curHunk == nil {
				curHunk = &Hunk{
					Lines: append([]DiffLine{}, pendingContext...),
				}
				pendingContext = nil
			}
			curHunk.Lines = append(curHunk.Lines, dl)
		}
	}

	if curHunk != nil {
		finalizeHunk(curHunk)
		hunks = append(hunks, *curHunk)
	}

	return DiffResult{
		Identical: false,
		Additions: totalAdd,
		Deletions: totalDel,
		Hunks:     hunks,
	}
}

func finalizeHunk(h *Hunk) {
	oldCount := 0
	newCount := 0
	oldStart := 0
	newStart := 0

	for _, l := range h.Lines {
		if l.Type == LineContext {
			if oldStart == 0 {
				oldStart = l.OldNum
			}
			if newStart == 0 {
				newStart = l.NewNum
			}
			oldCount++
			newCount++
		} else if l.Type == LineDelete {
			if oldStart == 0 {
				oldStart = l.OldNum
			}
			oldCount++
		} else if l.Type == LineInsert {
			if newStart == 0 {
				newStart = l.NewNum
			}
			newCount++
		}
	}

	if oldStart == 0 {
		oldStart = 1
	}
	if newStart == 0 {
		newStart = 1
	}

	h.OldStart = oldStart
	h.OldCount = oldCount
	h.NewStart = newStart
	h.NewCount = newCount
}

func normalizeLine(s string, opts Options) string {
	res := s
	if opts.IgnoreCase {
		res = strings.ToLower(res)
	}
	if opts.IgnoreAllSpace {
		res = strings.Join(strings.Fields(res), "")
	}
	return res
}

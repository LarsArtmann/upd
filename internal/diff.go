package internal

import "slices"

type diffOp int

const (
	opEqual diffOp = iota
	opInsert
	opDelete
)

type diffChunk struct {
	op   diffOp
	text string
}

// diffChars computes a character-level diff of oldStr vs newStr.
// Returns chunks where DELETE = in old only, INSERT = in new only, EQUAL = in both.
// Uses a simple LCS-based algorithm — sufficient for short version strings.
func diffChars(oldStr, newStr string) []diffChunk {
	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)
	m, n := len(oldRunes), len(newRunes)

	// Build LCS table
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldRunes[i-1] == newRunes[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else if lcs[i-1][j] >= lcs[i][j-1] {
				lcs[i][j] = lcs[i-1][j]
			} else {
				lcs[i][j] = lcs[i][j-1]
			}
		}
	}

	// Backtrack to produce diff
	var chunks []diffChunk
	i, j := m, n
	type step struct {
		op   diffOp
		char rune
	}
	var steps []step

	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && oldRunes[i-1] == newRunes[j-1]:
			steps = append(steps, step{opEqual, oldRunes[i-1]})
			i--
			j--
		case j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]):
			steps = append(steps, step{opInsert, newRunes[j-1]})
			j--
		case i > 0 && (j == 0 || lcs[i-1][j] > lcs[i][j-1]):
			steps = append(steps, step{opDelete, oldRunes[i-1]})
			i--
		}
	}

	// Reverse and coalesce consecutive same-op chunks
	for _, v := range slices.Backward(steps) {
		s := v
		if len(chunks) > 0 && chunks[len(chunks)-1].op == s.op {
			chunks[len(chunks)-1].text += string(s.char)
		} else {
			chunks = append(chunks, diffChunk{op: s.op, text: string(s.char)})
		}
	}

	return chunks
}

package upd

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

type diffStep struct {
	op   diffOp
	char rune
}

// diffChars computes a character-level diff of oldStr vs newStr.
// Returns chunks where DELETE = in old only, INSERT = in new only, EQUAL = in both.
// Uses a simple LCS-based algorithm — sufficient for short version strings.
func diffChars(oldStr, newStr string) []diffChunk {
	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)

	lcs := buildLCSTable(oldRunes, newRunes)
	steps := backtrackDiff(oldRunes, newRunes, lcs)

	return coalesceSteps(steps)
}

func buildLCSTable(oldRunes, newRunes []rune) [][]int {
	oldLen := len(oldRunes)
	newLen := len(newRunes)

	lcs := make([][]int, oldLen+1)
	for rowIdx := range lcs {
		lcs[rowIdx] = make([]int, newLen+1)
	}

	for i := 1; i <= oldLen; i++ {
		for j := 1; j <= newLen; j++ {
			lcs[i][j] = lcsCell(lcs, oldRunes, newRunes, i, j)
		}
	}

	return lcs
}

func lcsCell(lcs [][]int, oldRunes, newRunes []rune, i, j int) int {
	if oldRunes[i-1] == newRunes[j-1] {
		return lcs[i-1][j-1] + 1
	}

	if lcs[i-1][j] >= lcs[i][j-1] {
		return lcs[i-1][j]
	}

	return lcs[i][j-1]
}

func backtrackDiff(oldRunes, newRunes []rune, lcs [][]int) []diffStep {
	oldLen := len(oldRunes)
	newLen := len(newRunes)

	steps := make([]diffStep, 0, oldLen+newLen)

	i := oldLen
	j := newLen

	for i > 0 || j > 0 {
		next := nextDiffStep(oldRunes, newRunes, lcs, i, j)
		if next == nil {
			break
		}

		steps = append(steps, next.step)
		i, j = next.idx, next.jdx
	}

	return steps
}

type diffStepResult struct {
	step diffStep
	idx  int
	jdx  int
}

func nextDiffStep(oldRunes, newRunes []rune, lcs [][]int, i, j int) *diffStepResult {
	if i > 0 && j > 0 && oldRunes[i-1] == newRunes[j-1] {
		return &diffStepResult{
			step: diffStep{op: opEqual, char: oldRunes[i-1]},
			idx:  i - 1,
			jdx:  j - 1,
		}
	}

	if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
		return &diffStepResult{
			step: diffStep{op: opInsert, char: newRunes[j-1]},
			idx:  i,
			jdx:  j - 1,
		}
	}

	if i > 0 {
		return &diffStepResult{
			step: diffStep{op: opDelete, char: oldRunes[i-1]},
			idx:  i - 1,
			jdx:  j,
		}
	}

	return nil
}

func coalesceSteps(steps []diffStep) []diffChunk {
	chunks := make([]diffChunk, 0, len(steps))

	for _, step := range slices.Backward(steps) {
		if len(chunks) > 0 && chunks[len(chunks)-1].op == step.op {
			chunks[len(chunks)-1].text += string(step.char)
		} else {
			chunks = append(chunks, diffChunk{op: step.op, text: string(step.char)})
		}
	}

	return chunks
}

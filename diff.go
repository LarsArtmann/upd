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

	cols := len(newRunes) + 1
	lcs := buildLCSTable(oldRunes, newRunes, cols)
	steps := backtrackDiff(oldRunes, newRunes, lcs, cols)

	return coalesceSteps(steps)
}

// buildLCSTable builds a flattened (oldLen+1)×(newLen+1) LCS dynamic-programming
// table as a 1-D slice. The stride is cols (= newLen+1), so lcs[i*cols+j]
// corresponds to the 2-D position [i][j].
func buildLCSTable(oldRunes, newRunes []rune, cols int) []int {
	oldLen := len(oldRunes)
	newLen := len(newRunes)

	lcs := make([]int, 0, (oldLen+1)*cols)
	for row := 0; row <= oldLen; row++ {
		lcs = append(lcs, make([]int, cols)...)
	}

	for i := 1; i <= oldLen; i++ {
		for j := 1; j <= newLen; j++ {
			switch {
			case oldRunes[i-1] == newRunes[j-1]:
				lcs[i*cols+j] = lcs[(i-1)*cols+j-1] + 1
			case lcs[(i-1)*cols+j] >= lcs[i*cols+j-1]:
				lcs[i*cols+j] = lcs[(i-1)*cols+j]
			default:
				lcs[i*cols+j] = lcs[i*cols+j-1]
			}
		}
	}

	return lcs
}

func backtrackDiff(oldRunes, newRunes []rune, lcs []int, cols int) []diffStep {
	oldLen := len(oldRunes)
	newLen := len(newRunes)

	steps := make([]diffStep, 0, oldLen+newLen)

	i := oldLen
	j := newLen

	for i > 0 || j > 0 {
		next := nextDiffStep(oldRunes, newRunes, lcs, cols, i, j)
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

func nextDiffStep(oldRunes, newRunes []rune, lcs []int, cols, i, j int) *diffStepResult {
	if i > 0 && j > 0 && oldRunes[i-1] == newRunes[j-1] {
		return &diffStepResult{
			step: diffStep{op: opEqual, char: oldRunes[i-1]},
			idx:  i - 1,
			jdx:  j - 1,
		}
	}

	if j > 0 && (i == 0 || lcs[i*cols+j-1] >= lcs[(i-1)*cols+j]) {
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

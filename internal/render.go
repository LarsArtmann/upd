package internal

import (
	"fmt"
	"io"
	"strings"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiGrey   = "\x1b[90m"
	ansiBlue   = "\x1b[34m"
)

type Renderer struct {
	w       io.Writer
	noColor bool
}

func NewRenderer(w io.Writer, noColor bool) *Renderer {
	return &Renderer{w: w, noColor: noColor}
}

func (r *Renderer) color(code, text string) string {
	if r.noColor {
		return text
	}
	return code + text + ansiReset
}

func (r *Renderer) bold(text string) string {
	return r.color(ansiBold, text)
}

func (r *Renderer) red(text string) string {
	return r.color(ansiRed, text)
}

func (r *Renderer) green(text string) string {
	return r.color(ansiGreen, text)
}

func (r *Renderer) grey(text string) string {
	return r.color(ansiGrey, text)
}

func (r *Renderer) RenderTable(manifest Manifest, updates, errors int, showAll bool) {
	if updates == 0 && errors == 0 && !showAll {
		r.renderAllUpToDate()
		return
	}

	r.renderUpgradeTable(manifest, showAll)
}

func (r *Renderer) renderAllUpToDate() {
	border := strings.Repeat("─", 79)
	box := r.green("ALL PACKAGE DEPENDENCIES UP-TO-DATE")
	fmt.Fprintf(r.w, "┌%s┐\n", border)
	fmt.Fprintf(r.w, "│%s│\n", centerPad(box, 79))
	fmt.Fprintf(r.w, "└%s┘\n", border)
}

func (r *Renderer) renderUpgradeTable(manifest Manifest, showAll bool) {
	const (
		colName = 37
		colVer  = 14
		colState = 9
	)

	// Header
	fmt.Fprintf(r.w, "┌%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", colName),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colState),
	)

	fmt.Fprintf(r.w, "│%s│%s│%s│%s│\n",
		r.bold(padCell("MODULE NAME", colName)),
		r.bold(padCell("VERSION OLD", colVer)),
		r.bold(padCell("VERSION NEW", colVer)),
		r.bold(padCell("STATE", colState)),
	)

	fmt.Fprintf(r.w, "├%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", colName),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colState),
	)

	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if !showAll && spec.State != StateUpdated && spec.State != StateError {
				continue
			}

			var modName, oldVer, newVer, state string

			switch {
			case spec.State == StateUpdated:
				modName = name
				oldVer = r.markRed(spec.SNew, spec.SOld)
				newVer = r.markGreen(spec.SOld, spec.SNew)
				state = r.green(string(spec.State))
			case spec.State == StateError:
				modName = r.grey(name)
				oldVer = r.grey(spec.SOld)
				newVer = r.grey(spec.SNew)
				state = r.red(string(spec.State))
			default:
				modName = r.grey(name)
				oldVer = r.grey(spec.SOld)
				newVer = r.grey(spec.SNew)
				st := string(spec.State)
				if st == "" {
					st = "kept"
				}
				state = r.grey(st)
			}

			fmt.Fprintf(r.w, "│%s│%s│%s│%s│\n",
				padCell(modName, colName),
				padCell(oldVer, colVer),
				padCell(newVer, colVer),
				padCell(state, colState),
			)
		}
	}

	fmt.Fprintf(r.w, "└%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", colName),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colVer),
		strings.Repeat("─", colState),
	)
}

func (r *Renderer) markRed(text, other string) string {
	return r.diffHighlight(text, other, ansiRed)
}

func (r *Renderer) markGreen(text, other string) string {
	return r.diffHighlight(text, other, ansiGreen)
}

func (r *Renderer) diffHighlight(text, other string, color string) string {
	if r.noColor {
		return text
	}

	chunks := diffChars(text, other)
	var sb strings.Builder
	for _, c := range chunks {
		switch c.op {
		case opInsert:
			sb.WriteString(r.color(color, c.text))
		case opEqual:
			sb.WriteString(c.text)
		}
	}
	return sb.String()
}

func centerPad(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	total := width - len(text)
	left := total / 2
	right := total - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func padCell(text string, width int) string {
	visibleLen := visibleLength(text)
	if visibleLen >= width {
		return text
	}
	return text + strings.Repeat(" ", width-visibleLen)
}

func visibleLength(text string) int {
	// Strip ANSI escape sequences for length calculation
	out := make([]byte, 0, len(text))
	inEscape := false
	for _, c := range []byte(text) {
		if c == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		out = append(out, c)
	}
	return len(string(out))
}

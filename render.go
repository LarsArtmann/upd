package upd

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"strings"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
	ansiGrey  = "\x1b[90m"

	boxBorderChar = "─"
	terminalWidth = 79
	halfDivisor   = 2

	errorNameColumnWidth = 20
)

type Renderer struct {
	w       io.Writer
	noColor bool
	verbose bool
}

type RendererOptions struct {
	NoColor bool
	Verbose bool
}

func NewRenderer(w io.Writer, opts RendererOptions) *Renderer {
	return &Renderer{w: w, noColor: opts.NoColor, verbose: opts.Verbose}
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
	r.renderErrorDetails(manifest)
}

func (r *Renderer) renderAllUpToDate() {
	border := strings.Repeat(boxBorderChar, terminalWidth)
	box := r.green("ALL PACKAGE DEPENDENCIES UP-TO-DATE")
	_, _ = fmt.Fprintf(r.w, "┌%s┐\n", border)
	_, _ = fmt.Fprintf(r.w, "│%s│\n", centerPad(box, terminalWidth))
	_, _ = fmt.Fprintf(r.w, "└%s┘\n", border)
}

func (r *Renderer) renderErrorDetails(manifest Manifest) {
	type entry struct {
		name string
		err  error
	}

	var entries []entry

	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if spec.State == StateError && spec.Err != nil {
				entries = append(entries, entry{name: name, err: spec.Err})
			}
		}
	}

	if len(entries) == 0 {
		return
	}

	_, _ = fmt.Fprintf(r.w, "\n%s\n", r.bold(r.red(fmt.Sprintf("Errors (%d):", len(entries)))))

	nameWidth := errorNameColumnWidth

	for _, e := range entries {
		msg := e.err.Error()
		if r.verbose {
			msg = fmt.Sprintf("%+v", e.err)
		}

		_, _ = fmt.Fprintf(r.w, "  %s  %s\n", r.grey(padCell(e.name, nameWidth)), msg)
	}
}

func (r *Renderer) renderUpgradeTable(manifest Manifest, showAll bool) {
	const (
		colName  = 37
		colVer   = 14
		colState = 9
	)

	// Header
	r.renderBorder("top", colName, colVer, colVer, colState)

	_, _ = fmt.Fprintf(
		r.w, "│%s│%s│%s│%s│\n",
		r.bold(padCell("MODULE NAME", colName)),
		r.bold(padCell("VERSION OLD", colVer)),
		r.bold(padCell("VERSION NEW", colVer)),
		r.bold(padCell("STATE", colState)),
	)

	r.renderBorder("mid", colName, colVer, colVer, colState)

	r.renderRows(manifest, colName, colVer, colState, showAll)

	r.renderBorder("bottom", colName, colVer, colVer, colState)
}

func (r *Renderer) renderRows(manifest Manifest, colName, colVer, colState int, showAll bool) {
	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if !showAll && spec.State != StateUpdated && spec.State != StateError {
				continue
			}

			modName, oldVer, newVer, state := r.renderRow(name, spec)
			_, _ = fmt.Fprintf(
				r.w, "│%s│%s│%s│%s│\n",
				padCell(modName, colName),
				padCell(oldVer, colVer),
				padCell(newVer, colVer),
				padCell(state, colState),
			)
		}
	}
}

func (r *Renderer) renderRow(name string, spec *Spec) (string, string, string, string) {
	switch spec.State {
	case StateUpdated:
		modName := name
		oldVer := r.markRed(spec.SNew, spec.SOld)
		newVer := r.markGreen(spec.SOld, spec.SNew)
		state := r.green(string(spec.State))

		return modName, oldVer, newVer, state
	case StateError:
		modName := r.grey(name)
		oldVer := r.grey(spec.SOld)
		newVer := r.grey(spec.SNew)
		state := r.red(string(spec.State))

		return modName, oldVer, newVer, state
	case StateTodo, StateCheck, StateSkipped, StateKept, StateIgnored:
		modName := r.grey(name)
		oldVer := r.grey(spec.SOld)
		newVer := r.grey(spec.SNew)

		label := string(spec.State)
		if label == "" {
			label = "kept"
		}

		state := r.grey(label)

		return modName, oldVer, newVer, state
	}

	return "", "", "", ""
}

func borderChars(kind string) (string, string, string) {
	switch kind {
	case "top":
		return "┌", "┬", "┐"
	case "mid":
		return "├", "┼", "┤"
	case "bottom":
		return "└", "┴", "┘"
	}

	return "", "", ""
}

func (r *Renderer) renderBorder(kind string, widths ...int) {
	left, mid, right := borderChars(kind)
	r.writeBorder(left, mid, right, widths...)
}

func (r *Renderer) writeBorder(left, mid, right string, widths ...int) {
	segments := make([]string, 0, len(widths))
	for _, width := range widths {
		segments = append(segments, strings.Repeat(boxBorderChar, width))
	}

	_, _ = fmt.Fprintf(r.w, "%s%s%s\n", left, strings.Join(segments, mid), right)
}

func (r *Renderer) markRed(text, other string) string {
	return r.diffHighlight(text, other, ansiRed)
}

func (r *Renderer) markGreen(text, other string) string {
	return r.diffHighlight(text, other, ansiGreen)
}

func (r *Renderer) diffHighlight(text, other, color string) string {
	if r.noColor {
		return other
	}

	chunks := diffChars(text, other)

	var builder strings.Builder

	for _, chunk := range chunks {
		switch chunk.op {
		case opInsert:
			builder.WriteString(r.color(color, chunk.text))
		case opEqual:
			builder.WriteString(chunk.text)
		case opDelete:
			// deleted characters exist only in the old version, no need to render
		}
	}

	return builder.String()
}

func centerPad(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	total := width - len(text)
	left := total / halfDivisor
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

	for _, ch := range []byte(text) {
		if ch == '\x1b' {
			inEscape = true

			continue
		}

		if inEscape {
			if ch == 'm' {
				inEscape = false
			}

			continue
		}

		out = append(out, ch)
	}

	return len(string(out))
}

// --- JSON output ---

type jsonPackage struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	Old     string `json:"old"`
	New     string `json:"new"`
	State   string `json:"state"`
}

type jsonError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type jsonSummary struct {
	Updated int `json:"updated"`
	Kept    int `json:"kept"`
	Errors  int `json:"errors"`
	Total   int `json:"total"`
}

type jsonOutput struct {
	Summary  jsonSummary   `json:"summary"`
	Packages []jsonPackage `json:"packages"`
	Errors   []jsonError   `json:"errors,omitempty"`
}

// RenderJSON writes machine-readable JSON to w. Intended for CI pipelines and
// editor integrations where the table output is difficult to parse.
func RenderJSON(w io.Writer, manifest Manifest, updates int) error {
	summary := jsonSummary{Updated: updates, Kept: 0, Errors: 0, Total: 0}
	packages := make([]jsonPackage, 0, len(manifest))

	var jsonErrors []jsonError

	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			packages = append(packages, jsonPackage{
				Name:    name,
				Section: spec.Section,
				Old:     spec.SOld,
				New:     spec.SNew,
				State:   string(spec.State),
			})
			summary.Total++

			if spec.State == StateKept {
				summary.Kept++
			}

			if spec.State == StateError && spec.Err != nil {
				jsonErrors = append(jsonErrors, jsonError{
					Name:  name,
					Error: spec.Err.Error(),
				})
			}
		}
	}

	summary.Errors = len(jsonErrors)

	output := jsonOutput{
		Summary:  summary,
		Packages: packages,
		Errors:   jsonErrors,
	}

	enc := jsontext.NewEncoder(w, jsontext.WithIndent("  "))

	err := json.MarshalEncode(enc, output)
	if err != nil {
		return fmt.Errorf("encode JSON output: %w", err)
	}

	_, err = io.WriteString(w, "\n")
	if err != nil {
		return fmt.Errorf("write JSON newline: %w", err)
	}

	return nil
}

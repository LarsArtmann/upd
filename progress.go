package upd

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
)

const (
	progressComplete   = "█"
	progressIncomplete = "╌"
	progressWidth      = 24
	terminalResetWidth = 80
	percentMultiplier  = 100
)

type ProgressReporter struct {
	w       io.Writer
	total   int
	current atomic.Int64
}

func NewProgressReporter(w io.Writer, total int, _ bool) *ProgressReporter {
	return &ProgressReporter{w: w, total: total, current: atomic.Int64{}}
}

func (p *ProgressReporter) Start() {
	p.render("")
}

func (p *ProgressReporter) Tick(msg string, _ int) {
	p.render(msg)
}

func (p *ProgressReporter) Finish() {
	_, _ = fmt.Fprintf(p.w, "\r%s\r", strings.Repeat(" ", terminalResetWidth))
}

func (p *ProgressReporter) render(msg string) {
	current := min(int(p.current.Add(1)), p.total)

	filled := progressWidth * current / max(p.total, 1)
	if p.total == 0 {
		filled = progressWidth
	}

	bar := strings.Repeat(progressComplete, filled) + strings.Repeat(progressIncomplete, progressWidth-filled)
	percent := current * percentMultiplier / max(p.total, 1)

	_, _ = fmt.Fprintf(p.w, "\rchecking: %s %3d%% %s ", bar, percent, msg)
}

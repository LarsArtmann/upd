package internal

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
)

type ProgressReporter struct {
	w       io.Writer
	total   int
	current atomic.Int64
}

func NewProgressReporter(w io.Writer, total int, _ bool) *ProgressReporter {
	return &ProgressReporter{w: w, total: total}
}

func (p *ProgressReporter) Start() {
	p.render("")
}

func (p *ProgressReporter) Tick(msg string, _ int) {
	p.render(msg)
}

func (p *ProgressReporter) Finish() {
	fmt.Fprintf(p.w, "\r%s\r", strings.Repeat(" ", 80))
}

func (p *ProgressReporter) render(msg string) {
	current := int(p.current.Add(1))
	if current > p.total {
		current = p.total
	}

	filled := progressWidth * current / max(p.total, 1)
	if p.total == 0 {
		filled = progressWidth
	}

	bar := strings.Repeat(progressComplete, filled) + strings.Repeat(progressIncomplete, progressWidth-filled)
	percent := current * 100 / max(p.total, 1)

	fmt.Fprintf(p.w, "\rchecking: %s %3d%% %s ", bar, percent, msg)
}

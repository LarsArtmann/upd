package upd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	progressComplete   = "█"
	progressIncomplete = "╌"
	progressWidth      = 24
	defaultTermWidth   = 80
	percentMultiplier  = 100
)

type ProgressReporter struct {
	w       io.Writer
	total   int
	current atomic.Int64

	// mu serializes writes to w: Tick is called from concurrent fetch
	// goroutines, and concurrent Fprintf calls corrupt the output (and
	// trip the race detector on non-thread-safe writers).
	mu sync.Mutex
}

func NewProgressReporter(w io.Writer, total int, _ bool) *ProgressReporter {
	return &ProgressReporter{w: w, total: total, current: atomic.Int64{}, mu: sync.Mutex{}}
}

func (p *ProgressReporter) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, format, args...) //nolint:erraudit // progress-line write; write errors are not actionable
}

func (p *ProgressReporter) Start() {
	p.render("")
}

func (p *ProgressReporter) Tick(msg string, _ int) {
	p.render(msg)
}

func (p *ProgressReporter) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	width := clearWidth()
	p.printf("\r%s\r", strings.Repeat(" ", width))
}

// clearWidth returns the number of spaces to use when clearing the progress bar
// line. It checks the COLUMNS environment variable and falls back to 80.
func clearWidth() int {
	cols := os.Getenv("COLUMNS")
	if cols == "" {
		return defaultTermWidth
	}

	width, err := strconv.Atoi(cols)
	if err == nil && width > 0 {
		return width
	}

	return defaultTermWidth
}

func (p *ProgressReporter) render(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := min(int(p.current.Add(1)), p.total)

	filled := progressWidth * current / max(p.total, 1)
	if p.total == 0 {
		filled = progressWidth
	}

	bar := strings.Repeat(progressComplete, filled) + strings.Repeat(progressIncomplete, progressWidth-filled)
	percent := current * percentMultiplier / max(p.total, 1)

	p.printf("\rchecking: %s %3d%% %s ", bar, percent, msg)
}

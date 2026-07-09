package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/LarsArtmann/upd"
)

func TestExitCodeNilReturnsZero(t *testing.T) {
	t.Parallel()

	got := exitCode(nil)

	if got != 0 {
		t.Fatalf("exitCode(nil) = %d, want 0", got)
	}
}

func TestExitCodeRegistryUnavailableReturns75(t *testing.T) {
	t.Parallel()

	got := exitCode(upd.ErrRegistryUnavailable)

	if got != exitTransient {
		t.Fatalf("exitCode(ErrRegistryUnavailable) = %d, want %d", got, exitTransient)
	}
}

func TestExitCodeWrappedRegistryUnavailableReturns75(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("fetch failed: %w", upd.ErrRegistryUnavailable)

	got := exitCode(wrapped)

	if got != exitTransient {
		t.Fatalf("exitCode(wrapped ErrRegistryUnavailable) = %d, want %d", got, exitTransient)
	}
}

func TestExitCodeConcurrentModificationReturns1(t *testing.T) {
	t.Parallel()

	got := exitCode(upd.ErrConcurrentModification)

	if got != 1 {
		t.Fatalf("exitCode(ErrConcurrentModification) = %d, want 1", got)
	}
}

func TestExitCodeGenericErrorReturns1(t *testing.T) {
	t.Parallel()

	got := exitCode(upd.ErrInvalidJSON) //nolint:err113 // test-only: any non-registry sentinel proves exit 1

	if got != 1 {
		t.Fatalf("exitCode(generic error) = %d, want 1", got)
	}
}

func TestPrintWarningsEmptyProducesNoOutput(t *testing.T) {
	t.Parallel()

	output := capturePrintWarnings(t, nil)

	if output != "" {
		t.Fatalf("expected no output for empty warnings, got %q", output)
	}
}

func TestPrintWarningsOutputsFormattedLines(t *testing.T) {
	t.Parallel()

	warnings := []string{"invalid glob pattern", "malformed dependencies section"}
	output := capturePrintWarnings(t, warnings)

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	if len(lines) != len(warnings) {
		t.Fatalf("expected %d lines, got %d: %q", len(warnings), len(lines), output)
	}

	for i, want := range warnings {
		if !strings.Contains(lines[i], want) {
			t.Errorf("line %d: expected to contain %q, got %q", i, want, lines[i])
		}

		if !strings.Contains(lines[i], "\x1b[33mWARNING:\x1b[0m") {
			t.Errorf("line %d: missing yellow WARNING prefix, got %q", i, lines[i])
		}
	}
}

func capturePrintWarnings(t *testing.T, warnings []string) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	defer func() {
		err := r.Close()
		if err != nil {
			t.Errorf("close pipe reader: %v", err)
		}
	}()

	printWarnings(w, warnings)

	err = w.Close()
	if err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	return string(output)
}

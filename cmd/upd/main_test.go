package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

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

	for idx, want := range warnings {
		if !strings.Contains(lines[idx], want) {
			t.Errorf("line %d: expected to contain %q, got %q", idx, want, lines[idx])
		}

		if !strings.Contains(lines[idx], "\x1b[33mWARNING:\x1b[0m") {
			t.Errorf("line %d: missing yellow WARNING prefix, got %q", idx, lines[idx])
		}
	}
}

func capturePrintWarnings(t *testing.T, warnings []string) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			t.Errorf("close pipe reader: %v", closeErr)
		}
	}()

	printWarnings(writer, warnings)

	closeErr := writer.Close()
	if closeErr != nil {
		t.Fatalf("close pipe writer: %v", closeErr)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	return string(output)
}

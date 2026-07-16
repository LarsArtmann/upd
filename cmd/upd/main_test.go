package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/LarsArtmann/upd"
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

func TestVersionOutput(t *testing.T) {
	cmd, _ := upd.NewCommand(func(context.Context, *upd.Config) error { return nil })
	cmd.Version = "1.2.3-test"
	cmd.SetVersionTemplate(versionTemplate)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.2.3-test") {
		t.Errorf("version output missing version: %q", out)
	}

	if !strings.Contains(out, "Upgrade NPM Package Dependencies") {
		t.Errorf("version output missing description: %q", out)
	}
}

func TestCompletionBashOutput(t *testing.T) {
	cmd, _ := upd.NewCommand(func(context.Context, *upd.Config) error { return nil })

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "bash"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "# bash completion V2 for upd") {
		t.Errorf("unexpected bash completion prefix: %q", out)
	}

	if !strings.Contains(out, "upd") {
		t.Errorf("bash completion missing program name: %q", out)
	}
}

func TestManCommandOutput(t *testing.T) {
	cmd, _ := upd.NewCommand(func(context.Context, *upd.Config) error { return nil })
	cmd.SetArgs([]string{"man"})

	out := captureStdout(t, func() {
		if err := fang.Execute(context.Background(), cmd, fang.WithoutVersion()); err != nil {
			t.Errorf("Execute returned error: %v", err)
		}
	})

	if !strings.Contains(out, ".TH") {
		t.Errorf("man output missing .TH header: %q", out)
	}

	if !strings.Contains(out, "upd") {
		t.Errorf("man output missing program name: %q", out)
	}

	if !strings.Contains(out, "no-color") {
		t.Errorf("man output missing --no-color flag: %q", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w

	fn()

	os.Stdout = origStdout

	if err := w.Close(); err != nil {
		t.Errorf("close pipe writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	if err := r.Close(); err != nil {
		t.Errorf("close pipe reader: %v", err)
	}

	return string(out)
}

func TestColorSchemeFuncRespectsNoColor(t *testing.T) {
	cfg := upd.DefaultConfig()
	cfg.NoColor = true

	cs := colorSchemeFunc(cfg)(lipgloss.LightDark(true))
	assertAllNoColor(t, cs)
}

func TestColorSchemeFuncFallsBackToDefault(t *testing.T) {
	cfg := upd.DefaultConfig()
	cfg.NoColor = false

	cs := colorSchemeFunc(cfg)(lipgloss.LightDark(true))

	if _, ok := cs.Title.(lipgloss.NoColor); ok {
		t.Error("expected Title to have color when NoColor is false")
	}
}

func assertAllNoColor(t *testing.T, scheme fang.ColorScheme) {
	t.Helper()

	assertColorIsNoColor(t, "Base", scheme.Base)
	assertColorIsNoColor(t, "Title", scheme.Title)
	assertColorIsNoColor(t, "Description", scheme.Description)
	assertColorIsNoColor(t, "Codeblock", scheme.Codeblock)
	assertColorIsNoColor(t, "Program", scheme.Program)
	assertColorIsNoColor(t, "DimmedArgument", scheme.DimmedArgument)
	assertColorIsNoColor(t, "Comment", scheme.Comment)
	assertColorIsNoColor(t, "Flag", scheme.Flag)
	assertColorIsNoColor(t, "FlagDefault", scheme.FlagDefault)
	assertColorIsNoColor(t, "Command", scheme.Command)
	assertColorIsNoColor(t, "QuotedString", scheme.QuotedString)
	assertColorIsNoColor(t, "Argument", scheme.Argument)
	assertColorIsNoColor(t, "Help", scheme.Help)
	assertColorIsNoColor(t, "Dash", scheme.Dash)
	assertColorIsNoColor(t, "ErrorHeader[0]", scheme.ErrorHeader[0])
	assertColorIsNoColor(t, "ErrorHeader[1]", scheme.ErrorHeader[1])
	assertColorIsNoColor(t, "ErrorDetails", scheme.ErrorDetails)
}

func assertColorIsNoColor(t *testing.T, name string, c any) {
	t.Helper()

	if _, ok := c.(lipgloss.NoColor); !ok {
		t.Errorf("%s expected lipgloss.NoColor, got %T", name, c)
	}
}

func TestUnknownFlagReturnsError(t *testing.T) {
	cmd, _ := upd.NewCommand(func(context.Context, *upd.Config) error { return nil })
	cmd.SetArgs([]string{"--unknown-flag"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestDryRunAliasSetsNop(t *testing.T) {
	cfg, err := upd.ParseFlags([]string{"--dry-run"})
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	if !cfg.Nop {
		t.Error("expected Nop to be true for --dry-run")
	}
}

func TestNoColorAliasStillParses(t *testing.T) {
	cfg, err := upd.ParseFlags([]string{"--noColor"})
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	if !cfg.NoColor {
		t.Error("expected NoColor to be true for --noColor")
	}
}

func TestNoColorCanonicalFlagParses(t *testing.T) {
	cfg, err := upd.ParseFlags([]string{"--no-color"})
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	if !cfg.NoColor {
		t.Error("expected NoColor to be true for --no-color")
	}
}

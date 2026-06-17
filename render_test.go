package upd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTableUpdated(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	// Simulate an update
	for _, spec := range manifest["react"] {
		spec.SNew = "^19.0.0"
		spec.VNew = "19.0.0"
		spec.State = StateUpdated
	}

	var buf bytes.Buffer

	r := NewRenderer(&buf, true) // noColor=true for deterministic output
	r.RenderTable(manifest, 1, 0, false)

	output := buf.String()
	if !strings.Contains(output, "react") {
		t.Errorf("expected 'react' in output:\n%s", output)
	}

	if !strings.Contains(output, "^18.0.0") {
		t.Errorf("expected old version in output:\n%s", output)
	}

	if !strings.Contains(output, "^19.0.0") {
		t.Errorf("expected new version in output:\n%s", output)
	}

	if !strings.Contains(output, "updated") {
		t.Errorf("expected 'updated' state in output:\n%s", output)
	}
}

func TestRenderTableAllUpToDate(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	var buf bytes.Buffer

	r := NewRenderer(&buf, true)
	r.RenderTable(manifest, 0, 0, false)

	output := buf.String()
	if !strings.Contains(output, "UP-TO-DATE") {
		t.Errorf("expected 'UP-TO-DATE' message:\n%s", output)
	}
}

func TestRenderTableAllMode(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0", "lodash": "4.17.21"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	// Mark react as updated, lodash stays skipped
	for _, spec := range manifest["react"] {
		spec.State = StateUpdated
		spec.SNew = "^19.0.0"
		spec.VNew = "19.0.0"
	}

	var buf bytes.Buffer

	r := NewRenderer(&buf, true)
	r.RenderTable(manifest, 1, 0, true) // showAll=true

	output := buf.String()
	// Both packages should appear
	if !strings.Contains(output, "react") {
		t.Errorf("expected 'react' in all-mode output:\n%s", output)
	}

	if !strings.Contains(output, "lodash") {
		t.Errorf("expected 'lodash' in all-mode output:\n%s", output)
	}
}

func TestRenderTableErrorState(t *testing.T) {
	json := `{"dependencies": {"broken": "^1.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	for _, spec := range manifest["broken"] {
		spec.State = StateError
	}

	var buf bytes.Buffer

	r := NewRenderer(&buf, true)
	r.RenderTable(manifest, 0, 1, false)

	output := buf.String()
	if !strings.Contains(output, "error") {
		t.Errorf("expected 'error' state in output:\n%s", output)
	}
}

func TestRenderNoColorStripsANSI(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	for _, spec := range manifest["react"] {
		spec.State = StateUpdated
		spec.SNew = "^19.0.0"
		spec.VNew = "19.0.0"
	}

	var buf bytes.Buffer

	r := NewRenderer(&buf, true) // noColor
	r.RenderTable(manifest, 1, 0, false)

	output := buf.String()
	if strings.Contains(output, "\x1b[") {
		t.Errorf("noColor output contains ANSI codes:\n%s", output)
	}
}

func TestRenderWithColorContainsANSI(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	for _, spec := range manifest["react"] {
		spec.State = StateUpdated
		spec.SNew = "^19.0.0"
		spec.VNew = "19.0.0"
	}

	var buf bytes.Buffer

	r := NewRenderer(&buf, false) // color
	r.RenderTable(manifest, 1, 0, false)

	output := buf.String()
	if !strings.Contains(output, "\x1b[") {
		t.Errorf("color output missing ANSI codes:\n%s", output)
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"\x1b[31mred\x1b[0m", 3},
		{"\x1b[1mbold\x1b[0m text", 9},
		{"", 0},
	}

	for _, tt := range tests {
		got := visibleLength(tt.input)
		if got != tt.want {
			t.Errorf("visibleLength(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCenterPad(t *testing.T) {
	got := centerPad("hi", 10)
	if len(got) != 10 {
		t.Errorf("centerPad length = %d, want 10", len(got))
	}

	if !strings.Contains(got, "hi") {
		t.Errorf("centerPad lost content: %q", got)
	}
}

package upd

import (
	"bytes"
	"strings"
	"testing"
)

func markReactUpdated(manifest Manifest) {
	for _, spec := range manifest["react"] {
		spec.State = StateUpdated
		spec.SNew = "^19.0.0"
		spec.VNew = "19.0.0"
	}
}

func renderManifest(t *testing.T, manifest Manifest, updated, errored int, noColor, showAll bool) string {
	t.Helper()

	var buf bytes.Buffer

	r := NewRenderer(&buf, noColor)
	r.RenderTable(manifest, updated, errored, showAll)

	return buf.String()
}

func TestRenderTableUpdated(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)
	markReactUpdated(manifest)

	output := renderManifest(t, manifest, 1, 0, true, false)

	assertContains(t, output, "react", "package name")
	assertContains(t, output, "^18.0.0", "old version")
	assertContains(t, output, "^19.0.0", "new version")
	assertContains(t, output, "updated", "state label")
}

func TestRenderTableAllUpToDate(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)

	output := renderManifest(t, manifest, 0, 0, true, false)

	assertContains(t, output, "UP-TO-DATE", "status banner")
}

func TestRenderTableAllMode(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0", "lodash": "4.17.21"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)
	markReactUpdated(manifest)

	output := renderManifest(t, manifest, 1, 0, true, true)

	assertContains(t, output, "react", "package name")
	assertContains(t, output, "lodash", "package name")
}

func TestRenderTableErrorState(t *testing.T) {
	json := `{"dependencies": {"broken": "^1.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)

	for _, spec := range manifest["broken"] {
		spec.State = StateError
	}

	output := renderManifest(t, manifest, 0, 1, true, false)

	assertContains(t, output, "error", "state label")
}

func TestRenderNoColorStripsANSI(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)
	markReactUpdated(manifest)

	output := renderManifest(t, manifest, 1, 0, true, false)

	assertNotContains(t, output, "\x1b[", "ANSI escape sequences")
}

func TestRenderWithColorContainsANSI(t *testing.T) {
	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil, false)
	markReactUpdated(manifest)

	output := renderManifest(t, manifest, 1, 0, false, false)

	assertContains(t, output, "\x1b[", "ANSI escape sequences")
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

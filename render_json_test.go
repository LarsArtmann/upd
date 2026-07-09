package upd

import (
	"bytes"
	"encoding/json/v2"
	"strings"
	"testing"
)

func TestRenderJSONBasicOutput(t *testing.T) {
	pkgJSON := `{
		"dependencies": {
			"react": "^18.0.0",
			"lodash": "4.17.20"
		}
	}`

	pkg := &PackageFile{raw: []byte(pkgJSON)}
	manifest, _ := BuildManifest(pkg, nil, false)

	// Simulate updated react
	manifest["react"][0].State = StateUpdated
	manifest["react"][0].SNew = "^19.0.0"
	manifest["react"][0].VNew = "19.0.0"

	// Simulate kept lodash
	manifest["lodash"][0].State = StateKept

	var buf bytes.Buffer

	err := RenderJSON(&buf, manifest, 1)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	output := buf.String()

	var result jsonOutput

	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("JSON output is not valid JSON: %v\nOutput:\n%s", err, output)
	}

	if result.Summary.Updated != 1 {
		t.Errorf("summary.updated = %d, want 1", result.Summary.Updated)
	}

	if result.Summary.Kept != 1 {
		t.Errorf("summary.kept = %d, want 1", result.Summary.Kept)
	}

	if result.Summary.Total != 2 {
		t.Errorf("summary.total = %d, want 2", result.Summary.Total)
	}

	if len(result.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(result.Packages))
	}

	// SortedNames() returns alphabetical order: lodash, react
	lodash := result.Packages[0]
	if lodash.Name != "lodash" {
		t.Errorf("packages[0].name = %q, want lodash", lodash.Name)
	}

	if lodash.State != "kept" {
		t.Errorf("packages[0].state = %q, want kept", lodash.State)
	}

	react := result.Packages[1]
	if react.Name != "react" {
		t.Errorf("packages[1].name = %q, want react", react.Name)
	}

	if react.Old != "^18.0.0" {
		t.Errorf("packages[1].old = %q, want ^18.0.0", react.Old)
	}

	if react.New != "^19.0.0" {
		t.Errorf("packages[1].new = %q, want ^19.0.0", react.New)
	}

	if react.State != "updated" {
		t.Errorf("packages[1].state = %q, want updated", react.State)
	}
}

func TestRenderJSONIncludesErrors(t *testing.T) {
	pkgJSON := `{"dependencies": {"ghost": "^1.0.0"}}`
	pkg := &PackageFile{raw: []byte(pkgJSON)}
	manifest, _ := BuildManifest(pkg, nil, false)

	manifest["ghost"][0].State = StateError
	manifest["ghost"][0].Err = ErrPackageNotFound

	var buf bytes.Buffer

	err := RenderJSON(&buf, manifest, 0)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result jsonOutput

	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error entry, got %d", len(result.Errors))
	}

	if result.Errors[0].Name != "ghost" {
		t.Errorf("errors[0].name = %q, want ghost", result.Errors[0].Name)
	}

	if !strings.Contains(result.Errors[0].Error, "not found") {
		t.Errorf("errors[0].error should contain 'not found', got %q", result.Errors[0].Error)
	}
}

func TestRenderJSONNoErrorsOmitsField(t *testing.T) {
	pkgJSON := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(pkgJSON)}
	manifest, _ := BuildManifest(pkg, nil, false)
	manifest["react"][0].State = StateUpdated

	var buf bytes.Buffer

	err := RenderJSON(&buf, manifest, 1)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result jsonOutput

	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Top-level errors array must be nil/empty, not a populated array
	if len(result.Errors) != 0 {
		t.Errorf("expected no error entries, got %d", len(result.Errors))
	}
}

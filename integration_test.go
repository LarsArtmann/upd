package upd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullPipelineReadFetchWrite exercises the entire read→fetch→compare→write
// flow against a mock registry, verifying that the on-disk file is updated
// with byte-preserving edits.
func TestFullPipelineReadFetchWrite(t *testing.T) {
	originalJSON := `{
  "name": "my-project",
  "dependencies": {
    "react": "^18.0.0",
    "lodash": "4.17.20"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}
`

	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "package.json")

	err := os.WriteFile(pkgPath, []byte(originalJSON), 0o644)
	if err != nil {
		t.Fatalf("write temp package.json: %v", err)
	}

	// Mock registry
	registry := mockRegistry(map[string]map[string]any{
		"react":  packageVersion("19.0.0", "18.0.0", "19.0.0"),
		"lodash": packageVersion("4.17.21", "4.17.20", "4.17.21"),
		"jest":   packageVersion("30.0.0", "29.0.0", "30.0.0"),
	})
	t.Cleanup(registry.Close)

	// 1. Read package file
	pkg, err := ReadPackageFile(pkgPath)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	// 2. Build manifest (no patterns = check everything)
	manifest, warnings := BuildManifest(pkg, nil, false)
	if len(warnings) > 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	toCheck := manifest.ToCheck()
	if len(toCheck) != 3 {
		t.Fatalf("expected 3 packages to check, got %d", len(toCheck))
	}

	// 3. Set up engine
	cfg := DefaultConfig()
	cfg.Registry = registry.URL
	cfg.Retries = 0
	engine := NewEngine(cfg)

	// 4. Fetch + Apply
	results := engine.FetchAll(context.Background(), toCheck)
	updates, errCount := engine.ApplyUpdates(manifest, results, pkg)

	if errCount != 0 {
		t.Fatalf("expected 0 errors, got %d", errCount)
	}

	if updates != 3 {
		t.Fatalf("expected 3 updates, got %d", updates)
	}

	// 5. Write back
	err = pkg.Write(pkgPath)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// 6. Read the result and verify
	written, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	result := string(written)

	// New versions present
	assertContains(t, result, "19.0.0", "react new version")
	assertContains(t, result, "4.17.21", "lodash new version")
	assertContains(t, result, "30.0.0", "jest new version")

	// Old versions gone
	assertNotContains(t, result, "18.0.0", "react old version should be replaced")
	assertNotContains(t, result, "4.17.20", "lodash old version should be replaced")
	assertNotContains(t, result, "29.0.0", "jest old version should be replaced")

	// Formatting preserved: caret prefixes kept
	assertContains(t, result, "^19.0.0", "react caret preserved")
	assertContains(t, result, "^30.0.0", "jest caret preserved")

	// Key order and whitespace preserved
	assertContains(t, result, `"name": "my-project"`, "name field preserved")
	assertContains(t, result, "  \"devDependencies\":", "devDependencies indentation preserved")

	// Trailing newline preserved
	if !strings.HasSuffix(result, "\n") {
		t.Error("trailing newline should be preserved")
	}
}

// TestFullPipelineDryRunDoesNotWrite ensures that -n/--nop mode doesn't modify the file.
func TestFullPipelineDryRunDoesNotWrite(t *testing.T) {
	originalJSON := `{"dependencies": {"react": "^18.0.0"}}`

	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "package.json")

	err := os.WriteFile(pkgPath, []byte(originalJSON), 0o644)
	if err != nil {
		t.Fatalf("write temp package.json: %v", err)
	}

	registry := mockRegistry(map[string]map[string]any{
		"react": packageVersion("19.0.0", "19.0.0"),
	})
	t.Cleanup(registry.Close)

	pkg, err := ReadPackageFile(pkgPath)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	manifest, _ := BuildManifest(pkg, nil, false)

	cfg := DefaultConfig()
	cfg.Registry = registry.URL
	cfg.Retries = 0
	cfg.Nop = true
	engine := NewEngine(cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, _ := engine.ApplyUpdates(manifest, results, pkg)

	if updates != 1 {
		t.Fatalf("expected 1 update in dry-run, got %d", updates)
	}

	// File on disk must not have changed
	onDisk, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if string(onDisk) != originalJSON {
		t.Errorf("dry-run modified the file:\nwant: %q\ngot:  %q", originalJSON, string(onDisk))
	}
}

// TestScopedPackageURLEncoding verifies that scoped package names are correctly
// URL-encoded for NPM registry requests.
func TestScopedPackageURLEncoding(t *testing.T) {
	var requestedPath string

	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.EscapedPath()

		if strings.Contains(r.URL.Path, "@types/node") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"dist-tags":{"latest":"20.0.0"},"versions":{"20.0.0":{}}}`))

			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(registry.Close)

	cfg := DefaultConfig()
	cfg.Registry = registry.URL
	cfg.Retries = 0
	client := NewRegistryClient(cfg)

	pkg, _, err := client.FetchPackument(context.Background(), "@types/node")
	if err != nil {
		t.Fatalf("fetch scoped package failed: %v", err)
	}

	if pkg == nil {
		t.Fatal("expected non-nil packument")
	}

	// The scoped name must be path-escaped: %40 for @, %2F for /
	if !strings.Contains(requestedPath, "%2F") && !strings.Contains(requestedPath, "/") {
		t.Errorf("scoped package path unexpected: %s", requestedPath)
	}

	v, err := pkg.LatestVersion()
	if err != nil {
		t.Fatalf("LatestVersion: %v", err)
	}

	if v != "20.0.0" {
		t.Errorf("latest = %q, want 20.0.0", v)
	}
}

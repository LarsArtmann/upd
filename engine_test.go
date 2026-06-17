package upd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockRegistry(versions map[string]map[string]any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkgName := r.URL.Path[1:] // strip leading /

		data, ok := versions[pkgName]
		if !ok {
			data = versions["default"]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	}))
}

func packageVersion(latest string, versions ...string) map[string]any {
	versionMap := make(map[string]any, len(versions))
	for _, v := range versions {
		versionMap[v] = map[string]any{}
	}

	return map[string]any{
		"dist-tags": map[string]string{"latest": latest},
		"versions":  versionMap,
	}
}

func newTestEngine(t *testing.T, server *httptest.Server, cfg *Config) *Engine {
	t.Helper()

	engine := NewEngine(cfg)
	engine.registry.baseURL = server.URL

	return engine
}

func buildTestManifest(t *testing.T, json string) (*PackageFile, Manifest) {
	t.Helper()

	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	return pkg, manifest
}

func setupEngineTest(t *testing.T, json, name, latest string, versions ...string) (*Engine, *PackageFile, Manifest) {
	t.Helper()

	registry := mockRegistry(map[string]map[string]any{
		name: packageVersion(latest, versions...),
	})
	t.Cleanup(registry.Close)

	pkg, manifest := buildTestManifest(t, json)
	engine := newTestEngine(t, registry, DefaultConfig())

	return engine, pkg, manifest
}

func TestEngineFetchAllSuccess(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"react":  packageVersion("19.0.0", "18.0.0", "19.0.0"),
		"lodash": packageVersion("4.17.21", "4.17.20", "4.17.21"),
	})
	defer registry.Close()

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), []string{"react", "lodash"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	react := results["react"]
	if react.err != nil {
		t.Fatalf("react fetch error: %v", react.err)
	}

	v, err := react.pkg.LatestVersion()
	if err != nil {
		t.Fatalf("react latest error: %v", err)
	}

	if v != "19.0.0" {
		t.Errorf("react latest = %q, want 19.0.0", v)
	}
}

func TestEngineApplyUpdates(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"react":  packageVersion("19.0.0", "19.0.0"),
		"lodash": packageVersion("4.17.21", "4.17.21"),
	})
	defer registry.Close()

	pkg, manifest := buildTestManifest(t, `{
		"dependencies": {
			"react": "^18.0.0",
			"lodash": "4.17.20"
		}
	}`)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, errors := engine.ApplyUpdates(manifest, results, pkg)

	if errors != 0 {
		t.Errorf("expected 0 errors, got %d", errors)
	}

	if updates != 2 {
		t.Errorf("expected 2 updates, got %d", updates)
	}

	reactSpec := manifest["react"][0]
	if reactSpec.State != StateUpdated {
		t.Errorf("react state = %s, want updated", reactSpec.State)
	}

	if reactSpec.VNew != "19.0.0" {
		t.Errorf("react vNew = %q, want 19.0.0", reactSpec.VNew)
	}

	if reactSpec.SNew != "^19.0.0" {
		t.Errorf("react sNew = %q, want ^19.0.0", reactSpec.SNew)
	}
}

func TestEngineApplyUpdatesKept(t *testing.T) {
	engine, pkg, manifest := setupEngineTest(t, `{"dependencies": {"react": "^18.0.0"}}`, "react", "18.0.0", "18.0.0")

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, _ := engine.ApplyUpdates(manifest, results, pkg)

	if updates != 0 {
		t.Errorf("expected 0 updates for same version, got %d", updates)
	}

	if manifest["react"][0].State != StateKept {
		t.Errorf("react state = %s, want kept", manifest["react"][0].State)
	}
}

func TestEngineApplyUpdatesNop(t *testing.T) {
	engine, pkg, manifest := setupEngineTest(t, `{"dependencies": {"react": "^18.0.0"}}`, "react", "19.0.0", "19.0.0")
	engine.cfg.Nop = true

	json := `{"dependencies": {"react": "^18.0.0"}}`
	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, _ := engine.ApplyUpdates(manifest, results, pkg)

	if updates != 1 {
		t.Errorf("expected 1 update in nop mode, got %d", updates)
	}
	// package.json should NOT be modified in nop mode
	if string(pkg.raw) != json {
		t.Errorf("package.json was modified in nop mode")
	}
}

func TestEngineApplyUpdatesError(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer registry.Close()

	pkg, manifest := buildTestManifest(t, `{"dependencies": {"nonexistent": "^1.0.0"}}`)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	_, errors := engine.ApplyUpdates(manifest, results, pkg)

	if errors == 0 {
		t.Error("expected errors for 404 package, got 0")
	}

	if manifest["nonexistent"][0].State != StateError {
		t.Errorf("state = %s, want error", manifest["nonexistent"][0].State)
	}
}

func TestEngineApplyUpdatesWritesPackageJSON(t *testing.T) {
	engine, pkg, manifest := setupEngineTest(t, `{"dependencies": {"react": "^18.0.0"}}`, "react", "19.0.0", "19.0.0")

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	engine.ApplyUpdates(manifest, results, pkg)

	result := string(pkg.raw)
	assertContains(t, result, "19.0.0", "new version")
	assertNotContains(t, result, "18.0.0", "old version")
}

func TestEngineGreatestMode(t *testing.T) {
	engine, pkg, manifest := setupEngineTest(
		t,
		`{"dependencies": {"react": "^18.0.0"}}`,
		"react", "18.0.0",
		"18.0.0", "19.0.0-beta.1", "19.0.0",
	)
	engine.cfg.Greatest = true

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	engine.ApplyUpdates(manifest, results, pkg)

	if manifest["react"][0].VNew != "19.0.0" {
		t.Errorf("greatest vNew = %q, want 19.0.0", manifest["react"][0].VNew)
	}
}

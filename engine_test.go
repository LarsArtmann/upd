package upd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newTestEngine(t *testing.T, server *httptest.Server, cfg *Config) *Engine {
	t.Helper()

	engine := NewEngine(cfg)
	engine.registry.baseURL = server.URL

	return engine
}

func TestEngineFetchAllSuccess(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"react": {
			"dist-tags": map[string]string{"latest": "19.0.0"},
			"versions": map[string]any{
				"18.0.0": map[string]any{},
				"19.0.0": map[string]any{},
			},
		},
		"lodash": {
			"dist-tags": map[string]string{"latest": "4.17.21"},
			"versions": map[string]any{
				"4.17.20": map[string]any{},
				"4.17.21": map[string]any{},
			},
		},
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
		"react": {
			"dist-tags": map[string]string{"latest": "19.0.0"},
			"versions":  map[string]any{"19.0.0": map[string]any{}},
		},
		"lodash": {
			"dist-tags": map[string]string{"latest": "4.17.21"},
			"versions":  map[string]any{"4.17.21": map[string]any{}},
		},
	})
	defer registry.Close()

	json := `{
		"dependencies": {
			"react": "^18.0.0",
			"lodash": "4.17.20"
		}
	}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	toCheck := manifest.ToCheck()
	results := engine.FetchAll(context.Background(), toCheck)
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
	registry := mockRegistry(map[string]map[string]any{
		"react": {
			"dist-tags": map[string]string{"latest": "18.0.0"},
			"versions":  map[string]any{"18.0.0": map[string]any{}},
		},
	})
	defer registry.Close()

	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

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
	registry := mockRegistry(map[string]map[string]any{
		"react": {
			"dist-tags": map[string]string{"latest": "19.0.0"},
			"versions":  map[string]any{"19.0.0": map[string]any{}},
		},
	})
	defer registry.Close()

	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	cfg := DefaultConfig()
	cfg.Nop = true
	engine := newTestEngine(t, registry, cfg)

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

	json := `{"dependencies": {"nonexistent": "^1.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

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
	registry := mockRegistry(map[string]map[string]any{
		"react": {
			"dist-tags": map[string]string{"latest": "19.0.0"},
			"versions":  map[string]any{"19.0.0": map[string]any{}},
		},
	})
	defer registry.Close()

	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	engine.ApplyUpdates(manifest, results, pkg)

	result := string(pkg.raw)
	if !strings.Contains(result, "19.0.0") {
		t.Errorf("expected 19.0.0 in result after apply:\n%s", result)
	}

	if strings.Contains(result, "18.0.0") {
		t.Errorf("old version 18.0.0 still present:\n%s", result)
	}
}

func TestEngineGreatestMode(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"react": {
			"dist-tags": map[string]string{"latest": "18.0.0"},
			"versions": map[string]any{
				"18.0.0":        map[string]any{},
				"19.0.0-beta.1": map[string]any{},
				"19.0.0":        map[string]any{},
			},
		},
	})
	defer registry.Close()

	json := `{"dependencies": {"react": "^18.0.0"}}`
	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	cfg := DefaultConfig()
	cfg.Greatest = true
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	engine.ApplyUpdates(manifest, results, pkg)

	if manifest["react"][0].VNew != "19.0.0" {
		t.Errorf("greatest vNew = %q, want 19.0.0", manifest["react"][0].VNew)
	}
}

package upd

import (
	"context"
	"encoding/json/v2"
	"errors"
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
		_ = json.MarshalWrite(w, data)
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
	manifest, _ := BuildManifest(pkg, nil, false)

	return pkg, manifest
}

func setupEngineTest(t *testing.T, latest string, versions ...string) (*Engine, *PackageFile, Manifest) {
	t.Helper()

	registry := mockRegistry(map[string]map[string]any{
		testPackageName: packageVersion(latest, versions...),
	})
	t.Cleanup(registry.Close)

	pkg, manifest := buildTestManifest(t, singleDependencyPackageJSON)
	engine := newTestEngine(t, registry, DefaultConfig())

	return engine, pkg, manifest
}

const (
	testPackageName             = "react"
	singleDependencyPackageJSON = `{"dependencies": {"react": "^18.0.0"}}`
)

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
	engine, pkg, manifest := setupEngineTest(t, "18.0.0", "18.0.0")

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
	engine, pkg, manifest := setupEngineTest(t, "19.0.0", "19.0.0")
	engine.cfg.Nop = true

	originalJSON := `{"dependencies": {"react": "^18.0.0"}}`
	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, _ := engine.ApplyUpdates(manifest, results, pkg)

	if updates != 1 {
		t.Errorf("expected 1 update in nop mode, got %d", updates)
	}
	// package.json should NOT be modified in nop mode
	if string(pkg.raw) != originalJSON {
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
	engine, pkg, manifest := setupEngineTest(t, "19.0.0", "19.0.0")

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	engine.ApplyUpdates(manifest, results, pkg)

	result := string(pkg.raw)
	assertContains(t, result, "19.0.0", "new version")
	assertNotContains(t, result, "18.0.0", "old version")
}

func TestEngineGreatestMode(t *testing.T) {
	engine, pkg, manifest := setupEngineTest(
		t,
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

func TestEnginePinLatest(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"semver": packageVersion("7.7.4", "7.7.4"),
	})
	defer registry.Close()

	pkgJSON := `{"dependencies": {"semver": "latest"}}`
	pkg := &PackageFile{raw: []byte(pkgJSON)}
	manifest, _ := BuildManifest(pkg, nil, true)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, errors := engine.ApplyUpdates(manifest, results, pkg)

	if errors != 0 {
		t.Fatalf("expected 0 errors, got %d", errors)
	}

	if updates != 1 {
		t.Fatalf("expected 1 update, got %d", updates)
	}

	spec := manifest["semver"][0]
	if spec.State != StateUpdated {
		t.Errorf("state = %s, want updated", spec.State)
	}

	if spec.VNew != "7.7.4" {
		t.Errorf("vNew = %q, want 7.7.4", spec.VNew)
	}

	if spec.SNew != "7.7.4" {
		t.Errorf("sNew = %q, want 7.7.4", spec.SNew)
	}

	if !spec.IsLatest {
		t.Error("IsLatest = false, want true")
	}

	result := string(pkg.raw)
	assertContains(t, result, "7.7.4", "pinned version in package.json")
	assertNotContains(t, result, "latest", "old latest tag removed")
}

func TestEnginePinLatestDisabled(t *testing.T) {
	registry := mockRegistry(map[string]map[string]any{
		"semver": packageVersion("7.7.4", "7.7.4"),
	})
	defer registry.Close()

	pkgJSON := `{"dependencies": {"semver": "latest"}}`
	pkg := &PackageFile{raw: []byte(pkgJSON)}
	manifest, _ := BuildManifest(pkg, nil, false)

	cfg := DefaultConfig()
	engine := newTestEngine(t, registry, cfg)

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, errors := engine.ApplyUpdates(manifest, results, pkg)

	if errors != 0 || updates != 0 {
		t.Fatalf("expected no updates/errors for skipped latest, got %d/%d", updates, errors)
	}

	if manifest["semver"][0].State != StateSkipped {
		t.Errorf("state = %s, want skipped", manifest["semver"][0].State)
	}
}

func TestRegistryClassifiesNotFoundAsRejection(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer registry.Close()

	engine := newTestEngine(t, registry, DefaultConfig())
	_, _, err := engine.registry.FetchPackument(context.Background(), "ghost")

	if !errors.Is(err, ErrPackageNotFound) {
		t.Errorf("404 should wrap ErrPackageNotFound, got: %v", err)
	}

	if errors.Is(err, ErrRegistryUnavailable) {
		t.Errorf("404 must NOT wrap ErrRegistryUnavailable, got: %v", err)
	}
}

func TestRegistryClassifiesServerErrorAsTransient(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))
			defer registry.Close()

			engine := newTestEngine(t, registry, DefaultConfig())
			_, _, err := engine.registry.FetchPackument(context.Background(), "react")

			if !errors.Is(err, ErrRegistryUnavailable) {
				t.Errorf("status %d should wrap ErrRegistryUnavailable, got: %v", status, err)
			}
		})
	}
}

func TestApplyUpdatesPopulatesSpecErr(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer registry.Close()

	pkg, manifest := buildTestManifest(t, `{"dependencies": {"ghost": "^1.0.0"}}`)
	engine := newTestEngine(t, registry, DefaultConfig())

	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	_, errCount := engine.ApplyUpdates(manifest, results, pkg)

	if errCount == 0 {
		t.Fatal("expected errors for 404, got 0")
	}

	spec := manifest["ghost"][0]
	if spec.Err == nil {
		t.Fatal("expected spec.Err to be populated, got nil")
	}

	if !errors.Is(spec.Err, ErrPackageNotFound) {
		t.Errorf("spec.Err should wrap ErrPackageNotFound, got: %v", spec.Err)
	}
}

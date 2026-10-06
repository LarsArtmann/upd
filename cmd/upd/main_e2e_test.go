package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/upd"
	errorfamily "github.com/larsartmann/go-error-family"
)

// newE2ERegistry spins up a mock NPM registry serving one packument document
// per package name; unknown names get a plain 404 (non-retryable).
func newE2ERegistry(t *testing.T, packuments map[string]string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")

		packument, ok := packuments[name]
		if !ok {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(packument)) //nolint:erraudit // test server: client cancellation is not actionable
	}))
	t.Cleanup(server.Close)

	return server
}

func writeE2EPackage(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
}

func readE2EPackage(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}

	return string(data)
}

const (
	e2ePackument = `{"name":"left-pad","dist-tags":{"latest":"2.5.0"},"versions":{"1.0.0":{},"2.5.0":{}}}`

	e2ePkgBefore = "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.0.0\"\n  }\n}\n"
	e2ePkgAfter  = "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^2.5.0\"\n  }\n}\n"
)

func TestRunEEndToEndUpdatesAndPreservesFormatting(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--no-color"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if got := readE2EPackage(t, file); got != e2ePkgAfter {
		t.Errorf("package.json not updated byte-preserved:\nwant %q\ngot  %q", e2ePkgAfter, got)
	}

	out := stdout.String()
	if !strings.Contains(out, "left-pad") || !strings.Contains(out, "2.5.0") {
		t.Errorf("stdout missing update info: %q", out)
	}
}

func TestRunEEndToEndDryRunLeavesFileUntouched(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--no-color", "-n"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if got := readE2EPackage(t, file); got != e2ePkgBefore {
		t.Errorf("dry run modified package.json:\nwant %q\ngot  %q", e2ePkgBefore, got)
	}
}

func TestRunEEndToEndPartialFailureWritesUpdatesAndExitsOne(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")

	pkgBefore := "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.0.0\",\n    \"ghost-pkg\": \"^1.0.0\"\n  }\n}\n"
	pkgAfter := "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^2.5.0\",\n    \"ghost-pkg\": \"^1.0.0\"\n  }\n}\n"
	writeE2EPackage(t, file, pkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--no-color"}, &stdout, &stderr)
	if !errors.Is(err, upd.ErrPartialFailure) {
		t.Fatalf("expected ErrPartialFailure, got %v", err)
	}

	if code := errorfamily.ExitCode(err); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if got := readE2EPackage(t, file); got != pkgAfter {
		t.Errorf("successful update lost on partial failure:\nwant %q\ngot  %q", pkgAfter, got)
	}
}

func TestRunEEndToEndTotalFailureExitsOne(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--no-color"}, &stdout, &stderr)
	if !errors.Is(err, upd.ErrPartialFailure) {
		t.Fatalf("expected ErrPartialFailure, got %v", err)
	}

	if code := errorfamily.ExitCode(err); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if got := readE2EPackage(t, file); got != e2ePkgBefore {
		t.Errorf("failed run must not touch package.json:\nwant %q\ngot  %q", e2ePkgBefore, got)
	}
}

func TestRunEEndToEndJSONOutput(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--format=json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	var out struct {
		Summary struct {
			Updated int `json:"updated"`
			Errors  int `json:"errors"`
			Total   int `json:"total"`
		} `json:"summary"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}

	if out.Summary.Updated != 1 || out.Summary.Total != 1 || out.Summary.Errors != 0 {
		t.Errorf("unexpected JSON summary: %+v", out.Summary)
	}
}

func TestRunEDeprecatedJSONFlagWarnsAndEmitsJSON(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if !strings.Contains(stderr.String(), "--json is deprecated") {
		t.Errorf("stderr missing deprecation warning: %q", stderr.String())
	}

	var out struct {
		Summary struct {
			Updated int `json:"updated"`
		} `json:"summary"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("--json alias did not emit JSON: %v\n%s", err, stdout.String())
	}

	if out.Summary.Updated != 1 {
		t.Errorf("unexpected summary via --json alias: %+v", out.Summary)
	}
}

func TestRunESilentAliasSuppressesOutput(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--silent"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if stdout.Len() != 0 {
		t.Errorf("--silent wrote to stdout: %q", stdout.String())
	}
}

func TestRunEInvalidFormatFailsBeforeFetch(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--format=yaml"}, &stdout, &stderr)
	if !errors.Is(err, upd.ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}

	if !strings.Contains(err.Error(), "config.invalid_format") {
		t.Errorf("error should carry the invalid-format code, got: %v", err)
	}
}

func TestRunEEndToEndPositionalPatternFiltersUpdates(t *testing.T) {
	t.Parallel()

	rightPadPackument := `{"name":"right-pad","dist-tags":{"latest":"9.9.9"},"versions":{"9.9.9":{}}}`
	server := newE2ERegistry(t, map[string]string{
		"left-pad":  e2ePackument,
		"right-pad": rightPadPackument,
	})
	file := filepath.Join(t.TempDir(), "package.json")

	pkgBefore := "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.0.0\",\n    \"right-pad\": \"^1.0.0\"\n  }\n}\n"
	pkgAfter := "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^2.5.0\",\n    \"right-pad\": \"^1.0.0\"\n  }\n}\n"
	writeE2EPackage(t, file, pkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"left-pad*", "--registry", server.URL, "--file", file, "--no-color"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if got := readE2EPackage(t, file); got != pkgAfter {
		t.Errorf("positional pattern did not filter updates:\nwant %q\ngot  %q", pkgAfter, got)
	}
}

func TestRunEUnknownFlagWithPositionalArgSuggestsFlag(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--jso", "-f", file, "--registry", server.URL}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected an error for the unknown --jso flag")
	}

	if !strings.Contains(err.Error(), "Did you mean --format=json?") {
		t.Errorf("error missing flag suggestion: %v", err)
	}

	if got := readE2EPackage(t, file); got != e2ePkgBefore {
		t.Errorf("failed parse must not touch package.json:\nwant %q\ngot  %q", e2ePkgBefore, got)
	}
}

func TestRunEEndToEndQuietSuppressesOutput(t *testing.T) {
	t.Parallel()

	server := newE2ERegistry(t, map[string]string{"left-pad": e2ePackument})
	file := filepath.Join(t.TempDir(), "package.json")
	writeE2EPackage(t, file, e2ePkgBefore)

	var stdout, stderr bytes.Buffer
	err := runE([]string{"--registry", server.URL, "--file", file, "--quiet"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}

	if stdout.Len() != 0 {
		t.Errorf("quiet mode wrote to stdout: %q", stdout.String())
	}

	if got := readE2EPackage(t, file); got != e2ePkgAfter {
		t.Errorf("quiet mode still updates the file:\nwant %q\ngot  %q", e2ePkgAfter, got)
	}
}

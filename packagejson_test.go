package upd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func updateDep(t *testing.T, pkg *PackageFile, section, dep, version, label string) {
	t.Helper()

	err := pkg.UpdateDependency(section, dep, version)
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
}

func TestUpdateDependencyPreservesFormatting(t *testing.T) {
	original := `{
  "name": "test",
  "dependencies": {
    "react": "^18.0.0",
    "lodash": "4.17.21"
  }
}`

	pkg := &PackageFile{raw: []byte(original)}

	err := pkg.UpdateDependency("dependencies", "react", "^19.0.0")
	if err != nil {
		t.Fatalf("UpdateDependency failed: %v", err)
	}

	result := string(pkg.raw)

	assertContains(t, result, `"^19.0.0"`, "new react version")
	assertContains(t, result, `"4.17.21"`, "lodash version")
	assertContains(t, result, "  \"dependencies\"", "formatting")

	// The old value should NOT be present
	assertNotContains(t, result, `"^18.0.0"`, "old react version")
}

func TestUpdateDependencyMultipleSections(t *testing.T) {
	original := `{
  "dependencies": {
    "react": "^18.0.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`

	pkg := &PackageFile{raw: []byte(original)}

	updateDep(t, pkg, "dependencies", "react", "^19.0.0", "first update")
	updateDep(t, pkg, "devDependencies", "jest", "^30.0.0", "second update")

	result := string(pkg.raw)
	assertContains(t, result, `"^19.0.0"`, "react version")
	assertContains(t, result, `"^30.0.0"`, "jest version")
}

func TestUpdateDependencyNotFound(t *testing.T) {
	pkg := &PackageFile{raw: []byte(`{"dependencies": {"react": "^18.0.0"}}`)}

	err := pkg.UpdateDependency("dependencies", "nonexistent", "^1.0.0")
	if err == nil {
		t.Fatal("expected error for nonexistent dependency")
	}
}

func TestGetUpdArgs(t *testing.T) {
	t.Run("array form", func(t *testing.T) {
		pkg := &PackageFile{raw: []byte(`{"upd": ["react*", "!react-dom"], "dependencies": {}}`)}

		args, err := pkg.GetUpdArgs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(args))
		}

		if args[0] != "react*" || args[1] != "!react-dom" {
			t.Errorf("args = %v", args)
		}
	})

	t.Run("string form", func(t *testing.T) {
		pkg := &PackageFile{raw: []byte(`{"upd": "react*", "dependencies": {}}`)}

		args, err := pkg.GetUpdArgs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(args) != 1 || args[0] != "react*" {
			t.Errorf("args = %v", args)
		}
	})

	t.Run("missing", func(t *testing.T) {
		pkg := &PackageFile{raw: []byte(`{"dependencies": {}}`)}

		args, err := pkg.GetUpdArgs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(args) != 0 {
			t.Errorf("expected no args, got %v", args)
		}
	})
}

func TestGetDependencySection(t *testing.T) {
	t.Parallel()

	t.Run("valid section returns deps", func(t *testing.T) {
		t.Parallel()

		pkg := &PackageFile{raw: []byte(`{"dependencies": {"react": "^18.0.0", "vue": "^3.0.0"}}`)}

		deps, err := pkg.GetDependencySection("dependencies")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps["react"] != "^18.0.0" {
			t.Errorf("react = %q, want ^18.0.0", deps["react"])
		}

		if deps["vue"] != "^3.0.0" {
			t.Errorf("vue = %q, want ^3.0.0", deps["vue"])
		}
	})

	t.Run("missing section returns empty without error", func(t *testing.T) {
		t.Parallel()

		pkg := &PackageFile{raw: []byte(`{"name": "my-project"}`)}

		deps, err := pkg.GetDependencySection("dependencies")
		if err != nil {
			t.Fatalf("unexpected error for missing section: %v", err)
		}

		if len(deps) != 0 {
			t.Errorf("expected empty map, got %v", deps)
		}
	})

	t.Run("non-object section returns ErrInvalidJSON", func(t *testing.T) {
		t.Parallel()

		pkg := &PackageFile{raw: []byte(`{"dependencies": 42}`)}

		_, err := pkg.GetDependencySection("dependencies")
		if err == nil {
			t.Fatal("expected error for non-object section, got nil")
		}

		if !errors.Is(err, ErrInvalidJSON) {
			t.Errorf("expected error wrapping ErrInvalidJSON, got: %v", err)
		}
	})

	t.Run("non-string values return parse error", func(t *testing.T) {
		t.Parallel()

		pkg := &PackageFile{raw: []byte(`{"dependencies": {"react": 123}}`)}

		_, err := pkg.GetDependencySection("dependencies")
		if err == nil {
			t.Fatal("expected error for non-string dependency value, got nil")
		}
	})
}

// testFilePermissions is the mode used for package.json fixtures in the
// round-trip tests below. It is a named constant so the mnd linter does not
// flag it as a magic number, and it deliberately differs from 0o600 so the
// permission-preservation test proves upd does not downgrade an existing file.
const testFilePermissions os.FileMode = 0o644

func writePackageFixture(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), testFilePermissions)
	if err != nil {
		t.Fatalf("write test fixture %q: %v", path, err)
	}
}

func TestReadPackageFileComputesFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.json")
	data := `{"dependencies":{"react":"^18.0.0"}}`
	writePackageFixture(t, path, data)

	pkg, err := ReadPackageFile(path)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	if pkg.fingerprint.IsZero() {
		t.Fatal("fingerprint was not computed at read time")
	}

	if !pkg.fingerprint.Matches([]byte(data)) {
		t.Fatal("fingerprint does not match the original file content")
	}
}

func TestWriteRoundTripPreservesContentAndFormatting(t *testing.T) {
	original := `{
  "name": "test",
  "dependencies": {
    "react": "^18.0.0",
    "lodash": "4.17.21"
  }
}`
	path := filepath.Join(t.TempDir(), "package.json")
	writePackageFixture(t, path, original)

	pkg, err := ReadPackageFile(path)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	err = pkg.UpdateDependency("dependencies", "react", "^19.0.0")
	if err != nil {
		t.Fatalf("UpdateDependency: %v", err)
	}

	err = pkg.Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	result := string(written)
	assertContains(t, result, `"^19.0.0"`, "new version persisted")
	assertNotContains(t, result, `"^18.0.0"`, "old version removed")
	assertContains(t, result, `"4.17.21"`, "untouched dependency preserved")
	assertContains(t, result, "  \"name\"", "original indentation preserved")
}

func TestWritePreservesPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.json")
	writePackageFixture(t, path, `{"dependencies":{"react":"^18.0.0"}}`)

	pkg, err := ReadPackageFile(path)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	err = pkg.UpdateDependency("dependencies", "react", "^19.0.0")
	if err != nil {
		t.Fatalf("UpdateDependency: %v", err)
	}

	err = pkg.Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if mode := info.Mode().Perm(); mode != testFilePermissions {
		t.Errorf("permissions = %o, want %o", mode, testFilePermissions)
	}
}

func TestWriteRejectsConcurrentModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.json")
	writePackageFixture(t, path, `{"dependencies":{"react":"^18.0.0"}}`)

	pkg, err := ReadPackageFile(path)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	// Simulate another process (npm install, IDE formatter) editing the file
	// after upd read it but before upd writes it back.
	writePackageFixture(t, path, `{"dependencies":{"react":"^20.0.0"}}`)

	err = pkg.Write(path)
	if !errors.Is(err, ErrConcurrentModification) {
		t.Fatalf("expected ErrConcurrentModification, got %v", err)
	}

	// The on-disk file must be left exactly as the concurrent editor left it;
	// upd must not overwrite it with its now-stale in-memory bytes.
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	assertContains(t, string(written), `"^20.0.0"`, "concurrent editor's content preserved")
	assertNotContains(t, string(written), `"^18.0.0"`, "upd did not clobber with stale data")
}

func TestWriteLeavesNoLeftoverFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.json")
	writePackageFixture(t, path, `{"dependencies":{"react":"^18.0.0"}}`)

	pkg, err := ReadPackageFile(path)
	if err != nil {
		t.Fatalf("ReadPackageFile: %v", err)
	}

	err = pkg.UpdateDependency("dependencies", "react", "^19.0.0")
	if err != nil {
		t.Fatalf("UpdateDependency: %v", err)
	}

	err = pkg.Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// The library no longer creates .bak files; verify nothing is left behind.
	matches, globErr := filepath.Glob(path + ".*")
	if globErr != nil {
		t.Fatalf("glob: %v", globErr)
	}

	if len(matches) > 0 {
		t.Errorf("expected no leftover files beside %q, found: %v", path, matches)
	}
}

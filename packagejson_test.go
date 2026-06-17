package upd

import (
	"strings"
	"testing"
)

func assertContains(t *testing.T, haystack, needle, label string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected %q in:\n%s", label, needle, haystack)
	}
}

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
	if strings.Contains(result, `"^18.0.0"`) {
		t.Errorf("old value ^18.0.0 still present:\n%s", result)
	}
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

		args := pkg.GetUpdArgs()
		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(args))
		}

		if args[0] != "react*" || args[1] != "!react-dom" {
			t.Errorf("args = %v", args)
		}
	})

	t.Run("string form", func(t *testing.T) {
		pkg := &PackageFile{raw: []byte(`{"upd": "react*", "dependencies": {}}`)}

		args := pkg.GetUpdArgs()
		if len(args) != 1 || args[0] != "react*" {
			t.Errorf("args = %v", args)
		}
	})

	t.Run("missing", func(t *testing.T) {
		pkg := &PackageFile{raw: []byte(`{"dependencies": {}}`)}

		args := pkg.GetUpdArgs()
		if len(args) != 0 {
			t.Errorf("expected no args, got %v", args)
		}
	})
}

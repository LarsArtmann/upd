package internal

import (
	"strings"
	"testing"
)

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

	// The new value should be present
	if !strings.Contains(result, `"^19.0.0"`) {
		t.Errorf("expected ^19.0.0 in result:\n%s", result)
	}

	// The old value should NOT be present
	if strings.Contains(result, `"^18.0.0"`) {
		t.Errorf("old value ^18.0.0 still present:\n%s", result)
	}

	// Other deps should be unchanged
	if !strings.Contains(result, `"4.17.21"`) {
		t.Errorf("lodash version changed:\n%s", result)
	}

	// Formatting should be preserved (2-space indent)
	if !strings.Contains(result, "  \"dependencies\"") {
		t.Errorf("formatting lost:\n%s", result)
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

	if err := pkg.UpdateDependency("dependencies", "react", "^19.0.0"); err != nil {
		t.Fatalf("first update failed: %v", err)
	}
	if err := pkg.UpdateDependency("devDependencies", "jest", "^30.0.0"); err != nil {
		t.Fatalf("second update failed: %v", err)
	}

	result := string(pkg.raw)
	if !strings.Contains(result, `"^19.0.0"`) {
		t.Errorf("react not updated:\n%s", result)
	}
	if !strings.Contains(result, `"^30.0.0"`) {
		t.Errorf("jest not updated:\n%s", result)
	}
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

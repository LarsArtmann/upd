package upd

import (
	"testing"
)

func TestVersionRegex(t *testing.T) {
	tests := []struct {
		input string
		match bool
		ver   string
	}{
		{"1.0.0", true, "1.0.0"},
		{"^1.0.0", true, "1.0.0"},
		{"~2.3.4", true, "2.3.4"},
		{"^ 1.0.0", true, "1.0.0"},
		{">=1.0.0", false, ""},
		{"1.x", true, "1.x"},
		{"latest", false, ""},
		{"git://github.com/foo/bar.git", false, ""},
		{"file:../local-pkg", false, ""},
		{"1.0.0-beta.1", true, "1.0.0-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := versionRe.FindStringSubmatch(tt.input)
			if !tt.match {
				if m != nil {
					t.Fatalf("expected no match for %q, got %v", tt.input, m)
				}

				return
			}

			if m == nil {
				t.Fatalf("expected match for %q", tt.input)
			}

			if m[1] != tt.ver {
				t.Errorf("version = %q, want %q", m[1], tt.ver)
			}
		})
	}
}

func TestMatchesPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		expect   bool
	}{
		{"react", nil, true},
		{"react", []string{}, true},
		{"react", []string{"react"}, true},
		{"react-dom", []string{"react*"}, true},
		{"react-dom", []string{"react"}, false},
		{"react-dom", []string{"react", "!react-dom"}, false},
		{"vue", []string{"react", "!react-dom"}, false},
		{"lodash", []string{"!*-dom"}, true},
		{"react-dom", []string{"!*-dom"}, false},
		{"@scope/pkg", []string{"@scope/*"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPatterns(tt.name, tt.patterns)
			if got != tt.expect {
				t.Errorf("matchesPatterns(%q, %v) = %v, want %v", tt.name, tt.patterns, got, tt.expect)
			}
		})
	}
}

func TestBuildManifest(t *testing.T) {
	json := `{
		"dependencies": {
			"react": "^18.0.0",
			"lodash": "4.17.21",
			"local-pkg": "file:../local-pkg"
		},
		"devDependencies": {
			"jest": "^29.0.0"
		}
	}`

	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)

	if len(manifest) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(manifest))
	}

	reactSpecs := manifest["react"]
	if len(reactSpecs) != 1 {
		t.Fatalf("expected 1 spec for react, got %d", len(reactSpecs))
	}

	react := reactSpecs[0]
	if react.Section != "dependencies" {
		t.Errorf("section = %q, want dependencies", react.Section)
	}

	if react.SOld != "^18.0.0" {
		t.Errorf("sOld = %q, want ^18.0.0", react.SOld)
	}

	if react.VOld != "18.0.0" {
		t.Errorf("vOld = %q, want 18.0.0", react.VOld)
	}

	if react.State != StateCheck {
		t.Errorf("state = %q, want check", react.State)
	}

	localSpecs := manifest["local-pkg"]
	if len(localSpecs) != 1 {
		t.Fatalf("expected 1 spec for local-pkg, got %d", len(localSpecs))
	}

	if localSpecs[0].State != StateSkipped {
		t.Errorf("local-pkg state = %q, want skipped", localSpecs[0].State)
	}
}

func TestBuildManifestWithPatterns(t *testing.T) {
	json := `{
		"dependencies": {
			"react": "^18.0.0",
			"react-dom": "^18.0.0",
			"vue": "^3.0.0"
		}
	}`

	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, []string{"react*", "!react-dom"})

	cases := []struct {
		pkg   string
		state State
	}{
		{"react", StateCheck},
		{"react-dom", StateIgnored},
		{"vue", StateIgnored},
	}
	for _, c := range cases {
		if got := manifest[c.pkg][0].State; got != c.state {
			t.Errorf("%s state = %s, want %s", c.pkg, got, c.state)
		}
	}
}

func TestManifestToCheck(t *testing.T) {
	json := `{
		"dependencies": {
			"react": "^18.0.0",
			"local": "file:../local"
		}
	}`

	pkg := &PackageFile{raw: []byte(json)}
	manifest := BuildManifest(pkg, nil)
	toCheck := manifest.ToCheck()

	if len(toCheck) != 1 {
		t.Fatalf("expected 1 to check, got %d", len(toCheck))
	}

	if toCheck[0] != "react" {
		t.Errorf("expected react, got %s", toCheck[0])
	}
}

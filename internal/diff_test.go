package internal

import "testing"

func TestDiffCharsEqual(t *testing.T) {
	chunks := diffChars("1.0.0", "1.0.0")
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].op != opEqual {
		t.Errorf("expected opEqual, got %d", chunks[0].op)
	}
	if chunks[0].text != "1.0.0" {
		t.Errorf("text = %q", chunks[0].text)
	}
}

func TestDiffCharsReplace(t *testing.T) {
	// ^1.0.0 → ^2.0.0: common prefix "^", replaced "1.0.0" vs "2.0.0"
	chunks := diffChars("^1.0.0", "^2.0.0")

	// Expected: EQUAL "^", DELETE "1.0.0", INSERT "2.0.0"
	// But LCS may produce: EQUAL "^", DELETE "1", INSERT "2", EQUAL ".0.0"
	// Just verify total reconstructed strings are correct
	var oldStr, newStr string
	for _, c := range chunks {
		switch c.op {
		case opEqual:
			oldStr += c.text
			newStr += c.text
		case opDelete:
			oldStr += c.text
		case opInsert:
			newStr += c.text
		}
	}
	if oldStr != "^1.0.0" {
		t.Errorf("oldStr = %q, want ^1.0.0", oldStr)
	}
	if newStr != "^2.0.0" {
		t.Errorf("newStr = %q, want ^2.0.0", newStr)
	}
}

func TestDiffCharsInsert(t *testing.T) {
	// "1.0" → "1.0.0": trailing insert
	chunks := diffChars("1.0", "1.0.0")

	var oldStr, newStr string
	for _, c := range chunks {
		switch c.op {
		case opEqual:
			oldStr += c.text
			newStr += c.text
		case opDelete:
			oldStr += c.text
		case opInsert:
			newStr += c.text
		}
	}
	if oldStr != "1.0" {
		t.Errorf("oldStr = %q, want 1.0", oldStr)
	}
	if newStr != "1.0.0" {
		t.Errorf("newStr = %q, want 1.0.0", newStr)
	}
}

func TestReplaceVersion(t *testing.T) {
	tests := []struct {
		sOld, vOld, vNew, want string
	}{
		{"^1.0.0", "1.0.0", "2.0.0", "^2.0.0"},
		{"~1.2.3", "1.2.3", "1.2.4", "~1.2.4"},
		{"1.0.0", "1.0.0", "2.0.0", "2.0.0"},
		{">=1.0.0", "1.0.0", "2.0.0", ">=2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.sOld, func(t *testing.T) {
			got := replaceVersion(tt.sOld, tt.vOld, tt.vNew)
			if got != tt.want {
				t.Errorf("replaceVersion(%q, %q, %q) = %q, want %q",
					tt.sOld, tt.vOld, tt.vNew, got, tt.want)
			}
		})
	}
}

func TestResolveGreatest(t *testing.T) {
	json := `{
		"dist-tags": {"latest": "1.0.0"},
		"versions": {
			"1.0.0": {},
			"2.0.0-beta.1": {},
			"1.5.0": {}
		}
	}`

	pkg := &Packument{raw: []byte(json)}

	// Greatest should be 2.0.0-beta.1 (semver pre-release > release? No! 2.0.0-beta.1 < 1.5.0 in semver!)
	// Actually: 1.5.0 > 2.0.0-beta.1 because 2.0.0-beta.1 is a pre-release of 2.0.0
	// Semver: 2.0.0-beta.1 < 2.0.0 < ... but we only have 2.0.0-beta.1, 1.0.0, 1.5.0
	// 1.5.0 > 2.0.0-beta.1 because pre-release of higher version < release of lower? NO!
	// semver comparison: 2.0.0-beta.1 vs 1.5.0
	// Major: 2 > 1, so 2.0.0-beta.1 > 1.5.0
	got, err := pkg.GreatestVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2.0.0-beta.1" {
		t.Errorf("GreatestVersion() = %q, want 2.0.0-beta.1", got)
	}
}

func TestResolveLatest(t *testing.T) {
	json := `{
		"dist-tags": {"latest": "3.1.0"},
		"versions": {"3.1.0": {}}
	}`

	pkg := &Packument{raw: []byte(json)}
	got, err := pkg.LatestVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "3.1.0" {
		t.Errorf("LatestVersion() = %q, want 3.1.0", got)
	}
}

func TestResolveLatestMissing(t *testing.T) {
	pkg := &Packument{raw: []byte(`{"versions": {}}`)}
	_, err := pkg.LatestVersion()
	if err == nil {
		t.Fatal("expected error for missing dist-tags.latest")
	}
}

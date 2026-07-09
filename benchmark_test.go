package upd

import (
	"strings"
	"testing"
)

func BenchmarkDiffChars(b *testing.B) {
	benchmarks := []struct {
		name string
		old  string
		new  string
	}{
		{"same", "1.2.3", "1.2.3"},
		{"minor", "^18.0.0", "^19.0.0"},
		{"major", "1.0.0", "2.0.0"},
		{"prerelease", "1.0.0", "1.0.0-beta.1"},
		{"complete", "18.0.0-canary.1", "19.1.0-rc.2"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				diffChars(bm.old, bm.new)
			}
		})
	}
}

func BenchmarkCompilePatterns(b *testing.B) {
	benchmarks := []struct {
		name     string
		patterns []string
	}{
		{"none", nil},
		{"single", []string{"react*"}},
		{"multi", []string{"react*", "lodash*", "!*-dom", "@types/*"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				_, _ = compilePatterns(bm.patterns)
			}
		})
	}
}

func BenchmarkBuildManifest(b *testing.B) {
	// Simulate a realistic package.json with 30 deps
	var deps []string

	for i := range 30 {
		deps = append(deps, `"pkg-`+string(rune('a'+i))+`": "^1.0.0"`)
	}

	jsonStr := `{"dependencies": {` + strings.Join(deps, ",") + `}}`

	pkg := &PackageFile{raw: []byte(jsonStr)}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = BuildManifest(pkg, nil, false)
	}
}

func BenchmarkReplaceVersion(b *testing.B) {
	cases := []struct {
		name             string
		sOld, vOld, vNew string
	}{
		{"caret", "^18.0.0", "18.0.0", "19.0.0"},
		{"tilde", "~2.3.4", "2.3.4", "2.4.0"},
		{"exact", "1.0.0", "1.0.0", "1.2.0"},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				replaceVersion(c.sOld, c.vOld, c.vNew)
			}
		})
	}
}

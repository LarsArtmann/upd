package upd

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// Property tests for versionRe / latestRe classification. Generators are
// seeded and deterministic; the assertions hold for arbitrary generated
// inputs, not just a fixed table.

// randomSemverLike generates a version constraint versionRe must accept:
// optional ^/~ prefix, whitespace, and a semver-ish core (optionally with
// prerelease/build metadata). It returns the full constraint and the core
// the regex must capture.
func randomSemverLike(rng *rand.Rand) (string, string) {
	core := fmt.Sprintf("%d.%d.%d", rng.Intn(1000), rng.Intn(1000), rng.Intn(1000))

	switch rng.Intn(4) {
	case 1:
		core += fmt.Sprintf("-beta.%d", rng.Intn(10))
	case 2:
		core += fmt.Sprintf("+build.%d", rng.Intn(100))
	case 3:
		core += fmt.Sprintf("-rc.%d.%d", rng.Intn(5), rng.Intn(5))
	}

	prefix := ""
	switch rng.Intn(3) {
	case 1:
		prefix = "^"
	case 2:
		prefix = "~"
	}

	midWS := ""
	if prefix != "" && rng.Intn(2) == 0 {
		midWS = " "
	}

	return strings.Repeat(" ", rng.Intn(3)) + prefix + midWS + core + strings.Repeat(" ", rng.Intn(3)), core
}

// corruptSemverLike inserts one comparator character that must force the
// regex to reject the string.
func corruptSemverLike(rng *rand.Rand, full string) string {
	chars := []string{"<", ">", "|", "="}
	evil := chars[rng.Intn(len(chars))]
	pos := rng.Intn(len(full) + 1)

	return full[:pos] + evil + full[pos:]
}

func TestVersionReAcceptsGeneratedSemverConstraints(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(1))

	for range 500 {
		full, core := randomSemverLike(rng)

		m := versionRe.FindStringSubmatch(full)
		if m == nil {
			t.Fatalf("versionRe rejected %q (core %q)", full, core)
		}

		if m[1] != core {
			t.Fatalf("versionRe capture for %q = %q, want %q", full, m[1], core)
		}
	}
}

func TestVersionReRejectsGeneratedComparatorCorruptions(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(2))

	for range 500 {
		full, _ := randomSemverLike(rng)
		corrupted := corruptSemverLike(rng, full)

		if m := versionRe.FindStringSubmatch(corrupted); m != nil {
			t.Fatalf("versionRe accepted corrupted constraint %q (from %q)", corrupted, full)
		}
	}
}

func TestVersionReRejectsTagsAndRanges(t *testing.T) {
	t.Parallel()

	nonMatching := []string{
		"latest", "LATEST", " Latest ", "next", "beta", "*",
		">=1.0.0", "<2.0.0", "1.0.0 - 2.0.0", "^1 || ^2",
		"git+https://github.com/user/repo.git", "file:../local", "workspace:*", "",
	}

	for _, s := range nonMatching {
		if m := versionRe.FindStringSubmatch(s); m != nil {
			t.Errorf("versionRe unexpectedly accepted %q", s)
		}
	}
}

func TestLatestReMatchesCaseInsensitiveLatestOnly(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(3))

	matching := []string{"latest", "LATEST", "Latest", "  latest  ", "latest ", "\tlatest\t"}

	for _, s := range matching {
		if !latestRe.MatchString(s) {
			t.Errorf("latestRe rejected %q", s)
		}
	}

	nonMatching := []string{"latest-thing", "latestx", "xlatest", "flathest", "latest2", "", "thelatest"}

	for _, s := range nonMatching {
		if latestRe.MatchString(s) {
			t.Errorf("latestRe unexpectedly accepted %q", s)
		}
	}

	for range 200 {
		full, _ := randomSemverLike(rng)
		if latestRe.MatchString(full) {
			t.Fatalf("latestRe matched a semver constraint %q", full)
		}
	}
}

func TestBuildManifestClassifiesGeneratedConstraints(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(4))

	for range 50 {
		full, _ := randomSemverLike(rng)
		name := "pkg-" + strconv.Itoa(rng.Intn(1000000))
		jsonStr := `{"dependencies": {"` + name + `": "` + full + `"}}`
		pkg := &PackageFile{raw: []byte(jsonStr)}

		manifest, warnings := BuildManifest(pkg, nil, false)
		if len(warnings) != 0 {
			t.Fatalf("unexpected warnings for %q: %v", full, warnings)
		}

		spec := manifest[name][0]
		if spec.State != StateCheck {
			t.Fatalf("constraint %q: state = %q, want check", full, spec.State)
		}
	}
}

func TestBuildManifestPinLatestAcceptsAnyCaseLatest(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"latest", "LaTeSt", "  LATEST  "} {
		jsonStr := `{"dependencies": {"pkg": "` + tag + `"}}`
		pkg := &PackageFile{raw: []byte(jsonStr)}

		manifest, _ := BuildManifest(pkg, nil, true)
		spec := manifest["pkg"][0]

		if spec.State != StateCheck || !spec.IsLatest {
			t.Fatalf("pinLatest with %q: state = %q, isLatest = %v", tag, spec.State, spec.IsLatest)
		}

		manifestNoPin, _ := BuildManifest(pkg, nil, false)
		if got := manifestNoPin["pkg"][0].State; got != StateSkipped {
			t.Fatalf("without pinLatest %q: state = %q, want skipped", tag, got)
		}
	}
}

func TestReplaceVersionRoundTripPreservesDecoration(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(5))

	for range 500 {
		full, core := randomSemverLike(rng)
		want := strings.Replace(full, core, "9.9.9", 1)

		if got := replaceVersion(full, core, "9.9.9"); got != want {
			t.Fatalf("replaceVersion(%q, %q) = %q, want %q", full, core, got, want)
		}
	}
}

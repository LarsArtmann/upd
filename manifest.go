package upd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

type State string

const (
	StateTodo    State = "todo"
	StateCheck   State = "check"
	StateSkipped State = "skipped"
	StateKept    State = "kept"
	StateUpdated State = "updated"
	StateError   State = "error"
	StateIgnored State = "ignored"
)

func dependencySectionNames() []string {
	return []string{
		"optionalDependencies",
		"peerDependencies",
		"devDependencies",
		"dependencies",
	}
}

var (
	versionRe = regexp.MustCompile(`^\s*(?:[\^~]\s*)?(\d+[^\s<>|=]*)\s*$`)
	latestRe  = regexp.MustCompile(`(?i)^\s*latest\s*$`)
)

type Spec struct {
	Section  string
	Name     string
	SOld     string
	VOld     string
	SNew     string
	VNew     string
	State    State
	Err      error
	IsLatest bool
}

type Manifest map[string][]*Spec

func BuildManifest(pkg *PackageFile, patterns []string, pinLatest bool) (Manifest, []string) {
	manifest := make(Manifest)

	var warnings []string

	compiledPatterns, patternWarnings := compilePatterns(patterns)
	warnings = append(warnings, patternWarnings...)

	for _, section := range dependencySectionNames() {
		deps, err := pkg.GetDependencySection(section)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("section %q: %v", section, err))

			continue
		}

		for name, sOld := range deps {
			state := StateIgnored
			if matchesPatterns(name, compiledPatterns) {
				state = StateTodo
			}

			vOld := sOld
			isLatest := false

			if state == StateTodo {
				m := versionRe.FindStringSubmatch(sOld)
				switch {
				case m != nil:
					vOld = m[1]
					state = StateCheck
				case pinLatest && latestRe.MatchString(sOld):
					vOld = strings.TrimSpace(sOld)
					state = StateCheck
					isLatest = true
				default:
					state = StateSkipped
				}
			}

			spec := &Spec{
				Section:  section,
				Name:     name,
				SOld:     sOld,
				VOld:     vOld,
				SNew:     sOld,
				VNew:     vOld,
				State:    state,
				Err:      nil,
				IsLatest: isLatest,
			}
			manifest[name] = append(manifest[name], spec)
		}
	}

	return manifest, warnings
}

func (m Manifest) ToCheck() []string {
	seen := make(map[string]bool)

	var names []string

	for name, specs := range m {
		for _, spec := range specs {
			if spec.State == StateCheck && !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	return names
}

func (m Manifest) SortedNames() []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

type compiledPatterns struct {
	positive []glob.Glob
	negative []glob.Glob
}

func matchesPatterns(name string, patterns compiledPatterns) bool {
	if len(patterns.positive) == 0 && len(patterns.negative) == 0 {
		return true
	}

	hasPositive := len(patterns.positive) > 0

	matched := !hasPositive

	for _, p := range patterns.positive {
		if p.Match(name) {
			matched = true

			break
		}
	}

	if !matched {
		return false
	}

	for _, p := range patterns.negative {
		if p.Match(name) {
			return false
		}
	}

	return true
}

func compilePatterns(patterns []string) (compiledPatterns, []string) {
	var (
		compiled compiledPatterns
		warnings []string
	)

	for _, raw := range patterns {
		globExpr := raw

		isNegative := strings.HasPrefix(raw, "!")
		if isNegative {
			globExpr = raw[1:]
		}

		matcher, err := glob.Compile(globExpr)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("invalid glob pattern %q: %v", raw, err))

			continue
		}

		if isNegative {
			compiled.negative = append(compiled.negative, matcher)
		} else {
			compiled.positive = append(compiled.positive, matcher)
		}
	}

	return compiled, warnings
}

func (s *Spec) String() string {
	return fmt.Sprintf("%s[%s]: %s→%s (%s)", s.Name, s.Section, s.SOld, s.SNew, s.State)
}

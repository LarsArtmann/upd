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

var versionRe = regexp.MustCompile(`^\s*(?:[\^~]\s*)?(\d+[^\s<>|=]*)\s*$`)

type Spec struct {
	Section string
	Name    string
	SOld    string
	VOld    string
	SNew    string
	VNew    string
	State   State
}

type Manifest map[string][]*Spec

func BuildManifest(pkg *PackageFile, patterns []string) Manifest {
	manifest := make(Manifest)

	for _, section := range dependencySectionNames() {
		deps := pkg.GetDependencySection(section)
		for name, sOld := range deps {
			state := StateIgnored
			if matchesPatterns(name, patterns) {
				state = StateTodo
			}

			vOld := sOld
			if state == StateTodo {
				m := versionRe.FindStringSubmatch(sOld)
				if m != nil {
					vOld = m[1]
					state = StateCheck
				} else {
					state = StateSkipped
				}
			}

			spec := &Spec{
				Section: section,
				Name:    name,
				SOld:    sOld,
				VOld:    vOld,
				SNew:    sOld,
				VNew:    vOld,
				State:   state,
			}
			manifest[name] = append(manifest[name], spec)
		}
	}

	return manifest
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

func matchesPatterns(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	positive, negative := splitPatterns(patterns)
	hasPositive := len(positive) > 0

	matched := !hasPositive

	for _, p := range positive {
		if p.Match(name) {
			matched = true

			break
		}
	}

	if !matched {
		return false
	}

	for _, p := range negative {
		if p.Match(name) {
			return false
		}
	}

	return true
}

func splitPatterns(patterns []string) ([]glob.Glob, []glob.Glob) {
	var positive []glob.Glob
	var negative []glob.Glob

	for _, raw := range patterns {
		globExpr := raw
		isNegative := strings.HasPrefix(raw, "!")
		if isNegative {
			globExpr = raw[1:]
		}

		compiled, err := glob.Compile(globExpr)
		if err != nil {
			continue
		}

		if isNegative {
			negative = append(negative, compiled)
		} else {
			positive = append(positive, compiled)
		}
	}

	return positive, negative
}

func (s *Spec) String() string {
	return fmt.Sprintf("%s[%s]: %s→%s (%s)", s.Name, s.Section, s.SOld, s.SNew, s.State)
}

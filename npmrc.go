package upd

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

// Npmrc holds the registry settings parsed from npm configuration files.
// Only what upd consumes is modeled: the default registry URL and bearer
// tokens keyed by registry base URL. Token values are never logged,
// rendered, or included in errors.
type Npmrc struct {
	Registry   string
	AuthTokens map[string]string
}

// LoadNpmrc reads .npmrc from the user's home directory and then from dir
// (project-local), so project-local settings win over user-level ones,
// mirroring npm's precedence. Missing files are not an error. The second
// return value carries warnings for unreadable or malformed files.
func LoadNpmrc(dir string) (*Npmrc, []string) {
	paths := []string{}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".npmrc"))
	}

	paths = append(paths, filepath.Join(dir, ".npmrc"))

	merged := &Npmrc{Registry: "", AuthTokens: map[string]string{}}

	var warnings []string

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				warnings = append(warnings, fmt.Sprintf("could not read %s: %v", path, err))
			}

			continue
		}

		parsed, parseWarnings := ParseNpmrc(data)
		warnings = append(warnings, parseWarnings...)

		if parsed.Registry != "" {
			merged.Registry = parsed.Registry
		}

		maps.Copy(merged.AuthTokens, parsed.AuthTokens)
	}

	return merged, warnings
}

// ParseNpmrc parses .npmrc content (INI-like "key=value" lines with ";" or
// "#" comments). Malformed lines are returned as warnings and skipped so a
// bad line cannot hide the rest of the configuration. Keys upd does not
// consume (scope registries, strict-ssl, ...) are ignored silently.
func ParseNpmrc(data []byte) (*Npmrc, []string) {
	npmrc := &Npmrc{Registry: "", AuthTokens: map[string]string{}}

	var warnings []string

	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			warnings = append(warnings, fmt.Sprintf(
				"malformed line %d in .npmrc is ignored (expected key=value)", i+1,
			))

			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch {
		case key == "registry":
			npmrc.Registry = value
		case key == "_auth":
			warnings = append(warnings,
				"legacy _auth entry in .npmrc is not supported; use //registry-host/:_authToken instead")
		case strings.HasSuffix(key, ":_authToken"):
			registry := npmrcTokenRegistry(key)
			if registry == "" {
				warnings = append(warnings, fmt.Sprintf(
					"unsupported auth entry %q in .npmrc is ignored", key,
				))

				continue
			}

			npmrc.AuthTokens[registry] = value
		default:
			// Not consumed by upd.
		}
	}

	return npmrc, warnings
}

// npmrcTokenRegistry converts an .npmrc auth key like
// "//registry.npmjs.org/:_authToken" into the scheme-less registry host and
// path it scopes.
func npmrcTokenRegistry(key string) string {
	trimmed := strings.TrimSuffix(key, ":_authToken")
	if !strings.HasPrefix(trimmed, "//") {
		return ""
	}

	return strings.TrimPrefix(trimmed, "//")
}

// schemelessURL strips the leading scheme so registry URLs and scheme-less
// .npmrc auth keys can be compared.
func schemelessURL(raw string) string {
	before, _, found := strings.Cut(raw, "://")
	if found {
		return raw[len(before)+3:]
	}

	return raw
}

// TokenFor returns the bearer token registered for the most specific
// matching registry URL, if any. Matching ignores schemes and trailing
// slashes, and a path key never matches a different path that merely
// shares a prefix ("/team" does not match "/team-extra").
func (n *Npmrc) TokenFor(registryURL string) (string, bool) {
	if n == nil || len(n.AuthTokens) == 0 {
		return "", false
	}

	lookup := strings.TrimRight(schemelessURL(registryURL), "/")

	bestURL := ""
	bestToken := ""

	for registry, token := range n.AuthTokens {
		base := strings.TrimRight(registry, "/")
		if !strings.HasPrefix(lookup, base) {
			continue
		}

		if len(lookup) > len(base) && lookup[len(base)] != '/' {
			continue
		}

		if len(base) > len(bestURL) {
			bestURL, bestToken = base, token
		}
	}

	if bestURL == "" {
		return "", false
	}

	return bestToken, true
}

// ApplyNpmrc merges .npmrc settings into the configuration. Explicit user
// configuration (--registry flag or UPD_REGISTRY) wins over file values;
// file values win over the built-in default.
func (c *Config) ApplyNpmrc(n *Npmrc) {
	if n == nil {
		return
	}

	if c.Registry == defaultRegistryURL && n.Registry != "" {
		c.Registry = n.Registry
	}

	if token, ok := n.TokenFor(c.Registry); ok {
		c.RegistryToken = token
	}
}

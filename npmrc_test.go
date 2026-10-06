package upd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseNpmrcRegistryAndToken(t *testing.T) {
	t.Parallel()

	content := "# comment line\n" +
		"; also a comment\n" +
		"registry=https://registry.example.com/\n" +
		"//registry.example.com/:_authToken=secret-token\n" +
		"strict-ssl=true\n"

	npmrc, warnings := ParseNpmrc([]byte(content))

	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %q", warnings)
	}

	if npmrc.Registry != "https://registry.example.com/" {
		t.Errorf("Registry = %q, want https://registry.example.com/", npmrc.Registry)
	}

	token, ok := npmrc.TokenFor("https://registry.example.com/")
	if !ok || token != "secret-token" {
		t.Errorf("TokenFor = (%q, %v), want (secret-token, true)", token, ok)
	}
}

func TestParseNpmrcMalformedLineWarns(t *testing.T) {
	t.Parallel()

	npmrc, warnings := ParseNpmrc([]byte("registry=https://ok.example.com\nnot-a-kv-line\n"))

	if npmrc.Registry != "https://ok.example.com" {
		t.Errorf("Registry = %q, want the valid line to survive the malformed one", npmrc.Registry)
	}

	if len(warnings) != 1 || !strings.Contains(warnings[0], "line 2") {
		t.Errorf("expected one warning naming line 2, got %q", warnings)
	}
}

func TestParseNpmrcUnsupportedAuthEntryWarns(t *testing.T) {
	t.Parallel()

	_, warnings := ParseNpmrc([]byte("_auth=base64blob\n"))

	if len(warnings) != 1 || !strings.Contains(warnings[0], "_auth") {
		t.Errorf("expected one warning naming the unsupported entry, got %q", warnings)
	}
}

func TestTokenForLongestPrefixWins(t *testing.T) {
	t.Parallel()

	npmrc := &Npmrc{AuthTokens: map[string]string{
		"registry.example.com/":     "host-token",
		"registry.example.com/team": "team-token",
	}}

	token, ok := npmrc.TokenFor("https://registry.example.com/team")
	if !ok || token != "team-token" {
		t.Errorf("TokenFor = (%q, %v), want (team-token, true)", token, ok)
	}
}

func TestTokenForNoMatchAndNilReceiver(t *testing.T) {
	t.Parallel()

	var nilNpmrc *Npmrc
	if _, ok := nilNpmrc.TokenFor("https://registry.example.com"); ok {
		t.Error("nil receiver should never yield a token")
	}

	empty := &Npmrc{}
	if _, ok := empty.TokenFor("https://registry.example.com"); ok {
		t.Error("empty npmrc should never yield a token")
	}

	other := &Npmrc{AuthTokens: map[string]string{"other.example.com": "x"}}
	if _, ok := other.TokenFor("https://registry.example.com"); ok {
		t.Error("non-matching registry should not yield a token")
	}
}

func TestLoadNpmrcProjectOverridesUser(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	userNpmrc := filepath.Join(t.TempDir(), "home-marker")
	if err := os.MkdirAll(filepath.Dir(userNpmrc), 0o700); err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	userContent := "registry=https://user.example.com\n//user.example.com/:_authToken=user-token\n"
	if err := os.WriteFile(filepath.Join(home, ".npmrc"), []byte(userContent), 0o600); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	projectContent := "registry=https://project.example.com\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".npmrc"), []byte(projectContent), 0o600); err != nil {
		t.Fatal(err)
	}

	npmrc, warnings := LoadNpmrc(projectDir)

	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %q", warnings)
	}

	if npmrc.Registry != "https://project.example.com" {
		t.Errorf("Registry = %q, want project-local value to win", npmrc.Registry)
	}

	if token, ok := npmrc.TokenFor("https://user.example.com"); !ok || token != "user-token" {
		t.Errorf("user-level token should still be merged, got (%q, %v)", token, ok)
	}
}

func TestLoadNpmrcMissingFilesAreNoError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	npmrc, warnings := LoadNpmrc(t.TempDir())

	if len(warnings) != 0 {
		t.Errorf("expected no warnings for missing files, got %q", warnings)
	}

	if npmrc.Registry != "" || len(npmrc.AuthTokens) != 0 {
		t.Errorf("expected empty npmrc, got %+v", npmrc)
	}
}

func TestApplyNpmrcFlagAndEnvWinOverFile(t *testing.T) {
	t.Parallel()

	npmrc := &Npmrc{
		Registry:   "https://npmrc.example.com",
		AuthTokens: map[string]string{"npmrc.example.com": "file-token"},
	}

	cfg := DefaultConfig()
	cfg.ApplyNpmrc(npmrc)

	if cfg.Registry != "https://npmrc.example.com" {
		t.Errorf("Registry = %q, want npmrc value when user did not set one", cfg.Registry)
	}

	if cfg.RegistryToken != "file-token" {
		t.Errorf("RegistryToken = %q, want file-token", cfg.RegistryToken)
	}

	explicit := DefaultConfig()
	explicit.Registry = "https://explicit.example.com"
	explicit.ApplyNpmrc(npmrc)

	if explicit.Registry != "https://explicit.example.com" {
		t.Errorf("Registry = %q, want explicit value to win over npmrc", explicit.Registry)
	}

	if explicit.RegistryToken != "" {
		t.Errorf("RegistryToken = %q, want no token for a registry without one", explicit.RegistryToken)
	}
}

func TestApplyNpmrcNilIsNoop(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ApplyNpmrc(nil)

	if cfg.Registry != defaultRegistryURL {
		t.Errorf("Registry = %q, want unchanged default", cfg.Registry)
	}
}

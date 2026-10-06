//go:build integration

package upd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
)

// TestRealNPMRegistryFetch exercises the full registry client against the
// real NPM registry. It is excluded from the default suite; run it with:
//
//	GOEXPERIMENT=jsonv2 go test -tags integration -run TestRealNPMRegistry ./...
//
// or via the flake: nix run .#test-integration
func TestRealNPMRegistryFetch(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Timeout = 15 * time.Second

	client := NewRegistryClient(cfg)

	pkg, bytes, err := client.FetchPackument(context.Background(), "react")
	if err != nil {
		t.Fatalf("fetch react packument: %v", err)
	}

	if bytes == 0 {
		t.Error("expected non-empty packument body")
	}

	latest, err := pkg.LatestVersion()
	if err != nil {
		t.Fatalf("resolve dist-tags.latest: %v", err)
	}

	if _, err := semver.NewVersion(latest); err != nil {
		t.Fatalf("dist-tags.latest %q is not valid semver: %v", latest, err)
	}

	greatest, err := pkg.GreatestVersion()
	if err != nil {
		t.Fatalf("resolve greatest version: %v", err)
	}

	if _, err := semver.NewVersion(greatest); err != nil {
		t.Fatalf("greatest version %q is not valid semver: %v", greatest, err)
	}
}

func TestRealNPMRegistryNotFound(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Timeout = 15 * time.Second
	cfg.Retries = 0

	client := NewRegistryClient(cfg)

	_, _, err := client.FetchPackument(context.Background(), "upd-does-not-exist-anywhere-0123456789abcdef")
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}

	if !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("expected ErrPackageNotFound classification, got: %v", err)
	}
}

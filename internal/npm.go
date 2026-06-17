package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/tidwall/gjson"
)

type Packument struct {
	raw []byte
}

func FetchPackument(ctx context.Context, name, userAgent string) (*Packument, int, error) {
	reqURL := RegistryURL + "/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("package information retrieval failed: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("package information retrieval failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("registry returned status %d for %q: %w", resp.StatusCode, name, ErrPackageNotFound)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read packument for %q: %w", name, err)
	}

	return &Packument{raw: data}, len(data), nil
}

func (p *Packument) LatestVersion() (string, error) {
	r := gjson.GetBytes(p.raw, "dist-tags.latest")
	if !r.Exists() {
		return "", fmt.Errorf("no \"latest\" dist-tag found")
	}
	return r.String(), nil
}

func (p *Packument) GreatestVersion() (string, error) {
	versions := p.VersionKeys()
	if len(versions) == 0 {
		return "", fmt.Errorf("no valid versions found")
	}

	var greatest *semver.Version
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if greatest == nil || sv.GreaterThan(greatest) {
			greatest = sv
		}
	}
	if greatest == nil {
		return "", fmt.Errorf("no valid semver versions found")
	}
	return greatest.Original(), nil
}

func (p *Packument) VersionKeys() []string {
	result := gjson.GetBytes(p.raw, "versions")
	if !result.IsObject() {
		return nil
	}
	var keys []string
	result.ForEach(func(key, _ gjson.Result) bool {
		keys = append(keys, key.String())
		return true
	})
	sort.Strings(keys)
	return keys
}

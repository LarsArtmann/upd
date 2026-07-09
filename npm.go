package upd

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	defaultRegistryURL = "https://registry.npmjs.org"
	registryTimeout    = 20 * time.Second
)

type RegistryClient struct {
	baseURL   string
	userAgent string
	http      *http.Client
}

func NewRegistryClient(userAgent string) *RegistryClient {
	return &RegistryClient{
		baseURL:   defaultRegistryURL,
		userAgent: userAgent,
		http: &http.Client{
			Timeout: registryTimeout,
		},
	}
}

type Packument struct {
	raw []byte
}

func (c *RegistryClient) FetchPackument(ctx context.Context, name string) (*Packument, int, error) {
	reqURL := c.baseURL + "/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build registry request for %q: %w", name, err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send registry request for %q: %w", name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, classifyRegistryError(resp.StatusCode, name)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read packument for %q: %w", name, err)
	}

	return &Packument{raw: data}, len(data), nil
}

func classifyRegistryError(status int, name string) error {
	if status == http.StatusNotFound || status == http.StatusGone {
		return fmt.Errorf("registry returned status %d for %q: %w", status, name, ErrPackageNotFound)
	}

	return fmt.Errorf("registry returned status %d for %q: %w", status, name, ErrRegistryUnavailable)
}

func (p *Packument) LatestVersion() (string, error) {
	var v struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"` //nolint:tagliatelle // NPM registry uses kebab-case, not camelCase
	}

	err := json.Unmarshal(p.raw, &v)
	if err != nil {
		return "", fmt.Errorf("parse packument dist-tags: %w", err)
	}

	if v.DistTags.Latest == "" {
		return "", ErrNoLatestDistTag
	}

	return v.DistTags.Latest, nil
}

func (p *Packument) GreatestVersion() (string, error) {
	versions := p.VersionKeys()
	if len(versions) == 0 {
		return "", ErrNoValidVersions
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
		return "", ErrNoSemverVersions
	}

	return greatest.Original(), nil
}

func (p *Packument) VersionKeys() []string {
	var v struct {
		Versions map[string]struct{} `json:"versions"`
	}

	err := json.Unmarshal(p.raw, &v)
	if err != nil {
		return nil
	}

	keys := make([]string, 0, len(v.Versions))
	for k := range v.Versions {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

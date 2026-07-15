package upd

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	errorfamily "github.com/larsartmann/go-error-family"
)

const (
	backoffBase       = 1 * time.Second
	backoffMax        = 30 * time.Second
	backoffShiftCap   = 5
	transportMaxIdle  = 100
	transportIdleHost = 16
	transportIdleTime = 90 * time.Second
)

type RegistryClient struct {
	baseURL    string
	userAgent  string
	maxRetries int
	http       *http.Client
	sleep      sleeper
}

func NewRegistryClient(cfg *Config) *RegistryClient {
	return &RegistryClient{
		baseURL:    cfg.Registry,
		userAgent:  cfg.UserAgent(),
		maxRetries: cfg.Retries,
		http: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        transportMaxIdle,
				MaxIdleConnsPerHost: transportIdleHost,
				IdleConnTimeout:     transportIdleTime,
			},
		},
		sleep: sleepWithContext,
	}
}

type Packument struct {
	raw []byte
}

// sleeper abstracts time delays so tests can run without real sleeps.
// Production code uses sleepWithContext; tests inject a no-op.
type sleeper func(ctx context.Context, delay time.Duration) bool

// retryableError wraps a transient failure (429/5xx or network error) so the
// fetch loop knows to retry. It preserves the underlying cause for error
// classification via errors.Is/errors.As.
type retryableError struct {
	cause      error
	retryAfter time.Duration
}

func (e *retryableError) Error() string {
	return e.cause.Error()
}

func (e *retryableError) Unwrap() error {
	return e.cause
}

func (c *RegistryClient) FetchPackument(ctx context.Context, name string) (*Packument, int, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		pkg, bytes, err := c.fetchOnce(ctx, name)
		if err == nil {
			return pkg, bytes, nil
		}

		lastErr = err

		if attempt == c.maxRetries {
			break
		}

		var retryErr *retryableError
		if !errors.As(err, &retryErr) {
			break
		}

		delay := backoffDuration(attempt, retryErr.retryAfter)
		if !c.sleep(ctx, delay) {
			return nil, 0, errorfamily.WrapRejection(
				ctx.Err(),
				"registry.fetch_aborted",
				fmt.Sprintf("fetch %q aborted during backoff", name),
			)
		}
	}

	return nil, 0, lastErr
}

func (c *RegistryClient) fetchOnce(ctx context.Context, name string) (*Packument, int, error) {
	reqURL := c.baseURL + "/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, errorfamily.WrapRejection(
			err,
			"registry.request_build",
			fmt.Sprintf("build registry request for %q", name),
		)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, &retryableError{
			cause: errorfamily.WrapTransient(
				err,
				"registry.request_send",
				fmt.Sprintf("send registry request for %q", name),
			),
			retryAfter: 0,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errorFromStatus(resp, name)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, errorfamily.WrapTransient(
			err,
			"registry.read_body",
			fmt.Sprintf("read packument body for %q", name),
		)
	}

	return &Packument{raw: data}, len(data), nil
}

func errorFromStatus(resp *http.Response, name string) error {
	classified := classifyRegistryError(resp.StatusCode, name)
	if !isRetryableStatus(resp.StatusCode) {
		return classified
	}

	return &retryableError{
		cause:      classified,
		retryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
	}
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}

	return false
}

func backoffDuration(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return min(retryAfter, backoffMax)
	}

	shift := min(attempt, backoffShiftCap)
	delay := backoffBase * time.Duration(1<<shift)

	return min(delay, backoffMax)
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}

	seconds, err := strconv.Atoi(header)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}

	t, err := http.ParseTime(header)
	if err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}

	return 0
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func classifyRegistryError(status int, name string) error {
	if status == http.StatusNotFound || status == http.StatusGone {
		return ErrPackageNotFound.WithContext("package", name).WithContextf("status", "%d", status)
	}

	return ErrRegistryUnavailable.WithContext("package", name).WithContextf("status", "%d", status)
}

func (p *Packument) LatestVersion() (string, error) {
	var v struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"` //nolint:tagliatelle // NPM registry uses kebab-case, not camelCase
	}

	err := json.Unmarshal(p.raw, &v)
	if err != nil {
		return "", errorfamily.WrapCorruption(err, "registry.parse_dist_tags", "parse packument dist-tags")
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

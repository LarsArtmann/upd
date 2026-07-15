package upd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(registryURL string, retries int) *RegistryClient {
	cfg := DefaultConfig()
	cfg.Registry = registryURL
	cfg.Retries = retries
	cfg.Timeout = 5 * time.Second

	client := NewRegistryClient(cfg)
	client.sleep = func(_ context.Context, _ time.Duration) bool { return true }

	return client
}

func newCountingServer(
	t *testing.T,
	handler func(w http.ResponseWriter, attempt int32),
) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var attempts atomic.Int32

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, attempts.Add(1))
	}))
	t.Cleanup(s.Close)

	return s, &attempts
}

func fetchAndCaptureDelays(t *testing.T, registryURL string, retries int, pkgName string) []time.Duration {
	t.Helper()

	cfg := DefaultConfig()
	cfg.Registry = registryURL
	cfg.Retries = retries

	client := NewRegistryClient(cfg)

	var delays []time.Duration

	client.sleep = func(_ context.Context, d time.Duration) bool {
		delays = append(delays, d)

		return true
	}

	_, _, err := client.FetchPackument(context.Background(), pkgName)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}

	return delays
}

func TestFetchPackumentRetriesOn503(t *testing.T) {
	registry, attempts := newCountingServer(t, func(w http.ResponseWriter, n int32) {
		if n <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{}}}`))
	})

	client := newTestClient(registry.URL, 3)

	pkg, _, err := client.FetchPackument(context.Background(), "test-pkg")
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}

	if pkg == nil {
		t.Fatal("expected non-nil packument")
	}

	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts (2 retries + 1 success), got %d", got)
	}
}

func TestFetchPackumentDoesNotRetry404(t *testing.T) {
	registry, attempts := newCountingServer(t, func(w http.ResponseWriter, _ int32) {
		w.WriteHeader(http.StatusNotFound)
	})

	client := newTestClient(registry.URL, 3)

	_, _, err := client.FetchPackument(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for 404")
	}

	if got := attempts.Load(); got != 1 {
		t.Errorf("404 must not be retried, expected 1 attempt, got %d", got)
	}
}

func TestFetchPackumentRetries429ThenGivesUp(t *testing.T) {
	registry, attempts := newCountingServer(t, func(w http.ResponseWriter, _ int32) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	client := newTestClient(registry.URL, 2)

	_, _, err := client.FetchPackument(context.Background(), "rate-limited")
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}

	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts (initial + 2 retries), got %d", got)
	}
}

func TestBackoffDurationRespectsRetryAfter(t *testing.T) {
	got := backoffDuration(0, 5*time.Second)
	if got != 5*time.Second {
		t.Errorf("expected 5s from Retry-After, got %v", got)
	}
}

func TestBackoffDurationCapped(t *testing.T) {
	got := backoffDuration(10, 0)
	if got > backoffMax {
		t.Errorf("backoff %v exceeds max %v", got, backoffMax)
	}
}

func TestBackoffDurationExponential(t *testing.T) {
	d0 := backoffDuration(0, 0)
	d1 := backoffDuration(1, 0)

	if d1 <= d0 {
		t.Errorf("expected exponential growth: d0=%v d1=%v", d0, d1)
	}
}

func TestFetchPackumentBackoffScheduleRecorded(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(registry.Close)

	delays := fetchAndCaptureDelays(t, registry.URL, 2, "backoff-test")

	wantDelays := []time.Duration{1 * time.Second, 2 * time.Second}

	if len(delays) != len(wantDelays) {
		t.Fatalf("expected %d delays, got %d: %v", len(wantDelays), len(delays), delays)
	}

	for i, want := range wantDelays {
		if delays[i] != want {
			t.Errorf("delay[%d] = %v, want %v", i, delays[i], want)
		}
	}
}

func TestFetchPackumentSleepRecordsRetryAfter(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(registry.Close)

	delays := fetchAndCaptureDelays(t, registry.URL, 1, "retry-after-test")

	if len(delays) != 1 {
		t.Fatalf("expected 1 delay, got %d: %v", len(delays), delays)
	}

	if delays[0] != 3*time.Second {
		t.Errorf("delay = %v, want 3s (from Retry-After header)", delays[0])
	}
}

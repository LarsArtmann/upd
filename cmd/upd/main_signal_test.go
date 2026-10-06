package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestRunECancelsFetchOnSIGINT verifies the signal path end to end: fang
// installs a SIGINT notifier, the in-flight registry request is cancelled,
// and the run surfaces the failure without touching package.json.
func TestRunECancelsFetchOnSIGINT(t *testing.T) {
	requestSeen := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestSeen)

		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	file := filepath.Join(t.TempDir(), "package.json")
	pkgBefore := "{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.0.0\"\n  }\n}\n"
	if err := os.WriteFile(file, []byte(pkgBefore), 0o600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	var stdout, stderr bytes.Buffer
	runErr := make(chan error, 1)

	go func() {
		runErr <- runE([]string{"--registry", server.URL, "--file", file, "--quiet", "--retries", "0"}, &stdout, &stderr)
	}()

	select {
	case <-requestSeen:
	case <-time.After(10 * time.Second):
		t.Fatal("registry request never arrived")
	}

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}

	select {
	case err := <-runErr:
		if err == nil {
			t.Fatal("expected runE to fail after SIGINT cancellation")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("runE did not return after SIGINT")
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}

	if string(data) != pkgBefore {
		t.Errorf("cancelled run must not touch package.json:\nwant %q\ngot  %q", pkgBefore, string(data))
	}
}

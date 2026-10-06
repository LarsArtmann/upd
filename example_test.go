package upd_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/upd"
)

// Example demonstrates the library pipeline documented in doc.go against a
// mock registry: read package.json, build the manifest, fetch packuments,
// apply updates as a dry run, and inspect the result. It is executed by
// go test, so the flow stays compile- and behavior-verified.
func Example() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/") != "left-pad" {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		packument := `{"name":"left-pad","dist-tags":{"latest":"2.5.0"},"versions":{"1.0.0":{},"2.5.0":{}}}`
		_, _ = w.Write([]byte(packument))
	}))
	defer server.Close()

	dir, err := os.MkdirTemp("", "upd-example")
	if err != nil {
		fmt.Println("mkdir:", err)

		return
	}

	defer func() { _ = os.RemoveAll(dir) }()

	file := filepath.Join(dir, "package.json")
	content := []byte("{\n  \"name\": \"fixture\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.0.0\"\n  }\n}\n")
	if err := os.WriteFile(file, content, 0o600); err != nil {
		fmt.Println("write:", err)

		return
	}

	cfg := upd.DefaultConfig()
	cfg.File = file
	cfg.Registry = server.URL
	cfg.Nop = true // dry-run

	pkg, err := upd.ReadPackageFile(cfg.File)
	if err != nil {
		fmt.Println("read:", err)

		return
	}

	manifest, _ := upd.BuildManifest(pkg, nil, false)
	engine := upd.NewEngine(cfg)
	results := engine.FetchAll(context.Background(), manifest.ToCheck())
	updates, errCount := engine.ApplyUpdates(manifest, results, pkg)

	fmt.Println(updates, errCount)
	fmt.Println(manifest["left-pad"][0].State)

	// Output:
	// 1 0
	// updated
}

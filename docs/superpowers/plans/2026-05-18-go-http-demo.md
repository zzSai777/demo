# Go HTTP Demo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and deploy a minimal Go HTTP demo service to `10.235.106.18`.

**Architecture:** Use only the Go standard library. Keep handlers testable by constructing an `http.Handler` in a small function, then run that handler from `main`.

**Tech Stack:** Go, `net/http`, `httptest`, SSH, `systemd`.

---

## File Structure

- `go.mod`: module definition for the demo application.
- `main.go`: HTTP handlers and server startup.
- `main_test.go`: handler tests for `/` and `/healthz`.

### Task 1: Create Tested HTTP Demo

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `main_test.go`

- [ ] **Step 1: Write handler tests**

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	newServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "Hello from Go demo!\n" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	newServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail before implementation**

Run: `go test ./...`

Expected: FAIL because `newServer` is not defined.

- [ ] **Step 3: Implement the HTTP server**

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go demo!")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("go-demo listening on %s", addr)
	if err := http.ListenAndServe(addr, newServer()); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./...`

Expected: PASS.

### Task 2: Build and Deploy

**Files:**
- Use existing: `go.mod`
- Use existing: `main.go`
- Remote create: `/opt/go-demo/go-demo`
- Remote create: `/etc/systemd/system/go-demo.service`

- [ ] **Step 1: Build Linux binary**

Run: `$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o go-demo .`

Expected: `go-demo` binary exists locally.

- [ ] **Step 2: Verify SSH connectivity**

Run: `ssh 10.235.106.18 "uname -a"`

Expected: server prints Linux system information.

- [ ] **Step 3: Upload and install service**

Run commands that create `/opt/go-demo`, upload `go-demo`, write the `systemd` unit, reload daemon, enable and restart the service.

- [ ] **Step 4: Verify deployment**

Run: `ssh 10.235.106.18 "systemctl is-active go-demo"` and `curl http://10.235.106.18:8080/healthz`

Expected: service is `active` and health check returns `ok`.

package gamectl

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunStatusCallsControlPlane(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	var out strings.Builder
	code := Run([]string{"--addr", server.URL, "status"}, &out, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, out.String())
	}
	if requestedPath != "/control/v1/status" {
		t.Fatalf("expected status path, got %q", requestedPath)
	}
	if !strings.Contains(out.String(), `"status":"ok"`) {
		t.Fatalf("expected status response in output, got %s", out.String())
	}
}

func TestRunConfigSetSendsConfigPayload(t *testing.T) {
	var method, path, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		payload, _ := io.ReadAll(r.Body)
		body = string(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key":"landlord.base_score","value":"10"}`))
	}))
	defer server.Close()

	var out strings.Builder
	code := Run([]string{"--addr", server.URL, "config", "set", "landlord.base_score", "10", "--scope", "landlord"}, &out, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, out.String())
	}
	if method != http.MethodPut {
		t.Fatalf("expected method %s, got %s", http.MethodPut, method)
	}
	if path != "/control/v1/configs/landlord.base_score" {
		t.Fatalf("expected config path, got %q", path)
	}
	if !strings.Contains(body, `"scope":"landlord"`) || !strings.Contains(body, `"value":"10"`) {
		t.Fatalf("expected config payload, got %s", body)
	}
}

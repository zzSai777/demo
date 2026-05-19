package control

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlHandlerUpdatesConfigAndReturnsStatus(t *testing.T) {
	store := NewMemoryStore()
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodPut, "/control/v1/configs/landlord.base_score", strings.NewReader(`{"value":"10","scope":"landlord"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"key":"landlord.base_score"`) {
		t.Fatalf("expected config key in response, got %s", rec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/control/v1/status", nil)
	statusRec := httptest.NewRecorder()
	handler.ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected status endpoint to return %d, got %d", http.StatusOK, statusRec.Code)
	}
	if !strings.Contains(statusRec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok status, got %s", statusRec.Body.String())
	}
}

func TestControlHandlerCreatesABTestAndGrayRollout(t *testing.T) {
	store := NewMemoryStore()
	handler := NewHandler(store)

	abReq := httptest.NewRequest(http.MethodPost, "/control/v1/ab-tests", strings.NewReader(`{"name":"landlord_new_shuffle","feature_key":"landlord.shuffle","variants":["control","new"],"traffic_percent":5}`))
	abReq.Header.Set("Content-Type", "application/json")
	abRec := httptest.NewRecorder()
	handler.ServeHTTP(abRec, abReq)

	if abRec.Code != http.StatusCreated {
		t.Fatalf("expected AB test create status %d, got %d: %s", http.StatusCreated, abRec.Code, abRec.Body.String())
	}
	if !strings.Contains(abRec.Body.String(), `"traffic_percent":5`) {
		t.Fatalf("expected traffic percent in response, got %s", abRec.Body.String())
	}

	rolloutReq := httptest.NewRequest(http.MethodPost, "/control/v1/rollouts", strings.NewReader(`{"feature_key":"landlord.super_bomb","target_percent":10,"strategy":"user_id_hash"}`))
	rolloutReq.Header.Set("Content-Type", "application/json")
	rolloutRec := httptest.NewRecorder()
	handler.ServeHTTP(rolloutRec, rolloutReq)

	if rolloutRec.Code != http.StatusCreated {
		t.Fatalf("expected rollout create status %d, got %d: %s", http.StatusCreated, rolloutRec.Code, rolloutRec.Body.String())
	}
	if !strings.Contains(rolloutRec.Body.String(), `"strategy":"user_id_hash"`) {
		t.Fatalf("expected rollout strategy in response, got %s", rolloutRec.Body.String())
	}
}

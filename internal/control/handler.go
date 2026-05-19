package control

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/control/v1")
	switch {
	case path == "/status" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, map[string]any{
			"status":   "ok",
			"time":     time.Now().UTC(),
			"services": h.store.ListServices(),
			"nodes":    h.store.ListNodes(),
		})
	case path == "/services" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListServices())
	case strings.HasPrefix(path, "/services/") && strings.HasSuffix(path, "/actions") && r.Method == http.MethodPost:
		h.handleServiceAction(w, r, path)
	case path == "/configs" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListConfigs())
	case strings.HasPrefix(path, "/configs/"):
		h.handleConfig(w, r, strings.TrimPrefix(path, "/configs/"))
	case path == "/ab-tests" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListABTests())
	case path == "/ab-tests" && r.Method == http.MethodPost:
		h.handleCreateABTest(w, r)
	case path == "/rollouts" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListRollouts())
	case path == "/rollouts" && r.Method == http.MethodPost:
		h.handleCreateRollout(w, r)
	case path == "/updates" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListUpdatePlans())
	case path == "/updates" && r.Method == http.MethodPost:
		h.handleCreateUpdatePlan(w, r)
	case path == "/nodes" && r.Method == http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.store.ListNodes())
	default:
		h.writeError(w, http.StatusNotFound, "not_found")
	}
}

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request, key string) {
	if key == "" {
		h.writeError(w, http.StatusNotFound, "config key required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		config, ok := h.store.GetConfig(key)
		if !ok {
			h.writeError(w, http.StatusNotFound, "config not found")
			return
		}
		h.writeJSON(w, http.StatusOK, config)
	case http.MethodPut:
		var req struct {
			Value string `json:"value"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid config payload")
			return
		}
		h.writeJSON(w, http.StatusOK, h.store.UpsertConfig(key, req.Value, req.Scope))
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleCreateABTest(w http.ResponseWriter, r *http.Request) {
	var test ABTest
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid ab test payload")
		return
	}
	if test.Name == "" || test.FeatureKey == "" || len(test.Variants) == 0 {
		h.writeError(w, http.StatusBadRequest, "name, feature_key, and variants are required")
		return
	}
	h.writeJSON(w, http.StatusCreated, h.store.CreateABTest(test))
}

func (h *Handler) handleCreateRollout(w http.ResponseWriter, r *http.Request) {
	var rollout Rollout
	if err := json.NewDecoder(r.Body).Decode(&rollout); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid rollout payload")
		return
	}
	if rollout.FeatureKey == "" || rollout.Strategy == "" {
		h.writeError(w, http.StatusBadRequest, "feature_key and strategy are required")
		return
	}
	h.writeJSON(w, http.StatusCreated, h.store.CreateRollout(rollout))
}

func (h *Handler) handleCreateUpdatePlan(w http.ResponseWriter, r *http.Request) {
	var plan UpdatePlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid update payload")
		return
	}
	if plan.Service == "" || plan.Version == "" {
		h.writeError(w, http.StatusBadRequest, "service and version are required")
		return
	}
	h.writeJSON(w, http.StatusCreated, h.store.CreateUpdatePlan(plan))
}

func (h *Handler) handleServiceAction(w http.ResponseWriter, r *http.Request, path string) {
	name := strings.TrimSuffix(strings.TrimPrefix(path, "/services/"), "/actions")
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid action payload")
		return
	}
	service, err := h.store.ApplyServiceAction(name, req.Action)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, service)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

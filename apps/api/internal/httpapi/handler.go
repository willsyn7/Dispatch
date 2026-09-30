package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/service"
)

type Handler struct {
	logger *log.Logger
	jobs   *service.JobService
	origin string
}

func NewHandler(logger *log.Logger, jobs *service.JobService, origin string) http.Handler {
	h := &Handler{logger: logger, jobs: jobs, origin: origin}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/v1/jobs/resolve", h.resolveJob)
	return h.cors(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type resolveJobRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Description string `json:"description"`
}

func (h *Handler) resolveJob(w http.ResponseWriter, r *http.Request) {
	var request resolveJobRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	posting, err := h.jobs.Resolve(r.Context(), request.URL, request.Title, request.Company, request.Description)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingJobInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ats.ErrUnsupportedURL):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			h.logger.Printf("job adapter unavailable error=%q", err)
			writeError(w, http.StatusNotImplemented, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, posting)
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

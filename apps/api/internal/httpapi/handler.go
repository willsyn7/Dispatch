package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
	"github.com/example/job-application-assistant/apps/api/internal/service"
)

type Handler struct {
	logger       *log.Logger
	jobs         *service.JobService
	applications *service.ApplicationService
	origin       string
}

func NewHandler(logger *log.Logger, jobs *service.JobService, origin string, applications ...*service.ApplicationService) http.Handler {
	var applicationService *service.ApplicationService
	if len(applications) > 0 {
		applicationService = applications[0]
	}
	h := &Handler{logger: logger, jobs: jobs, applications: applicationService, origin: origin}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/v1/jobs/resolve", h.resolveJob)
	if applicationService != nil {
		mux.HandleFunc("GET /api/v1/profile", h.getProfile)
		mux.HandleFunc("PUT /api/v1/profile", h.putProfile)
		mux.HandleFunc("POST /api/v1/drafts", h.generateDraft)
		mux.HandleFunc("GET /api/v1/applications", h.listApplications)
		mux.HandleFunc("PATCH /api/v1/applications/{applicationID}", h.updateApplication)
		mux.HandleFunc("GET /api/v1/applications/{applicationID}/drafts", h.listDrafts)
	}
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
	if err := decodeJSON(w, r, &request); err != nil {
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
			h.logger.Printf("job adapter failed error=%q", err)
			writeError(w, http.StatusBadGateway, "the ATS posting could not be loaded")
		}
		return
	}
	writeJSON(w, http.StatusOK, posting)
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.applications.GetProfile(r.Context())
	if errors.Is(err, service.ErrProfileRequired) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h *Handler) putProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ResumeText string `json:"resumeText"`
		Background string `json:"background"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	profile, err := h.applications.SaveProfile(r.Context(), request.ResumeText, request.Background)
	if errors.Is(err, service.ErrResumeRequired) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h *Handler) generateDraft(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Posting domain.JobPosting `json:"posting"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	draft, err := h.applications.GenerateDraft(r.Context(), request.Posting)
	switch {
	case errors.Is(err, service.ErrProfileRequired):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrPostingRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		h.internalError(w, err)
	default:
		writeJSON(w, http.StatusCreated, draft)
	}
}

func (h *Handler) listApplications(w http.ResponseWriter, r *http.Request) {
	applications, err := h.applications.ListApplications(r.Context())
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, applications)
}

func (h *Handler) updateApplication(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Status domain.ApplicationStatus `json:"status"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	application, err := h.applications.UpdateStatus(r.Context(), r.PathValue("applicationID"), request.Status)
	switch {
	case errors.Is(err, service.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrApplicationNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case err != nil:
		h.internalError(w, err)
	default:
		writeJSON(w, http.StatusOK, application)
	}
}

func (h *Handler) listDrafts(w http.ResponseWriter, r *http.Request) {
	drafts, err := h.applications.ListDrafts(r.Context(), r.PathValue("applicationID"))
	if errors.Is(err, service.ErrApplicationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, drafts)
}

func (h *Handler) internalError(w http.ResponseWriter, err error) {
	h.logger.Printf("request failed error=%q", err)
	writeError(w, http.StatusInternalServerError, "request could not be completed")
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

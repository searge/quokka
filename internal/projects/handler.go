package projects

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/searge/quokka/internal/platform"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{service: service, log: logger}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}

	project, err := h.service.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.As(err, &validator.ValidationErrors{}):
			platform.RespondValidationError(w, err)
		case errors.Is(err, ErrProjectExists):
			platform.RespondError(w, http.StatusConflict, "PROJECT_EXISTS", err.Error())
		case errors.Is(err, ErrInvalidUnixName):
			platform.RespondError(w, http.StatusBadRequest, "INVALID_UNIX_NAME", err.Error())
		default:
			h.log.Error("internal err", "error", err)
			platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		h.log.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.List(r.Context(), 100, 0)
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		h.log.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	project, err := h.service.Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			platform.RespondError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "project not found")
		case errors.Is(err, ErrInvalidProjectID):
			platform.RespondError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project id")
		default:
			h.log.Error("internal err", "error", err)
			platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		h.log.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}

	project, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			platform.RespondError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "project not found")
		case errors.Is(err, ErrInvalidProjectID):
			platform.RespondError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project id")
		default:
			h.log.Error("internal err", "error", err)
			platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		h.log.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			platform.RespondError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "project not found")
		case errors.Is(err, ErrInvalidProjectID):
			platform.RespondError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project id")
		default:
			h.log.Error("internal err", "error", err)
			platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"encoding/json"
	"net/http"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
	"github.com/go-chi/chi/v5"
)

// TargetHandler manages HTTP requests for target management.
type TargetHandler struct {
	repo *database.TargetRepo
}

// NewTargetHandler creates a new TargetHandler.
func NewTargetHandler(repo *database.TargetRepo) *TargetHandler {
	return &TargetHandler{repo: repo}
}

// CreateTargetRequest represents the body payload to create a new target.
type CreateTargetRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Notes  string `json:"notes"`
}

// Create handles creating a new target program.
func (h *TargetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	req.Domain = tools.CleanDomain(req.Domain)
	if req.Domain == "" {
		ErrorJSON(w, http.StatusBadRequest, "Valid domain is required")
		return
	}

	if req.Name == "" {
		req.Name = req.Domain
	}

	target, err := h.repo.Create(req.Name, req.Domain, req.Notes)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, target)
}

// List returns all registered targets.
func (h *TargetHandler) List(w http.ResponseWriter, r *http.Request) {
	targets, err := h.repo.List()
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if targets == nil {
		targets = []models.Target{} // Ensure JSON array instead of null
	}

	JSON(w, http.StatusOK, targets)
}

// GetByID returns target details by ID.
func (h *TargetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	target, err := h.repo.GetByID(id)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if target == nil {
		ErrorJSON(w, http.StatusNotFound, "Target not found")
		return
	}

	JSON(w, http.StatusOK, target)
}

// Delete removes a target.
func (h *TargetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(id); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Target deleted"})
}

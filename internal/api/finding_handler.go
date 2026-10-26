package api

import (
	"net/http"
	"strconv"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
)

// FindingHandler handles queries and operations on vulnerability findings.
type FindingHandler struct {
	repo *database.FindingRepo
}

// NewFindingHandler creates a new FindingHandler.
func NewFindingHandler(repo *database.FindingRepo) *FindingHandler {
	return &FindingHandler{repo: repo}
}

// FindingListResponse provides paginated finding results metadata and severity statistics.
type FindingListResponse struct {
	Findings []models.Finding      `json:"findings"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	Limit    int                   `json:"limit"`
	Stats    database.FindingStats `json:"stats"`
}

// ListByTarget returns paginated findings for a target scope.
func (h *FindingHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	search := r.URL.Query().Get("search")
	severity := r.URL.Query().Get("severity")

	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	offset := (page - 1) * limit
	findings, total, err := h.repo.ListByTarget(targetID, severity, search, limit, offset)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats, _ := h.repo.GetStats(targetID)

	if findings == nil {
		findings = []models.Finding{}
	}

	JSON(w, http.StatusOK, FindingListResponse{
		Findings: findings,
		Total:    total,
		Page:     page,
		Limit:    limit,
		Stats:    stats,
	})
}

// Delete removes a specific finding by ID.
func (h *FindingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(id); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "Finding deleted successfully"})
}

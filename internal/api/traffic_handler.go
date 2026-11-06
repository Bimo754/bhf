package api

import (
	"net/http"
	"strconv"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
)

// TrafficHandler handles queries for intercepted HTTP traffic transactions.
type TrafficHandler struct {
	repo *database.TrafficRepo
}

// NewTrafficHandler creates a new TrafficHandler.
func NewTrafficHandler(repo *database.TrafficRepo) *TrafficHandler {
	return &TrafficHandler{repo: repo}
}

// TrafficListResponse provides paginated results metadata for traffic records.
type TrafficListResponse struct {
	Traffic []models.TrafficRecord `json:"traffic"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}

// ListByTarget returns paginated traffic records for a given target scope.
func (h *TrafficHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	category := r.URL.Query().Get("category")
	method := r.URL.Query().Get("method")
	search := r.URL.Query().Get("search")

	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	offset := (page - 1) * limit
	traffic, total, err := h.repo.ListByTarget(targetID, category, method, search, limit, offset)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if traffic == nil {
		traffic = []models.TrafficRecord{}
	}

	JSON(w, http.StatusOK, TrafficListResponse{
		Traffic: traffic,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

// GetByID returns a single traffic transaction record.
func (h *TrafficHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	record, err := h.repo.GetByID(id)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if record == nil {
		ErrorJSON(w, http.StatusNotFound, "traffic record not found")
		return
	}

	JSON(w, http.StatusOK, record)
}

// GetStats returns traffic category metrics for a target scope.
func (h *TrafficHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	stats, err := h.repo.GetCategoryStats(targetID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if stats == nil {
		stats = map[string]int{}
	}

	JSON(w, http.StatusOK, map[string]any{
		"target_id":  targetID,
		"categories": stats,
	})
}

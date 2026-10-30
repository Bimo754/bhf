package api

import (
	"net/http"
	"strconv"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
)

// EndpointHandler handles queries for discovered web endpoints and paths.
type EndpointHandler struct {
	repo *database.EndpointRepo
}

// NewEndpointHandler creates a new EndpointHandler.
func NewEndpointHandler(repo *database.EndpointRepo) *EndpointHandler {
	return &EndpointHandler{repo: repo}
}

// EndpointListResponse provides paginated results metadata.
type EndpointListResponse struct {
	Endpoints []models.Endpoint `json:"endpoints"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

// ListByTarget returns paginated endpoints for a given target scope.
func (h *EndpointHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	search := r.URL.Query().Get("search")
	all := r.URL.Query().Get("all")

	// If all=true, return all endpoints for building sitemap hierarchy without pagination
	if all == "true" || all == "1" {
		endpoints, err := h.repo.GetAllForTarget(targetID)
		if err != nil {
			ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		if endpoints == nil {
			endpoints = []models.Endpoint{}
		}
		JSON(w, http.StatusOK, EndpointListResponse{
			Endpoints: endpoints,
			Total:     len(endpoints),
			Page:      1,
			Limit:     len(endpoints),
		})
		return
	}

	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	offset := (page - 1) * limit
	endpoints, total, err := h.repo.ListByTarget(targetID, search, limit, offset)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if endpoints == nil {
		endpoints = []models.Endpoint{}
	}

	JSON(w, http.StatusOK, EndpointListResponse{
		Endpoints: endpoints,
		Total:     total,
		Page:      page,
		Limit:     limit,
	})
}

package api

import (
	"net/http"
	"strconv"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
)

// AssetHandler handles queries for discovered subdomains and assets.
type AssetHandler struct {
	repo *database.AssetRepo
}

// NewAssetHandler creates a new AssetHandler.
func NewAssetHandler(repo *database.AssetRepo) *AssetHandler {
	return &AssetHandler{repo: repo}
}

// AssetListResponse provides paginated results metadata.
type AssetListResponse struct {
	Assets []models.Asset      `json:"assets"`
	Total  int                 `json:"total"`
	Page   int                 `json:"page"`
	Limit  int                 `json:"limit"`
	Stats  database.AssetStats `json:"stats"`
}

// ListByTarget returns paginated assets for a given target scope.
func (h *AssetHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	search := r.URL.Query().Get("search")
	filter := r.URL.Query().Get("filter")
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	offset := (page - 1) * limit
	assets, total, err := h.repo.ListByTarget(targetID, search, filter, sortBy, order, limit, offset)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats, _ := h.repo.GetStats(targetID)

	if assets == nil {
		assets = []models.Asset{}
	}

	JSON(w, http.StatusOK, AssetListResponse{
		Assets: assets,
		Total:  total,
		Page:   page,
		Limit:  limit,
		Stats:  stats,
	})
}

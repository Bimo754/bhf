package api

import (
	"net/http"
	"time"

	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AssessmentHandler coordinates deterministic assessment execution and rule inspection.
type AssessmentHandler struct {
	engine      *assessment.AssessmentEngine
	trafficRepo *database.TrafficRepo
	targetRepo  *database.TargetRepo
	scanRepo    *database.ScanRepo
}

// NewAssessmentHandler creates a new AssessmentHandler.
func NewAssessmentHandler(
	engine *assessment.AssessmentEngine,
	trafficRepo *database.TrafficRepo,
	targetRepo *database.TargetRepo,
	scanRepo *database.ScanRepo,
) *AssessmentHandler {
	return &AssessmentHandler{
		engine:      engine,
		trafficRepo: trafficRepo,
		targetRepo:  targetRepo,
		scanRepo:    scanRepo,
	}
}

// RunResponse contains assessment execution metrics and identified findings.
type RunResponse struct {
	TargetID        string                          `json:"target_id"`
	ScanID          string                          `json:"scan_id"`
	TrafficAnalyzed int                             `json:"traffic_analyzed"`
	FindingsCount   int                             `json:"findings_count"`
	Findings        []*assessment.AssessmentFinding `json:"findings"`
}

// RunTargetAssessment executes all registered deterministic security rules against target traffic.
func (h *AssessmentHandler) RunTargetAssessment(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	target, err := h.targetRepo.GetByID(targetID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	traffic, _, err := h.trafficRepo.ListByTarget(targetID, "", "", "", 1000, 0)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(traffic) == 0 {
		ErrorJSON(w, http.StatusBadRequest, "no intercepted traffic records found for target scope. Route traffic through proxy or browser first.")
		return
	}

	now := time.Now()
	scan := &models.Scan{
		ID:        uuid.New().String(),
		TargetID:  targetID,
		ToolName:  "deterministic_assessment",
		Status:    models.StatusRunning,
		StartedAt: now,
		Arguments: []string{},
	}
	if err := h.scanRepo.Create(scan); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to initialize assessment scan: "+err.Error())
		return
	}

	findings, err := h.engine.Run(r.Context(), target, traffic, scan.ID)
	completedAt := time.Now()
	if err != nil {
		_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusFailed, &completedAt, err.Error())
		ErrorJSON(w, http.StatusInternalServerError, "assessment execution error: "+err.Error())
		return
	}

	_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusCompleted, &completedAt, "")

	JSON(w, http.StatusOK, RunResponse{
		TargetID:        targetID,
		ScanID:          scan.ID,
		TrafficAnalyzed: len(traffic),
		FindingsCount:   len(findings),
		Findings:        findings,
	})
}

// RuleInfo describes an assessment verification check.
type RuleInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListRules returns all loaded deterministic assessment rules.
func (h *AssessmentHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules := make([]RuleInfo, 0)
	for _, rule := range h.engine.Rules() {
		rules = append(rules, RuleInfo{
			ID:          rule.ID(),
			Name:        rule.Name(),
			Description: rule.Description(),
		})
	}
	JSON(w, http.StatusOK, rules)
}

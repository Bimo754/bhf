package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/engine"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ScanHandler manages tool execution requests and history.
type ScanHandler struct {
	scanRepo    *database.ScanRepo
	targetRepo  *database.TargetRepo
	assetRepo   *database.AssetRepo
	findingRepo *database.FindingRepo
	logRepo     *database.LogRepo
	dispatcher  *engine.Dispatcher
	runner      *engine.Runner
	registry    *tools.Registry
	hub         *engine.Hub
}

// NewScanHandler creates a new ScanHandler.
func NewScanHandler(
	scanRepo *database.ScanRepo,
	targetRepo *database.TargetRepo,
	assetRepo *database.AssetRepo,
	findingRepo *database.FindingRepo,
	logRepo *database.LogRepo,
	dispatcher *engine.Dispatcher,
	runner *engine.Runner,
	registry *tools.Registry,
	hub *engine.Hub,
) *ScanHandler {
	return &ScanHandler{
		scanRepo:    scanRepo,
		targetRepo:  targetRepo,
		assetRepo:   assetRepo,
		findingRepo: findingRepo,
		logRepo:     logRepo,
		dispatcher:  dispatcher,
		runner:      runner,
		registry:    registry,
		hub:         hub,
	}
}

// StartScanRequest payload.
type StartScanRequest struct {
	TargetID   string   `json:"target_id"`
	ToolName   string   `json:"tool_name"`
	CustomArgs []string `json:"custom_args,omitempty"`
}

// Start initiates a tool execution session on demand.
func (h *ScanHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req StartScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	target, err := h.targetRepo.GetByID(req.TargetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "Target not found")
		return
	}

	tool, err := h.registry.Get(req.ToolName)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("Tool %s not supported", req.ToolName))
		return
	}

	discoveredHosts, _ := h.assetRepo.GetDistinctHosts(target.ID)
	liveURLs, _ := h.assetRepo.GetLiveURLs(target.ID)
	scanID := uuid.New().String()
	outputFile := tool.OutputFile(scanID)

	args := req.CustomArgs
	if len(args) == 0 {
		args = tool.DefaultArgs(target, discoveredHosts, liveURLs, outputFile)
	}

	fullCmd := tool.BinaryName() + " " + strings.Join(args, " ")

	scan := &models.Scan{
		ID:        scanID,
		TargetID:  target.ID,
		ToolName:  tool.Name(),
		Status:    models.StatusRunning,
		Command:   fullCmd,
		Arguments: args,
		StartedAt: time.Now().UTC(),
	}

	if err := h.scanRepo.Create(scan); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	h.dispatcher.RegisterActive(scanID, cancel)

	// Broadcast initial running status
	h.hub.Broadcast(engine.StreamMessage{
		Type:      "status",
		ScanID:    scanID,
		Status:    models.StatusRunning,
		Timestamp: scan.StartedAt,
	})

	// Run process asynchronously
	go h.executeScanAsync(ctx, scan, target, tool, args, discoveredHosts, liveURLs, outputFile)

	JSON(w, http.StatusAccepted, scan)
}

func (h *ScanHandler) executeScanAsync(
	ctx context.Context,
	scan *models.Scan,
	target *models.Target,
	tool tools.Tool,
	args []string,
	discoveredHosts []string,
	liveURLs []string,
	outputFile string,
) {
	defer h.dispatcher.Unregister(scan.ID)

	rawOutput, runErr := h.runner.Run(ctx, scan.ID, tool.BinaryName(), args)

	now := time.Now().UTC()
	var finalStatus models.ScanStatus = models.StatusCompleted
	var errMsg string

	if runErr != nil {
		if ctx.Err() == context.Canceled {
			finalStatus = models.StatusStopped
			errMsg = "Cancelled by user"
		} else {
			finalStatus = models.StatusFailed
			errMsg = runErr.Error()
		}
		_ = h.logRepo.Append(&models.LogEntry{
			ScanID:    scan.ID,
			Stream:    "stderr",
			Line:      errMsg,
			Timestamp: now,
		})
	}

	// Always parse findings even if stopped by user, to preserve all assets discovered so far
	execCtx := tools.ExecutionContext{
		Ctx:             ctx,
		Scan:            scan,
		Target:          target,
		DiscoveredHosts: discoveredHosts,
		LiveURLs:        liveURLs,
		OutputFile:      outputFile,
		CustomArgs:      args,
	}

	result, parseErr := tool.Run(execCtx, rawOutput)
	if parseErr == nil {
		if len(result.Assets) > 0 {
			for i := range result.Assets {
				if tools.IsInScope(result.Assets[i].Host, target.Domain) {
					_ = h.assetRepo.Upsert(&result.Assets[i])
				}
			}
		}
		if len(result.Findings) > 0 {
			for i := range result.Findings {
				_ = h.findingRepo.Insert(&result.Findings[i])
			}
		}
	}

	_ = h.scanRepo.UpdateStatus(scan.ID, finalStatus, &now, errMsg)

	h.hub.Broadcast(engine.StreamMessage{
		Type:      "complete",
		ScanID:    scan.ID,
		Status:    finalStatus,
		Line:      errMsg,
		Timestamp: now,
	})
}

// Stop halts an actively running scan session.
func (h *ScanHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.dispatcher.Cancel(id); err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Scan stopped successfully"})
}

// ListByTarget returns history of scans for a target.
func (h *ScanHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	scans, err := h.scanRepo.ListByTarget(targetID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if scans == nil {
		scans = []models.Scan{}
	}

	JSON(w, http.StatusOK, scans)
}

// GetLogs returns recorded logs for a scan.
func (h *ScanHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logs, err := h.logRepo.GetByScan(id)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if logs == nil {
		logs = []models.LogEntry{}
	}

	JSON(w, http.StatusOK, logs)
}

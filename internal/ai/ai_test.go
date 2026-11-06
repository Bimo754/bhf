package ai_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/ai"
	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
)

func TestParseProposals(t *testing.T) {
	rawJSON := "```json\n" + `{"proposals":[{"endpoint":"/api/users","method":"GET","test_type":"idor","reasoning":"numeric user_id param","priority":1}]}` + "\n```"
	proposals, err := ai.ParseProposals(rawJSON)
	if err != nil {
		t.Fatalf("failed to parse proposals: %v", err)
	}

	if len(proposals) != 1 {
		t.Fatalf("expected 1 proposal, got %d", len(proposals))
	}
	if proposals[0].Endpoint != "/api/users" || proposals[0].TestType != "idor" {
		t.Errorf("unexpected proposal attributes: %+v", proposals[0])
	}
}

func TestBuildTrafficAnalysisPrompt(t *testing.T) {
	records := []models.TrafficRecord{
		{
			Method:             "GET",
			Path:               "/api/v1/profile",
			Category:           "api",
			ResponseStatusCode: 200,
			AuthContext:        "session_cookie",
		},
	}
	graph := appintel.BuildParameterGraph(records)

	sys, user := ai.BuildTrafficAnalysisPrompt(records, graph)
	if !strings.Contains(sys, "BHF Security Assessment Planner") {
		t.Errorf("system prompt missing planner title")
	}
	if !strings.Contains(user, "/api/v1/profile") {
		t.Errorf("user prompt missing endpoint: %s", user)
	}
}

func TestAssessmentPlanner(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ai_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()
	_ = database.MigrateSchema(db)

	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	findingRepo := database.NewFindingRepo(db)

	target, _ := targetRepo.Create("AI Target", "ai.example.com", "")
	_ = scanRepo.Create(&models.Scan{
		ID:        "scan-ai-1",
		TargetID:  target.ID,
		ToolName:  "assessment",
		Status:    models.StatusCompleted,
		Command:   "assessment",
		StartedAt: time.Now().UTC(),
	})

	mockClient := ai.NewMockClient()
	engine := assessment.NewAssessmentEngine(findingRepo, nil)
	planner := ai.NewPlanner(mockClient, engine)

	traffic := []models.TrafficRecord{
		{
			URL:                "https://ai.example.com/api/login?password=admin",
			Method:             "GET",
			Path:               "/api/login",
			ResponseStatusCode: 200,
		},
	}

	ctx := context.Background()
	result, err := planner.PlanAndAssess(ctx, target, traffic, "scan-ai-1")
	if err != nil {
		t.Fatalf("PlanAndAssess failed: %v", err)
	}

	if len(result.Proposals) == 0 {
		t.Errorf("expected at least 1 proposal from AI client")
	}
	if len(result.VerifiedFindings) == 0 {
		t.Errorf("expected deterministic findings to be verified and returned")
	}
}

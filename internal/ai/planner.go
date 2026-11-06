package ai

import (
	"context"
	"fmt"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/models"
)

// PlanResult encapsulates the AI-generated proposals and deterministically verified findings.
type PlanResult struct {
	Proposals        []AssessmentProposal            `json:"proposals"`
	VerifiedFindings []*assessment.AssessmentFinding `json:"verified_findings"`
	ParameterGraph   *appintel.ParameterGraph        `json:"parameter_graph"`
}

// AssessmentPlanner coordinates AI-driven traffic analysis with deterministic verification execution.
type AssessmentPlanner struct {
	client Client
	engine *assessment.AssessmentEngine
}

// NewPlanner creates a new AssessmentPlanner instance.
func NewPlanner(client Client, engine *assessment.AssessmentEngine) *AssessmentPlanner {
	if client == nil {
		client = NewMockClient()
	}
	return &AssessmentPlanner{
		client: client,
		engine: engine,
	}
}

// PlanAndAssess analyzes traffic using the AI client and executes deterministic verification checks.
func (p *AssessmentPlanner) PlanAndAssess(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, scanID string) (*PlanResult, error) {
	// 1. Build Parameter Graph
	graph := appintel.BuildParameterGraph(traffic)

	// 2. Generate AI Proposals
	systemPrompt, userPrompt := BuildTrafficAnalysisPrompt(traffic, graph)
	rawResponse, err := p.client.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI planner generation failed: %w", err)
	}

	proposals, err := ParseProposals(rawResponse)
	if err != nil {
		// Non-fatal: if AI response parsing failed, proceed with deterministic engine checks
		proposals = []AssessmentProposal{}
	}

	// 3. Deterministic Validation Gate
	var verifiedFindings []*assessment.AssessmentFinding
	if p.engine != nil && target != nil {
		findings, err := p.engine.Run(ctx, target, traffic, scanID)
		if err != nil {
			return nil, fmt.Errorf("deterministic assessment failed: %w", err)
		}
		verifiedFindings = findings
	}

	return &PlanResult{
		Proposals:        proposals,
		VerifiedFindings: verifiedFindings,
		ParameterGraph:   graph,
	}, nil
}

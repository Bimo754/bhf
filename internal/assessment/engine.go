package assessment

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
)

// AssessmentEngine coordinates deterministic security check execution across targets and traffic.
type AssessmentEngine struct {
	rules       []Rule
	findingRepo *database.FindingRepo
	client      *http.Client
}

// NewAssessmentEngine creates a new AssessmentEngine with default verification rules.
func NewAssessmentEngine(findingRepo *database.FindingRepo, client *http.Client) *AssessmentEngine {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		}
	}

	engine := &AssessmentEngine{
		rules:       make([]Rule, 0),
		findingRepo: findingRepo,
		client:      client,
	}

	// Register core deterministic verification rules
	engine.RegisterRule(NewAuthSecurityRule())
	engine.RegisterRule(NewIDORSecurityRule())
	engine.RegisterRule(NewUploadSecurityRule())
	engine.RegisterRule(NewInputValidationRule())

	return engine
}

// RegisterRule registers a security check rule.
func (e *AssessmentEngine) RegisterRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

// Rules returns all registered assessment rules.
func (e *AssessmentEngine) Rules() []Rule {
	return e.rules
}

// Run executes all registered assessment rules against the given traffic slice and persists findings.
func (e *AssessmentEngine) Run(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, scanID string) ([]*AssessmentFinding, error) {
	allFindings := make([]*AssessmentFinding, 0)

	for _, rule := range e.rules {
		select {
		case <-ctx.Done():
			return allFindings, ctx.Err()
		default:
		}

		findings, err := rule.Evaluate(ctx, target, traffic, e.client)
		if err != nil {
			return allFindings, fmt.Errorf("rule %s failed: %w", rule.ID(), err)
		}

		for _, f := range findings {
			allFindings = append(allFindings, f)
			if e.findingRepo != nil && target != nil {
				model := f.ToFindingModel(target.ID, scanID)
				_ = e.findingRepo.Insert(model)
			}
		}
	}

	return allFindings, nil
}

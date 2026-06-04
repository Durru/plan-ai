package change

import (
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type DeepImpactReport struct {
	ID                   string
	ProjectID            string
	ChangeType           ChangeType
	Summary              string
	ArchitectureConcerns []string
	BackendConcerns      []string
	MigrationConcerns    []string
	DocsConcerns         []string
	APIConcerns          []string
	PlanConcerns         []string
	ValidationCommands   []string
	RollbackStrategy     []string
	AffectedPlans        []string
	AffectedTasks        []string
	Severity             Severity
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type DeepImpactRepository interface {
	SaveDeepImpact(DeepImpactReport) (DeepImpactReport, error)
	ListDeepImpacts(projectID string) ([]DeepImpactReport, error)
}

type DeepImpactService struct{ repo DeepImpactRepository }

func NewDeepImpactService(repo DeepImpactRepository) DeepImpactService {
	return DeepImpactService{repo: repo}
}

func (s DeepImpactService) Analyze(projectID string, changeType ChangeType, summary string) (DeepImpactReport, error) {
	return s.repo.SaveDeepImpact(BuildDeepImpactReport(projectID, changeType, summary))
}

func (s DeepImpactService) List(projectID string) ([]DeepImpactReport, error) {
	return s.repo.ListDeepImpacts(projectID)
}

func BuildDeepImpactReport(projectID string, changeType ChangeType, summary string) DeepImpactReport {
	if changeType == "" {
		changeType = TechnologyChanged
	}
	cleanSummary := strings.TrimSpace(summary)
	if cleanSummary == "" {
		cleanSummary = "Change requires V2 impact analysis before targeted regeneration."
	}
	severity := ClassifySeverity(changeType)
	report := DeepImpactReport{
		ID:                   domain.NewID("impactv2"),
		ProjectID:            projectID,
		ChangeType:           changeType,
		Summary:              cleanSummary,
		ArchitectureConcerns: []string{"Review architecture boundaries and dependency direction."},
		BackendConcerns:      []string{"Check persistence, service contracts, and command behavior."},
		MigrationConcerns:    []string{"Use additive migrations and compatibility views only."},
		DocsConcerns:         []string{"Update relevant workflow and architecture docs after approval."},
		APIConcerns:          []string{"Preserve CLI/MCP compatibility unless explicitly versioned."},
		PlanConcerns:         []string{"Regenerate only affected plan sections."},
		ValidationCommands:   []string{"go test ./...", "go vet ./...", "go build ./...", "bash scripts/test-sandbox.sh"},
		RollbackStrategy:     []string{"Revert additive source and migration changes for this impact slice."},
		AffectedPlans:        []string{"master_plan", "specific_plan", "implementation_package"},
		AffectedTasks:        []string{"validation", "documentation", "sandbox verification"},
		Severity:             severity,
		Status:               "draft",
	}
	if strings.Contains(strings.ToLower(cleanSummary), "postgresql") && strings.Contains(strings.ToLower(cleanSummary), "mariadb") {
		report.MigrationConcerns = append(report.MigrationConcerns, "Audit SQL dialect differences between PostgreSQL and MariaDB.")
		report.BackendConcerns = append(report.BackendConcerns, "Review driver, transaction, and JSON behavior changes.")
		report.Severity = SeverityHigh
	}
	return report
}

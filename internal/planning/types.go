package planning

import "time"

type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
	StatusArchived Status = "archived"
)

type MasterPlan struct {
	ID                       string
	ProjectID                string
	Title                    string
	VisionReference          string
	Objectives               []string
	Scope                    []string
	OutOfScope               []string
	RecommendedSpecificPlans []string
	Risks                    []string
	Assumptions              []string
	Status                   Status
	Version                  int
	SupersedesID             string
	SupersededByID           string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type SpecificPlan struct {
	ID                     string
	ProjectID              string
	MasterPlanID           string
	Title                  string
	Goal                   string
	Requirements           []string
	Constraints            []string
	Decisions              []string
	KnowledgeUsed          []string
	ResearchUsed           []string
	ImplementationStrategy string
	Risks                  []string
	ValidationCriteria     []string
	Status                 Status
	Version                int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ImplementationDocument struct {
	ID                  string
	ProjectID           string
	SpecificPlanID      string
	Objective           string
	Architecture        string
	ExpectedFiles       []string
	ExpectedDirectories []string
	Validations         []string
	KnownRisks          []string
	TestingStrategy     string
	RollbackStrategy    string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PlanningInput struct {
	ProjectID            string
	VisionReference      string
	ApprovedRequirements []string
	ApprovedConstraints  []string
	ApprovedDecisions    []string
	ResearchIDs          []string
	KnowledgeIDs         []string
}

type SpecificPlanInput struct {
	ProjectID              string
	Title                  string
	Goal                   string
	Requirements           []string
	Constraints            []string
	Decisions              []string
	KnowledgeUsed          []string
	ResearchUsed           []string
	ImplementationStrategy string
	Risks                  []string
	ValidationCriteria     []string
}

type ImplementationDocumentInput struct {
	ProjectID           string
	Objective           string
	Architecture        string
	ExpectedFiles       []string
	ExpectedDirectories []string
	Validations         []string
	KnownRisks          []string
	TestingStrategy     string
	RollbackStrategy    string
}

// Repository is the canonical service-layer interface for planning
// operations. It diverges from domain.PlanRepository:
//   - Uses Create* prefix instead of Save* for write semantics.
//   - Includes ImplementationDocument methods (domain.PlanRepository
//     delegates those to a separate ImplDocRepository).
//   - Lacks ListSpecificsByMaster and UpdatePlanStatus/Delete
//     currently present only in domain.PlanRepository.
// Use planning.MasterPlan / planning.SpecificPlan types with this
// interface; domain.MasterPlan / domain.SpecificPlan remain for
// store-direct compatibility.
type Repository interface {
	CreateMasterPlan(MasterPlan) (MasterPlan, error)
	CreateSpecificPlan(SpecificPlan) (SpecificPlan, error)
	CreateImplementationDocument(ImplementationDocument) (ImplementationDocument, error)
	GetMasterPlan(string) (MasterPlan, error)
	GetSpecificPlan(string) (SpecificPlan, error)
	GetImplementationDocument(string) (ImplementationDocument, error)
	ListMasterPlans(projectID string) ([]MasterPlan, error)
}

package domain

// ProjectRepository defines persistence operations for Project entities.
type ProjectRepository interface {
	Save(project Project) error
	GetByID(id string) (Project, error)
	FindByName(name string) (Project, error)
	List() ([]Project, error)
	UpdateStatus(id string, status ProjectStatus) error
	Delete(id string) error
}

// VisionRepository defines persistence operations for Vision entities.
type VisionRepository interface {
	Save(vision Vision) error
	GetByID(id string) (Vision, error)
	ListByProject(projectID string) ([]Vision, error)
	Approve(id string) error
	Delete(id string) error
}

// RequirementRepository defines persistence operations for Requirement entities.
type RequirementRepository interface {
	Save(req Requirement) error
	GetByID(id string) (Requirement, error)
	ListByProject(projectID string) ([]Requirement, error)
	ListByType(projectID string, reqType RequirementType) ([]Requirement, error)
	Approve(id string) error
	Delete(id string) error
}

// DecisionRepository defines persistence operations for Decision entities.
type DecisionRepository interface {
	Save(decision Decision) error
	GetByID(id string) (Decision, error)
	ListByProject(projectID string) ([]Decision, error)
	UpdateStatus(id string, status Status) error
	Delete(id string) error
}

// ResearchRepository defines persistence operations for Research entities.
type ResearchRepository interface {
	Save(research Research) error
	GetByID(id string) (Research, error)
	ListByProject(projectID string) ([]Research, error)
	Search(query string) ([]Research, error)
	Delete(id string) error
}

// KnowledgeRepository defines persistence operations for KnowledgeObject entities.
type KnowledgeRepository interface {
	Save(knowledge KnowledgeObject) error
	Update(knowledge KnowledgeObject) error
	GetByID(id string) (KnowledgeObject, error)
	GetByTopic(topic string) (KnowledgeObject, error)
	List() ([]KnowledgeObject, error)
	ListByCategory(category KnowledgeCategory) ([]KnowledgeObject, error)
	Search(query string) ([]KnowledgeObject, error)
	IncrementReuseCount(id string) (KnowledgeObject, error)
	AddTag(knowledgeID, tag string) error
	ListTags(knowledgeID string) ([]KnowledgeTag, error)
	AddRelation(sourceID, targetID string, relationType KnowledgeRelationType) error
	ListRelations(knowledgeID string) ([]KnowledgeRelation, error)
	AddReference(knowledgeID string, referenceType KnowledgeReferenceType, referenceID string) error
	ListReferences(knowledgeID string) ([]KnowledgeReference, error)
	Delete(id string) error
}

// PlanRepository defines persistence operations for MasterPlan and SpecificPlan entities.
type PlanRepository interface {
	SaveMaster(plan MasterPlan) error
	SaveSpecific(plan SpecificPlan) error
	GetMasterByID(id string) (MasterPlan, error)
	GetSpecificByID(id string) (SpecificPlan, error)
	ListMastersByProject(projectID string) ([]MasterPlan, error)
	ListSpecificsByMaster(masterPlanID string) ([]SpecificPlan, error)
	UpdatePlanStatus(id string, status Status) error
	Delete(id string) error
}

// TaskRepository defines persistence operations for Task entities.
type TaskRepository interface {
	Save(task Task) error
	GetByID(id string) (Task, error)
	ListByPhase(phaseID string) ([]Task, error)
	UpdateStatus(id string, status Status) error
	Delete(id string) error
}

// SnapshotRepository defines persistence operations for Snapshot entities.
type SnapshotRepository interface {
	Save(snapshot Snapshot) error
	GetByID(id string) (Snapshot, error)
	ListByProject(projectID string) ([]Snapshot, error)
	Delete(id string) error
}

// ChangeRepository defines persistence operations for ChangeRequest and
// ImpactReport entities. These are managed together because an ImpactReport
// is always derived from a ChangeRequest.
type ChangeRepository interface {
	SaveChangeRequest(cr ChangeRequest) error
	GetChangeRequest(id string) (ChangeRequest, error)
	ListChangeRequests(projectID string) ([]ChangeRequest, error)
	UpdateChangeRequestStatus(id string, status ChangeRequestStatus) error
	SaveImpactReport(report ImpactReport) error
	GetImpactReportByChangeRequest(changeRequestID string) (ImpactReport, error)
	DeleteChangeRequest(id string) error
}

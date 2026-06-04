package change

// Registry maintains the set of valid change types and their metadata.
type Registry struct {
	types map[ChangeType]TypeMeta
}

// TypeMeta describes a change type.
type TypeMeta struct {
	Type        ChangeType `json:"type"`
	DisplayName string     `json:"display_name"`
	Description string     `json:"description"`
	Severity    Severity   `json:"severity"`
}

// NewRegistry creates a registry populated with all known change types.
func NewRegistry() *Registry {
	r := &Registry{types: make(map[ChangeType]TypeMeta)}
	for _, meta := range defaultTypes {
		r.types[meta.Type] = meta
	}
	return r
}

// List returns all registered change types.
func (r *Registry) List() []TypeMeta {
	out := make([]TypeMeta, 0, len(r.types))
	for _, meta := range r.types {
		out = append(out, meta)
	}
	return out
}

// Get returns metadata for a specific change type.
func (r *Registry) Get(ct ChangeType) (TypeMeta, bool) {
	meta, ok := r.types[ct]
	return meta, ok
}

// Register adds a custom change type.
func (r *Registry) Register(meta TypeMeta) {
	r.types[meta.Type] = meta
}

var defaultTypes = []TypeMeta{
	{Type: VisionChanged, DisplayName: "Vision Changed", Description: "The project vision or high-level direction changed", Severity: SeverityHigh},
	{Type: RequirementAdded, DisplayName: "Requirement Added", Description: "A new requirement was added to the project", Severity: SeverityMedium},
	{Type: RequirementRemoved, DisplayName: "Requirement Removed", Description: "An existing requirement was removed", Severity: SeverityHigh},
	{Type: ConstraintChanged, DisplayName: "Constraint Changed", Description: "A project constraint was modified", Severity: SeverityMedium},
	{Type: DecisionChanged, DisplayName: "Decision Changed", Description: "A design or technical decision was changed", Severity: SeverityMedium},
	{Type: ResearchUpdated, DisplayName: "Research Updated", Description: "Research findings or conclusions were updated", Severity: SeverityLow},
	{Type: KnowledgeUpdated, DisplayName: "Knowledge Updated", Description: "Knowledge base content was updated", Severity: SeverityLow},
	{Type: PlanChanged, DisplayName: "Plan Changed", Description: "A plan structure or content changed", Severity: SeverityHigh},
	{Type: TechnologyChanged, DisplayName: "Technology Changed", Description: "Technology choices or dependencies changed", Severity: SeverityLow},
	{Type: ImplementationFeedback, DisplayName: "Implementation Feedback", Description: "Feedback from implementation affecting future plans", Severity: SeverityLow},
}

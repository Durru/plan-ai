package confidencev3

// Score is a 0-100 confidence score where higher means better understood.
type Score int

type ConfidenceReport struct {
	IntentID          string
	IntentScore       Score
	VisionScore       Score
	UXScore           Score
	BusinessScore     Score
	RequirementsScore Score
	ConstraintsScore  Score
	IntentConfidence  Score
	Strengths         []string
	Weaknesses        []string
	Recommendations   []string
}

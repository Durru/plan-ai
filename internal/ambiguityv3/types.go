package ambiguityv3

import "github.com/Durru/plan-ai/internal/discoveryv3"

// AmbiguityScore is a 0-100 score where higher means more ambiguous.
type AmbiguityScore int

type MissingInformation struct {
	Field  string
	Reason string
}

type Assumption struct {
	ID     string
	Reason string
}

type Conflict struct {
	ID       string
	Evidence string
}

type UnknownArea struct {
	Level      discoveryv3.DiscoveryLevel
	QuestionID discoveryv3.QuestionID
	Question   string
	Required   bool
}

type AmbiguityReport struct {
	IntentID           string
	Score              AmbiguityScore
	KnownAreas         []string
	MissingInformation []MissingInformation
	Assumptions        []Assumption
	Conflicts          []Conflict
	UnknownAreas       []UnknownArea
	NeedsToKnow        []string
}

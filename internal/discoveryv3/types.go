package discoveryv3

import "time"

// ──────────────────────────────────────────────
// Discovery Level — ordered progressive levels
// ──────────────────────────────────────────────

type DiscoveryLevel string

const (
	LevelProject      DiscoveryLevel = "project"
	LevelMasterPlan   DiscoveryLevel = "master_plan"
	LevelSpecificPlan DiscoveryLevel = "specific_plan"
	LevelPhase        DiscoveryLevel = "phase"
	LevelTask         DiscoveryLevel = "task"
)

// AllLevels returns levels in progressive order.
func AllLevels() []DiscoveryLevel {
	return []DiscoveryLevel{
		LevelProject,
		LevelMasterPlan,
		LevelSpecificPlan,
		LevelPhase,
		LevelTask,
	}
}

// LevelIndex returns the progressive index of a level (0-4).
func LevelIndex(l DiscoveryLevel) int {
	for i, v := range AllLevels() {
		if v == l {
			return i
		}
	}
	return -1
}

// NextLevel returns the next progressive level.
func NextLevel(l DiscoveryLevel) (DiscoveryLevel, bool) {
	idx := LevelIndex(l)
	if idx < 0 || idx >= len(AllLevels())-1 {
		return "", false
	}
	return AllLevels()[idx+1], true
}

// ──────────────────────────────────────────────
// Discovery Question
// ──────────────────────────────────────────────

type QuestionID string

type Question struct {
	ID            QuestionID
	IntentID      string
	Level         DiscoveryLevel
	Question      string
	Reason        string // why we're asking this
	Required      bool
	RelatedFields []string // intent fields this relates to
	Position      int
	CreatedAt     time.Time
}

// ──────────────────────────────────────────────
// Answer
// ──────────────────────────────────────────────

type Answer struct {
	ID         string
	QuestionID QuestionID
	IntentID   string
	Level      DiscoveryLevel
	Answer     string
	CreatedAt  time.Time
}

// ──────────────────────────────────────────────
// Session status
// ──────────────────────────────────────────────

type SessionStatus struct {
	IntentID       string
	CurrentLevel   DiscoveryLevel
	TotalQuestions int
	AnsweredCount  int
	RemainingCount int
	SuggestedLevel string // progression hint
	AllLevels      []string
	AnsweredLevels []string
}

// ──────────────────────────────────────────────
// Repository interfaces
// ──────────────────────────────────────────────

type QuestionRepository interface {
	SaveQuestions([]Question) error
	GetQuestions(intentID string, level DiscoveryLevel) ([]Question, error)
	GetAllQuestions(intentID string) ([]Question, error)
	GetQuestion(id QuestionID) (Question, error)
}

type AnswerRepository interface {
	SaveAnswer(Answer) (Answer, error)
	GetAnswers(intentID string) ([]Answer, error)
	GetAnswersByLevel(intentID string, level DiscoveryLevel) ([]Answer, error)
}

package discoveryv3

import (
	"fmt"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// Service implements Phase 53 Progressive Discovery System.
// It manages deterministic discovery questions across 5 levels
// (project → master_plan → specific_plan → phase → task) with
// persistence of answers and progression tracking.
type Service struct {
	questionRepo QuestionRepository
	answerRepo   AnswerRepository
	now          func() time.Time
}

func NewService(qr QuestionRepository, ar AnswerRepository) Service {
	return Service{
		questionRepo: qr,
		answerRepo:   ar,
		now:          time.Now,
	}
}

// ──────────────────────────────────────────────
// Phase 53: Progressive Discovery
// ──────────────────────────────────────────────

// Initialize creates and persists the deterministic questions for an intent.
// Idempotent — safe to call multiple times.
func (s Service) Initialize(intentID string) error {
	if intentID == "" {
		return fmt.Errorf("intent_id is required")
	}
	existing, err := s.questionRepo.GetAllQuestions(intentID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil // already initialized
	}
	qs := deterministicQuestions(intentID)
	now := s.now().UTC()
	for i := range qs {
		qs[i].ID = QuestionID(domain.NewID("discq"))
		qs[i].CreatedAt = now
	}
	return s.questionRepo.SaveQuestions(qs)
}

// GetNextQuestions returns unanswered questions for the given level.
// If no level is provided, it starts from the first incomplete level.
func (s Service) GetNextQuestions(intentID string, level DiscoveryLevel) ([]Question, error) {
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	if level == "" {
		// Find the first incomplete level
		lvl, err := s.firstIncompleteLevel(intentID)
		if err != nil {
			return nil, err
		}
		level = lvl
	}
	questions, err := s.questionRepo.GetQuestions(intentID, level)
	if err != nil {
		return nil, err
	}
	answers, err := s.answerRepo.GetAnswers(intentID)
	if err != nil {
		return nil, err
	}
	answeredMap := make(map[QuestionID]bool)
	for _, a := range answers {
		answeredMap[a.QuestionID] = true
	}
	unanswered := make([]Question, 0, len(questions))
	for _, q := range questions {
		if !answeredMap[q.ID] {
			unanswered = append(unanswered, q)
		}
	}
	return unanswered, nil
}

// Answer stores an answer to a discovery question.
func (s Service) Answer(intentID string, questionID QuestionID, answer string) (Answer, error) {
	if intentID == "" {
		return Answer{}, fmt.Errorf("intent_id is required")
	}
	if questionID == "" {
		return Answer{}, fmt.Errorf("question_id is required")
	}
	if answer == "" {
		return Answer{}, fmt.Errorf("answer is required")
	}

	// Look up the question to get its level
	q, err := s.questionRepo.GetQuestion(questionID)
	if err != nil {
		return Answer{}, fmt.Errorf("question not found: %w", err)
	}
	if q.IntentID != intentID {
		return Answer{}, fmt.Errorf("question %s does not belong to intent %s", questionID, intentID)
	}

	a := Answer{
		ID:         domain.NewID("discans"),
		QuestionID: questionID,
		IntentID:   intentID,
		Level:      q.Level,
		Answer:     answer,
		CreatedAt:  s.now().UTC(),
	}
	return s.answerRepo.SaveAnswer(a)
}

// Status returns the current discovery session status for an intent.
func (s Service) Status(intentID string) (SessionStatus, error) {
	if intentID == "" {
		return SessionStatus{}, fmt.Errorf("intent_id is required")
	}
	allQs, err := s.questionRepo.GetAllQuestions(intentID)
	if err != nil {
		return SessionStatus{}, err
	}
	allAnswers, err := s.answerRepo.GetAnswers(intentID)
	if err != nil {
		return SessionStatus{}, err
	}

	// Count answered questions
	answeredMap := make(map[QuestionID]bool)
	for _, a := range allAnswers {
		answeredMap[a.QuestionID] = true
	}

	answeredCount := 0
	for _, q := range allQs {
		if answeredMap[q.ID] {
			answeredCount++
		}
	}

	// Find current level and answered levels
	answeredLevels := make(map[DiscoveryLevel]bool)
	for _, a := range allAnswers {
		answeredLevels[a.Level] = true
	}

	var currentLevel DiscoveryLevel
	var answeredLevelList []string
	var allLevelList []string

	for _, lvl := range AllLevels() {
		allLevelList = append(allLevelList, string(lvl))
		if answeredLevels[lvl] {
			answeredLevelList = append(answeredLevelList, string(lvl))
		}
		if currentLevel == "" {
			// Check if this level's questions are fully answered
			levelQs := filterByLevel(allQs, lvl)
			allAnswered := true
			for _, q := range levelQs {
				if !answeredMap[q.ID] {
					allAnswered = false
					break
				}
			}
			if !allAnswered {
				currentLevel = lvl
			}
		}
	}

	if currentLevel == "" {
		currentLevel = LevelTask // all answered
	}

	remaining := len(allQs) - answeredCount

	nextLevel, hasNext := NextLevel(currentLevel)
	suggestion := string(currentLevel)
	if remaining == 0 && hasNext {
		suggestion = fmt.Sprintf("move to %s", nextLevel)
	}

	return SessionStatus{
		IntentID:       intentID,
		CurrentLevel:   currentLevel,
		TotalQuestions: len(allQs),
		AnsweredCount:  answeredCount,
		RemainingCount: remaining,
		SuggestedLevel: suggestion,
		AllLevels:      allLevelList,
		AnsweredLevels: answeredLevelList,
	}, nil
}

// firstIncompleteLevel finds the first level with unanswered required questions.
func (s Service) firstIncompleteLevel(intentID string) (DiscoveryLevel, error) {
	allQs, err := s.questionRepo.GetAllQuestions(intentID)
	if err != nil {
		return "", err
	}
	allAnswers, err := s.answerRepo.GetAnswers(intentID)
	if err != nil {
		return "", err
	}
	answeredMap := make(map[QuestionID]bool)
	for _, a := range allAnswers {
		answeredMap[a.QuestionID] = true
	}

	for _, lvl := range AllLevels() {
		levelQs := filterByLevel(allQs, lvl)
		for _, q := range levelQs {
			if q.Required && !answeredMap[q.ID] {
				return lvl, nil
			}
		}
		// All required answered, move to next level
	}
	return LevelTask, nil
}

func filterByLevel(qs []Question, level DiscoveryLevel) []Question {
	var out []Question
	for _, q := range qs {
		if q.Level == level {
			out = append(out, q)
		}
	}
	return out
}

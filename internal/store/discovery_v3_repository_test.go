package store

import (
	"testing"

	"github.com/plan-ai/plan-ai/internal/discoveryv3"
)

func TestDiscoveryV3QuestionRepository(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)

	svc := discoveryv3.NewService(qr, ar)

	// Initialize
	intentID := "pintent_test123"
	if err := svc.Initialize(intentID); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	// Verify all questions were created
	allQs, err := qr.GetAllQuestions(intentID)
	if err != nil {
		t.Fatalf("GetAllQuestions: %v", err)
	}

	expectedTotal := 4 + 5 + 4 + 3 + 3 // project + master_plan + specific_plan + phase + task
	if len(allQs) != expectedTotal {
		t.Fatalf("expected %d questions, got %d", expectedTotal, len(allQs))
	}

	// Verify init is idempotent
	if err := svc.Initialize(intentID); err != nil {
		t.Fatalf("Initialize (idempotent): %v", err)
	}
	allQs2, _ := qr.GetAllQuestions(intentID)
	if len(allQs2) != expectedTotal {
		t.Fatalf("expected %d questions after re-init, got %d", expectedTotal, len(allQs2))
	}
}

func TestDiscoveryV3Progression(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)

	svc := discoveryv3.NewService(qr, ar)
	intentID := "pintent_progression"

	svc.Initialize(intentID)

	// Initially should get project-level questions
	qs, err := svc.GetNextQuestions(intentID, "")
	if err != nil {
		t.Fatalf("GetNextQuestions (first): %v", err)
	}
	if len(qs) == 0 {
		t.Fatal("expected questions")
	}
	if qs[0].Level != discoveryv3.LevelProject {
		t.Fatalf("expected level project, got %s", qs[0].Level)
	}

	// Answer all project questions
	for _, q := range qs {
		_, err := svc.Answer(intentID, q.ID, "test answer")
		if err != nil {
			t.Fatalf("Answer(%s): %v", q.ID, err)
		}
	}

	// Next batch should be master_plan
	qs2, err := svc.GetNextQuestions(intentID, "")
	if err != nil {
		t.Fatalf("GetNextQuestions (second): %v", err)
	}
	if len(qs2) == 0 {
		t.Fatal("expected master_plan questions")
	}
	if qs2[0].Level != discoveryv3.LevelMasterPlan {
		t.Fatalf("expected level master_plan, got %s", qs2[0].Level)
	}
}

func TestDiscoveryV3NextQuestionsExcludesAnsweredQuestions(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)
	svc := discoveryv3.NewService(qr, ar)
	intentID := "pintent_unanswered"

	if err := svc.Initialize(intentID); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	projectQuestions, err := svc.GetNextQuestions(intentID, discoveryv3.LevelProject)
	if err != nil {
		t.Fatalf("GetNextQuestions(project): %v", err)
	}
	if len(projectQuestions) < 2 {
		t.Fatalf("expected at least 2 project questions, got %d", len(projectQuestions))
	}

	answeredID := projectQuestions[0].ID
	if _, err := svc.Answer(intentID, answeredID, "already answered"); err != nil {
		t.Fatalf("Answer: %v", err)
	}

	nextQuestions, err := svc.GetNextQuestions(intentID, discoveryv3.LevelProject)
	if err != nil {
		t.Fatalf("GetNextQuestions(project after answer): %v", err)
	}
	for _, q := range nextQuestions {
		if q.ID == answeredID {
			t.Fatalf("answered question %s was returned again", answeredID)
		}
	}
	if len(nextQuestions) != len(projectQuestions)-1 {
		t.Fatalf("expected %d unanswered questions, got %d", len(projectQuestions)-1, len(nextQuestions))
	}
}

func TestDiscoveryV3Status(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)

	svc := discoveryv3.NewService(qr, ar)
	intentID := "pintent_status"

	svc.Initialize(intentID)

	// Status before answering
	s, err := svc.Status(intentID)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if s.AnsweredCount != 0 {
		t.Fatalf("expected 0 answered, got %d", s.AnsweredCount)
	}
	if s.TotalQuestions != 19 {
		t.Fatalf("expected 19 total, got %d", s.TotalQuestions)
	}
	if s.RemainingCount != 19 {
		t.Fatalf("expected 19 remaining, got %d", s.RemainingCount)
	}

	// Answer one question
	qs, _ := svc.GetNextQuestions(intentID, "")
	svc.Answer(intentID, qs[0].ID, "answer one")

	s2, _ := svc.Status(intentID)
	if s2.AnsweredCount != 1 {
		t.Fatalf("expected 1 answered, got %d", s2.AnsweredCount)
	}
	if s2.RemainingCount != 18 {
		t.Fatalf("expected 18 remaining, got %d", s2.RemainingCount)
	}
}

func TestDiscoveryV3AnswerPersistence(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)

	svc := discoveryv3.NewService(qr, ar)
	intentID := "pintent_persist"

	svc.Initialize(intentID)
	qs, _ := svc.GetNextQuestions(intentID, "")

	a, err := svc.Answer(intentID, qs[0].ID, "persistent answer")
	if err != nil {
		t.Fatalf("Answer: %v", err)
	}
	if a.ID == "" {
		t.Fatal("expected non-empty answer id")
	}
	if a.Answer != "persistent answer" {
		t.Fatalf("answer = %q, want 'persistent answer'", a.Answer)
	}

	// Verify via repo
	answers, err := ar.GetAnswers(intentID)
	if err != nil {
		t.Fatalf("GetAnswers: %v", err)
	}
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}
	if answers[0].QuestionID != qs[0].ID {
		t.Fatalf("question_id = %q, want %q", answers[0].QuestionID, qs[0].ID)
	}
}

func TestDiscoveryV3Validation(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	qr := NewDiscoveryV3QuestionRepository(ps.DB)
	ar := NewDiscoveryV3AnswerRepository(ps.DB)

	svc := discoveryv3.NewService(qr, ar)

	// Empty intent_id
	if err := svc.Initialize(""); err == nil {
		t.Fatal("expected error for empty intent_id")
	}

	// Answer with empty values
	if _, err := svc.Answer("", "q1", "answer"); err == nil {
		t.Fatal("expected error for empty intent_id")
	}
	if _, err := svc.Answer("intent", "", "answer"); err == nil {
		t.Fatal("expected error for empty question_id")
	}
	if _, err := svc.Answer("intent", "q1", ""); err == nil {
		t.Fatal("expected error for empty answer")
	}

	// Status with empty intent_id
	if _, err := svc.Status(""); err == nil {
		t.Fatal("expected error for empty intent_id")
	}
}

func TestDiscoveryV3Levels(t *testing.T) {
	// Verify level progression
	if discoveryv3.LevelIndex(discoveryv3.LevelProject) != 0 {
		t.Fatal("LevelProject should be index 0")
	}
	if discoveryv3.LevelIndex(discoveryv3.LevelTask) != 4 {
		t.Fatal("LevelTask should be index 4")
	}

	next, ok := discoveryv3.NextLevel(discoveryv3.LevelProject)
	if !ok || next != discoveryv3.LevelMasterPlan {
		t.Fatalf("next after project should be master_plan, got %s", next)
	}

	_, ok = discoveryv3.NextLevel(discoveryv3.LevelTask)
	if ok {
		t.Fatal("next after task should be false")
	}

	if discoveryv3.LevelIndex("invalid") != -1 {
		t.Fatal("invalid level should return -1")
	}
}

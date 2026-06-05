package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/discoveryv3"
)

// ──────────────────────────────────────────────
// Question Repository
// ──────────────────────────────────────────────

type DiscoveryV3QuestionRepository struct{ db *sql.DB }

func NewDiscoveryV3QuestionRepository(db *sql.DB) DiscoveryV3QuestionRepository {
	return DiscoveryV3QuestionRepository{db: db}
}

var _ discoveryv3.QuestionRepository = DiscoveryV3QuestionRepository{}

func (r DiscoveryV3QuestionRepository) SaveQuestions(qs []discoveryv3.Question) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO discovery_v3_questions (id, intent_id, level, question, reason, required, related_fields, position, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, q := range qs {
		c := nowString()
		if !q.CreatedAt.IsZero() {
			c = q.CreatedAt.Format(time.RFC3339)
		}
		if q.CreatedAt.IsZero() {
			q.CreatedAt = parseRFC3339(c)
		}
		_, err := stmt.Exec(
			string(q.ID), q.IntentID, string(q.Level),
			q.Question, q.Reason, boolToInt(q.Required),
			mustJSON(q.RelatedFields), q.Position, c,
		)
		if err != nil {
			return fmt.Errorf("save question %s: %w", q.ID, err)
		}
	}
	return tx.Commit()
}

func (r DiscoveryV3QuestionRepository) GetQuestions(intentID string, level discoveryv3.DiscoveryLevel) ([]discoveryv3.Question, error) {
	return r.listQuestions("WHERE intent_id = ? AND level = ? ORDER BY position ASC", intentID, string(level))
}

func (r DiscoveryV3QuestionRepository) GetAllQuestions(intentID string) ([]discoveryv3.Question, error) {
	return r.listQuestions("WHERE intent_id = ? ORDER BY position ASC", intentID)
}

func (r DiscoveryV3QuestionRepository) GetQuestion(id discoveryv3.QuestionID) (discoveryv3.Question, error) {
	items, err := r.listQuestions("WHERE id = ?", string(id))
	if err != nil {
		return discoveryv3.Question{}, err
	}
	if len(items) == 0 {
		return discoveryv3.Question{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r DiscoveryV3QuestionRepository) listQuestions(where string, args ...any) ([]discoveryv3.Question, error) {
	rows, err := r.db.Query(`SELECT id, intent_id, level, question, reason, required, related_fields, position, created_at FROM discovery_v3_questions `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []discoveryv3.Question
	for rows.Next() {
		var q discoveryv3.Question
		var id, intentID, level, question, reason, rf, c string
		var required int
		if err := rows.Scan(&id, &intentID, &level, &question, &reason, &required, &rf, &q.Position, &c); err != nil {
			return nil, err
		}
		q.ID = discoveryv3.QuestionID(id)
		q.IntentID = intentID
		q.Level = discoveryv3.DiscoveryLevel(level)
		q.Question = question
		q.Reason = reason
		q.Required = required == 1
		decodeJSON(rf, &q.RelatedFields)
		q.CreatedAt = parseRFC3339(c)
		out = append(out, q)
	}
	return out, rows.Err()
}

// ──────────────────────────────────────────────
// Answer Repository
// ──────────────────────────────────────────────

type DiscoveryV3AnswerRepository struct{ db *sql.DB }

func NewDiscoveryV3AnswerRepository(db *sql.DB) DiscoveryV3AnswerRepository {
	return DiscoveryV3AnswerRepository{db: db}
}

var _ discoveryv3.AnswerRepository = DiscoveryV3AnswerRepository{}

func (r DiscoveryV3AnswerRepository) SaveAnswer(a discoveryv3.Answer) (discoveryv3.Answer, error) {
	c := nowString()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = parseRFC3339(c)
	}
	_, err := r.db.Exec(`INSERT INTO discovery_v3_answers (id, question_id, intent_id, level, answer, created_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET answer=excluded.answer`,
		a.ID, string(a.QuestionID), a.IntentID, string(a.Level), a.Answer, c)
	if err != nil {
		return discoveryv3.Answer{}, err
	}
	a.CreatedAt = parseRFC3339(c)
	return a, nil
}

func (r DiscoveryV3AnswerRepository) GetAnswers(intentID string) ([]discoveryv3.Answer, error) {
	return r.listAnswers("WHERE intent_id = ? ORDER BY created_at ASC", intentID)
}

func (r DiscoveryV3AnswerRepository) GetAnswersByLevel(intentID string, level discoveryv3.DiscoveryLevel) ([]discoveryv3.Answer, error) {
	return r.listAnswers("WHERE intent_id = ? AND level = ? ORDER BY created_at ASC", intentID, string(level))
}

func (r DiscoveryV3AnswerRepository) listAnswers(where string, args ...any) ([]discoveryv3.Answer, error) {
	rows, err := r.db.Query(`SELECT id, question_id, intent_id, level, answer, created_at FROM discovery_v3_answers `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []discoveryv3.Answer
	for rows.Next() {
		var a discoveryv3.Answer
		var id, qID, intentID, level, answer, c string
		if err := rows.Scan(&id, &qID, &intentID, &level, &answer, &c); err != nil {
			return nil, err
		}
		a.ID = id
		a.QuestionID = discoveryv3.QuestionID(qID)
		a.IntentID = intentID
		a.Level = discoveryv3.DiscoveryLevel(level)
		a.Answer = answer
		a.CreatedAt = parseRFC3339(c)
		out = append(out, a)
	}
	return out, rows.Err()
}

// ──────────────────────────────────────────────
// Migration 0041 constant
// ──────────────────────────────────────────────

const projectV3DiscoveryProgressiveSQL = `
-- 0041_v3_discovery_progressive: Phase 53 Progressive Discovery System
CREATE TABLE IF NOT EXISTS discovery_v3_questions (
  id TEXT PRIMARY KEY,
  intent_id TEXT NOT NULL,
  level TEXT NOT NULL,
  question TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  required INTEGER NOT NULL DEFAULT 0,
  related_fields TEXT NOT NULL DEFAULT '[]',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_questions_intent ON discovery_v3_questions(intent_id);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_questions_level ON discovery_v3_questions(level);

CREATE TABLE IF NOT EXISTS discovery_v3_answers (
  id TEXT PRIMARY KEY,
  question_id TEXT NOT NULL,
  intent_id TEXT NOT NULL,
  level TEXT NOT NULL,
  answer TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_answers_intent ON discovery_v3_answers(intent_id);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_answers_question ON discovery_v3_answers(question_id);
`

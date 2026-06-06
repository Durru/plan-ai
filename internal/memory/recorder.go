package memory

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// Recorder writes memory records automatically when important project
// events occur: approved context, approved research, applied proposals,
// change events. It is the single authority for "what gets remembered."
//
// All writes use the memory.Service (and Store) underneath, ensuring
// FTS-backed storage. The Recorder also supports supersession and
// topic-key-based lookup.
type Recorder struct {
	svc *Service
}

// NewRecorder creates a recorder backed by the given store.
func NewRecorder(store Store) *Recorder {
	return &Recorder{svc: NewService(store)}
}

// RecordApprovedContext creates a memory entry for an approved context item.
func (r *Recorder) RecordApprovedContext(projectID, atype, content string) (Entry, error) {
	entryType := contextTypeToMemory(atype)
	title := truncate(content, 120)
	e := NewEntry(projectID, entryType, title, content, atype, "approved-context")
	e.Answer = content
	e.Source = e.Source + " topic:" + topicKey(atype, content)
	if _, err := r.svc.store.Add(e); err != nil {
		return e, err
	}
	return e, nil
}

// RecordApprovedResearch creates a memory entry when research is approved.
func (r *Recorder) RecordApprovedResearch(projectID, researchID, topic, summary string) (Entry, error) {
	e := NewEntry(projectID, TypeResearch, truncate(topic, 120), summary, researchID, "approved-research")
	e.Answer = summary
	e.Source = e.Source + " topic:" + topicKey("research", topic)
	if _, err := r.svc.store.Add(e); err != nil {
		return e, err
	}
	return e, nil
}

// RecordAppliedProposal creates a memory entry from an applied proposal.
func (r *Recorder) RecordAppliedProposal(projectID, proposalID, summary string) (Entry, error) {
	e := NewEntry(projectID, TypeChange, truncate(summary, 120), summary, proposalID, "applied-proposal")
	e.Answer = summary
	e.Source = e.Source + " topic:" + topicKey("proposal", proposalID)
	if _, err := r.svc.store.Add(e); err != nil {
		return e, err
	}
	return e, nil
}

// RecordChangeEvent creates a memory entry from a change event.
func (r *Recorder) RecordChangeEvent(projectID, eventType, summary string) (Entry, error) {
	e := NewEntry(projectID, TypeChange, truncate(summary, 120), summary, eventType, "change-event")
	e.Answer = summary
	e.Source = e.Source + " topic:" + topicKey("change", eventType)
	if _, err := r.svc.store.Add(e); err != nil {
		return e, err
	}
	return e, nil
}

// FindByTopicKey searches memory entries by their topic_key in source.
func (r *Recorder) FindByTopicKey(projectID, topicKey string) ([]Entry, error) {
	return r.svc.Search(projectID, topicKey)
}

// Search returns memory entries matching the given query using
// LIKE on title, content, and question fields.
func (r *Recorder) Search(projectID, query string) ([]Entry, error) {
	return r.svc.Search(projectID, query)
}

// Supersede marks an existing entry as "superseded" and creates a
// replacement. Returns the new entry.
func (r *Recorder) Supersede(projectID, oldTopicKey string, newEntry Entry) (Entry, error) {
	old, err := r.FindByTopicKey(projectID, oldTopicKey)
	if err != nil {
		return newEntry, err
	}
	for _, o := range old {
		o.Status = "superseded"
		if _, err := r.svc.store.Update(o); err != nil {
			return newEntry, fmt.Errorf("supersede %s: %w", o.ID, err)
		}
	}
	return r.svc.store.Add(newEntry)
}

// ── helpers ──

func (r *Recorder) upsertWithTopicKey(e Entry, tk string) error {
	e.Source = e.Source + " topic:" + tk
	if _, err := r.svc.store.Add(e); err != nil {
		return err
	}
	return nil
}

func contextTypeToMemory(at string) EntryType {
	switch at {
	case "requirement":
		return TypeDecision
	case "constraint":
		return TypeReference
	case "decision":
		return TypeDecision
	case "preference":
		return TypeReference
	case "goal":
		return TypePlan
	case "reference":
		return TypeReference
	default:
		return TypeDecision
	}
}

func topicKey(prefix, content string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(content)))
	key := fmt.Sprintf("%x", h[:8])
	return fmt.Sprintf("%s:%s", prefix, key)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func parseRFC3339(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

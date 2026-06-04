package context

import "fmt"

type Registry struct{ repo Repository }

func NewRegistry(repo Repository) Registry { return Registry{repo: repo} }

func (r Registry) StoreApproved(item ApprovedItem) (ApprovedItem, error) {
	if item.ProjectID == "" {
		return ApprovedItem{}, fmt.Errorf("project id is required")
	}
	if item.Type == "" {
		return ApprovedItem{}, fmt.Errorf("approved type is required")
	}
	if item.Content == "" {
		return ApprovedItem{}, fmt.Errorf("content is required")
	}
	item.State = StateApproved
	return r.repo.StoreApproved(item)
}

func (r Registry) GetApproved(itemType ApprovedType, id string) (ApprovedItem, error) {
	return r.repo.GetApproved(itemType, id)
}

func (r Registry) ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error) {
	return r.repo.ListApproved(projectID, itemType)
}

func (r Registry) FindApproved(projectID string, itemType ApprovedType, query string) ([]ApprovedItem, error) {
	return r.repo.FindApproved(projectID, itemType, query)
}

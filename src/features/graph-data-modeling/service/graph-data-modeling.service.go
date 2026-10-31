package service

import (
	"context"
	"time"

	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	graphPorts "knowledge-graph-platform/src/features/graph-data-modeling/infrastructure/ports"
	"knowledge-graph-platform/src/shared/errors"
)

type GraphDataModelingService struct {
	repo graphPorts.GraphRepositoryPort
}

func NewGraphDataModelingService(repo graphPorts.GraphRepositoryPort) *GraphDataModelingService {
	return &GraphDataModelingService{
		repo: repo,
	}
}

func (s *GraphDataModelingService) CreateNode(ctx context.Context, input graphContracts.CreateNodeInput) (*graphContracts.GraphNode, error) {
	if err := graphContracts.ValidateCreateNodeInput(input); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	properties := input.Properties
	if properties == "" {
		properties = "{}"
	}
	node := &graphContracts.GraphNode{
		ID:         input.ID,
		TenantID:   input.TenantID,
		Label:      input.Label,
		Properties: properties,
		Status:     graphContracts.NodeStatusActive,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.CreateNode(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *GraphDataModelingService) GetNode(ctx context.Context, tenantID string, id string) (*graphContracts.GraphNode, error) {
	if tenantID == "" || id == "" {
		return nil, errors.NewValidationError("tenantId and id are required")
	}
	return s.repo.GetNodeByID(ctx, tenantID, id)
}

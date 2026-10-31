package ports

import (
	"context"

	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
)

type GraphRepositoryPort interface {
	GetNodeByID(ctx context.Context, tenantID string, id string) (*graphContracts.GraphNode, error)
	CreateNode(ctx context.Context, node *graphContracts.GraphNode) error
}

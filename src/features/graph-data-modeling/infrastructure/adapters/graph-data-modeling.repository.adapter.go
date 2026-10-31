package adapters

import (
	"context"
	"sync"

	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	"knowledge-graph-platform/src/shared/errors"
)

type MemoryGraphRepositoryAdapter struct {
	mu    sync.RWMutex
	store map[string]*graphContracts.GraphNode
}

func NewMemoryGraphRepositoryAdapter() *MemoryGraphRepositoryAdapter {
	return &MemoryGraphRepositoryAdapter{
		store: make(map[string]*graphContracts.GraphNode),
	}
}

func (a *MemoryGraphRepositoryAdapter) GetNodeByID(ctx context.Context, tenantID string, id string) (*graphContracts.GraphNode, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	key := tenantID + ":" + id
	node, exists := a.store[key]
	if !exists {
		return nil, errors.NewNotFoundError("graph node not found")
	}
	return node, nil
}

func (a *MemoryGraphRepositoryAdapter) CreateNode(ctx context.Context, node *graphContracts.GraphNode) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := node.TenantID + ":" + node.ID
	if _, exists := a.store[key]; exists {
		return errors.NewConflictError("graph node already exists")
	}
	a.store[key] = node
	return nil
}

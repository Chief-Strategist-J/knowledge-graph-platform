package machines

import (
	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	graphRules "knowledge-graph-platform/src/features/graph-data-modeling/domain/rules"
	"knowledge-graph-platform/src/shared/errors"
)

type GraphNodeStateMachine struct{}

func NewGraphNodeStateMachine() *GraphNodeStateMachine {
	return &GraphNodeStateMachine{}
}

func (m *GraphNodeStateMachine) TransitionTo(current graphContracts.NodeStatus, next graphContracts.NodeStatus) (graphContracts.NodeStatus, error) {
	if !graphRules.CanTransitionNodeStatus(current, next) {
		return current, errors.NewConflictError("invalid graph node status transition")
	}
	return next, nil
}

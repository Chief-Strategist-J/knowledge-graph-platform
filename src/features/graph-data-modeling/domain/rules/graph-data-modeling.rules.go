package rules

import (
	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
)

func CanTransitionNodeStatus(from graphContracts.NodeStatus, to graphContracts.NodeStatus) bool {
	switch from {
	case graphContracts.NodeStatusDraft:
		return to == graphContracts.NodeStatusActive || to == graphContracts.NodeStatusArchived
	case graphContracts.NodeStatusActive:
		return to == graphContracts.NodeStatusDeprecated || to == graphContracts.NodeStatusArchived
	case graphContracts.NodeStatusDeprecated:
		return to == graphContracts.NodeStatusArchived
	default:
		return false
	}
}

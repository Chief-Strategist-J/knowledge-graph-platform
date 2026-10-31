package tests

import (
	"testing"

	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	graphRules "knowledge-graph-platform/src/features/graph-data-modeling/domain/rules"
)

func TestCanTransitionNodeStatus(t *testing.T) {
	if !graphRules.CanTransitionNodeStatus(graphContracts.NodeStatusDraft, graphContracts.NodeStatusActive) {
		t.Errorf("expected transition draft -> active to be valid")
	}
	if graphRules.CanTransitionNodeStatus(graphContracts.NodeStatusDraft, graphContracts.NodeStatusDeprecated) {
		t.Errorf("expected transition draft -> deprecated to be invalid")
	}
}

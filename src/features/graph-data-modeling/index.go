package graphdatamodeling

import (
	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	graphService "knowledge-graph-platform/src/features/graph-data-modeling/service"
)

type Service = graphService.GraphDataModelingService
type Node = graphContracts.GraphNode
type Status = graphContracts.NodeStatus

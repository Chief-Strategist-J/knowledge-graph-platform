package queries

const (
	FlowInsertNode = "INSERT INTO graph_nodes (id, tenant_id, label, properties, status, version, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	FlowSelectNodeByID = "SELECT id, tenant_id, label, properties, status, version, created_at, updated_at FROM graph_nodes WHERE tenant_id = $1 AND id = $2"
)

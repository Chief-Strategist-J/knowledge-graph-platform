package graphdatamodeling

type NodeStatus string

const (
	NodeStatusDraft      NodeStatus = "DRAFT"
	NodeStatusActive     NodeStatus = "ACTIVE"
	NodeStatusDeprecated NodeStatus = "DEPRECATED"
	NodeStatusArchived   NodeStatus = "ARCHIVED"
)

type GraphNode struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenantId"`
	Label      string     `json:"label"`
	Properties string     `json:"properties"`
	Status     NodeStatus `json:"status"`
	Version    int64      `json:"version"`
	CreatedAt  string     `json:"createdAt"`
	UpdatedAt  string     `json:"updatedAt"`
}

type CreateNodeInput struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenantId"`
	Label      string `json:"label"`
	Properties string `json:"properties"`
}

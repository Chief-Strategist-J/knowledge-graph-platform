package routers

import (
	"net/http"

	"knowledge-graph-platform/src/api/rest/v1/handlers"
)

type GraphDataModelingRouter struct {
	handler *handlers.GraphDataModelingHandler
}

func NewGraphDataModelingRouter(h *handlers.GraphDataModelingHandler) *GraphDataModelingRouter {
	return &GraphDataModelingRouter{handler: h}
}

func (r *GraphDataModelingRouter) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.handler.CreateNode(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	graphContracts "knowledge-graph-platform/src/features/graph-data-modeling/contracts"
	graphService "knowledge-graph-platform/src/features/graph-data-modeling/service"
	"knowledge-graph-platform/src/shared/constants"
	appErrors "knowledge-graph-platform/src/shared/errors"
	"knowledge-graph-platform/src/shared/types"
)

type GraphDataModelingHandler struct {
	service *graphService.GraphDataModelingService
}

func NewGraphDataModelingHandler(s *graphService.GraphDataModelingService) *GraphDataModelingHandler {
	return &GraphDataModelingHandler{service: s}
}

func (h *GraphDataModelingHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	w.Header().Set("Content-Type", "application/json")
	traceparent := r.Header.Get(constants.HeaderTraceparent)
	reqID := r.Header.Get(constants.HeaderXRequestID)
	corrID := r.Header.Get(constants.HeaderXCorrelationID)

	var input graphContracts.CreateNodeInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrCodeValidationFailed, "malformed request payload", start, reqID, corrID)
		return
	}

	result, err := h.service.CreateNode(r.Context(), input)
	if err != nil {
		if appErr, ok := err.(*appErrors.ApplicationError); ok {
			h.writeError(w, appErr.StatusCode, appErr.CanonicalCode, appErr.Message, start, reqID, corrID)
			return
		}
		h.writeError(w, http.StatusInternalServerError, constants.ErrCodeInternalError, "internal error", start, reqID, corrID)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set(constants.HeaderTraceparent, traceparent)
	resp := types.ApiResponse{
		Success:    true,
		StatusCode: http.StatusCreated,
		Data:       result,
		Meta: types.ResponseMeta{
			RequestID:       reqID,
			CorrelationID:   corrID,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			ExecutionTimeMs: time.Since(start).Milliseconds(),
			APIVersion:      "v1",
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *GraphDataModelingHandler) writeError(w http.ResponseWriter, code int, errCode string, msg string, start time.Time, reqID string, corrID string) {
	w.WriteHeader(code)
	resp := types.ApiResponse{
		Success:    false,
		StatusCode: code,
		Error: &types.ResponseError{
			Code:      errCode,
			Message:   msg,
			Retryable: code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable,
		},
		Meta: types.ResponseMeta{
			RequestID:       reqID,
			CorrelationID:   corrID,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			ExecutionTimeMs: time.Since(start).Milliseconds(),
			APIVersion:      "v1",
		},
	}
	json.NewEncoder(w).Encode(resp)
}

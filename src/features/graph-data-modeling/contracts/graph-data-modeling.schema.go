package graphdatamodeling

import (
	"knowledge-graph-platform/src/shared/errors"
)

func ValidateCreateNodeInput(input CreateNodeInput) error {
	if input.ID == "" {
		return errors.NewValidationError("id is required")
	}
	if input.TenantID == "" {
		return errors.NewValidationError("tenantId is required")
	}
	if input.Label == "" {
		return errors.NewValidationError("label is required")
	}
	return nil
}

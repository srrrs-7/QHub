package admin

import (
	"api/src/services/batchservice"
)

// AdminHandler serves administrative batch operation endpoints.
type AdminHandler struct {
	batch *batchservice.BatchService
}

// NewAdminHandler creates a new AdminHandler with the given batch service.
func NewAdminHandler(batch *batchservice.BatchService) *AdminHandler {
	return &AdminHandler{batch: batch}
}

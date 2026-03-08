package response

import (
	"api/src/domain/apperror"
	"encoding/json"
	"errors"
	"net/http"
)

func OK(w http.ResponseWriter, body any) {
	writeJSON(w, http.StatusOK, body)
}

func Created(w http.ResponseWriter, body any) {
	writeJSON(w, http.StatusCreated, body)
}

func Accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func MapSlice[S any, D any](src []S, fn func(S) D) []D {
	result := make([]D, 0, len(src))
	for _, s := range src {
		result = append(result, fn(s))
	}
	return result
}

// HandleError maps any error to an HTTP response.
func HandleError(w http.ResponseWriter, err error) {
	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		handleAppError(w, appErr)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

var errorStatusMap = map[string]int{
	apperror.ValidationErrorName:     http.StatusBadRequest,
	apperror.BadRequestErrorName:     http.StatusBadRequest,
	apperror.NotFoundErrorName:       http.StatusNotFound,
	apperror.UnauthorizedErrorName:   http.StatusUnauthorized,
	apperror.ForbiddenErrorName:      http.StatusForbidden,
	apperror.ConflictErrorName:       http.StatusConflict,
	apperror.DatabaseErrorName:       http.StatusInternalServerError,
	apperror.InternalServerErrorName: http.StatusInternalServerError,
}

func handleAppError(w http.ResponseWriter, err apperror.AppError) {
	status, ok := errorStatusMap[err.ErrorName()]
	if !ok {
		status = http.StatusInternalServerError
	}
	writeError(w, status, err)
}

type errorResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Domain  string `json:"domain"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, err apperror.AppError) {
	writeJSON(w, status, errorResponse{
		Message: err.Error(),
		Type:    err.ErrorName(),
		Domain:  err.DomainName(),
	})
}

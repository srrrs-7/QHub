package response

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"
)

func OK(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Created(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// handleAppError - AppErrorを網羅的に処理し、適切なHTTPレスポンスを返す
func HandleAppError(w http.ResponseWriter, err apperror.AppError) {
	errName := err.ErrorName()

	switch errName {
	case apperror.ValidationErrorName:
		badRequest(w, err)
	case apperror.NotFoundErrorName:
		notFound(w, err)
	case apperror.UnauthorizedErrorName:
		unauthorized(w, err)
	case apperror.ForbiddenErrorName:
		forbidden(w, err)
	case apperror.BadRequestErrorName:
		badRequest(w, err)
	case apperror.ConflictErrorName:
		conflict(w, err)
	case apperror.DatabaseErrorName:
		internalError(w, err)
	case apperror.InternalServerErrorName:
		internalError(w, err)
	default:
		unexpectedError(w, err)
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Domain  string `json:"domain"`
}

func newErrorResponse(err apperror.AppError) ErrorResponse {
	return ErrorResponse{
		Message: err.Error(),
		Type:    err.ErrorName(),
		Domain:  err.DomainName(),
	}
}

func writeError(w http.ResponseWriter, status int, err apperror.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(newErrorResponse(err)); encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

func badRequest(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusBadRequest, err)
}

func notFound(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusNotFound, err)
}

func unauthorized(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusUnauthorized, err)
}

func internalError(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusInternalServerError, err)
}

func forbidden(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusForbidden, err)
}

func conflict(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusConflict, err)
}

func unexpectedError(w http.ResponseWriter, err apperror.AppError) {
	writeError(w, http.StatusInternalServerError, err)
}

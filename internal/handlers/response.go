package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/domain"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	WriteJSON(w, status, resp)
}

func WriteAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*domain.AppError); ok {
		status := getStatusCode(appErr.Code)
		WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	WriteError(w, http.StatusInternalServerError, domain.ErrCodeInternalError, err.Error())
}

func getStatusCode(errCode string) int {
	switch errCode {
	case domain.ErrCodeTeamExists:
		return http.StatusBadRequest
	case domain.ErrCodePRExists, domain.ErrCodePRMerged, domain.ErrCodeNotAssigned, domain.ErrCodeNoCandidate:
		return http.StatusConflict
	case domain.ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

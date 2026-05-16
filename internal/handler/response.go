package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tvizzi/org-api-test-task/internal/service"
)

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg, Code: code})
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	default:
		slog.Error("internal server error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

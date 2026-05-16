package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tvizzi/org-api-test-task/internal/dto"
	"github.com/tvizzi/org-api-test-task/internal/service"
)

type EmployeeHandler struct {
	svc *service.EmployeeService
}

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /departments/{id}/employees/", h.Create)
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	deptID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid department id format")
		return
	}

	var req dto.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	resp, err := h.svc.Create(r.Context(), deptID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

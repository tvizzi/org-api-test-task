package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tvizzi/org-api-test-task/internal/dto"
	"github.com/tvizzi/org-api-test-task/internal/service"
)

type DepartmentHandler struct {
	svc *service.DepartmentService
}

func NewDepartmentHandler(svc *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

func (h *DepartmentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /departments/", h.Create)
	mux.HandleFunc("GET /departments/{id}", h.GetTree)
	mux.HandleFunc("PATCH /departments/{id}", h.Update)
	mux.HandleFunc("DELETE /departments/{id}", h.Delete)
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	resp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *DepartmentHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid department id")
		return
	}

	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		parsedDepth, err := strconv.Atoi(d)
		if err != nil || parsedDepth < 1 {
			writeError(w, http.StatusBadRequest, "invalid_depth", "depth must be a positive integer")
			return
		}
		depth = parsedDepth
	}

	includeEmp := true
	if inc := r.URL.Query().Get("include_employees"); inc != "" {
		parsedInc, err := strconv.ParseBool(inc)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_query", "include_employees must be a boolean (true/false)")
			return
		}
		includeEmp = parsedInc
	}

	resp, err := h.svc.GetTree(r.Context(), id, depth, includeEmp)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid department id")
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	resp, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid department id")
		return
	}

	mode := r.URL.Query().Get("mode")
	var reassignID *int

	if rIDStr := r.URL.Query().Get("reassign_to_department_id"); rIDStr != "" {
		if rID, err := strconv.Atoi(rIDStr); err == nil {
			reassignID = &rID
		} else {
			writeError(w, http.StatusBadRequest, "invalid_query", "invalid reassign_to_department_id")
			return
		}
	}

	if err := h.svc.Delete(r.Context(), id, mode, reassignID); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

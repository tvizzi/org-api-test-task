package service

import (
	"context"
	"strings"
	"time"

	"github.com/tvizzi/org-api-test-task/internal/dto"
	"github.com/tvizzi/org-api-test-task/internal/model"
	"github.com/tvizzi/org-api-test-task/internal/repository"
)

type EmployeeService struct {
	empRepo  repository.EmployeeRepository
	deptRepo repository.DepartmentRepository
}

func NewEmployeeService(er repository.EmployeeRepository, dr repository.DepartmentRepository) *EmployeeService {
	return &EmployeeService{empRepo: er, deptRepo: dr}
}

func (s *EmployeeService) Create(ctx context.Context, deptID int, req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	req.Position = strings.TrimSpace(req.Position)

	if len(req.FullName) == 0 || len(req.FullName) > 200 || len(req.Position) == 0 || len(req.Position) > 200 {
		return nil, ErrInvalidInput
	}

	exists, err := s.deptRepo.ExistsByID(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	var hiredAt *time.Time
	if req.HiredAt != nil {
		parsed, err := time.Parse("2006-01-02", *req.HiredAt)
		if err != nil {
			return nil, ErrInvalidInput
		}
		hiredAt = &parsed
	}

	emp := &model.Employee{
		DepartmentID: deptID,
		FullName:     req.FullName,
		Position:     req.Position,
		HiredAt:      hiredAt,
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}

	return &dto.EmployeeResponse{
		ID:           emp.ID,
		DepartmentID: emp.DepartmentID,
		FullName:     emp.FullName,
		Position:     emp.Position,
		HiredAt:      emp.HiredAt,
		CreatedAt:    emp.CreatedAt,
	}, nil
}

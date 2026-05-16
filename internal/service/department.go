package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tvizzi/org-api-test-task/internal/dto"
	"github.com/tvizzi/org-api-test-task/internal/model"
	"github.com/tvizzi/org-api-test-task/internal/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input data")
	ErrNotFound     = repository.ErrNotFound
	ErrConflict     = repository.ErrConflict
)

type DepartmentService struct {
	deptRepo repository.DepartmentRepository
	empRepo  repository.EmployeeRepository
	tx       repository.Transactor
}

func NewDepartmentService(dr repository.DepartmentRepository, er repository.EmployeeRepository, tx repository.Transactor) *DepartmentService {
	return &DepartmentService{deptRepo: dr, empRepo: er, tx: tx}
}

func (s *DepartmentService) Create(ctx context.Context, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 || len(req.Name) > 200 {
		return nil, ErrInvalidInput
	}

	if req.ParentID != nil {
		exists, err := s.deptRepo.ExistsByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}

	dept := &model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if err := s.deptRepo.Create(ctx, dept); err != nil {
		return nil, err
	}

	return &dto.DepartmentResponse{
		ID:        dept.ID,
		Name:      dept.Name,
		ParentID:  dept.ParentID,
		CreatedAt: dept.CreatedAt,
	}, nil
}

func (s *DepartmentService) GetTree(ctx context.Context, id int, depth int, includeEmp bool) (*dto.DepartmentTreeResponse, error) {
	if depth < 1 {
		depth = 1
	} else if depth > 5 {
		depth = 5
	}

	dept, err := s.deptRepo.GetTree(ctx, id, depth, includeEmp)
	if err != nil {
		return nil, err
	}

	resp := toTreeResponse(dept, includeEmp)
	return &resp, nil
}

func toTreeResponse(d *model.Department, includeEmp bool) dto.DepartmentTreeResponse {
	node := dto.DepartmentTreeResponse{
		Department: dto.DepartmentResponse{
			ID:        d.ID,
			Name:      d.Name,
			ParentID:  d.ParentID,
			CreatedAt: d.CreatedAt,
		},
		Children: []dto.DepartmentTreeResponse{},
	}
	if includeEmp {
		node.Employees = make([]dto.EmployeeResponse, 0, len(d.Employees))
		for _, e := range d.Employees {
			node.Employees = append(node.Employees, dto.EmployeeResponse{
				ID:           e.ID,
				DepartmentID: e.DepartmentID,
				FullName:     e.FullName,
				Position:     e.Position,
				HiredAt:      e.HiredAt,
				CreatedAt:    e.CreatedAt,
			})
		}
	}
	for i := range d.Children {
		node.Children = append(node.Children, toTreeResponse(&d.Children[i], includeEmp))
	}
	return node
}

func (s *DepartmentService) Update(ctx context.Context, id int, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error) {
	var updatedDept *model.Department

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		dept, err := s.deptRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if req.ParentIDSet {
			if req.ParentID != nil {
				if *req.ParentID == id {
					return fmt.Errorf("%w: cannot set department as its own parent", ErrInvalidInput)
				}
				exists, err := s.deptRepo.ExistsByID(txCtx, *req.ParentID)
				if err != nil {
					return err
				}
				if !exists {
					return ErrNotFound
				}

				isAncestor, err := s.deptRepo.IsAncestor(txCtx, id, *req.ParentID)
				if err != nil {
					return err
				}
				if isAncestor {
					return ErrConflict
				}
			}
			dept.ParentID = req.ParentID
		}

		if req.NameSet && req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if len(name) == 0 || len(name) > 200 {
				return ErrInvalidInput
			}
			dept.Name = name
		}

		if err := s.deptRepo.Update(txCtx, dept); err != nil {
			return err
		}
		updatedDept = dept
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.DepartmentResponse{
		ID:        updatedDept.ID,
		Name:      updatedDept.Name,
		ParentID:  updatedDept.ParentID,
		CreatedAt: updatedDept.CreatedAt,
	}, nil
}

func (s *DepartmentService) Delete(ctx context.Context, id int, mode string, reassignID *int) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		exists, err := s.deptRepo.ExistsByID(txCtx, id)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}

		if mode == "reassign" {
			if reassignID == nil || *reassignID == id {
				return ErrInvalidInput
			}

			targetExists, err := s.deptRepo.ExistsByID(txCtx, *reassignID)
			if err != nil {
				return err
			}
			if !targetExists {
				return ErrNotFound
			}

			isAncestor, err := s.deptRepo.IsAncestor(txCtx, id, *reassignID)
			if err != nil {
				return err
			}
			if isAncestor {
				return ErrConflict
			}

			if err := s.empRepo.ReassignByDepartment(txCtx, id, *reassignID); err != nil {
				return err
			}
			if err := s.deptRepo.ReassignChildren(txCtx, id, *reassignID); err != nil {
				return err
			}
		} else if mode != "cascade" {
			return ErrInvalidInput
		}

		return s.deptRepo.Delete(txCtx, id)
	})
}

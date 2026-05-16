package repository

import (
	"context"

	"github.com/tvizzi/org-api-test-task/internal/model"

	"gorm.io/gorm"
)

type empRepo struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &empRepo{db: db}
}

func (r *empRepo) Create(ctx context.Context, e *model.Employee) error {
	db := extractDB(ctx, r.db)
	return db.Create(e).Error
}

func (r *empRepo) ReassignByDepartment(ctx context.Context, fromID, toID int) error {
	db := extractDB(ctx, r.db)
	return db.Model(&model.Employee{}).
		Where("department_id = ?", fromID).
		Update("department_id", toID).Error
}

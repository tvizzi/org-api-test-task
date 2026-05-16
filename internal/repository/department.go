package repository

import (
	"context"
	"errors"

	"github.com/tvizzi/org-api-test-task/internal/model"

	"gorm.io/gorm"
)

type deptRepo struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &deptRepo{db: db}
}

func (r *deptRepo) Create(ctx context.Context, d *model.Department) error {
	db := extractDB(ctx, r.db)
	if err := db.Create(d).Error; err != nil {
		if IsUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *deptRepo) GetByID(ctx context.Context, id int) (*model.Department, error) {
	db := extractDB(ctx, r.db)
	var d model.Department
	if err := db.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *deptRepo) GetTree(ctx context.Context, id int, depth int, includeEmp bool) (*model.Department, error) {
	db := extractDB(ctx, r.db)
	query := db.Where("id = ?", id)

	if includeEmp {
		query = query.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		})
	}

	preloadStr := "Children"
	for i := 1; i < depth; i++ {
		query = query.Preload(preloadStr)
		preloadStr += ".Children"
	}

	var d model.Department
	if err := query.First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *deptRepo) Update(ctx context.Context, d *model.Department) error {
	db := extractDB(ctx, r.db)
	if err := db.Select("*").Updates(d).Error; err != nil {
		if IsUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *deptRepo) Delete(ctx context.Context, id int) error {
	db := extractDB(ctx, r.db)
	return db.Delete(&model.Department{}, id).Error
}

func (r *deptRepo) ExistsByID(ctx context.Context, id int) (bool, error) {
	db := extractDB(ctx, r.db)
	var count int64
	err := db.Model(&model.Department{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *deptRepo) ReassignChildren(ctx context.Context, oldParentID, newParentID int) error {
	db := extractDB(ctx, r.db)
	return db.Model(&model.Department{}).
		Where("parent_id = ?", oldParentID).
		Update("parent_id", newParentID).Error
}

func (r *deptRepo) IsAncestor(ctx context.Context, ancestorID, descendantID int) (bool, error) {
	if ancestorID == descendantID {
		return true, nil
	}
	db := extractDB(ctx, r.db)
	current := descendantID

	for i := 0; i < 1000; i++ {
		var d model.Department
		err := db.Select("id, parent_id").First(&d, current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if d.ParentID == nil {
			return false, nil
		}
		if *d.ParentID == ancestorID {
			return true, nil
		}
		current = *d.ParentID
	}
	return false, errors.New("tree depth exceeded safety bound")
}

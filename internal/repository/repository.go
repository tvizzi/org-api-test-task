package repository

import (
	"context"
	"errors"

	"github.com/tvizzi/org-api-test-task/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type DepartmentRepository interface {
	Create(ctx context.Context, d *model.Department) error
	GetByID(ctx context.Context, id int) (*model.Department, error)
	GetTree(ctx context.Context, id int, depth int, includeEmp bool) (*model.Department, error)
	Update(ctx context.Context, d *model.Department) error
	Delete(ctx context.Context, id int) error
	ExistsByID(ctx context.Context, id int) (bool, error)
	IsAncestor(ctx context.Context, ancestorID, descendantID int) (bool, error)
	ReassignChildren(ctx context.Context, oldParentID, newParentID int) error
}

type EmployeeRepository interface {
	Create(ctx context.Context, e *model.Employee) error
	ReassignByDepartment(ctx context.Context, fromID, toID int) error
}

type txKey struct{}

func extractDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

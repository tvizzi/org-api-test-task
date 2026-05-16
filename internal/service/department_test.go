package service_test

import (
	"context"
	"testing"

	"github.com/tvizzi/org-api-test-task/internal/dto"
	"github.com/tvizzi/org-api-test-task/internal/model"
	"github.com/tvizzi/org-api-test-task/internal/repository"
	"github.com/tvizzi/org-api-test-task/internal/service"

	"github.com/stretchr/testify/assert"
)

type mockDeptRepo struct {
	repository.DepartmentRepository
	exists     bool
	isAncestor bool
	lastDepth  int
}

func (m *mockDeptRepo) ExistsByID(ctx context.Context, id int) (bool, error) {
	return m.exists, nil
}
func (m *mockDeptRepo) GetByID(ctx context.Context, id int) (*model.Department, error) {
	return &model.Department{ID: id}, nil
}
func (m *mockDeptRepo) IsAncestor(ctx context.Context, anc, desc int) (bool, error) {
	return m.isAncestor, nil
}
func (m *mockDeptRepo) Update(ctx context.Context, d *model.Department) error {
	return nil
}
func (m *mockDeptRepo) GetTree(ctx context.Context, id int, depth int, includeEmp bool) (*model.Department, error) {
	m.lastDepth = depth
	return &model.Department{ID: id}, nil
}
func (m *mockDeptRepo) ReassignChildren(ctx context.Context, oldID, newID int) error {
	return nil
}
func (m *mockDeptRepo) Delete(ctx context.Context, id int) error {
	return nil
}

type mockEmpRepo struct {
	repository.EmployeeRepository
}

func (m *mockEmpRepo) ReassignByDepartment(ctx context.Context, fromID, toID int) error {
	return nil
}

type mockTx struct{}

func (m *mockTx) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestDepartmentService_Update_DetectsCycle(t *testing.T) {
	mockRepo := &mockDeptRepo{exists: true, isAncestor: true}
	svc := service.NewDepartmentService(mockRepo, nil, &mockTx{})

	parentID := 2
	req := dto.UpdateDepartmentRequest{
		ParentIDSet: true,
		ParentID:    &parentID,
	}

	_, err := svc.Update(context.Background(), 1, &req)

	assert.ErrorIs(t, err, service.ErrConflict)
}

func TestDepartmentService_Delete_ReassignMovesChildren(t *testing.T) {
	mockDept := &mockDeptRepo{exists: true, isAncestor: false}
	mockEmp := &mockEmpRepo{}
	svc := service.NewDepartmentService(mockDept, mockEmp, &mockTx{})

	reassignID := 2
	err := svc.Delete(context.Background(), 1, "reassign", &reassignID)

	assert.NoError(t, err)
}

func TestDepartmentService_GetTree_DepthClamping(t *testing.T) {
	mockRepo := &mockDeptRepo{}
	svc := service.NewDepartmentService(mockRepo, nil, nil)

	_, _ = svc.GetTree(context.Background(), 1, -5, false)
	assert.Equal(t, 1, mockRepo.lastDepth)

	_, _ = svc.GetTree(context.Background(), 1, 999, false)
	assert.Equal(t, 5, mockRepo.lastDepth)
}

package dto

import "time"

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at,omitempty"`
}

type EmployeeResponse struct {
	ID           int        `json:"id"`
	DepartmentID int        `json:"department_id"`
	FullName     string     `json:"full_name"`
	Position     string     `json:"position"`
	HiredAt      *time.Time `json:"hired_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

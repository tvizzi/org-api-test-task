package dto

import (
	"encoding/json"
	"time"
)

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id,omitempty"`
}

type UpdateDepartmentRequest struct {
	Name        *string
	NameSet     bool
	ParentID    *int
	ParentIDSet bool
}

func (r *UpdateDepartmentRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; ok {
		r.NameSet = true
		if string(v) != "null" {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}
			r.Name = &s
		}
	}
	if v, ok := raw["parent_id"]; ok {
		r.ParentIDSet = true
		if string(v) != "null" {
			var i int
			if err := json.Unmarshal(v, &i); err != nil {
				return err
			}
			r.ParentID = &i
		}
	}
	return nil
}

type DepartmentResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type DepartmentTreeResponse struct {
	Department DepartmentResponse       `json:"department"`
	Employees  []EmployeeResponse       `json:"employees,omitempty"`
	Children   []DepartmentTreeResponse `json:"children"`
}

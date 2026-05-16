package model

import "time"

type Department struct {
	ID        int          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string       `gorm:"type:varchar(200);not null" json:"name"`
	ParentID  *int         `json:"parent_id"`
	CreatedAt time.Time    `json:"created_at"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

type Employee struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID int        `gorm:"not null" json:"department_id"`
	FullName     string     `gorm:"type:varchar(200);not null" json:"full_name"`
	Position     string     `gorm:"type:varchar(200);not null" json:"position"`
	HiredAt      *time.Time `gorm:"type:date" json:"hired_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

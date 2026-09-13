package model

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Username   string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email      string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password   string    `gorm:"size:255;not null" json:"-"`
	RoleID     uint      `gorm:"not null" json:"role_id"`
	Role       *Role     `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"role,omitempty"`
	EmployeeID *uint     `gorm:"index" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"employee,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

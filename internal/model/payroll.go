package model

import (
	"time"

	"gorm.io/datatypes"
)

type Overtime struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmployeeID uint      `gorm:"index;not null" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"employee,omitempty"`
	Date       time.Time `gorm:"type:date;not null" json:"date"`
	Hours      float64   `gorm:"not null" json:"hours"`
	IsHoliday  bool      `gorm:"default:false" json:"is_holiday"`
	Amount     float64   `gorm:"not null;default:0" json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Payroll struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	EmployeeID           uint           `gorm:"index;uniqueIndex:idx_payroll_employee_period;not null" json:"employee_id"`
	Employee             *Employee      `gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"employee,omitempty"`
	Period               string         `gorm:"size:7;index;uniqueIndex:idx_payroll_employee_period;not null" json:"period"` // YYYY-MM
	BasicSalary          float64        `gorm:"not null" json:"basic_salary"`
	Allowances           float64        `gorm:"not null" json:"allowances"`
	OvertimePay          float64        `gorm:"not null" json:"overtime_pay"`
	Bonus                float64        `gorm:"not null;default:0" json:"bonus"`
	THR                  float64        `gorm:"not null;default:0" json:"thr"`
	GrossIncome          float64        `gorm:"not null" json:"gross_income"`
	EmployeeDeduction    float64        `gorm:"not null" json:"employee_deduction"`
	PPh21                float64        `gorm:"column:pph21;not null" json:"pph21"`
	NetSalary            float64        `gorm:"not null" json:"net_salary"`
	EmployerContribution float64        `gorm:"not null;default:0" json:"employer_contribution"`
	Details              datatypes.JSON `gorm:"type:jsonb" json:"details"`
	Status               string         `gorm:"size:20;default:draft" json:"status"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

const (
	PayrollDraft    = "draft"
	PayrollApproved = "approved"
	PayrollPaid     = "paid"
)

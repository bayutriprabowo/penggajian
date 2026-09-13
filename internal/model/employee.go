package model

import "time"

type Employee struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	NIK        string      `gorm:"uniqueIndex;size:50;not null" json:"nik"`
	FullName   string      `gorm:"size:200;not null" json:"full_name"`
	Email      string      `gorm:"size:100" json:"email"`
	Phone      string      `gorm:"size:30" json:"phone"`
	Address    string      `gorm:"type:text" json:"address"`
	Position   string      `gorm:"size:100" json:"position"`
	Department string      `gorm:"size:100" json:"department"`
	JoinDate   time.Time   `gorm:"type:date" json:"join_date"`
	Status     string      `gorm:"size:20;default:active" json:"status"`
	PTKPStatus string      `gorm:"size:5;default:TK0" json:"ptkp_status"`
	BaseSalary float64     `gorm:"not null;default:0" json:"base_salary"`
	BPJSHealth bool        `gorm:"default:true" json:"bpjs_health"`
	BPJSTK     bool        `gorm:"default:true" json:"bpjs_tk"`
	JKKRisk    float64     `gorm:"default:0.54" json:"jkk_risk"`
	Allowances []Allowance `gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"allowances,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type Allowance struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmployeeID uint      `gorm:"index;not null" json:"employee_id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Type       string    `gorm:"size:20;not null" json:"type"` // fixed | non_fixed
	Amount     float64   `gorm:"not null" json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

const (
	AllowanceFixed    = "fixed"
	AllowanceNonFixed = "non_fixed"
)

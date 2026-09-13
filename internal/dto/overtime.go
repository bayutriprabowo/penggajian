package dto

import "time"

type OvertimeRequest struct {
	EmployeeID uint    `json:"employee_id"`
	Date       string  `json:"date"` // YYYY-MM-DD
	Hours      float64 `json:"hours"`
	IsHoliday  bool    `json:"is_holiday"`
}

type OvertimeDTO struct {
	ID         uint      `json:"id"`
	EmployeeID uint      `json:"employee_id"`
	Date       time.Time `json:"date"`
	Hours      float64   `json:"hours"`
	IsHoliday  bool      `json:"is_holiday"`
	Amount     float64   `json:"amount"`
}

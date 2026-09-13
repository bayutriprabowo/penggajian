package dto

import "time"

type PayrollRunRequest struct {
	Period     string  `json:"period"`      // YYYY-MM
	EmployeeID uint    `json:"employee_id"` // 0 = semua karyawan aktif
	Bonus      float64 `json:"bonus"`
	THR        float64 `json:"thr"` // 0 = hitung otomatis, >0 = override
}

type PayrollDTO struct {
	ID                   uint               `json:"id"`
	EmployeeID           uint               `json:"employee_id"`
	EmployeeName         string             `json:"employee_name"`
	NIK                  string             `json:"nik"`
	Position             string             `json:"position"`
	Period               string             `json:"period"`
	BasicSalary          float64            `json:"basic_salary"`
	Allowances           float64            `json:"allowances"`
	OvertimePay          float64            `json:"overtime_pay"`
	Bonus                float64            `json:"bonus"`
	THR                  float64            `json:"thr"`
	GrossIncome          float64            `json:"gross_income"`
	EmployeeDeduction    float64            `json:"employee_deduction"`
	PPh21                float64            `json:"pph21"`
	NetSalary            float64            `json:"net_salary"`
	EmployerContribution float64            `json:"employer_contribution"`
	Details              []PayslipDetailDTO `json:"details"`
	Status               string             `json:"status"`
	CreatedAt            time.Time          `json:"created_at"`
}

type PayrollListDTO struct {
	Payrolls   []PayrollDTO `json:"payrolls"`
	Pagination Pagination   `json:"pagination"`
}

type PayslipDetailDTO struct {
	Category string  `json:"category"` // earning | deduction | tax | employer
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
}

type THRCalculateRequest struct {
	EmployeeID uint   `json:"employee_id"`
	Period     string `json:"period"` // YYYY-MM, basis masa kerja
}

type THRResponse struct {
	EmployeeID      uint    `json:"employee_id"`
	EmployeeName    string  `json:"employee_name"`
	MonthsOfService int     `json:"months_of_service"`
	FullTHR         bool    `json:"full_thr"`
	BaseSalary      float64 `json:"base_salary"`
	FixedAllowance  float64 `json:"fixed_allowance"`
	THRAmount       float64 `json:"thr_amount"`
	PPh21OnTHR      float64 `json:"pph21_on_thr"`
	NetTHR          float64 `json:"net_thr"`
}

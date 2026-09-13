package dto

import "time"

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Success membangun envelope respons sukses.
func Success(message string, data any) Response {
	return Response{Status: "success", Message: message, Data: data}
}

// Error membangun envelope respons gagal.
func Error(message string) Response {
	return Response{Status: "error", Message: message}
}

type AllowanceDTO struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

type EmployeeRequest struct {
	NIK        string         `json:"nik"`
	FullName   string         `json:"full_name"`
	Email      string         `json:"email"`
	Phone      string         `json:"phone"`
	Address    string         `json:"address"`
	Position   string         `json:"position"`
	Department string         `json:"department"`
	JoinDate   string         `json:"join_date"` // YYYY-MM-DD
	Status     string         `json:"status"`
	PTKPStatus string         `json:"ptkp_status"`
	BaseSalary float64        `json:"base_salary"`
	BPJSHealth *bool          `json:"bpjs_health"`
	BPJSTK     *bool          `json:"bpjs_tk"`
	JKKRisk    float64        `json:"jkk_risk"`
	Allowances []AllowanceDTO `json:"allowances"`
}

type EmployeeDTO struct {
	ID         uint           `json:"id"`
	NIK        string         `json:"nik"`
	FullName   string         `json:"full_name"`
	Email      string         `json:"email"`
	Phone      string         `json:"phone"`
	Address    string         `json:"address"`
	Position   string         `json:"position"`
	Department string         `json:"department"`
	JoinDate   time.Time      `json:"join_date"`
	Status     string         `json:"status"`
	PTKPStatus string         `json:"ptkp_status"`
	BaseSalary float64        `json:"base_salary"`
	BPJSHealth bool           `json:"bpjs_health"`
	BPJSTK     bool           `json:"bpjs_tk"`
	JKKRisk    float64        `json:"jkk_risk"`
	Allowances []AllowanceDTO `json:"allowances,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type EmployeeListDTO struct {
	Employees  []EmployeeDTO `json:"employees"`
	Pagination Pagination    `json:"pagination"`
}

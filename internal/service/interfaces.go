package service

import (
	"context"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserDTO, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	ListUsers(ctx context.Context, page, limit int) (*dto.UserListDTO, error)
	GenerateToken(user *model.User) (string, error)
	ParseToken(tokenString string) (*jwt.Token, error)
}

type EmployeeService interface {
	Create(ctx context.Context, req dto.EmployeeRequest) (*dto.EmployeeDTO, error)
	GetByID(ctx context.Context, id uint) (*dto.EmployeeDTO, error)
	List(ctx context.Context, page, limit int, status string) (*dto.EmployeeListDTO, error)
	Update(ctx context.Context, id uint, req dto.EmployeeRequest) (*dto.EmployeeDTO, error)
	Delete(ctx context.Context, id uint) error
}

type OvertimeService interface {
	Create(ctx context.Context, req dto.OvertimeRequest) (*dto.OvertimeDTO, error)
	ListByEmployee(ctx context.Context, employeeID uint, period string) ([]dto.OvertimeDTO, error)
}

type PayrollService interface {
	Run(ctx context.Context, req dto.PayrollRunRequest) ([]dto.PayrollDTO, error)
	GetByID(ctx context.Context, id uint, requester *model.User) (*dto.PayrollDTO, error)
	ListByEmployee(ctx context.Context, employeeID uint, requester *model.User) ([]dto.PayrollDTO, error)
	ListByPeriod(ctx context.Context, period string, page, limit int, requester *model.User) (*dto.PayrollListDTO, error)
	Approve(ctx context.Context, id uint) error
	MarkPaid(ctx context.Context, id uint) error
	CalculateTHR(ctx context.Context, req dto.THRCalculateRequest) (*dto.THRResponse, error)
}

type TaxService interface {
	TERTables() []dto.TERTableDTO
	TERInfo(ptkpStatus string, monthlyGross float64) (*dto.TERInfoDTO, error)
	AnnualRecap(ctx context.Context, req dto.AnnualRecapRequest) (*dto.AnnualRecapDTO, error)
}

type Services struct {
	Auth     AuthService
	Employee EmployeeService
	Overtime OvertimeService
	Payroll  PayrollService
	Tax      TaxService
}

// TokenClaims adalah klaim JWT yang dipakai aplikasi.
type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

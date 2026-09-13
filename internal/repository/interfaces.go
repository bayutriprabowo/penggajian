package repository

import (
	"context"

	"github.com/bayutriprabowo/penggajian/internal/model"
)

type Repos struct {
	Users     UserRepository
	Roles     RoleRepository
	Employees EmployeeRepository
	Overtimes OvertimeRepository
	Payrolls  PayrollRepository
}

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	WithRole(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	List(ctx context.Context, page, limit int) ([]model.User, int64, error)
}

type RoleRepository interface {
	CreateRole(ctx context.Context, r *model.Role) error
	FindRoleByName(ctx context.Context, name string) (*model.Role, error)
	FindPermissionsByName(ctx context.Context, names []string) ([]model.Permission, error)
	CreatePermissions(ctx context.Context, perms []model.Permission) error
	AssignPermissions(ctx context.Context, roleID uint, perms []model.Permission) error
	RoleWithPermissions(ctx context.Context, roleID uint) (*model.Role, error)
	PermissionsByRole(ctx context.Context, roleID uint) ([]string, error)
}

type EmployeeRepository interface {
	Create(ctx context.Context, e *model.Employee) error
	FindByID(ctx context.Context, id uint) (*model.Employee, error)
	FindByNIK(ctx context.Context, nik string) (*model.Employee, error)
	List(ctx context.Context, page, limit int, status string) ([]model.Employee, int64, error)
	Update(ctx context.Context, e *model.Employee) error
	UpdateWithAllowances(ctx context.Context, e *model.Employee, allowances *[]model.Allowance) error
	Delete(ctx context.Context, id uint) error
	ListActive(ctx context.Context) ([]model.Employee, error)
}

type OvertimeRepository interface {
	Create(ctx context.Context, o *model.Overtime) error
	SumByEmployeePeriod(ctx context.Context, employeeID uint, from, to string) (float64, error)
	ListByEmployee(ctx context.Context, employeeID uint, from, to string) ([]model.Overtime, error)
}

type PayrollRepository interface {
	Create(ctx context.Context, p *model.Payroll) error
	FindByID(ctx context.Context, id uint) (*model.Payroll, error)
	FindByEmployeePeriod(ctx context.Context, employeeID uint, period string) (*model.Payroll, error)
	ListByEmployee(ctx context.Context, employeeID uint) ([]model.Payroll, error)
	ListByPeriod(ctx context.Context, period string, page, limit int) ([]model.Payroll, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	SumPPh21ByEmployeeYear(ctx context.Context, employeeID uint, year int) (float64, error)
}

package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/config"
	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var allPermissions = []string{
	"user:read", "user:write", "user:list",
	"employee:read", "employee:write",
	"overtime:write",
	"payroll:read", "payroll:run", "payroll:approve",
	"tax:read",
}

var rolePermissions = map[string][]string{
	"admin":    allPermissions,
	"hr":       {"user:read", "user:list", "employee:read", "employee:write", "overtime:write", "payroll:read", "payroll:run", "tax:read"},
	"finance":  {"user:read", "employee:read", "payroll:read", "payroll:approve", "tax:read"},
	"employee": {"user:read", "payroll:read"},
}

var roleDescriptions = map[string]string{
	"admin":    "Akses penuh sistem",
	"hr":       "Mengelola data karyawan dan menjalankan payroll",
	"finance":  "Melihat dan menyetujui payroll",
	"employee": "Melihat slip gaji milik sendiri",
}

// SeedRBAC membuat permission, role, dan user admin default (idempotent).
func SeedRBAC(ctx context.Context, roles repository.RoleRepository, users repository.UserRepository, cfg *config.Config) error {
	existing, err := roles.FindPermissionsByName(ctx, allPermissions)
	if err != nil {
		return err
	}
	existingMap := make(map[string]bool, len(existing))
	for _, p := range existing {
		existingMap[p.Name] = true
	}
	var toCreate []model.Permission
	for _, name := range allPermissions {
		if !existingMap[name] {
			toCreate = append(toCreate, model.Permission{Name: name})
		}
	}
	if err := roles.CreatePermissions(ctx, toCreate); err != nil {
		return err
	}
	allPerms, err := roles.FindPermissionsByName(ctx, allPermissions)
	if err != nil {
		return err
	}
	permMap := make(map[string]model.Permission, len(allPerms))
	for _, p := range allPerms {
		permMap[p.Name] = p
	}

	for roleName, permNames := range rolePermissions {
		role, err := roles.FindRoleByName(ctx, roleName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = &model.Role{Name: roleName, Description: roleDescriptions[roleName]}
			if err := roles.CreateRole(ctx, role); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		var perms []model.Permission
		for _, n := range permNames {
			perms = append(perms, permMap[n])
		}
		if err := roles.AssignPermissions(ctx, role.ID, perms); err != nil {
			return err
		}
	}

	if _, err := users.FindByUsername(ctx, cfg.DefaultAdminUser); errors.Is(err, gorm.ErrRecordNotFound) {
		role, err := roles.FindRoleByName(ctx, "admin")
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := &model.User{
			Username: cfg.DefaultAdminUser,
			Email:    cfg.DefaultAdminUser + "@penggajian.local",
			Password: string(hash),
			RoleID:   role.ID,
			Role:     role,
		}
		if err := users.Create(ctx, admin); err != nil {
			return err
		}
		log.Printf("[seeder] user admin default dibuat: %s", cfg.DefaultAdminUser)
	} else if err != nil {
		return err
	}
	log.Println("[seeder] RBAC siap")
	return nil
}

// ---------- Seeder data demo ----------

type demoEmployee struct {
	nik        string
	fullName   string
	position   string
	department string
	joinDate   string
	ptkp       string
	baseSalary float64
	allowances []dto.AllowanceDTO
	username   string
}

var demoEmployees = []demoEmployee{
	{
		nik: "EMP-001", fullName: "Budi Santoso", position: "Backend Engineer",
		department: "IT", joinDate: "2025-03-01", ptkp: "K0", baseSalary: 10_000_000,
		allowances: []dto.AllowanceDTO{
			{Name: "Tunjangan Jabatan", Type: "fixed", Amount: 500_000},
			{Name: "Uang Makan", Type: "non_fixed", Amount: 300_000},
		},
		username: "budi",
	},
	{
		nik: "EMP-002", fullName: "Siti Aminah", position: "Finance Staff",
		department: "Finance", joinDate: "2025-10-15", ptkp: "TK0", baseSalary: 8_000_000,
		username: "siti",
	},
	{
		nik: "EMP-003", fullName: "Andi Wijaya", position: "HR Manager",
		department: "HR", joinDate: "2024-06-01", ptkp: "K2", baseSalary: 15_000_000,
		allowances: []dto.AllowanceDTO{
			{Name: "Tunjangan Jabatan", Type: "fixed", Amount: 1_000_000},
		},
		username: "andi",
	},
	{
		nik: "EMP-004", fullName: "Dewi Lestari", position: "Marketing Specialist",
		department: "Marketing", joinDate: "2026-01-10", ptkp: "TK1", baseSalary: 7_000_000,
		username: "dewi",
	},
}

// SeedDemo membuat data contoh (karyawan, user, lembur, payroll) secara idempotent.
// Diaktifkan via env SEED_DEMO=true — jangan digunakan di produksi.
func SeedDemo(
	ctx context.Context,
	repos *repository.Repos,
	auth AuthService,
	employees EmployeeService,
	overtimes OvertimeService,
	payroll PayrollService,
	cfg *config.Config,
) error {
	// guard 1: bila sudah ada karyawan, seluruh seeder dilewati
	_, total, err := repos.Employees.List(ctx, 1, 1, "")
	if err != nil {
		return err
	}
	if total > 0 {
		log.Println("[seed-demo] data karyawan sudah ada, seeder demo dilewati")
		return nil
	}

	log.Println("[seed-demo] membuat data contoh...")
	for _, d := range demoEmployees {
		emp, err := employees.Create(ctx, dto.EmployeeRequest{
			NIK:        d.nik,
			FullName:   d.fullName,
			Email:      d.username + "@perusahaan.com",
			Position:   d.position,
			Department: d.department,
			JoinDate:   d.joinDate,
			Status:     "active",
			PTKPStatus: d.ptkp,
			BaseSalary: d.baseSalary,
			BPJSHealth: boolPtr(true),
			BPJSTK:     boolPtr(true),
			JKKRisk:    0.54,
			Allowances: d.allowances,
		})
		if err != nil {
			return err
		}
		if _, err := auth.Register(ctx, dto.RegisterRequest{
			Username:   d.username,
			Email:      d.username + "@perusahaan.com",
			Password:   cfg.DemoUserPass,
			RoleName:   "employee",
			EmployeeID: &emp.ID,
		}); err != nil {
			return err
		}
	}

	// periode demo = bulan lalu
	periodDate := time.Now().AddDate(0, -1, 0)
	period := periodDate.Format("2006-01")

	// lembur demo
	overtimeSamples := []struct {
		employeeID uint
		day        int
		hours      float64
		isHoliday  bool
	}{
		{1, 5, 2, false},
		{1, 12, 3, true},
		{2, 8, 2.5, false},
		{4, 15, 4, true},
	}
	for _, o := range overtimeSamples {
		date := time.Date(periodDate.Year(), periodDate.Month(), o.day, 0, 0, 0, 0, time.Local)
		if _, err := overtimes.Create(ctx, dto.OvertimeRequest{
			EmployeeID: o.employeeID,
			Date:       date.Format("2006-01-02"),
			Hours:      o.hours,
			IsHoliday:  o.isHoliday,
		}); err != nil {
			return err
		}
	}

	// guard 2: payroll demo hanya bila periode belum pernah di-generate
	_, payrollTotal, err := repos.Payrolls.ListByPeriod(ctx, period, 1, 1)
	if err != nil {
		return err
	}
	if payrollTotal == 0 {
		if _, err := payroll.Run(ctx, dto.PayrollRunRequest{
			Period: period, EmployeeID: 0, Bonus: 0, THR: 0,
		}); err != nil {
			return err
		}
		log.Printf("[seed-demo] payroll periode %s di-generate (status draft)", period)
	} else {
		log.Printf("[seed-demo] payroll periode %s sudah ada, dilewati", period)
	}

	// user finance untuk demo alur approve
	if _, err := auth.Register(ctx, dto.RegisterRequest{
		Username: "finance1", Email: "finance1@perusahaan.com",
		Password: cfg.DemoUserPass, RoleName: "finance",
	}); err != nil && !errors.Is(err, ErrConflict) {
		return err
	}

	log.Println("[seed-demo] selesai.")
	log.Printf("[seed-demo] user karyawan: %s | finance1 (password: %s)", demoUserList(), cfg.DemoUserPass)
	log.Println("[seed-demo] login admin: admin / admin123")
	return nil
}

// demoUserList merangkum username karyawan demo untuk log.
func demoUserList() string {
	out := ""
	for i, d := range demoEmployees {
		if i > 0 {
			out += ", "
		}
		out += d.username
	}
	return out
}

// boolPtr mengembalikan pointer ke bool (helper request DTO).
func boolPtr(v bool) *bool {
	return &v
}

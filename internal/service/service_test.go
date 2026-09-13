package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/config"
	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"gorm.io/gorm"
)

// ---------- fake repositories ----------

type fakeUserRepo struct {
	users map[string]*model.User
}

// Create (fake) menyimpan user atau menolak username duplikat.
func (f *fakeUserRepo) Create(_ context.Context, u *model.User) error {
	if _, ok := f.users[u.Username]; ok {
		return errors.New("duplicate")
	}
	f.users[u.Username] = u
	return nil
}

// FindByID (fake) mencari user berdasarkan ID.
func (f *fakeUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// FindByUsername (fake) mencari user berdasarkan username.
func (f *fakeUserRepo) FindByUsername(_ context.Context, username string) (*model.User, error) {
	if u, ok := f.users[username]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// FindByEmail (fake) mencari user berdasarkan email.
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// WithRole (fake) tidak melakukan apa-apa.
func (f *fakeUserRepo) WithRole(_ context.Context, u *model.User) error { return nil }

// Update (fake) tidak melakukan apa-apa.
func (f *fakeUserRepo) Update(_ context.Context, u *model.User) error { return nil }

// List (fake) mengembalikan daftar kosong.
func (f *fakeUserRepo) List(_ context.Context, _, _ int) ([]model.User, int64, error) {
	return nil, 0, nil
}

type fakeRoleRepo struct {
	roles map[string]*model.Role
	perms []string
}

// CreateRole (fake) menyimpan role ke map.
func (f *fakeRoleRepo) CreateRole(_ context.Context, r *model.Role) error {
	f.roles[r.Name] = r
	return nil
}

// FindRoleByName (fake) mencari role dari map.
func (f *fakeRoleRepo) FindRoleByName(_ context.Context, name string) (*model.Role, error) {
	if r, ok := f.roles[name]; ok {
		return r, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// FindPermissionsByName (fake) membuat permission dari nama.
func (f *fakeRoleRepo) FindPermissionsByName(_ context.Context, names []string) ([]model.Permission, error) {
	var out []model.Permission
	for _, n := range names {
		out = append(out, model.Permission{Name: n})
	}
	return out, nil
}

// CreatePermissions (fake) mencatat nama permission.
func (f *fakeRoleRepo) CreatePermissions(_ context.Context, perms []model.Permission) error {
	for _, p := range perms {
		f.perms = append(f.perms, p.Name)
	}
	return nil
}

// AssignPermissions (fake) tidak melakukan apa-apa.
func (f *fakeRoleRepo) AssignPermissions(_ context.Context, _ uint, _ []model.Permission) error {
	return nil
}

// RoleWithPermissions (fake) tidak digunakan dalam test.
func (f *fakeRoleRepo) RoleWithPermissions(_ context.Context, roleID uint) (*model.Role, error) {
	return nil, nil
}

// PermissionsByRole (fake) mengembalikan daftar permission tercatat.
func (f *fakeRoleRepo) PermissionsByRole(_ context.Context, _ uint) ([]string, error) {
	return f.perms, nil
}

type fakeEmployeeRepo struct {
	employees map[uint]*model.Employee
	nextID    uint
}

// newFakeEmployeeRepo membuat repo karyawan dalam memori.
func newFakeEmployeeRepo() *fakeEmployeeRepo {
	return &fakeEmployeeRepo{employees: map[uint]*model.Employee{}, nextID: 1}
}

// Create (fake) menyimpan karyawan dan memberi ID berurutan.
func (f *fakeEmployeeRepo) Create(_ context.Context, e *model.Employee) error {
	e.ID = f.nextID
	f.nextID++
	f.employees[e.ID] = e
	return nil
}

// FindByID (fake) mencari karyawan berdasarkan ID.
func (f *fakeEmployeeRepo) FindByID(_ context.Context, id uint) (*model.Employee, error) {
	if e, ok := f.employees[id]; ok {
		return e, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// FindByNIK (fake) mencari karyawan berdasarkan NIK.
func (f *fakeEmployeeRepo) FindByNIK(_ context.Context, nik string) (*model.Employee, error) {
	for _, e := range f.employees {
		if e.NIK == nik {
			return e, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// List (fake) mengembalikan seluruh karyawan.
func (f *fakeEmployeeRepo) List(_ context.Context, _, _ int, _ string) ([]model.Employee, int64, error) {
	var out []model.Employee
	for _, e := range f.employees {
		out = append(out, *e)
	}
	return out, int64(len(out)), nil
}

// Update (fake) menimpa data karyawan di map.
func (f *fakeEmployeeRepo) Update(_ context.Context, e *model.Employee) error {
	f.employees[e.ID] = e
	return nil
}

// UpdateWithAllowances (fake) menyimpan karyawan dan tunjangan baru.
func (f *fakeEmployeeRepo) UpdateWithAllowances(_ context.Context, e *model.Employee, allowances *[]model.Allowance) error {
	f.employees[e.ID] = e
	if allowances != nil {
		e.Allowances = *allowances
	}
	return nil
}

// Delete (fake) menghapus karyawan dari map.
func (f *fakeEmployeeRepo) Delete(_ context.Context, id uint) error {
	delete(f.employees, id)
	return nil
}

// ListActive (fake) mengembalikan karyawan berstatus aktif.
func (f *fakeEmployeeRepo) ListActive(_ context.Context) ([]model.Employee, error) {
	var out []model.Employee
	for _, e := range f.employees {
		if e.Status == "active" {
			out = append(out, *e)
		}
	}
	return out, nil
}

type fakeOvertimeRepo struct {
	overtimes []model.Overtime
}

// Create (fake) menambah catatan lembur ke slice.
func (f *fakeOvertimeRepo) Create(_ context.Context, o *model.Overtime) error {
	f.overtimes = append(f.overtimes, *o)
	return nil
}

// SumByEmployeePeriod (fake) menjumlahkan upah lembur karyawan.
func (f *fakeOvertimeRepo) SumByEmployeePeriod(_ context.Context, employeeID uint, from, to string) (float64, error) {
	var sum float64
	for _, o := range f.overtimes {
		if o.EmployeeID == employeeID {
			sum += o.Amount
		}
	}
	return sum, nil
}

// ListByEmployee (fake) memfilter lembur milik karyawan.
func (f *fakeOvertimeRepo) ListByEmployee(_ context.Context, employeeID uint, from, to string) ([]model.Overtime, error) {
	var out []model.Overtime
	for _, o := range f.overtimes {
		if o.EmployeeID == employeeID {
			out = append(out, o)
		}
	}
	return out, nil
}

type fakePayrollRepo struct {
	payrolls map[uint]*model.Payroll
	nextID   uint
}

// newFakePayrollRepo membuat repo payroll dalam memori.
func newFakePayrollRepo() *fakePayrollRepo {
	return &fakePayrollRepo{payrolls: map[uint]*model.Payroll{}, nextID: 1}
}

// Create (fake) menyimpan slip gaji dan memberi ID berurutan.
func (f *fakePayrollRepo) Create(_ context.Context, p *model.Payroll) error {
	p.ID = f.nextID
	f.nextID++
	f.payrolls[p.ID] = p
	return nil
}

// FindByID (fake) mencari slip gaji berdasarkan ID.
func (f *fakePayrollRepo) FindByID(_ context.Context, id uint) (*model.Payroll, error) {
	if p, ok := f.payrolls[id]; ok {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// FindByEmployeePeriod (fake) mencari slip per karyawan dan periode.
func (f *fakePayrollRepo) FindByEmployeePeriod(_ context.Context, employeeID uint, period string) (*model.Payroll, error) {
	for _, p := range f.payrolls {
		if p.EmployeeID == employeeID && p.Period == period {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// ListByEmployee (fake) mengembalikan slip milik karyawan.
func (f *fakePayrollRepo) ListByEmployee(_ context.Context, employeeID uint) ([]model.Payroll, error) {
	var out []model.Payroll
	for _, p := range f.payrolls {
		if p.EmployeeID == employeeID {
			out = append(out, *p)
		}
	}
	return out, nil
}

// ListByPeriod (fake) mengembalikan slip satu periode.
func (f *fakePayrollRepo) ListByPeriod(_ context.Context, period string, page, limit int) ([]model.Payroll, int64, error) {
	var out []model.Payroll
	for _, p := range f.payrolls {
		if p.Period == period {
			out = append(out, *p)
		}
	}
	return out, int64(len(out)), nil
}

// UpdateStatus (fake) mengubah status slip gaji.
func (f *fakePayrollRepo) UpdateStatus(_ context.Context, id uint, status string) error {
	if p, ok := f.payrolls[id]; ok {
		p.Status = status
		return nil
	}
	return errors.New("not found")
}

// SumPPh21ByEmployeeYear (fake) menjumlahkan PPh 21 karyawan.
func (f *fakePayrollRepo) SumPPh21ByEmployeeYear(_ context.Context, employeeID uint, year int) (float64, error) {
	var sum float64
	for _, p := range f.payrolls {
		if p.EmployeeID == employeeID {
			sum += p.PPh21
		}
	}
	return sum, nil
}

var _ repository.UserRepository = (*fakeUserRepo)(nil)
var _ repository.RoleRepository = (*fakeRoleRepo)(nil)
var _ repository.EmployeeRepository = (*fakeEmployeeRepo)(nil)
var _ repository.OvertimeRepository = (*fakeOvertimeRepo)(nil)
var _ repository.PayrollRepository = (*fakePayrollRepo)(nil)

// ---------- helpers ----------

// testRoles menyediakan role dasar untuk test.
func testRoles() *fakeRoleRepo {
	employee := &model.Role{ID: 4, Name: "employee"}
	hr := &model.Role{ID: 2, Name: "hr"}
	finance := &model.Role{ID: 3, Name: "finance"}
	admin := &model.Role{ID: 1, Name: "admin"}
	return &fakeRoleRepo{roles: map[string]*model.Role{"employee": employee, "hr": hr, "finance": finance, "admin": admin}}
}

// testCfg menyediakan konfigurasi lengkap untuk test.
func testCfg() *config.Config {
	return &config.Config{
		JWTSecret:        "secret-tes",
		JWTExpireHours:   1,
		BPJSHealthCap:    12_000_000,
		BPJSHealthWorker: 0.01,
		BPJSHealthEmploy: 0.04,
		JPCap:            10_042_300,
		JPWorker:         0.01,
		JPEmployer:       0.02,
		JHTWorker:        0.02,
		JHTEmployer:      0.037,
		JKMEmployer:      0.003,
	}
}

// testEmployee menyediakan karyawan contoh untuk test.
func testEmployee() *model.Employee {
	join, _ := time.Parse("2006-01-02", "2025-10-15")
	return &model.Employee{
		NIK: "EMP-001", FullName: "Budi Santoso", JoinDate: join,
		Status: "active", PTKPStatus: "K0", BaseSalary: 10_000_000,
		BPJSHealth: true, BPJSTK: true, JKKRisk: 0.54,
		Allowances: []model.Allowance{
			{Name: "Tunjangan Jabatan", Type: "fixed", Amount: 500_000},
		},
	}
}

// ---------- tests ----------

// TestAuthRegisterAndLogin menguji validasi, konflik, dan login.
func TestAuthRegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	auth := NewAuthService(&fakeUserRepo{users: map[string]*model.User{}}, testRoles(), "secret-tes", 1)

	// password terlalu pendek
	_, err := auth.Register(ctx, dto.RegisterRequest{Username: "a", Email: "a@b.c", Password: "123"})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
	// registrasi sukses
	u, err := auth.Register(ctx, dto.RegisterRequest{Username: "budi", Email: "budi@x.com", Password: "rahasia123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.RoleName != "employee" {
		t.Fatalf("want employee, got %s", u.RoleName)
	}
	// username duplikat
	_, err = auth.Register(ctx, dto.RegisterRequest{Username: "budi", Email: "lain@x.com", Password: "rahasia123"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	// login salah password
	_, err = auth.Login(ctx, dto.LoginRequest{Username: "budi", Password: "salah"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
	// login sukses
	resp, err := auth.Login(ctx, dto.LoginRequest{Username: "budi", Password: "rahasia123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("token kosong")
	}
}

// TestEmployeeUpdateKeepsBaseSalary memastikan gaji pokok tidak berubah saat update parsial.
func TestEmployeeUpdateKeepsBaseSalary(t *testing.T) {
	ctx := context.Background()
	empRepo := newFakeEmployeeRepo()
	svc := NewEmployeeService(empRepo)

	created, err := svc.Create(ctx, dto.EmployeeRequest{
		NIK: "EMP-001", FullName: "Budi", JoinDate: "2025-01-01",
		BaseSalary: 10_000_000, PTKPStatus: "K0",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// update hanya posisi (tanpa base_salary)
	updated, err := svc.Update(ctx, created.ID, dto.EmployeeRequest{Position: "Senior"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.BaseSalary != 10_000_000 {
		t.Fatalf("gaji pokok berubah: %v", updated.BaseSalary)
	}
	if updated.Position != "Senior" {
		t.Fatalf("posisi tidak berubah: %v", updated.Position)
	}
}

// TestPayrollRunDuplicateConflict memastikan payroll duplikat ditolak.
func TestPayrollRunDuplicateConflict(t *testing.T) {
	ctx := context.Background()
	empRepo := newFakeEmployeeRepo()
	empRepo.Create(ctx, testEmployee())
	svc := NewPayrollService(newFakePayrollRepo(), empRepo, &fakeOvertimeRepo{}, testCfg())

	req := dto.PayrollRunRequest{Period: "2026-03", EmployeeID: 1}
	if _, err := svc.Run(ctx, req); err != nil {
		t.Fatalf("run pertama: %v", err)
	}
	_, err := svc.Run(ctx, req)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

// TestPayrollTHRConsistency memastikan masa kerja dan nominal THR.
func TestPayrollTHRConsistency(t *testing.T) {
	ctx := context.Background()
	empRepo := newFakeEmployeeRepo()
	empRepo.Create(ctx, testEmployee())
	svc := NewPayrollService(newFakePayrollRepo(), empRepo, &fakeOvertimeRepo{}, testCfg())

	resp, err := svc.CalculateTHR(ctx, dto.THRCalculateRequest{EmployeeID: 1, Period: "2026-03"})
	if err != nil {
		t.Fatalf("thr: %v", err)
	}
	// join 2025-10-15 -> basis akhir Maret 2026 = 5 bulan
	if resp.MonthsOfService != 5 {
		t.Fatalf("want 5 bulan, got %d", resp.MonthsOfService)
	}
	want := float64(10_500_000*5) / 12
	if resp.THRAmount != want {
		t.Fatalf("want %v, got %v", want, resp.THRAmount)
	}
}

// TestPayrollRunEmptyEmployees memastikan error bila tidak ada karyawan aktif.
func TestPayrollRunEmptyEmployees(t *testing.T) {
	ctx := context.Background()
	svc := NewPayrollService(newFakePayrollRepo(), newFakeEmployeeRepo(), &fakeOvertimeRepo{}, testCfg())
	_, err := svc.Run(ctx, dto.PayrollRunRequest{Period: "2026-03"})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
}

// TestSeedDemoIdempotent memastikan seeder demo tidak membuat data ganda.
func TestSeedDemoIdempotent(t *testing.T) {
	ctx := context.Background()
	roles := testRoles()
	users := &fakeUserRepo{users: map[string]*model.User{}}
	empRepo := newFakeEmployeeRepo()
	otRepo := &fakeOvertimeRepo{}
	payRepo := newFakePayrollRepo()
	repos := &repository.Repos{
		Users: users, Roles: roles, Employees: empRepo,
		Overtimes: otRepo, Payrolls: payRepo,
	}
	cfg := testCfg()
	cfg.DemoUserPass = "rahasia123"

	auth := NewAuthService(users, roles, cfg.JWTSecret, cfg.JWTExpireHours)
	empSvc := NewEmployeeService(empRepo)
	otSvc := NewOvertimeService(otRepo, empRepo)
	paySvc := NewPayrollService(payRepo, empRepo, otRepo, cfg)

	if err := SeedDemo(ctx, repos, auth, empSvc, otSvc, paySvc, cfg); err != nil {
		t.Fatalf("seed demo: %v", err)
	}
	if len(empRepo.employees) != 4 {
		t.Fatalf("want 4 karyawan, got %d", len(empRepo.employees))
	}
	if len(users.users) != 5 { // 4 karyawan + 1 finance
		t.Fatalf("want 5 user, got %d", len(users.users))
	}
	if len(otRepo.overtimes) != 4 {
		t.Fatalf("want 4 lembur, got %d", len(otRepo.overtimes))
	}
	if len(payRepo.payrolls) != 4 {
		t.Fatalf("want 4 payroll, got %d", len(payRepo.payrolls))
	}
	for _, p := range payRepo.payrolls {
		if p.Status != model.PayrollDraft {
			t.Fatalf("status want draft, got %s", p.Status)
		}
	}

	// jalankan ulang: harus idempotent (tidak ada data ganda)
	if err := SeedDemo(ctx, repos, auth, empSvc, otSvc, paySvc, cfg); err != nil {
		t.Fatalf("seed demo kedua: %v", err)
	}
	if len(empRepo.employees) != 4 || len(otRepo.overtimes) != 4 || len(payRepo.payrolls) != 4 {
		t.Fatal("seeder tidak idempotent: data bertambah")
	}
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/dberr"
	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"gorm.io/gorm"
)

type employeeService struct {
	employees repository.EmployeeRepository
}

func NewEmployeeService(employees repository.EmployeeRepository) EmployeeService {
	return &employeeService{employees: employees}
}

var validPTKP = map[string]bool{"TK0": true, "TK1": true, "TK2": true, "TK3": true, "K0": true, "K1": true, "K2": true, "K3": true}

func (s *employeeService) Create(ctx context.Context, req dto.EmployeeRequest) (*dto.EmployeeDTO, error) {
	if req.NIK == "" || req.FullName == "" || req.BaseSalary <= 0 {
		return nil, ErrBadRequest
	}
	if err := validateEmployeeRequest(req); err != nil {
		return nil, err
	}
	joinDate, err := time.Parse("2006-01-02", req.JoinDate)
	if err != nil {
		return nil, ErrBadRequest
	}
	ptkp := strings.ToUpper(strings.TrimSpace(req.PTKPStatus))
	if ptkp == "" {
		ptkp = "TK0"
	}
	if !validPTKP[ptkp] {
		return nil, ErrBadRequest
	}
	if _, err := s.employees.FindByNIK(ctx, req.NIK); err == nil {
		return nil, ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	emp := &model.Employee{
		NIK:        req.NIK,
		FullName:   req.FullName,
		Email:      req.Email,
		Phone:      req.Phone,
		Address:    req.Address,
		Position:   req.Position,
		Department: req.Department,
		JoinDate:   joinDate,
		Status:     defaultString(req.Status, "active"),
		PTKPStatus: ptkp,
		BaseSalary: req.BaseSalary,
		BPJSHealth: req.BPJSHealth == nil || *req.BPJSHealth,
		BPJSTK:     req.BPJSTK == nil || *req.BPJSTK,
		JKKRisk:    req.JKKRisk,
		Allowances: toAllowanceModels(req.Allowances),
	}
	if emp.JKKRisk == 0 {
		emp.JKKRisk = 0.54
	}
	if err := s.employees.Create(ctx, emp); err != nil {
		if dberr.IsUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return toEmployeeDTO(emp), nil
}

func (s *employeeService) GetByID(ctx context.Context, id uint) (*dto.EmployeeDTO, error) {
	emp, err := s.employees.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toEmployeeDTO(emp), nil
}

func (s *employeeService) List(ctx context.Context, page, limit int, status string) (*dto.EmployeeListDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	emps, total, err := s.employees.List(ctx, page, limit, status)
	if err != nil {
		return nil, err
	}
	res := &dto.EmployeeListDTO{Pagination: dto.Pagination{Page: page, Limit: limit, Total: int(total)}}
	for i := range emps {
		res.Employees = append(res.Employees, *toEmployeeDTO(&emps[i]))
	}
	return res, nil
}

func (s *employeeService) Update(ctx context.Context, id uint, req dto.EmployeeRequest) (*dto.EmployeeDTO, error) {
	if err := validateEmployeeRequest(req); err != nil {
		return nil, err
	}
	emp, err := s.employees.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.NIK != "" && req.NIK != emp.NIK {
		if _, err := s.employees.FindByNIK(ctx, req.NIK); err == nil {
			return nil, ErrConflict
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		emp.NIK = req.NIK
	}
	if req.FullName != "" {
		emp.FullName = req.FullName
	}
	if req.Email != "" {
		emp.Email = req.Email
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
	}
	if req.Address != "" {
		emp.Address = req.Address
	}
	if req.Position != "" {
		emp.Position = req.Position
	}
	if req.Department != "" {
		emp.Department = req.Department
	}
	if req.JoinDate != "" {
		jd, err := time.Parse("2006-01-02", req.JoinDate)
		if err != nil {
			return nil, ErrBadRequest
		}
		emp.JoinDate = jd
	}
	if req.Status != "" {
		emp.Status = req.Status
	}
	if req.PTKPStatus != "" {
		ptkp := strings.ToUpper(strings.TrimSpace(req.PTKPStatus))
		if !validPTKP[ptkp] {
			return nil, ErrBadRequest
		}
		emp.PTKPStatus = ptkp
	}
	if req.BaseSalary > 0 {
		emp.BaseSalary = req.BaseSalary
	}
	if req.BPJSHealth != nil {
		emp.BPJSHealth = *req.BPJSHealth
	}
	if req.BPJSTK != nil {
		emp.BPJSTK = *req.BPJSTK
	}
	if req.JKKRisk > 0 {
		emp.JKKRisk = req.JKKRisk
	}
	var allowances *[]model.Allowance
	if req.Allowances != nil {
		list := toAllowanceModels(req.Allowances)
		allowances = &list
	}
	if err := s.employees.UpdateWithAllowances(ctx, emp, allowances); err != nil {
		if dberr.IsUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	updated, err := s.employees.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEmployeeDTO(updated), nil
}

func (s *employeeService) Delete(ctx context.Context, id uint) error {
	if err := s.employees.Delete(ctx, id); err != nil {
		if dberr.IsForeignKeyViolation(err) {
			return fmt.Errorf("%w: karyawan masih memiliki data terkait (payroll/lembur)", ErrConflict)
		}
		return ErrNotFound
	}
	return nil
}

func toAllowanceModels(items []dto.AllowanceDTO) []model.Allowance {
	out := make([]model.Allowance, 0, len(items))
	for _, it := range items {
		t := it.Type
		if t != model.AllowanceNonFixed {
			t = model.AllowanceFixed
		}
		out = append(out, model.Allowance{Name: it.Name, Type: t, Amount: it.Amount})
	}
	return out
}

func toEmployeeDTO(e *model.Employee) *dto.EmployeeDTO {
	res := &dto.EmployeeDTO{
		ID:         e.ID,
		NIK:        e.NIK,
		FullName:   e.FullName,
		Email:      e.Email,
		Phone:      e.Phone,
		Address:    e.Address,
		Position:   e.Position,
		Department: e.Department,
		JoinDate:   e.JoinDate,
		Status:     e.Status,
		PTKPStatus: e.PTKPStatus,
		BaseSalary: e.BaseSalary,
		BPJSHealth: e.BPJSHealth,
		BPJSTK:     e.BPJSTK,
		JKKRisk:    e.JKKRisk,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
	for _, a := range e.Allowances {
		res.Allowances = append(res.Allowances, dto.AllowanceDTO{ID: a.ID, Name: a.Name, Type: a.Type, Amount: a.Amount})
	}
	return res
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func validateEmployeeRequest(req dto.EmployeeRequest) error {
	if req.Email != "" && !strings.Contains(req.Email, "@") {
		return ErrBadRequest
	}
	if req.JKKRisk != 0 && (req.JKKRisk < 0.24 || req.JKKRisk > 1.74) {
		return ErrBadRequest
	}
	if req.Status != "" && req.Status != "active" && req.Status != "inactive" {
		return ErrBadRequest
	}
	for _, a := range req.Allowances {
		if a.Name == "" || a.Amount < 0 {
			return ErrBadRequest
		}
	}
	return nil
}

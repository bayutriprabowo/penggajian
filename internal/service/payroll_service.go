package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/config"
	"github.com/bayutriprabowo/penggajian/internal/dberr"
	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/payroll"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"gorm.io/gorm"
)

type payrollService struct {
	payrolls  repository.PayrollRepository
	employees repository.EmployeeRepository
	overtimes repository.OvertimeRepository
	cfg       *config.Config
}

func NewPayrollService(
	payrolls repository.PayrollRepository,
	employees repository.EmployeeRepository,
	overtimes repository.OvertimeRepository,
	cfg *config.Config,
) PayrollService {
	return &payrollService{payrolls: payrolls, employees: employees, overtimes: overtimes, cfg: cfg}
}

func (s *payrollService) rates() payroll.BPJSRates {
	return payroll.BPJSRates{
		HealthCap:      s.cfg.BPJSHealthCap,
		HealthWorker:   s.cfg.BPJSHealthWorker,
		HealthEmployer: s.cfg.BPJSHealthEmploy,
		JPCap:          s.cfg.JPCap,
		JPWorker:       s.cfg.JPWorker,
		JPEmployer:     s.cfg.JPEmployer,
		JHTWorker:      s.cfg.JHTWorker,
		JHTEmployer:    s.cfg.JHTEmployer,
		JKMEmployer:    s.cfg.JKMEmployer,
	}
}

func (s *payrollService) Run(ctx context.Context, req dto.PayrollRunRequest) ([]dto.PayrollDTO, error) {
	if _, err := time.Parse("2006-01", req.Period); err != nil {
		return nil, ErrBadRequest
	}
	if req.Bonus < 0 || req.THR < -1 {
		return nil, ErrBadRequest
	}

	var employees []model.Employee
	if req.EmployeeID > 0 {
		emp, err := s.employees.FindByID(ctx, req.EmployeeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		employees = append(employees, *emp)
	} else {
		list, err := s.employees.ListActive(ctx)
		if err != nil {
			return nil, err
		}
		employees = list
	}
	if len(employees) == 0 {
		return nil, fmt.Errorf("%w: tidak ada karyawan aktif", ErrBadRequest)
	}

	// validasi konflik dulu agar tidak terjadi pembuatan sebagian
	for i := range employees {
		existing, err := s.payrolls.FindByEmployeePeriod(ctx, employees[i].ID, req.Period)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("%w: payroll periode %s untuk karyawan %s sudah ada (id: %d)", ErrConflict, req.Period, employees[i].NIK, existing.ID)
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	var result []dto.PayrollDTO
	for i := range employees {
		emp := &employees[i]
		slip, err := s.buildPayslip(ctx, emp, req.Period, req.Bonus, req.THR)
		if err != nil {
			return nil, err
		}
		detailsJSON, _ := json.Marshal(slip.Details)
		p := &model.Payroll{
			EmployeeID:           emp.ID,
			Period:               req.Period,
			BasicSalary:          slip.BasicSalary,
			Allowances:           slip.FixedAllowance + slip.NonFixedAllowance,
			OvertimePay:          slip.OvertimePay,
			Bonus:                slip.Bonus,
			THR:                  slip.THR,
			GrossIncome:          slip.GrossIncome,
			EmployeeDeduction:    slip.EmployeeDeduction,
			PPh21:                slip.PPh21,
			NetSalary:            slip.NetSalary,
			EmployerContribution: slip.EmployerContribution,
			Details:              detailsJSON,
			Status:               model.PayrollDraft,
		}
		if err := s.payrolls.Create(ctx, p); err != nil {
			if dberr.IsUniqueViolation(err) {
				return nil, fmt.Errorf("%w: payroll periode %s untuk karyawan %s sudah ada", ErrConflict, req.Period, emp.NIK)
			}
			return nil, err
		}
		result = append(result, *toPayrollDTO(p, emp))
	}
	return result, nil
}

func (s *payrollService) buildPayslip(ctx context.Context, emp *model.Employee, period string, bonus, thrOverride float64) (*payroll.Payslip, error) {
	var fixed, nonFixed float64
	for _, a := range emp.Allowances {
		if a.Type == model.AllowanceFixed {
			fixed += a.Amount
		} else {
			nonFixed += a.Amount
		}
	}

	from := period + "-01"
	to := lastDayOfMonth(period)
	overtimeTotal, err := s.overtimes.SumByEmployeePeriod(ctx, emp.ID, from, to)
	if err != nil {
		return nil, err
	}

	thr := thrOverride
	if thr == -1 { // -1 = hitung THR otomatis berdasarkan masa kerja
		thrResult := payroll.CalculateTHR(emp.BaseSalary, fixed, emp.JoinDate, periodEndDate(period))
		thr = thrResult.Amount
	}
	if thr < 0 {
		thr = 0
	}

	input := payroll.EmployeePayrollInput{
		BaseSalary:        emp.BaseSalary,
		FixedAllowance:    fixed,
		NonFixedAllowance: nonFixed,
		OvertimePay:       overtimeTotal,
		Bonus:             bonus,
		THR:               thr,
		PTKPStatus:        emp.PTKPStatus,
		BPJSHealth:        emp.BPJSHealth,
		BPJSTK:            emp.BPJSTK,
		JKKRisk:           emp.JKKRisk,
		Rates:             s.rates(),
	}
	input.Rates.JKKRisk = emp.JKKRisk
	return payroll.ComputePayslip(input)
}

func (s *payrollService) GetByID(ctx context.Context, id uint, requester *model.User) (*dto.PayrollDTO, error) {
	p, err := s.payrolls.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if requester.Role != nil && requester.Role.Name == "employee" &&
		(requester.EmployeeID == nil || *requester.EmployeeID != p.EmployeeID) {
		return nil, ErrForbidden
	}
	return toPayrollDTO(p, nil), nil
}

func (s *payrollService) ListByEmployee(ctx context.Context, employeeID uint, requester *model.User) ([]dto.PayrollDTO, error) {
	if requester.Role != nil && requester.Role.Name == "employee" &&
		(requester.EmployeeID == nil || *requester.EmployeeID != employeeID) {
		return nil, ErrForbidden
	}
	list, err := s.payrolls.ListByEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PayrollDTO, 0, len(list))
	for i := range list {
		out = append(out, *toPayrollDTO(&list[i], nil))
	}
	return out, nil
}

func (s *payrollService) ListByPeriod(ctx context.Context, period string, page, limit int, requester *model.User) (*dto.PayrollListDTO, error) {
	if requester.Role != nil && requester.Role.Name == "employee" {
		return nil, ErrForbidden
	}
	if _, err := time.Parse("2006-01", period); err != nil {
		return nil, ErrBadRequest
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	list, total, err := s.payrolls.ListByPeriod(ctx, period, page, limit)
	if err != nil {
		return nil, err
	}
	res := &dto.PayrollListDTO{Pagination: dto.Pagination{Page: page, Limit: limit, Total: int(total)}}
	for i := range list {
		res.Payrolls = append(res.Payrolls, *toPayrollDTO(&list[i], nil))
	}
	return res, nil
}

func (s *payrollService) Approve(ctx context.Context, id uint) error {
	p, err := s.payrolls.FindByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if p.Status != model.PayrollDraft {
		return ErrBadRequest
	}
	return s.payrolls.UpdateStatus(ctx, id, model.PayrollApproved)
}

func (s *payrollService) MarkPaid(ctx context.Context, id uint) error {
	p, err := s.payrolls.FindByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if p.Status != model.PayrollApproved {
		return ErrBadRequest
	}
	return s.payrolls.UpdateStatus(ctx, id, model.PayrollPaid)
}

func (s *payrollService) CalculateTHR(ctx context.Context, req dto.THRCalculateRequest) (*dto.THRResponse, error) {
	if req.EmployeeID == 0 || req.Period == "" {
		return nil, ErrBadRequest
	}
	emp, err := s.employees.FindByID(ctx, req.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	periodDate, err := time.Parse("2006-01", req.Period)
	if err != nil {
		return nil, ErrBadRequest
	}
	// basis masa kerja: akhir periode, konsisten dengan saat payroll di-generate
	periodDate = periodEndDate(req.Period)
	var fixed float64
	for _, a := range emp.Allowances {
		if a.Type == model.AllowanceFixed {
			fixed += a.Amount
		}
	}
	res := payroll.CalculateTHR(emp.BaseSalary, fixed, emp.JoinDate, periodDate)
	regularGross := emp.BaseSalary + fixed
	pph21THR, err := payroll.PPh21OnTHR(emp.PTKPStatus, regularGross, res.Amount)
	if err != nil {
		return nil, err
	}
	return &dto.THRResponse{
		EmployeeID:      emp.ID,
		EmployeeName:    emp.FullName,
		MonthsOfService: res.MonthsOfService,
		FullTHR:         res.FullTHR,
		BaseSalary:      emp.BaseSalary,
		FixedAllowance:  fixed,
		THRAmount:       res.Amount,
		PPh21OnTHR:      pph21THR,
		NetTHR:          res.Amount - pph21THR,
	}, nil
}

func toPayrollDTO(p *model.Payroll, emp *model.Employee) *dto.PayrollDTO {
	res := &dto.PayrollDTO{
		ID:                   p.ID,
		EmployeeID:           p.EmployeeID,
		Period:               p.Period,
		BasicSalary:          p.BasicSalary,
		Allowances:           p.Allowances,
		OvertimePay:          p.OvertimePay,
		Bonus:                p.Bonus,
		THR:                  p.THR,
		GrossIncome:          p.GrossIncome,
		EmployeeDeduction:    p.EmployeeDeduction,
		PPh21:                p.PPh21,
		NetSalary:            p.NetSalary,
		EmployerContribution: p.EmployerContribution,
		Status:               p.Status,
		CreatedAt:            p.CreatedAt,
	}
	if p.Employee != nil {
		res.EmployeeName = p.Employee.FullName
		res.NIK = p.Employee.NIK
		res.Position = p.Employee.Position
	} else if emp != nil {
		res.EmployeeName = emp.FullName
		res.NIK = emp.NIK
		res.Position = emp.Position
	}
	if len(p.Details) > 0 {
		var details []dto.PayslipDetailDTO
		if json.Unmarshal(p.Details, &details) == nil {
			res.Details = details
		}
	}
	return res
}

func lastDayOfMonth(period string) string {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return period + "-31"
	}
	return t.AddDate(0, 1, -1).Format("2006-01-02")
}

func periodEndDate(period string) time.Time {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Now()
	}
	return t.AddDate(0, 1, -1)
}

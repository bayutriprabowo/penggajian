package service

import (
	"context"
	"errors"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/payroll"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"gorm.io/gorm"
)

type overtimeService struct {
	overtimes repository.OvertimeRepository
	employees repository.EmployeeRepository
}

// NewOvertimeService membuat implementasi OvertimeService.
func NewOvertimeService(overtimes repository.OvertimeRepository, employees repository.EmployeeRepository) OvertimeService {
	return &overtimeService{overtimes: overtimes, employees: employees}
}

// Create memvalidasi lembur lalu menghitung upahnya sesuai PP 35/2021.
func (s *overtimeService) Create(ctx context.Context, req dto.OvertimeRequest) (*dto.OvertimeDTO, error) {
	if req.EmployeeID == 0 || req.Hours <= 0 || req.Date == "" {
		return nil, ErrBadRequest
	}
	if req.Hours > 24 {
		return nil, ErrBadRequest
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, ErrBadRequest
	}
	emp, err := s.employees.FindByID(ctx, req.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var fixed float64
	for _, a := range emp.Allowances {
		if a.Type == model.AllowanceFixed {
			fixed += a.Amount
		}
	}
	amount := payroll.OvertimePay(emp.BaseSalary, fixed, req.Hours, req.IsHoliday)
	o := &model.Overtime{
		EmployeeID: req.EmployeeID,
		Date:       date,
		Hours:      req.Hours,
		IsHoliday:  req.IsHoliday,
		Amount:     amount,
	}
	if err := s.overtimes.Create(ctx, o); err != nil {
		return nil, err
	}
	return &dto.OvertimeDTO{
		ID:         o.ID,
		EmployeeID: o.EmployeeID,
		Date:       o.Date,
		Hours:      o.Hours,
		IsHoliday:  o.IsHoliday,
		Amount:     o.Amount,
	}, nil
}

// ListByEmployee menampilkan riwayat lembur (opsional filter periode).
func (s *overtimeService) ListByEmployee(ctx context.Context, employeeID uint, period string) ([]dto.OvertimeDTO, error) {
	from, to := "", ""
	if period != "" {
		if _, err := time.Parse("2006-01", period); err != nil {
			return nil, ErrBadRequest
		}
		from = period + "-01"
		to = lastDayOfMonth(period)
	}
	list, err := s.overtimes.ListByEmployee(ctx, employeeID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]dto.OvertimeDTO, 0, len(list))
	for _, o := range list {
		out = append(out, dto.OvertimeDTO{
			ID:         o.ID,
			EmployeeID: o.EmployeeID,
			Date:       o.Date,
			Hours:      o.Hours,
			IsHoliday:  o.IsHoliday,
			Amount:     o.Amount,
		})
	}
	return out, nil
}

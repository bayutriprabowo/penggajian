package service

import (
	"context"
	"errors"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/payroll"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"gorm.io/gorm"
)

type taxService struct {
	payrolls  repository.PayrollRepository
	employees repository.EmployeeRepository
}

// NewTaxService membuat implementasi TaxService.
func NewTaxService(payrolls repository.PayrollRepository, employees repository.EmployeeRepository) TaxService {
	return &taxService{payrolls: payrolls, employees: employees}
}

// TERTables mengembalikan tabel TER A/B/C untuk ditampilkan ke API.
func (s *taxService) TERTables() []dto.TERTableDTO {
	tables := payroll.TERTables()
	out := make([]dto.TERTableDTO, 0, 3)
	for _, cat := range []string{payroll.TERCategoryA, payroll.TERCategoryB, payroll.TERCategoryC} {
		t := dto.TERTableDTO{Category: cat}
		for _, b := range tables[cat] {
			t.Rows = append(t.Rows, dto.TERRowDTO{From: b.From, UpTo: b.UpTo, Rate: b.Rate})
		}
		out = append(out, t)
	}
	return out
}

// TERInfo menyimulasikan PPh 21 TER bulanan.
func (s *taxService) TERInfo(ptkpStatus string, monthlyGross float64) (*dto.TERInfoDTO, error) {
	category, rate, err := payroll.MonthlyPPh21Rate(ptkpStatus, monthlyGross)
	if err != nil {
		return nil, ErrBadRequest
	}
	tax, err := payroll.MonthlyPPh21(ptkpStatus, monthlyGross)
	if err != nil {
		return nil, ErrBadRequest
	}
	return &dto.TERInfoDTO{
		PTKPStatus:   ptkpStatus,
		TERCategory:  category,
		MonthlyGross: monthlyGross,
		TERRate:      rate,
		MonthlyPPh21: tax,
	}, nil
}

// AnnualRecap menghitung rekap PPh 21 tahunan dan selisihnya dengan potongan TER.
func (s *taxService) AnnualRecap(ctx context.Context, req dto.AnnualRecapRequest) (*dto.AnnualRecapDTO, error) {
	if req.Year < 2000 || req.Year > 2100 {
		return nil, ErrBadRequest
	}
	emp, err := s.employees.FindByID(ctx, req.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	withheld, err := s.payrolls.SumPPh21ByEmployeeYear(ctx, req.EmployeeID, req.Year)
	if err != nil {
		return nil, err
	}
	result := payroll.AnnualPPh21(req.AnnualGross, req.JHTWorker, emp.PTKPStatus)
	diff := result.AnnualPPh21 - withheld
	return &dto.AnnualRecapDTO{
		EmployeeID:        req.EmployeeID,
		Year:              req.Year,
		AnnualGross:       req.AnnualGross,
		PositionAllowance: result.PositionAllowance,
		PensionDeduction:  result.PensionDeduction,
		NetIncome:         result.NetIncome,
		PTKP:              result.PTKP,
		PKP:               result.PKP,
		AnnualPPh21:       result.AnnualPPh21,
		WithheldPPh21:     withheld,
		Underpaid:         diff,
	}, nil
}

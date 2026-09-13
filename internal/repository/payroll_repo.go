package repository

import (
	"context"
	"fmt"

	"github.com/bayutriprabowo/penggajian/internal/model"
	"gorm.io/gorm"
)

type payrollRepo struct {
	db *gorm.DB
}

func NewPayrollRepository(db *gorm.DB) PayrollRepository {
	return &payrollRepo{db: db}
}

func (r *payrollRepo) Create(ctx context.Context, p *model.Payroll) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *payrollRepo) FindByID(ctx context.Context, id uint) (*model.Payroll, error) {
	var p model.Payroll
	if err := r.db.WithContext(ctx).Preload("Employee").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *payrollRepo) FindByEmployeePeriod(ctx context.Context, employeeID uint, period string) (*model.Payroll, error) {
	var p model.Payroll
	err := r.db.WithContext(ctx).Where("employee_id = ? AND period = ?", employeeID, period).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *payrollRepo) ListByEmployee(ctx context.Context, employeeID uint) ([]model.Payroll, error) {
	var list []model.Payroll
	err := r.db.WithContext(ctx).Preload("Employee").Where("employee_id = ?", employeeID).Order("period DESC").Find(&list).Error
	return list, err
}

func (r *payrollRepo) ListByPeriod(ctx context.Context, period string, page, limit int) ([]model.Payroll, int64, error) {
	var list []model.Payroll
	q := r.db.WithContext(ctx).Model(&model.Payroll{}).Where("period = ?", period)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Preload("Employee").Order("id ASC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payrollRepo) UpdateStatus(ctx context.Context, id uint, status string) error {
	res := r.db.WithContext(ctx).Model(&model.Payroll{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("payroll id %d tidak ditemukan", id)
	}
	return nil
}

func (r *payrollRepo) SumPPh21ByEmployeeYear(ctx context.Context, employeeID uint, year int) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&model.Payroll{}).
		Where("employee_id = ? AND period LIKE ?", employeeID, fmt.Sprintf("%d-%%", year)).
		Select("COALESCE(SUM(pph21), 0)").Scan(&sum).Error
	return sum, err
}

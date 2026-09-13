package repository

import (
	"context"
	"errors"

	"github.com/bayutriprabowo/penggajian/internal/model"
	"gorm.io/gorm"
)

type employeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepo{db: db}
}

func (r *employeeRepo) Create(ctx context.Context, e *model.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *employeeRepo) FindByID(ctx context.Context, id uint) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.WithContext(ctx).Preload("Allowances").First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepo) FindByNIK(ctx context.Context, nik string) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.WithContext(ctx).Preload("Allowances").Where("nik = ?", nik).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepo) List(ctx context.Context, page, limit int, status string) ([]model.Employee, int64, error) {
	var employees []model.Employee
	q := r.db.WithContext(ctx).Model(&model.Employee{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := q.Preload("Allowances").Order("id ASC").Offset(offset).Limit(limit).Find(&employees).Error; err != nil {
		return nil, 0, err
	}
	return employees, total, nil
}

func (r *employeeRepo) ListActive(ctx context.Context) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.WithContext(ctx).Preload("Allowances").Where("status = ?", "active").Find(&employees).Error
	return employees, err
}

func (r *employeeRepo) Update(ctx context.Context, e *model.Employee) error {
	return r.db.WithContext(ctx).Omit("Allowances").Save(e).Error
}

// UpdateWithAllowances memperbarui data karyawan sekaligus mengganti
// daftar tunjangan dalam satu transaksi (allowances nil = tidak diubah).
func (r *employeeRepo) UpdateWithAllowances(ctx context.Context, e *model.Employee, allowances *[]model.Allowance) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Allowances").Save(e).Error; err != nil {
			return err
		}
		if allowances == nil {
			return nil
		}
		if err := tx.Where("employee_id = ?", e.ID).Delete(&model.Allowance{}).Error; err != nil {
			return err
		}
		list := *allowances
		for i := range list {
			list[i].ID = 0
			list[i].EmployeeID = e.ID
		}
		if len(list) > 0 {
			return tx.Create(&list).Error
		}
		return nil
	})
}

func (r *employeeRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Employee{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("karyawan tidak ditemukan")
	}
	return nil
}

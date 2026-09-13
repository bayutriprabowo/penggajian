package repository

import (
	"context"

	"github.com/bayutriprabowo/penggajian/internal/model"
	"gorm.io/gorm"
)

type overtimeRepo struct {
	db *gorm.DB
}

// NewOvertimeRepository membuat implementasi OvertimeRepository berbasis GORM.
func NewOvertimeRepository(db *gorm.DB) OvertimeRepository {
	return &overtimeRepo{db: db}
}

// Create menyimpan catatan lembur.
func (r *overtimeRepo) Create(ctx context.Context, o *model.Overtime) error {
	return r.db.WithContext(ctx).Create(o).Error
}

// SumByEmployeePeriod menjumlahkan upah lembur karyawan pada rentang tanggal.
func (r *overtimeRepo) SumByEmployeePeriod(ctx context.Context, employeeID uint, from, to string) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&model.Overtime{}).
		Where("employee_id = ? AND date >= ? AND date <= ?", employeeID, from, to).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

// ListByEmployee menampilkan riwayat lembur karyawan (opsional filter rentang tanggal).
func (r *overtimeRepo) ListByEmployee(ctx context.Context, employeeID uint, from, to string) ([]model.Overtime, error) {
	var list []model.Overtime
	q := r.db.WithContext(ctx).Where("employee_id = ?", employeeID)
	if from != "" && to != "" {
		q = q.Where("date >= ? AND date <= ?", from, to)
	}
	err := q.Order("date DESC").Find(&list).Error
	return list, err
}

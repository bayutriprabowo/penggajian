package repository

import (
	"context"

	"github.com/bayutriprabowo/penggajian/internal/model"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository membuat implementasi UserRepository berbasis GORM.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create menyimpan user baru.
func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// FindByID mencari user berdasarkan ID (termasuk role).
func (r *userRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByUsername mencari user berdasarkan username (termasuk role).
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmail mencari user berdasarkan email (termasuk role).
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// WithRole memuat relasi role ke user.
func (r *userRepo) WithRole(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Preload("Role").First(u, u.ID).Error
}

// Update menyimpan perubahan data user.
func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// List menampilkan user dengan paginasi beserta totalnya.
func (r *userRepo) List(ctx context.Context, page, limit int) ([]model.User, int64, error) {
	var users []model.User
	q := r.db.WithContext(ctx).Model(&model.User{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Preload("Role").Order("id ASC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

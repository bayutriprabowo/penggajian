package repository

import (
	"context"

	"github.com/bayutriprabowo/penggajian/internal/model"
	"gorm.io/gorm"
)

type roleRepo struct {
	db *gorm.DB
}

// NewRoleRepository membuat implementasi RoleRepository berbasis GORM.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepo{db: db}
}

// CreateRole menyimpan role baru.
func (r *roleRepo) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// FindRoleByName mencari role berdasarkan nama.
func (r *roleRepo) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindPermissionsByName mencari permission yang sudah ada dari daftar nama.
func (r *roleRepo) FindPermissionsByName(ctx context.Context, names []string) ([]model.Permission, error) {
	var perms []model.Permission
	if len(names) == 0 {
		return perms, nil
	}
	if err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// CreatePermissions menyimpan permission baru.
func (r *roleRepo) CreatePermissions(ctx context.Context, perms []model.Permission) error {
	if len(perms) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&perms).Error
}

// AssignPermissions mengganti seluruh permission milik sebuah role.
func (r *roleRepo) AssignPermissions(ctx context.Context, roleID uint, perms []model.Permission) error {
	return r.db.WithContext(ctx).Model(&model.Role{ID: roleID}).Association("Permissions").Replace(perms)
}

// RoleWithPermissions memuat role beserta daftar permissionnya.
func (r *roleRepo) RoleWithPermissions(ctx context.Context, roleID uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// PermissionsByRole mengembalikan nama permission sebuah role.
func (r *roleRepo) PermissionsByRole(ctx context.Context, roleID uint) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID).
		Pluck("permissions.name", &names).Error
	return names, err
}

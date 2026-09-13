package service

import (
	"context"
	"errors"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/dberr"
	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authService struct {
	users       repository.UserRepository
	roles       repository.RoleRepository
	jwtSecret   string
	jwtExpHours int
}

// NewAuthService membuat implementasi AuthService.
func NewAuthService(users repository.UserRepository, roles repository.RoleRepository, jwtSecret string, jwtExpHours int) AuthService {
	return &authService{users: users, roles: roles, jwtSecret: jwtSecret, jwtExpHours: jwtExpHours}
}

// Register membuat user baru dengan password bcrypt dan role tertentu.
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserDTO, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, ErrBadRequest
	}
	if len(req.Password) < 6 {
		return nil, ErrBadRequest
	}
	if len(req.Username) > 100 || len(req.Email) > 100 {
		return nil, ErrBadRequest
	}
	if _, err := s.users.FindByUsername(ctx, req.Username); err == nil {
		return nil, ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if _, err := s.users.FindByEmail(ctx, req.Email); err == nil {
		return nil, ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	roleName := req.RoleName
	if roleName == "" {
		roleName = "employee"
	}
	role, err := s.roles.FindRoleByName(ctx, roleName)
	if err != nil {
		return nil, ErrBadRequest
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   string(hash),
		RoleID:     role.ID,
		EmployeeID: req.EmployeeID,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if dberr.IsUniqueViolation(err) {
			return nil, ErrConflict
		}
		if dberr.IsForeignKeyViolation(err) {
			return nil, ErrBadRequest
		}
		return nil, err
	}
	user.Role = role
	return toUserDTO(user), nil
}

// Login memverifikasi kredensial dan menerbitkan token JWT.
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrBadRequest
	}
	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, ErrUnauthorized
	}
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{Token: token, User: *toUserDTO(user)}, nil
}

// ListUsers menampilkan daftar user dengan paginasi.
func (s *authService) ListUsers(ctx context.Context, page, limit int) (*dto.UserListDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	users, total, err := s.users.List(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	res := &dto.UserListDTO{Pagination: dto.Pagination{Page: page, Limit: limit, Total: int(total)}}
	for i := range users {
		res.Users = append(res.Users, *toUserDTO(&users[i]))
	}
	return res, nil
}

// GenerateToken membuat JWT berisi identitas user.
func (s *authService) GenerateToken(user *model.User) (string, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	claims := TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ParseToken memvalidasi tanda tangan dan struktur token JWT.
func (s *authService) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return []byte(s.jwtSecret), nil
	})
}

// toUserDTO mengubah model user menjadi DTO (tanpa field sensitif).
func toUserDTO(u *model.User) *dto.UserDTO {
	res := &dto.UserDTO{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		RoleID:     u.RoleID,
		EmployeeID: u.EmployeeID,
	}
	if u.Role != nil {
		res.RoleName = u.Role.Name
	}
	return res
}

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"github.com/bayutriprabowo/penggajian/internal/service"
	"gorm.io/gorm"
)

type contextKey string

const (
	userKey        contextKey = "user"
	permissionsKey contextKey = "permissions"
)

// Auth memvalidasi Bearer token, memuat user + permission ke context.
// Jika tidak ada token, request ditolak (401).
func Auth(auth service.AuthService, users repository.UserRepository, roles repository.RoleRepository) func(http.Handler) http.Handler {
	return authMiddleware(auth, users, roles, false)
}

// OptionalAuth memuat user + permission bila token valid; tanpa token tetap lanjut
// (user di context bernilai nil). Dipakai untuk endpoint yang mengubah perilaku
// berdasarkan identitas pemanggil, misalnya register.
func OptionalAuth(auth service.AuthService, users repository.UserRepository, roles repository.RoleRepository) func(http.Handler) http.Handler {
	return authMiddleware(auth, users, roles, true)
}

func authMiddleware(auth service.AuthService, users repository.UserRepository, roles repository.RoleRepository, optional bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			tokenString := strings.TrimPrefix(header, "Bearer ")
			if tokenString == "" || tokenString == header {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "token tidak ditemukan")
				return
			}
			token, err := auth.ParseToken(tokenString)
			if err != nil || !token.Valid {
				if optional {
					// token kedaluwarsa/tidak valid: lanjut sebagai anonim
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "token tidak valid")
				return
			}
			claims, ok := token.Claims.(*service.TokenClaims)
			if !ok {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "klaim token tidak valid")
				return
			}
			user, err := users.FindByID(r.Context(), claims.UserID)
			if err != nil {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					writeError(w, http.StatusUnauthorized, "user tidak ditemukan")
				} else {
					writeError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
			perms, err := roles.PermissionsByRole(r.Context(), user.RoleID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			ctx := context.WithValue(r.Context(), userKey, user)
			ctx = context.WithValue(ctx, permissionsKey, perms)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission membatasi akses endpoint berdasarkan permission.
func RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasPermission(r.Context(), perm) {
				writeError(w, http.StatusForbidden, "tidak memiliki permission: "+perm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission memperbolehkan salah satu dari daftar permission.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range perms {
				if HasPermission(r.Context(), p) {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "tidak memiliki permission yang dibutuhkan")
		})
	}
}

// HasPermission mengecek permission pada context request.
func HasPermission(ctx context.Context, perm string) bool {
	granted, _ := ctx.Value(permissionsKey).([]string)
	for _, p := range granted {
		if p == perm {
			return true
		}
	}
	return false
}

// PermissionsFromContext mengambil daftar permission pemanggil.
func PermissionsFromContext(ctx context.Context) []string {
	perms, _ := ctx.Value(permissionsKey).([]string)
	return perms
}

// UserFromContext mengambil user dari context (untuk handler).
func UserFromContext(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey).(*model.User)
	return user
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.Error(message))
}

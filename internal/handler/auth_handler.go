package handler

import (
	"net/http"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/middleware"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register membuat user baru.
// Publik: role dipaksa "employee" (tanpa user:write).
// Dengan token user yang punya permission user:write: boleh memilih role lain.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	if !middleware.HasPermission(r.Context(), "user:write") {
		req.RoleName = "employee"
		req.EmployeeID = nil
	}
	user, err := h.auth.Register(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, "registrasi berhasil", user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	resp, err := h.auth.Login(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "login berhasil", resp)
}

type UserHandler struct {
	auth service.AuthService
}

func NewUserHandler(auth service.AuthService) *UserHandler {
	return &UserHandler{auth: auth}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, service.ErrUnauthorized)
		return
	}
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	writeOK(w, "profil user", dto.UserDTO{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		RoleID:     user.RoleID,
		RoleName:   roleName,
		EmployeeID: user.EmployeeID,
	})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 10)
	list, err := h.auth.ListUsers(r.Context(), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "daftar user", list)
}

package dto

type RegisterRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	RoleName   string `json:"role_name,omitempty"`
	EmployeeID *uint  `json:"employee_id,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserDTO struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	RoleID     uint   `json:"role_id"`
	RoleName   string `json:"role_name"`
	EmployeeID *uint  `json:"employee_id"`
}

type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type UserListDTO struct {
	Users      []UserDTO  `json:"users"`
	Pagination Pagination `json:"pagination"`
}

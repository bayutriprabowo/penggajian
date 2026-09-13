package handler

import (
	"net/http"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type EmployeeHandler struct {
	employees service.EmployeeService
}

func NewEmployeeHandler(employees service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employees: employees}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.EmployeeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	emp, err := h.employees.Create(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, "karyawan berhasil ditambahkan", emp)
}

func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	emp, err := h.employees.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "detail karyawan", emp)
}

func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 10)
	status := r.URL.Query().Get("status")
	list, err := h.employees.List(r.Context(), page, limit, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "daftar karyawan", list)
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req dto.EmployeeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	emp, err := h.employees.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "karyawan berhasil diperbarui", emp)
}

func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.employees.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "karyawan berhasil dihapus", nil)
}

package handler

import (
	"net/http"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/middleware"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type PayrollHandler struct {
	payroll service.PayrollService
}

// NewPayrollHandler membuat handler untuk payroll.
func NewPayrollHandler(payroll service.PayrollService) *PayrollHandler {
	return &PayrollHandler{payroll: payroll}
}

// Run men-generate payroll untuk periode tertentu.
func (h *PayrollHandler) Run(w http.ResponseWriter, r *http.Request) {
	var req dto.PayrollRunRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	list, err := h.payroll.Run(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, "payroll berhasil di-generate", list)
}

// GetByID menampilkan satu slip gaji.
func (h *PayrollHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	user := middleware.UserFromContext(r.Context())
	p, err := h.payroll.GetByID(r.Context(), id, user)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "slip gaji", p)
}

// ListByEmployee menampilkan riwayat slip gaji karyawan.
func (h *PayrollHandler) ListByEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	user := middleware.UserFromContext(r.Context())
	list, err := h.payroll.ListByEmployee(r.Context(), id, user)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "riwayat slip gaji", list)
}

// ListByPeriod menampilkan semua slip gaji pada satu periode.
func (h *PayrollHandler) ListByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 10)
	user := middleware.UserFromContext(r.Context())
	list, err := h.payroll.ListByPeriod(r.Context(), period, page, limit, user)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "daftar payroll periode "+period, list)
}

// Approve mengubah status payroll draft menjadi approved.
func (h *PayrollHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.payroll.Approve(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "payroll disetujui", nil)
}

// MarkPaid mengubah status payroll approved menjadi paid.
func (h *PayrollHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.payroll.MarkPaid(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "payroll ditandai lunas", nil)
}

// CalculateTHR menghitung THR karyawan (penuh/proporsional + pajaknya).
func (h *PayrollHandler) CalculateTHR(w http.ResponseWriter, r *http.Request) {
	req := dto.THRCalculateRequest{
		EmployeeID: queryUint(r, "employee_id", 0),
		Period:     r.URL.Query().Get("period"),
	}
	resp, err := h.payroll.CalculateTHR(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "kalkulasi THR", resp)
}

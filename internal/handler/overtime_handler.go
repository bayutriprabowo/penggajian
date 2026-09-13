package handler

import (
	"net/http"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type OvertimeHandler struct {
	overtimes service.OvertimeService
}

// NewOvertimeHandler membuat handler untuk data lembur.
func NewOvertimeHandler(overtimes service.OvertimeService) *OvertimeHandler {
	return &OvertimeHandler{overtimes: overtimes}
}

// Create mencatat lembur baru (upah dihitung otomatis sesuai PP 35/2021).
func (h *OvertimeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.OvertimeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, service.ErrBadRequest)
		return
	}
	o, err := h.overtimes.Create(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, "lembur berhasil dicatat", o)
}

// ListByEmployee menampilkan riwayat lembur karyawan (bisa difilter per periode).
func (h *OvertimeHandler) ListByEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	period := r.URL.Query().Get("period")
	list, err := h.overtimes.ListByEmployee(r.Context(), id, period)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "daftar lembur", list)
}

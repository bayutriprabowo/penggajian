package handler

import (
	"net/http"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type TaxHandler struct {
	tax service.TaxService
}

// NewTaxHandler membuat handler untuk pajak.
func NewTaxHandler(tax service.TaxService) *TaxHandler {
	return &TaxHandler{tax: tax}
}

// TERTables menampilkan tabel TER A/B/C lengkap.
func (h *TaxHandler) TERTables(w http.ResponseWriter, r *http.Request) {
	writeOK(w, "tabel TER PPh 21", h.tax.TERTables())
}

// TERInfo menyimulasikan PPh 21 TER untuk status PTKP dan bruto tertentu.
func (h *TaxHandler) TERInfo(w http.ResponseWriter, r *http.Request) {
	ptkp := r.URL.Query().Get("ptkp")
	gross := queryFloat(r, "gross", 0)
	info, err := h.tax.TERInfo(ptkp, gross)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "info TER", info)
}

// AnnualRecap menghitung rekap PPh 21 tahunan (Pasal 17).
func (h *TaxHandler) AnnualRecap(w http.ResponseWriter, r *http.Request) {
	req := dto.AnnualRecapRequest{
		EmployeeID:  queryUint(r, "employee_id", 0),
		Year:        queryInt(r, "year", 0),
		AnnualGross: queryFloat(r, "annual_gross", 0),
		JHTWorker:   queryFloat(r, "jht_worker", 0),
	}
	resp, err := h.tax.AnnualRecap(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, "rekap PPh 21 tahunan", resp)
}

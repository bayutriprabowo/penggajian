package dto

type TERTableDTO struct {
	Category string      `json:"category"`
	Rows     []TERRowDTO `json:"rows"`
}

type TERRowDTO struct {
	From float64 `json:"from"`  // juta rupiah, 0 jika baris pertama
	UpTo float64 `json:"up_to"` // juta rupiah, -1 artinya di atas From tanpa batas
	Rate float64 `json:"rate"`  // persen
}

type TERInfoDTO struct {
	PTKPStatus   string  `json:"ptkp_status"`
	TERCategory  string  `json:"ter_category"`
	MonthlyGross float64 `json:"monthly_gross"`
	TERRate      float64 `json:"ter_rate"`
	MonthlyPPh21 float64 `json:"monthly_pph21"`
}

type AnnualRecapRequest struct {
	EmployeeID  uint    `json:"employee_id"`
	Year        int     `json:"year"`
	AnnualGross float64 `json:"annual_gross"` // penghasilan bruto setahun (gaji+tunjangan+bonus+THR)
	JHTWorker   float64 `json:"jht_worker"`   // iuran JHT+JP dibayar pekerja setahun (pengurang neto, optional)
}

type AnnualRecapDTO struct {
	EmployeeID        uint    `json:"employee_id"`
	Year              int     `json:"year"`
	AnnualGross       float64 `json:"annual_gross"`
	PositionAllowance float64 `json:"position_allowance"` // biaya jabatan 5%, max 6jt
	PensionDeduction  float64 `json:"pension_deduction"`
	NetIncome         float64 `json:"net_income"`
	PTKP              float64 `json:"ptkp"`
	PKP               float64 `json:"pkp"`
	AnnualPPh21       float64 `json:"annual_pph21"`
	WithheldPPh21     float64 `json:"withheld_pph21"` // jumlah yang sudah dipotong TER selama setahun
	Underpaid         float64 `json:"underpaid"`      // kurang bayar (positif) / lebih bayar (negatif)
}

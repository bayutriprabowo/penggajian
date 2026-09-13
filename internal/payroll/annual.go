package payroll

import "math"

// Rekap PPh 21 tahunan sesuai UU HPP (Pasal 17) + PTKP PMK 101/2016.

type AnnualTaxResult struct {
	PositionAllowance float64 // biaya jabatan 5%, maks Rp 6.000.000
	PensionDeduction  float64 // biaya pensiun (iuran JHT+JP pekerja), maks Rp 2.400.000
	NetIncome         float64
	PTKP              float64
	PKP               float64
	AnnualPPh21       float64
}

const (
	positionAllowanceMax = 6_000_000
	pensionDeductionMax  = 2_400_000
	basePTKP             = 54_000_000
	ptkpPerDependent     = 4_500_000
	ptkpMarried          = 4_500_000
)

// PTKPAmount menghitung PTKP berdasarkan status (TK0..K3).
func PTKPAmount(ptkpStatus string) float64 {
	switch ptkpStatus {
	case "TK0":
		return basePTKP
	case "TK1", "K0":
		return basePTKP + ptkpPerDependent
	case "TK2", "K1":
		return basePTKP + 2*ptkpPerDependent
	case "TK3", "K2":
		return basePTKP + 3*ptkpPerDependent
	case "K3":
		return basePTKP + ptkpMarried + 3*ptkpPerDependent
	default:
		return basePTKP
	}
}

// AnnualPPh21 menghitung PPh 21 setahun (tarif progresif Pasal 17 UU HPP).
func AnnualPPh21(annualGross, jhtWorkerPaid float64, ptkpStatus string) AnnualTaxResult {
	position := math.Min(annualGross*0.05, positionAllowanceMax)
	pension := math.Min(jhtWorkerPaid, pensionDeductionMax)
	netIncome := annualGross - position - pension
	ptkp := PTKPAmount(ptkpStatus)
	pkp := math.Max(netIncome-ptkp, 0)

	var tax float64
	switch {
	case pkp <= 60_000_000:
		tax = pkp * 0.05
	case pkp <= 250_000_000:
		tax = 60_000_000*0.05 + (pkp-60_000_000)*0.15
	case pkp <= 500_000_000:
		tax = 60_000_000*0.05 + 190_000_000*0.15 + (pkp-250_000_000)*0.25
	case pkp <= 5_000_000_000:
		tax = 60_000_000*0.05 + 190_000_000*0.15 + 250_000_000*0.25 + (pkp-500_000_000)*0.30
	default:
		tax = 60_000_000*0.05 + 190_000_000*0.15 + 250_000_000*0.25 + 4_500_000_000*0.30 + (pkp-5_000_000_000)*0.35
	}

	return AnnualTaxResult{
		PositionAllowance: math.Floor(position),
		PensionDeduction:  math.Floor(pension),
		NetIncome:         math.Floor(netIncome),
		PTKP:              ptkp,
		PKP:               math.Floor(pkp),
		AnnualPPh21:       math.Floor(tax),
	}
}

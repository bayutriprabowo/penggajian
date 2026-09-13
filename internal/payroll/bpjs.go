package payroll

import "math"

type BPJSRates struct {
	HealthCap      float64 // upah maksimum BPJS Kesehatan
	HealthWorker   float64 // 1%
	HealthEmployer float64 // 4%
	JPCap          float64 // upah maksimum JP
	JPWorker       float64 // 1%
	JPEmployer     float64 // 2%
	JHTWorker      float64 // 2%
	JHTEmployer    float64 // 3.7%
	JKMEmployer    float64 // 0.3%
	JKKRisk        float64 // 0.24 - 1.74 (%)
}

type BPJSResult struct {
	HealthWorker  float64 `json:"health_worker"`
	HealthEmploy  float64 `json:"health_employer"`
	JHTWorker     float64 `json:"jht_worker"`
	JHTEmployer   float64 `json:"jht_employer"`
	JKKEmployer   float64 `json:"jkk_employer"`
	JKMEmployer   float64 `json:"jkm_employer"`
	JPWorker      float64 `json:"jp_worker"`
	JPEmployer    float64 `json:"jp_employer"`
	WorkerTotal   float64 `json:"worker_total"`
	EmployerTotal float64 `json:"employer_total"`
}

// CalculateBPJS menghitung iuran BPJS Kesehatan & Ketenagakerjaan.
// enabledHealth/enabledTK: keikutsertaan karyawan.
func CalculateBPJS(r BPJSRates, monthlyWage float64, enabledHealth, enabledTK bool) BPJSResult {
	var res BPJSResult
	if enabledHealth {
		healthBase := math.Min(monthlyWage, r.HealthCap)
		res.HealthWorker = math.Floor(healthBase*r.HealthWorker + 0.5)
		res.HealthEmploy = math.Floor(healthBase*r.HealthEmployer + 0.5)
	}
	if enabledTK {
		res.JHTWorker = math.Floor(monthlyWage*r.JHTWorker + 0.5)
		res.JHTEmployer = math.Floor(monthlyWage*r.JHTEmployer + 0.5)
		res.JKKEmployer = math.Floor(monthlyWage*r.JKKRisk/100 + 0.5)
		res.JKMEmployer = math.Floor(monthlyWage*r.JKMEmployer + 0.5)
		jpBase := math.Min(monthlyWage, r.JPCap)
		res.JPWorker = math.Floor(jpBase*r.JPWorker + 0.5)
		res.JPEmployer = math.Floor(jpBase*r.JPEmployer + 0.5)
	}
	res.WorkerTotal = res.HealthWorker + res.JHTWorker + res.JPWorker
	res.EmployerTotal = res.HealthEmploy + res.JHTEmployer + res.JKKEmployer + res.JKMEmployer + res.JPEmployer
	return res
}

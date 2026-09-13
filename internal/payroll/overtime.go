package payroll

import "math"

// Upah lembur sesuai PP 35/2021.
// Upah sejam = 1/173 x (gaji pokok + tunjangan tetap).

const OvertimeDivisor = 173

// hourlyRate menghitung upah sejam = 1/173 x (gaji pokok + tunjangan tetap).
func hourlyRate(baseSalary, fixedAllowance float64) float64 {
	return (baseSalary + fixedAllowance) / OvertimeDivisor
}

// OvertimePay menghitung upah lembur.
// Hari kerja: jam pertama 1.5x, berikutnya 2x.
// Hari libur (asumsi 6 hari kerja/minggu): 7 jam pertama 2x, jam ke-8 3x, jam 9-10 4x.
func OvertimePay(baseSalary, fixedAllowance, hours float64, isHoliday bool) float64 {
	if hours <= 0 {
		return 0
	}
	rate := hourlyRate(baseSalary, fixedAllowance)
	if !isHoliday {
		if hours <= 1 {
			return math.Floor(rate * 1.5 * hours)
		}
		return math.Floor(rate*1.5 + rate*2*(hours-1))
	}
	switch {
	case hours <= 7:
		return math.Floor(rate * 2 * hours)
	case hours <= 8:
		return math.Floor(rate*2*7 + rate*3*(hours-7))
	case hours <= 10:
		return math.Floor(rate*2*7 + rate*3 + rate*4*(hours-8))
	default:
		return math.Floor(rate*2*7 + rate*3 + rate*4*2)
	}
}

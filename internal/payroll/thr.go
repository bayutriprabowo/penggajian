package payroll

import (
	"math"
	"time"
)

// THR sesuai Permenaker 6/2016.
// Masa kerja >= 12 bulan: 1x (gaji pokok + tunjangan tetap).
// Masa kerja < 12 bulan: (masa kerja bulan / 12) x (gaji pokok + tunjangan tetap).

type THRResult struct {
	MonthsOfService int
	FullTHR         bool
	Amount          float64
}

// CalculateTHR menghitung THR sesuai Permenaker 6/2016:
// masa kerja >= 12 bulan mendapat 1x gaji, di bawahnya proporsional (bulan/12).
func CalculateTHR(baseSalary, fixedAllowance float64, joinDate, periodDate time.Time) THRResult {
	months := fullMonthsBetween(joinDate, periodDate)
	monthly := baseSalary + fixedAllowance
	if months >= 12 {
		return THRResult{MonthsOfService: months, FullTHR: true, Amount: math.Floor(monthly)}
	}
	return THRResult{
		MonthsOfService: months,
		FullTHR:         false,
		Amount:          math.Floor(monthly * float64(months) / 12),
	}
}

// fullMonthsBetween menghitung jumlah bulan penuh antara dua tanggal.
func fullMonthsBetween(from, to time.Time) int {
	months := (to.Year()-from.Year())*12 + int(to.Month()-from.Month())
	if to.Day() < from.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

package payroll

import (
	"fmt"
	"math"
)

// TER (Tarif Efektif Rata-rata) bulanan sesuai PP 58/2023 / PMK 168/2023.
// Rentang dalam juta rupiah. UpTo == -1 berarti tanpa batas atas.

type TERBracket struct {
	From float64
	UpTo float64
	Rate float64 // persen
}

const (
	TERCategoryA = "A"
	TERCategoryB = "B"
	TERCategoryC = "C"
)

var terTableA = []TERBracket{
	{0, 5.40, 0}, {5.40, 5.65, 0.25}, {5.65, 5.95, 0.50}, {5.95, 6.30, 0.75},
	{6.30, 6.75, 1.00}, {6.75, 7.50, 1.25}, {7.50, 8.55, 1.50}, {8.55, 9.65, 1.75},
	{9.65, 10.05, 2.00}, {10.05, 10.35, 2.25}, {10.35, 10.70, 2.50}, {10.70, 11.05, 3.00},
	{11.05, 11.60, 3.50}, {11.60, 12.50, 4.00}, {12.50, 13.75, 5.00}, {13.75, 15.10, 6.00},
	{15.10, 16.95, 7.00}, {16.95, 19.75, 8.00}, {19.75, 24.15, 9.00}, {24.15, 26.45, 10.00},
	{26.45, 28.00, 11.00}, {28.00, 30.05, 12.00}, {30.05, 32.40, 13.00}, {32.40, 35.40, 14.00},
	{35.40, 39.10, 15.00}, {39.10, 43.85, 16.00}, {43.85, 47.80, 17.00}, {47.80, 51.40, 18.00},
	{51.40, 56.30, 19.00}, {56.30, 62.20, 20.00}, {62.20, 68.60, 21.00}, {68.60, 77.50, 22.00},
	{77.50, 89.00, 23.00}, {89.00, 103.00, 24.00}, {103.00, 125.00, 25.00}, {125.00, 157.00, 26.00},
	{157.00, 206.00, 27.00}, {206.00, 337.00, 28.00}, {337.00, 454.00, 29.00}, {454.00, 550.00, 30.00},
	{550.00, 695.00, 31.00}, {695.00, 910.00, 32.00}, {910.00, 1400.00, 33.00}, {1400.00, -1, 34.00},
}

var terTableB = []TERBracket{
	{0, 6.20, 0}, {6.20, 6.50, 0.25}, {6.50, 6.85, 0.50}, {6.85, 7.30, 0.75},
	{7.30, 9.20, 1.00}, {9.20, 10.75, 1.50}, {10.75, 11.25, 2.00}, {11.25, 11.60, 2.50},
	{11.60, 12.60, 3.00}, {12.60, 13.60, 4.00}, {13.60, 14.95, 5.00}, {14.95, 16.40, 6.00},
	{16.40, 18.45, 7.00}, {18.45, 21.65, 8.00}, {21.65, 26.45, 9.00}, {26.45, 31.35, 10.00},
	{31.35, 32.60, 11.00}, {32.60, 36.25, 12.00}, {36.25, 40.55, 13.00}, {40.55, 44.75, 14.00},
	{44.75, 50.05, 15.00}, {50.05, 53.55, 16.00}, {53.55, 55.15, 17.00}, {55.15, 58.55, 18.00},
	{58.55, 62.45, 19.00}, {62.45, 68.05, 20.00}, {68.05, 74.25, 21.00}, {74.25, 80.45, 22.00},
	{80.45, 85.75, 23.00}, {85.75, 93.55, 24.00}, {93.55, 102.30, 25.00}, {102.30, 111.15, 26.00},
	{111.15, 119.90, 27.00}, {119.90, 129.15, 28.00}, {129.15, 140.30, 29.00}, {140.30, 150.30, 30.00},
	{150.30, 160.00, 31.00}, {160.00, 169.50, 32.00}, {169.50, 180.20, 33.00}, {180.20, -1, 34.00},
}

var terTableC = []TERBracket{
	{0, 6.60, 0}, {6.60, 6.95, 0.25}, {6.95, 7.35, 0.50}, {7.35, 7.80, 0.75},
	{7.80, 8.85, 1.00}, {8.85, 9.80, 1.25}, {9.80, 10.95, 1.50}, {10.95, 11.20, 1.75},
	{11.20, 12.05, 2.00}, {12.05, 12.95, 3.00}, {12.95, 14.15, 4.00}, {14.15, 15.55, 5.00},
	{15.55, 17.00, 6.00}, {17.00, 19.55, 7.00}, {19.55, 22.25, 8.00}, {22.25, 26.60, 9.00},
	{26.60, 32.05, 10.00}, {32.05, 33.75, 11.00}, {33.75, 35.35, 12.00}, {35.35, 39.15, 13.00},
	{39.15, 41.85, 14.00}, {41.85, 46.85, 15.00}, {46.85, 50.25, 16.00}, {50.25, 52.85, 17.00},
	{52.85, 56.65, 18.00}, {56.65, 59.55, 19.00}, {59.55, 64.55, 20.00}, {64.55, 70.05, 21.00},
	{70.05, 76.35, 22.00}, {76.35, 83.65, 23.00}, {83.65, 91.60, 24.00}, {91.60, 101.40, 25.00},
	{101.40, 110.40, 26.00}, {110.40, 119.70, 27.00}, {119.70, 128.90, 28.00}, {128.90, 139.50, 29.00},
	{139.50, 151.40, 30.00}, {151.40, 166.30, 31.00}, {166.30, 175.80, 32.00}, {175.80, 185.60, 33.00},
	{185.60, -1, 34.00},
}

// TERTables mengembalikan seluruh tabel TER per kategori.
func TERTables() map[string][]TERBracket {
	return map[string][]TERBracket{
		TERCategoryA: terTableA,
		TERCategoryB: terTableB,
		TERCategoryC: terTableC,
	}
}

// TERCategory memetakan status PTKP ke kategori TER.
// A: TK/0, TK/1, K/0 — B: TK/2, TK/3, K/1, K/2 — C: K/3
func TERCategory(ptkpStatus string) (string, error) {
	switch ptkpStatus {
	case "TK0", "TK1", "K0":
		return TERCategoryA, nil
	case "TK2", "TK3", "K1", "K2":
		return TERCategoryB, nil
	case "K3":
		return TERCategoryC, nil
	default:
		return "", fmt.Errorf("status PTKP tidak valid: %s", ptkpStatus)
	}
}

// terRate mencari tarif TER yang berlaku untuk bruto bulanan pada kategori tertentu.
func terRate(category string, monthlyGross float64) float64 {
	table := TERTables()[category]
	grossMillions := monthlyGross / 1_000_000
	for _, b := range table {
		if grossMillions >= b.From && (b.UpTo == -1 || grossMillions < b.UpTo) {
			return b.Rate
		}
	}
	return 0
}

// MonthlyPPh21 menghitung PPh 21 bulanan metode TER.
// Dikembalikan dibulatkan ke bawah rupiah penuh.
func MonthlyPPh21(ptkpStatus string, monthlyGross float64) (float64, error) {
	category, err := TERCategory(ptkpStatus)
	if err != nil {
		return 0, err
	}
	rate := terRate(category, monthlyGross)
	return math.Floor(monthlyGross * rate / 100), nil
}

// MonthlyPPh21Rate mengembalikan tarif TER yang berlaku.
func MonthlyPPh21Rate(ptkpStatus string, monthlyGross float64) (string, float64, error) {
	category, err := TERCategory(ptkpStatus)
	if err != nil {
		return "", 0, err
	}
	return category, terRate(category, monthlyGross), nil
}

// PPh21OnTHR = TER(bruto+THR) - TER(bruto), sesuai ketentuan pajak THR.
func PPh21OnTHR(ptkpStatus string, regularGross, thr float64) (float64, error) {
	taxRegular, err := MonthlyPPh21(ptkpStatus, regularGross)
	if err != nil {
		return 0, err
	}
	taxWithTHR, err := MonthlyPPh21(ptkpStatus, regularGross+thr)
	if err != nil {
		return 0, err
	}
	return math.Floor(taxWithTHR - taxRegular), nil
}

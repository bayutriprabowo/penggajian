package payroll

import (
	"testing"
	"time"
)

// TestMonthlyPPh21TER memastikan tarif TER A dan C untuk bruto 10 juta.
func TestMonthlyPPh21TER(t *testing.T) {
	// TK0, gaji 10 juta -> kategori A, bracket 9.65-10.05 = 2%
	tax, err := MonthlyPPh21("TK0", 10_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if tax != 200_000 {
		t.Fatalf("want 200000 got %v", tax)
	}
	// K3, gaji 10 juta -> kategori C, bracket 9.8-10.95 = 1.5%
	tax, err = MonthlyPPh21("K3", 10_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if tax != 150_000 {
		t.Fatalf("want 150000 got %v", tax)
	}
}

// TestOvertime memverifikasi rumus lembur hari kerja.
func TestOvertime(t *testing.T) {
	// gaji 10jt, lembur 2 jam hari kerja: 1.5x + 2x dari (10jt/173), dibulatkan ke bawah
	r := 10_000_000.0 / 173
	want := float64(int64(r*1.5 + r*2))
	if got := OvertimePay(10_000_000, 0, 2, false); got != want {
		t.Fatalf("want %v got %v", want, got)
	}
}

// TestBPJS memverifikasi iuran BPJS Kesehatan dan JHT/JP.
func TestBPJS(t *testing.T) {
	r := BPJSRates{
		HealthCap: 12_000_000, HealthWorker: 0.01, HealthEmployer: 0.04,
		JPCap: 10_042_300, JPWorker: 0.01, JPEmployer: 0.02,
		JHTWorker: 0.02, JHTEmployer: 0.037, JKMEmployer: 0.003, JKKRisk: 0.54,
	}
	res := CalculateBPJS(r, 10_000_000, true, true)
	if res.HealthWorker != 100_000 || res.JHTWorker != 200_000 || res.JPWorker != 100_000 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

// TestTHR memverifikasi THR penuh dan proporsional.
func TestTHR(t *testing.T) {
	join, _ := time.Parse("2006-01-02", "2025-03-01")
	period, _ := time.Parse("2006-01", "2026-03")
	res := CalculateTHR(8_000_000, 1_000_000, join, period)
	if !res.FullTHR || res.Amount != 9_000_000 {
		t.Fatalf("want full 9000000 got %+v", res)
	}
	// masa kerja 5 bulan (basis akhir Maret)
	join2, _ := time.Parse("2006-01-02", "2025-10-15")
	period2, _ := time.Parse("2006-01-02", "2026-03-31")
	res2 := CalculateTHR(8_000_000, 1_000_000, join2, period2)
	if res2.FullTHR || res2.Amount != 3_750_000 {
		t.Fatalf("want proporsional 3750000 got %+v", res2)
	}
}

// TestAnnual memverifikasi PPh 21 tahunan untuk PKP 57,6 juta.
func TestAnnual(t *testing.T) {
	res := AnnualPPh21(120_000_000, 3_600_000, "TK0")
	// neto = 120jt - 6jt (jabatan max) - 2.4jt... 3.6jt pension > cap 2.4jt
	// neto = 120 - 6 - 2.4 = 111.6jt ; PTKP 54jt ; PKP = 57.6jt -> 5% x 57.6jt = 2.88jt
	if res.AnnualPPh21 != 2_880_000 {
		t.Fatalf("want 2880000 got %v", res.AnnualPPh21)
	}
}

// TestTERCategory memverifikasi pemetaan status PTKP ke kategori TER.
func TestTERCategory(t *testing.T) {
	cases := map[string]string{
		"TK0": "A", "TK1": "A", "K0": "A",
		"TK2": "B", "TK3": "B", "K1": "B", "K2": "B",
		"K3": "C",
	}
	for ptkp, want := range cases {
		got, err := TERCategory(ptkp)
		if err != nil {
			t.Fatalf("%s: %v", ptkp, err)
		}
		if got != want {
			t.Fatalf("%s: want %s got %s", ptkp, want, got)
		}
	}
	if _, err := TERCategory("XX9"); err == nil {
		t.Fatal("want error untuk status PTKP tidak valid")
	}
}

// TestPTKPAmount memverifikasi nilai PTKP seluruh status.
func TestPTKPAmount(t *testing.T) {
	cases := map[string]float64{
		"TK0": 54_000_000,
		"TK1": 58_500_000,
		"TK2": 63_000_000,
		"TK3": 67_500_000,
		"K0":  58_500_000,
		"K1":  63_000_000,
		"K2":  67_500_000,
		"K3":  72_000_000,
	}
	for ptkp, want := range cases {
		if got := PTKPAmount(ptkp); got != want {
			t.Fatalf("%s: want %v got %v", ptkp, want, got)
		}
	}
}

// TestAnnualBrackets memverifikasi tarif progresif Pasal 17.
func TestAnnualBrackets(t *testing.T) {
	// TK0, neto bersih 200jt -> PKP 146jt: 5%x60jt + 15%x86jt = 3jt + 12.9jt = 15.9jt
	res := AnnualPPh21(260_000_000, 0, "TK0")
	// neto = 260 - 6 - 0 = 254jt; PKP = 254 - 54 = 200jt
	// 5%x60 + 15%x140 = 3 + 21 = 24jt
	if res.AnnualPPh21 != 24_000_000 {
		t.Fatalf("want 24000000 got %v", res.AnnualPPh21)
	}
}

// TestTHRUnder12MonthsEdge memverifikasi masa kerja 11 bulan penuh.
func TestTHRUnder12MonthsEdge(t *testing.T) {
	join, _ := time.Parse("2006-01-02", "2025-03-01")
	period, _ := time.Parse("2006-01-02", "2026-02-28")
	res := CalculateTHR(8_000_000, 1_000_000, join, period)
	// 11 bulan penuh
	if res.MonthsOfService != 11 || res.FullTHR {
		t.Fatalf("unexpected: %+v", res)
	}
	if res.Amount != 8_250_000 {
		t.Fatalf("want 8250000 got %v", res.Amount)
	}
}

// TestOvertimeHoliday memverifikasi lembur hari libur dan batas 10 jam.
func TestOvertimeHoliday(t *testing.T) {
	r := 10_000_000.0 / 173
	// 10 jam hari libur: 7x2 + 1x3 + 2x4 = 25x rate
	want := float64(int64(r * 25))
	if got := OvertimePay(10_000_000, 0, 10, true); got != want {
		t.Fatalf("want %v got %v", want, got)
	}
	// lebih dari 10 jam dikenakan maksimal 10 jam (kebijakan aplikasi)
	if got := OvertimePay(10_000_000, 0, 12, true); got != want {
		t.Fatalf("cap want %v got %v", want, got)
	}
}

// TestComputePayslip memverifikasi slip lengkap: bruto, potongan, PPh 21, net.
func TestComputePayslip(t *testing.T) {
	in := EmployeePayrollInput{
		BaseSalary: 10_000_000, FixedAllowance: 500_000, NonFixedAllowance: 300_000,
		OvertimePay: 200_000, PTKPStatus: "TK0", BPJSHealth: true, BPJSTK: true,
		Rates: BPJSRates{
			HealthCap: 12_000_000, HealthWorker: 0.01, HealthEmployer: 0.04,
			JPCap: 10_042_300, JPWorker: 0.01, JPEmployer: 0.02,
			JHTWorker: 0.02, JHTEmployer: 0.037, JKMEmployer: 0.003, JKKRisk: 0.54,
		},
	}
	slip, err := ComputePayslip(in)
	if err != nil {
		t.Fatal(err)
	}
	gross := 10_000_000 + 500_000 + 300_000 + 200_000
	if slip.GrossIncome != float64(gross) {
		t.Fatalf("gross want %v got %v", gross, slip.GrossIncome)
	}
	// potongan (basis upah = gaji pokok + tunjangan tetap = 10.5jt):
	// BPJS-Kes 1% x 10.5jt = 105000, JHT 2% x 10.5jt = 210000, JP 1% x cap 10.042.300 = 100423
	wantDeduction := 105_000 + 210_000 + 100_423
	if slip.EmployeeDeduction != float64(wantDeduction) {
		t.Fatalf("deduction want %v got %v", wantDeduction, slip.EmployeeDeduction)
	}
	// TER A: 11jt masuk bracket 10.70-11.05 = 3% -> 330000
	if slip.PPh21 != 330_000 {
		t.Fatalf("pph21 want 330000 got %v", slip.PPh21)
	}
	if slip.NetSalary != float64(gross)-float64(wantDeduction)-330_000 {
		t.Fatalf("net got %v", slip.NetSalary)
	}
}

package payroll

// Orchestrator perhitungan slip gaji bulanan.

type EmployeePayrollInput struct {
	BaseSalary        float64
	FixedAllowance    float64
	NonFixedAllowance float64
	OvertimePay       float64
	Bonus             float64
	THR               float64
	PTKPStatus        string
	BPJSHealth        bool
	BPJSTK            bool
	JKKRisk           float64
	Rates             BPJSRates
}

type DetailItem struct {
	Category string  `json:"category"` // earning | deduction | tax | employer
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
}

type Payslip struct {
	BasicSalary          float64      `json:"basic_salary"`
	FixedAllowance       float64      `json:"fixed_allowance"`
	NonFixedAllowance    float64      `json:"non_fixed_allowance"`
	OvertimePay          float64      `json:"overtime_pay"`
	Bonus                float64      `json:"bonus"`
	THR                  float64      `json:"thr"`
	GrossIncome          float64      `json:"gross_income"`
	BPJS                 BPJSResult   `json:"bpjs"`
	EmployeeDeduction    float64      `json:"employee_deduction"`
	PPh21                float64      `json:"pph21"`
	NetSalary            float64      `json:"net_salary"`
	EmployerContribution float64      `json:"employer_contribution"`
	TERCategory          string       `json:"ter_category"`
	TERRate              float64      `json:"ter_rate"`
	Details              []DetailItem `json:"details"`
}

func ComputePayslip(in EmployeePayrollInput) (*Payslip, error) {
	// Basis upah BPJS = gaji pokok + tunjangan tetap (PP 46/2015)
	bpjs := CalculateBPJS(in.Rates, in.BaseSalary+in.FixedAllowance, in.BPJSHealth, in.BPJSTK)
	gross := in.BaseSalary + in.FixedAllowance + in.NonFixedAllowance + in.OvertimePay + in.Bonus + in.THR

	terCategory, terRate, err := MonthlyPPh21Rate(in.PTKPStatus, gross)
	if err != nil {
		return nil, err
	}
	pph21, err := MonthlyPPh21(in.PTKPStatus, gross)
	if err != nil {
		return nil, err
	}

	deduction := bpjs.WorkerTotal
	net := gross - deduction - pph21

	slip := &Payslip{
		BasicSalary:          in.BaseSalary,
		FixedAllowance:       in.FixedAllowance,
		NonFixedAllowance:    in.NonFixedAllowance,
		OvertimePay:          in.OvertimePay,
		Bonus:                in.Bonus,
		THR:                  in.THR,
		GrossIncome:          gross,
		BPJS:                 bpjs,
		EmployeeDeduction:    deduction,
		PPh21:                pph21,
		NetSalary:            net,
		EmployerContribution: bpjs.EmployerTotal,
		TERCategory:          terCategory,
		TERRate:              terRate,
	}

	slip.Details = append(slip.Details,
		DetailItem{"earning", "Gaji Pokok", in.BaseSalary},
		DetailItem{"earning", "Tunjangan Tetap", in.FixedAllowance},
		DetailItem{"earning", "Tunjangan Tidak Tetap", in.NonFixedAllowance},
		DetailItem{"earning", "Upah Lembur", in.OvertimePay},
		DetailItem{"earning", "Bonus", in.Bonus},
		DetailItem{"earning", "THR", in.THR},
		DetailItem{"deduction", "BPJS Kesehatan (pekerja)", bpjs.HealthWorker},
		DetailItem{"deduction", "BPJS TK - JHT (pekerja)", bpjs.JHTWorker},
		DetailItem{"deduction", "BPJS TK - JP (pekerja)", bpjs.JPWorker},
		DetailItem{"tax", "PPh 21 (TER " + terCategory + ")", pph21},
		DetailItem{"employer", "BPJS Kesehatan (perusahaan)", bpjs.HealthEmploy},
		DetailItem{"employer", "BPJS TK - JHT (perusahaan)", bpjs.JHTEmployer},
		DetailItem{"employer", "BPJS TK - JKK (perusahaan)", bpjs.JKKEmployer},
		DetailItem{"employer", "BPJS TK - JKM (perusahaan)", bpjs.JKMEmployer},
		DetailItem{"employer", "BPJS TK - JP (perusahaan)", bpjs.JPEmployer},
	)
	return slip, nil
}

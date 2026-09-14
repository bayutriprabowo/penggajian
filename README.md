# Aplikasi Penggajian (Payroll) Indonesia — Go Backend

Backend penggajian lengkap berbasis **Golang** dengan input/output **JSON**,
database **PostgreSQL** via **GORM**, autentikasi **JWT + RBAC**, arsitektur
berbasis **interface** (repository/service) dan **DTO**.

Perhitungan mengikuti regulasi Indonesia:

| Komponen | Dasar Hukum |
|---|---|
| PPh 21 bulanan metode **TER** (A/B/C) | PP 58/2023, PMK 168/2023 |
| Upah lembur (1/173) | PP 35/2021 |
| BPJS Kesehatan (5%, cap Rp 12 jt) | PP 28/2024 |
| BPJS TK: JHT 5,7%, JKK 0,24–1,74%, JKM 0,3%, JP 3% | PP 46/2015 |
| THR penuh & proporsional + pajak THR | Permenaker 6/2016 |
| Rekap PPh 21 tahunan (Pasal 17, PTKP) | UU HPP |

---

## Daftar Isi

1. [Fungsi Kode (Arsitektur)](#1-fungsi-kode-arsitektur)
2. [Alur Pemakaian](#2-alur-pemakaian)
3. [Autentikasi & RBAC](#3-autentikasi--rbac)
4. [Referensi API](#4-referensi-api)
5. [Detail Perhitungan](#5-detail-perhitungan)
6. [Konfigurasi](#6-konfigurasi)
7. [Test & Kualitas](#7-test--kualitas)
8. [Catatan Produksi](#8-catatan-produksi)

---

# 1. Fungsi Kode (Arsitektur)

## 1.1 Alur Request

```
HTTP request
   └─> router        : memilih handler + memasang middleware (auth, RBAC)
        └─> handler  : baca/validasi JSON (DTO) -> panggil service -> tulis JSON
             └─> service      : business logic (validasi, orkestrasi perhitungan)
                  └─> repository : akses database via GORM
                       └─> PostgreSQL
```

Lapisan **service** dan **repository** dideklarasikan sebagai **interface**
(`internal/service/interfaces.go`, `internal/repository/interfaces.go`)
sehingga mudah diganti atau di-mock saat test.

## 1.2 Struktur & Fungsi Tiap Package

### `cmd/server/main.go` — entrypoint

| Fungsi | Penjelasan |
|---|---|
| `main()` | Memuat konfigurasi, konek database, jalankan `AutoMigrate`, seed RBAC/demo, lalu HTTP server dengan graceful shutdown |
| `newRepos()` | Membuat seluruh implementasi repository dengan satu koneksi `*gorm.DB` |
| `newServices()` | Menyusun service dengan dependensi repository + konfigurasi |

### `internal/config` — konfigurasi environment

| Fungsi | Penjelasan |
|---|---|
| `Load()` | Membaca semua env (PORT, DATABASE_URL, JWT, tarif BPJS, seeder) menjadi satu struct `Config` |
| `getEnv / getEnvInt / getEnvFloat / getEnvBool` | Helper baca env dengan nilai default |

### `internal/model` — model database (GORM)

Entitas: `User`, `Role`, `Permission`, `Employee`, `Allowance`, `Overtime`, `Payroll`.
`Payroll` memiliki unique index `(employee_id, period)` dan detail slip disimpan
sebagai **JSONB** agar fleksibel.

### `internal/dto` — Data Transfer Object

Struktur request/response API agar field internal (mis. `Password`) tidak bocor ke JSON.

| Fungsi | Penjelasan |
|---|---|
| `Success(message, data)` / `Error(message)` | Membangun envelope respons `{"status","message","data"}` |

### `internal/repository` — akses database (interface + GORM)

| File | Fungsi utama |
|---|---|
| `user_repo.go` | `Create`, `FindByID`, `FindByUsername`, `FindByEmail`, `List` |
| `role_repo.go` | `CreateRole`, `FindRoleByName`, `AssignPermissions`, `PermissionsByRole` |
| `employee_repo.go` | `Create`, `FindByID`, `FindByNIK`, `List`, `ListActive`, `UpdateWithAllowances` (satu transaksi), `Delete` |
| `overtime_repo.go` | `Create`, `SumByEmployeePeriod` (total lembur 1 periode), `ListByEmployee` |
| `payroll_repo.go` | `Create`, `FindByID`, `FindByEmployeePeriod`, `ListByEmployee`, `ListByPeriod`, `UpdateStatus`, `SumPPh21ByEmployeeYear` |

### `internal/payroll` — engine perhitungan (pure function, tanpa DB)

| Fungsi | Penjelasan |
|---|---|
| `ComputePayslip()` | Orkestrator slip gaji: bruto, potongan BPJS, PPh 21 TER, net salary, detail per komponen |
| `MonthlyPPh21()` | PPh 21 bulanan = tarif TER × bruto, dibulatkan ke bawah |
| `MonthlyPPh21Rate()` | Mencari kategori & tarif TER (A/B/C) untuk status PTKP |
| `PPh21OnTHR()` | Pajak THR = TER(bruto+THR) − TER(bruto) |
| `TERTables()` / `TERCategory()` | Tabel TER lengkap & pemetaan status PTKP |
| `CalculateBPJS()` | Iuran BPJS Kesehatan + Ketenagakerjaan (JHT/JKK/JKM/JP), pekerja & perusahaan |
| `OvertimePay()` | Upah lembur 1/173, hari kerja & hari libur (PP 35/2021) |
| `CalculateTHR()` | THR penuh (≥12 bulan) / proporsional (<12 bulan) |
| `PTKPAmount()` / `AnnualPPh21()` | PTKP tahunan & rekap PPh 21 setahun (Pasal 17) |

### `internal/service` — business logic (interface + implementasi)

| Service | Fungsi utama |
|---|---|
| `AuthService` | `Register` (bcrypt + anti-duplikat), `Login` (JWT), `ListUsers`, `GenerateToken`, `ParseToken` |
| `EmployeeService` | `Create/GetByID/List/Update/Delete` karyawan + validasi (email, PTKP, JKK) |
| `OvertimeService` | `Create` (hitung upah lembur otomatis), `ListByEmployee` (filter periode) |
| `PayrollService` | `Run` (generate slip 1/semua karyawan, pre-check anti duplikat), `GetByID`, `ListByEmployee`, `ListByPeriod`, `Approve`, `MarkPaid`, `CalculateTHR`, `buildPayslip` |
| `TaxService` | `TERTables`, `TERInfo` (simulasi), `AnnualRecap` (rekap tahunan) |
| `SeedRBAC()` | Seeder permission/role/admin default — selalu jalan saat startup |
| `SeedDemo()` | Seeder data contoh (idempotent) — aktif bila `SEED_DEMO=true` |

### `internal/handler` — HTTP handler JSON

| Handler | Fungsi utama |
|---|---|
| `AuthHandler` | `Register`, `Login` |
| `UserHandler` | `Me` (profil login), `List` (daftar user) |
| `EmployeeHandler` | CRUD karyawan |
| `OvertimeHandler` | Input & riwayat lembur |
| `PayrollHandler` | `Run`, slip, approve/paid, kalkulasi THR |
| `TaxHandler` | Tabel TER, simulasi, rekap tahunan |
| `HealthHandler` | Health check + ping database |

### `internal/middleware` — keamanan

| Fungsi | Penjelasan |
|---|---|
| `Auth()` | Wajib token valid; memuat user + permission ke context |
| `OptionalAuth()` | Token opsional (endpoint publik), token rusak dianggap anonim |
| `RequirePermission()` / `RequireAnyPermission()` | Gerbang RBAC per endpoint |
| `HasPermission()`, `UserFromContext()` | Helper untuk handler |

### `internal/router` — tabel route

`New()` mendaftarkan seluruh endpoint + middleware, `recoverer()` menangkap panic,
`logging()` mencatat setiap request.

### `internal/dberr` — pemetaan error PostgreSQL

| Fungsi | Penjelasan |
|---|---|
| `IsUniqueViolation()` | SQLSTATE 23505 → HTTP 409 |
| `IsForeignKeyViolation()` | SQLSTATE 23503/23001 → 400/409 |

---

# 2. Alur Pemakaian

## 2.1 Prasyarat

- Go 1.22+
- PostgreSQL lokal (versi 14+)
- `make` (opsional)

## 2.2 Siapkan PostgreSQL (sekali saja)

```bash
psql -d postgres -c "CREATE ROLE payroll LOGIN PASSWORD 'payroll';"
createdb -O payroll payroll

# atau dengan make
make db-create
```

> Default koneksi: `host=localhost user=payroll password=payroll dbname=payroll port=5432 sslmode=disable`
> (ubah via env `DATABASE_URL` bila berbeda).

## 2.3 Jalankan Server

```bash
go run ./cmd/server      # atau: make run
```

Saat startup otomatis: **AutoMigrate** (buat tabel), **seed RBAC** (4 role + 10 permission),
**user admin default** `admin` / `admin123`.

Cek kesehatan:

```bash
curl http://localhost:8080/health
# {"status":"success","message":"ok","data":{"database":"ok"}}
```

## 2.4 Seeder Data Demo (opsional)

Seed sekali jalan tanpa menjalankan server:

```bash
make seed    # = SEED_DEMO=true SEED_ONLY=true go run ./cmd/server
```

Membuat: 4 karyawan + user (`budi`, `siti`, `andi`, `dewi` — password `rahasia123`),
user `finance1`, 4 lembur, dan payroll bulan lalu (draft). Idempotent — aman dijalankan ulang.

## 2.5 Login & Simpan Token

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# respons: {"status":"success","message":"login berhasil","data":{"token":"eyJ...","user":{...}}}
```

Simpan token untuk request berikutnya:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .data.token)
```

## 2.6 Skenario HR: Karyawan → Lembur → Payroll

**1) Tambah karyawan:**

```bash
curl -s -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{
    "nik": "EMP-001", "full_name": "Budi Santoso",
    "email": "budi@perusahaan.com", "position": "Backend Engineer",
    "department": "IT", "join_date": "2025-03-01", "status": "active",
    "ptkp_status": "K0", "base_salary": 10000000,
    "bpjs_health": true, "bpjs_tk": true, "jkk_risk": 0.54,
    "allowances": [
      {"name": "Tunjangan Jabatan", "type": "fixed", "amount": 500000},
      {"name": "Uang Makan", "type": "non_fixed", "amount": 300000}
    ]
  }'
```

`ptkp_status`: `TK0 TK1 TK2 TK3 K0 K1 K2 K3` — menentukan kategori TER
(A = TK0/TK1/K0, B = TK2/TK3/K1/K2, C = K3).

**2) Input lembur** (upah dihitung otomatis PP 35/2021):

```bash
curl -s -X POST http://localhost:8080/api/v1/overtimes \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"employee_id":1,"date":"2026-03-05","hours":2.5,"is_holiday":false}'
```

**3) Generate payroll periode:**

```bash
curl -s -X POST http://localhost:8080/api/v1/payrolls/run \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"period":"2026-03","employee_id":0,"bonus":0,"thr":-1}'
```

- `employee_id: 0` → semua karyawan aktif; `>0` → satu karyawan
- `bonus` → bonus bulan berjalan (kena PPh 21 TER)
- `thr` → `0` tanpa THR, `-1` hitung otomatis dari masa kerja, `>0` nilai manual

Contoh respons slip (ringkas):

```json
{
  "status": "success",
  "message": "payroll berhasil di-generate",
  "data": [{
    "id": 1, "employee_name": "Budi Santoso", "period": "2026-03",
    "basic_salary": 10000000, "allowances": 800000, "overtime_pay": 202312,
    "gross_income": 11002312, "employee_deduction": 424423,
    "pph21": 330069, "net_salary": 10247820, "status": "draft",
    "details": [
      {"category":"earning","name":"Gaji Pokok","amount":10000000},
      {"category":"deduction","name":"BPJS Kesehatan (pekerja)","amount":108000},
      {"category":"tax","name":"PPh 21 (TER A)","amount":330069}
    ]
  }]
}
```

## 2.7 Skenario Finance: Approve → Paid

```bash
# login finance1 (dari seeder demo)
curl -s -X PATCH http://localhost:8080/api/v1/payrolls/1/approve \
  -H "Authorization: Bearer $FTOKEN"

curl -s -X PATCH http://localhost:8080/api/v1/payrolls/1/paid \
  -H "Authorization: Bearer $FTOKEN"
```

Alur status: `draft` → `approved` → `paid`.

## 2.8 Skenario Karyawan: Lihat Slip Sendiri

```bash
# login budi, lalu lihat slip miliknya
curl -s http://localhost:8080/api/v1/payrolls/employee/1 \
  -H "Authorization: Bearer $BTOKEN"
```

Role `employee` otomatis dilarang melihat slip karyawan lain atau mengubah data.

## 2.9 THR & Pajak

```bash
# kalkulasi THR (penuh/proporsional + pajak THR)
curl -s "http://localhost:8080/api/v1/thr/calculate?employee_id=1&period=2026-03" \
  -H "Authorization: Bearer $TOKEN"

# tabel TER lengkap
curl -s http://localhost:8080/api/v1/tax/ter -H "Authorization: Bearer $TOKEN"

# simulasi PPh 21 TER
curl -s "http://localhost:8080/api/v1/tax/ter-info?ptkp=K0&gross=15000000" \
  -H "Authorization: Bearer $TOKEN"

# rekap PPh 21 tahunan (Pasal 17)
curl -s "http://localhost:8080/api/v1/tax/annual-recap?employee_id=1&year=2026&annual_gross=150000000&jht_worker=3600000" \
  -H "Authorization: Bearer $TOKEN"
```

---

# 3. Autentikasi & RBAC

Semua endpoint (kecuali register/login/health) butuh header `Authorization: Bearer <token>`.

| Role | Hak |
|---|---|
| admin | semua akses, kelola user & role |
| hr | kelola karyawan, input lembur, jalankan payroll |
| finance | lihat & approve/mark-paid payroll, akses pajak |
| employee | profil sendiri + slip gaji milik sendiri |

**Register:**
- Publik (tanpa token) → selalu role `employee` (anti eskalasi role).
- Dengan token `user:write` (admin) → boleh memilih `role_name` & `employee_id`.

```bash
# admin membuat user HR
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"budi.hr","email":"budi@perusahaan.com","password":"rahasia123","role_name":"hr"}'
```

---

# 4. Referensi API

| Method | Path | Permission | Keterangan |
|---|---|---|---|
| POST | /api/v1/auth/register | publik / user:write | daftar user |
| POST | /api/v1/auth/login | publik | login → JWT |
| GET | /api/v1/users/me | user:read | profil user login |
| GET | /api/v1/users | user:list | daftar user (paginasi) |
| POST | /api/v1/employees | employee:write | tambah karyawan |
| GET | /api/v1/employees | employee:read | list karyawan (?page,&limit,&status) |
| GET | /api/v1/employees/{id} | employee:read | detail karyawan |
| PUT | /api/v1/employees/{id} | employee:write | ubah karyawan |
| DELETE | /api/v1/employees/{id} | employee:write | hapus (ditolak bila punya payroll) |
| POST | /api/v1/overtimes | overtime:write | input lembur |
| GET | /api/v1/overtimes/employee/{id} | employee:read | riwayat lembur (?period=) |
| POST | /api/v1/payrolls/run | payroll:run | generate payroll |
| GET | /api/v1/payrolls | payroll:read | list per periode (?period=) |
| GET | /api/v1/payrolls/{id} | payroll:read | slip gaji |
| GET | /api/v1/payrolls/employee/{id} | payroll:read | riwayat slip |
| PATCH | /api/v1/payrolls/{id}/approve | payroll:approve | draft → approved |
| PATCH | /api/v1/payrolls/{id}/paid | payroll:approve | approved → paid |
| GET | /api/v1/thr/calculate | payroll:read | kalkulasi THR |
| GET | /api/v1/tax/ter | tax:read | tabel TER A/B/C |
| GET | /api/v1/tax/ter-info | tax:read | simulasi TER |
| GET | /api/v1/tax/annual-recap | tax:read | rekap PPh 21 tahunan |
| GET | /health | publik | status server + database |

Envelope sukses: `{"status":"success","message":"...","data":{...}}`
Envelope error: `{"status":"error","message":"..."}`

---

# 5. Detail Perhitungan

## 5.1 Lembur (PP 35/2021)
- Upah sejam = `1/173 × (gaji pokok + tunjangan tetap)`
- Hari kerja: jam ke-1 = 1,5×, jam berikutnya = 2×
- Hari libur (6 hari kerja/minggu): 7 jam pertama 2×, jam ke-8 = 3×, jam 9–10 = 4×

## 5.2 BPJS Kesehatan (PP 28/2024)
- Total 5% upah: 4% perusahaan + 1% pekerja, cap upah Rp 12.000.000/bulan

## 5.3 BPJS Ketenagakerjaan (PP 46/2015)
| Program | Pekerja | Perusahaan | Cap |
|---|---|---|---|
| JHT | 2% | 3,7% | — |
| JKK | — | 0,24%–1,74% (kelas risiko) | — |
| JKM | — | 0,3% | — |
| JP | 1% | 2% | upah maks Rp 10.042.300 |

Basis upah BPJS = gaji pokok + tunjangan tetap.

## 5.4 PPh 21 Bulanan — TER (PP 58/2023)
- Kategori A: TK/0, TK/1, K/0 — B: TK/2, TK/3, K/1, K/2 — C: K/3
- `PPh21 = tarif TER × bruto bulanan`, dibulatkan ke bawah
- Tabel lengkap (38 baris/kategori) di `internal/payroll/ter.go`

## 5.5 THR (Permenaker 6/2016)
- Masa kerja ≥ 12 bulan: `1 × (gaji pokok + tunjangan tetap)`
- Masa kerja < 12 bulan: `(bulan/12) × (gaji pokok + tunjangan tetap)`
- Pajak THR = `TER(bruto+THR) − TER(bruto)`

## 5.6 Rekap Tahunan (Pasal 17 UU HPP)
- Neto = bruto − biaya jabatan (5%, maks Rp 6 jt) − iuran pensiun pekerja (maks Rp 2,4 jt)
- PKP = neto − PTKP (TK/0 = Rp 54 jt, +Rp 4,5 jt per tanggungan/kawin)
- Tarif progresif: 5% / 15% / 25% / 30% / 35%

---

# 6. Konfigurasi

Salin `.env.example` → export sebelum menjalankan:

| Env | Default | Keterangan |
|---|---|---|
| PORT | 8080 | port server |
| DATABASE_URL | `host=localhost user=payroll password=payroll dbname=payroll port=5432 sslmode=disable` | koneksi PostgreSQL |
| JWT_SECRET | change-me-in-production | secret JWT (ganti untuk produksi!) |
| JWT_EXPIRE_HOURS | 24 | masa berlaku token |
| BPJS_HEALTH_CAP | 12000000 | cap upah BPJS Kesehatan |
| BPJS_HEALTH_WORKER / _EMPLOYER | 0.01 / 0.04 | iuran kesehatan |
| JP_CAP | 10042300 | cap upah JP |
| JP_WORKER / _EMPLOYER | 0.01 / 0.02 | iuran JP |
| JHT_WORKER / _EMPLOYER | 0.02 / 0.037 | iuran JHT |
| JKM_EMPLOYER | 0.003 | iuran JKM |
| DEFAULT_ADMIN_USER / _PASS | admin / admin123 | admin default |
| SEED_DEMO | false | aktifkan seeder demo |
| DEMO_USER_PASS | rahasia123 | password user demo |
| SEED_ONLY | false | migrasi + seed lalu keluar |

---

# 7. Test & Kualitas

```bash
make test   # go test ./...  (17 test: engine perhitungan + service dengan fake repo)
make vet    # go vet ./...
make build  # binary di bin/penggajian
```

---

# 8. Catatan Produksi

- Ganti `JWT_SECRET` dan password admin default sebelum deploy.
- Nilai cap BPJS (kesehatan, JP) dapat disesuaikan via env bila regulasi berubah.
- Slip gaji disimpan dengan detail JSONB sehingga mudah dikembangkan (misal: komponen baru).
- Hapus karyawan yang masih punya riwayat payroll otomatis ditolak (riwayat aman).
- Payroll tidak bisa diduplikasi untuk (karyawan, periode) yang sama — termasuk saat request paralel.

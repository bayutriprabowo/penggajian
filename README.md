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

## 1. Prasyarat

- Go 1.22+
- PostgreSQL lokal (versi 14+)
- `make` (opsional)

## 2. Menyiapkan PostgreSQL

Buat role dan database sekali saja (via `psql`/`createdb`):

```bash
# opsional: lewati bila role/database sudah ada
psql -d postgres -c "CREATE ROLE payroll LOGIN PASSWORD 'payroll';"
createdb -O payroll payroll

# atau dengan make
make db-create
```

> Default koneksi: `host=localhost user=payroll password=payroll dbname=payroll port=5432 sslmode=disable`
> (ubah lewat env `DATABASE_URL` bila PostgreSQL Anda memakai konfigurasi lain).

## 3. Menjalankan Aplikasi

```bash
# migrasi + seed RBAC otomatis saat startup
go run ./cmd/server

# atau dengan make
make run
```

Server berjalan di `http://localhost:8080`. Cek kesehatan (termasuk status database):

```bash
curl http://localhost:8080/health
# {"status":"success","message":"ok","data":{"database":"ok"}}
```

Saat startup, sistem otomatis:
- menjalankan migrasi tabel (AutoMigrate)
- membuat permission, role (`admin`, `hr`, `finance`, `employee`), dan
  **user admin default: `admin` / `admin123`** (ubah lewat env `DEFAULT_ADMIN_USER` / `DEFAULT_ADMIN_PASS`)
- (opsional) membuat data demo bila `SEED_DEMO=true`

Server mendukung graceful shutdown (Ctrl+C / SIGTERM). Konfigurasi lain dapat
disalin dari `.env.example` (export manual atau `set -a; source .env; set +a`).

## 3a. Seeder Data Demo (opsional)

Aktifkan dengan env `SEED_DEMO=true` (sekali jalan, idempotent — dilewati bila
data karyawan sudah ada). Membuat:

- 4 karyawan: Budi (K0, 10 jt), Siti (TK0, 8 jt), Andi (K2, 15 jt), Dewi (TK1, 7 jt)
- User login untuk masing-masing: `budi`, `siti`, `andi`, `dewi` (password `DEMO_USER_PASS`, default `rahasia123`)
- User finance: `finance1` (untuk mencoba alur approve)
- 4 catatan lembur + payroll periode bulan lalu (status `draft`)

```bash
SEED_DEMO=true go run ./cmd/server
```

## 4. Autentikasi & RBAC

Semua endpoint (kecuali register/login/health) butuh header:

```
Authorization: Bearer <token>
```

| Role | Hak |
|---|---|
| admin | semua akses, kelola user & role |
| hr | kelola karyawan, input lembur, jalankan payroll |
| finance | lihat & approve/mark-paid payroll, akses pajak |
| employee | profil sendiri + lihat slip gaji milik sendiri |

### Login

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Respons:

```json
{
  "status": "success",
  "message": "login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { "id": 1, "username": "admin", "email": "admin@penggajian.local", "role_id": 1, "role_name": "admin" }
  }
}
```

Simpan token: `TOKEN=$(curl -s ... | jq -r .data.token)`

### Register user

- **Publik (tanpa token):** selalu dibuat dengan role `employee` — aman dari
  eskalasi role.
- **Dengan token user yang punya `user:write` (admin):** boleh memilih
  `role_name` apa pun dan menautkan `employee_id`.

```bash
# admin membuat user HR
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"budi.hr","email":"budi@perusahaan.com","password":"rahasia123","role_name":"hr"}'

# publik mendaftar sebagai karyawan
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"budi.karyawan","email":"budi2@perusahaan.com","password":"rahasia123"}'
```

### Profil & daftar user

```bash
curl -s http://localhost:8080/api/v1/users/me -H "Authorization: Bearer $TOKEN"
curl -s "http://localhost:8080/api/v1/users?page=1&limit=10" -H "Authorization: Bearer $TOKEN"   # admin/hr
```

## 5. Data Master Karyawan

### Tambah karyawan

```bash
curl -s -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "nik": "EMP-001",
    "full_name": "Budi Santoso",
    "email": "budi@perusahaan.com",
    "position": "Backend Engineer",
    "department": "IT",
    "join_date": "2025-03-01",
    "status": "active",
    "ptkp_status": "K0",
    "base_salary": 10000000,
    "bpjs_health": true,
    "bpjs_tk": true,
    "jkk_risk": 0.54,
    "allowances": [
      {"name": "Tunjangan Jabatan", "type": "fixed", "amount": 500000},
      {"name": "Uang Makan", "type": "non_fixed", "amount": 300000}
    ]
  }'
```

Keterangan `ptkp_status`: `TK0 TK1 TK2 TK3 K0 K1 K2 K3`
(menentukan kategori TER: A = TK0/TK1/K0, B = TK2/TK3/K1/K2, C = K3).

### List / detail / ubah / hapus

```bash
curl -s "http://localhost:8080/api/v1/employees?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"

curl -s http://localhost:8080/api/v1/employees/1 \
  -H "Authorization: Bearer $TOKEN"

curl -s -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"position":"Senior Backend Engineer","base_salary":12000000}'

curl -s -X DELETE http://localhost:8080/api/v1/employees/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 6. Lembur (PP 35/2021)

Upah sejam = `1/173 × (gaji pokok + tunjangan tetap)`.
Hari kerja: jam 1 = 1,5×, jam berikutnya 2×. Hari libur: 7 jam pertama 2×, jam 8 = 3×, jam 9–10 = 4×.

```bash
curl -s -X POST http://localhost:8080/api/v1/overtimes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"employee_id":1,"date":"2026-03-05","hours":2.5,"is_holiday":false}'

# semua riwayat, atau filter periode tertentu (YYYY-MM)
curl -s http://localhost:8080/api/v1/overtimes/employee/1 \
  -H "Authorization: Bearer $TOKEN"
curl -s "http://localhost:8080/api/v1/overtimes/employee/1?period=2026-03" \
  -H "Authorization: Bearer $TOKEN"
```

## 7. Generate Payroll (Slip Gaji)

```bash
curl -s -X POST http://localhost:8080/api/v1/payrolls/run \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"period":"2026-03","employee_id":0,"bonus":0,"thr":-1}'
```

- `employee_id: 0` → semua karyawan aktif
- `bonus` → nilai bonus bulan berjalan (kena PPh 21 TER)
- `thr` → `0` tidak dihitung, `-1` hitung otomatis berdasarkan masa kerja, `>0` nilai manual

Contoh respons:

```json
{
  "status": "success",
  "message": "payroll berhasil di-generate",
  "data": [
    {
      "id": 1,
      "employee_id": 1,
      "employee_name": "Budi Santoso",
      "nik": "EMP-001",
      "period": "2026-03",
      "basic_salary": 10000000,
      "allowances": 800000,
      "overtime_pay": 202312,
      "bonus": 0,
      "thr": 0,
      "gross_income": 11002312,
      "employee_deduction": 424423,
      "pph21": 330069,
      "net_salary": 10247820,
      "employer_contribution": 1030454,
      "details": [
        {"category":"earning","name":"Gaji Pokok","amount":10000000},
        {"category":"earning","name":"Tunjangan Tetap","amount":500000},
        {"category":"deduction","name":"BPJS Kesehatan (pekerja)","amount":108000},
        {"category":"tax","name":"PPh 21 (TER A)","amount":330069}
      ],
      "status": "draft"
    }
  ]
}
```

### Lihat, approve & tandai lunas

```bash
# slip per id
curl -s http://localhost:8080/api/v1/payrolls/1 \
  -H "Authorization: Bearer $TOKEN"

# riwayat slip per karyawan
curl -s http://localhost:8080/api/v1/payrolls/employee/1 \
  -H "Authorization: Bearer $TOKEN"

# semua slip pada satu periode (admin/hr/finance, paginasi)
curl -s "http://localhost:8080/api/v1/payrolls?period=2026-03&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"

# alur status: draft -> approve (finance) -> paid (finance)
curl -s -X PATCH http://localhost:8080/api/v1/payrolls/1/approve \
  -H "Authorization: Bearer $TOKEN"
curl -s -X PATCH http://localhost:8080/api/v1/payrolls/1/paid \
  -H "Authorization: Bearer $TOKEN"
```

## 8. THR (Permenaker 6/2016)

Masa kerja ≥ 12 bulan → 1× (gaji pokok + tunjangan tetap); di bawahnya proporsional `(bulan/12) × gaji`.
Pajak THR = `TER(bruto+THR) − TER(bruto)`.

```bash
curl -s "http://localhost:8080/api/v1/thr/calculate?employee_id=1&period=2026-03" \
  -H "Authorization: Bearer $TOKEN"
```

```json
{
  "status": "success",
  "message": "kalkulasi THR",
  "data": {
    "employee_id": 1,
    "employee_name": "Budi Santoso",
    "months_of_service": 12,
    "full_thr": true,
    "base_salary": 10000000,
    "fixed_allowance": 500000,
    "thr_amount": 10500000,
    "pph21_on_thr": 1912500,
    "net_thr": 8587500
  }
}
```

## 9. Pajak (PPh 21)

### Tabel TER lengkap

```bash
curl -s http://localhost:8080/api/v1/tax/ter -H "Authorization: Bearer $TOKEN"
```

### Simulasi TER

```bash
curl -s "http://localhost:8080/api/v1/tax/ter-info?ptkp=K0&gross=15000000" \
  -H "Authorization: Bearer $TOKEN"
```

### Rekap PPh 21 tahunan (Pasal 17 UU HPP)

```bash
curl -s "http://localhost:8080/api/v1/tax/annual-recap?employee_id=1&year=2026&annual_gross=150000000&jht_worker=3600000" \
  -H "Authorization: Bearer $TOKEN"
```

`jht_worker` = total iuran JHT+JP dibayar pekerja selama setahun (pengurang neto, maks Rp 2,4 jt).

## 10. Struktur Proyek

```
cmd/server/main.go        # entrypoint (migrasi, seed, server)
internal/
  config/                 # konfigurasi env
  model/                  # model GORM
  dto/                    # DTO request/response
  repository/             # interface + implementasi GORM
  service/                # interface + business logic
  payroll/                # engine perhitungan (pure functions)
  handler/                # HTTP handler JSON
  middleware/             # JWT auth + RBAC
  router/                 # tabel route
openspec/                 # proposal, design, tasklist, archive
```

## 11. Test & Kualitas

```bash
make test   # go test ./...  (verifikasi engine perhitungan)
make vet    # go vet ./...
make build  # binary di bin/penggajian
```

## 12. Catatan Produksi

- Ganti `JWT_SECRET` dan password admin default sebelum deploy.
- Nilai cap BPJS (kesehatan, JP) dapat disesuaikan via env bila regulasi berubah.
- Slip gaji disimpan dengan detail JSONB sehingga mudah dikembangkan (misal: komponen baru).

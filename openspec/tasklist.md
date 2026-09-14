# Tasklist: Sistem Penggajian (Payroll) Indonesia

Status: `[x]` selesai, `[ ]` belum

## Phase 1 — Pondasi

- [x] Inisialisasi modul Go, `go.mod` + dependensi (GORM, postgres driver, JWT, bcrypt)
- [x] Konfigurasi environment (`internal/config`) + `.env.example`
- [x] Model GORM: User, Role, Permission, Employee, Allowance, Overtime, Payroll
- [x] DTO request/response + envelope respons JSON

## Phase 2 — Repository & Service (interface-based)

- [x] Interface repository: user, role/permission, employee, overtime, payroll
- [x] Implementasi repository dengan GORM
- [x] Interface service: auth, employee, payroll, tax
- [x] Implementasi service
- [x] Seeder RBAC: role, permission, user admin default

## Phase 3 — Engine Perhitungan

- [x] Tabel TER A/B/C lengkap (PP 58/2023)
- [x] PPh 21 bulanan metode TER + kategori PTKP
- [x] BPJS Kesehatan (5%, cap 12 jt)
- [x] BPJS Ketenagakerjaan: JHT, JKK, JKM, JP
- [x] Lembur PP 35/2021 (1/173, hari kerja & hari libur)
- [x] THR penuh & proporsional + pajak THR
- [x] Rekap PPh 21 tahunan (Pasal 17, PTKP, biaya jabatan)

## Phase 4 — HTTP API

- [x] Middleware JWT auth
- [x] Middleware RBAC (permission check)
- [x] Router dengan method pattern Go 1.22
- [x] Handler: auth (register/login), user, employee, overtime, payroll, THR, tax
- [x] `cmd/server/main.go` (migrasi otomatis, seed, start server)

## Phase 5 — Operasional & Dokumentasi

- [x] `Makefile` (termasuk target `db-create` untuk PostgreSQL lokal)
- [x] `README.md` (langkah penggunaan + contoh request/response)
- [x] `go build ./...`, `go vet ./...` lolos
- [x] Unit test engine perhitungan (11 test: TER, BPJS, lembur, THR, pajak tahunan)

## Phase 6 — Penyempurnaan

- [x] Register publik dipaksa role employee (anti eskalasi role)
- [x] Validasi input ketat: password, email, jam lembur, nominal, status PTKP
- [x] Pre-check konflik payroll run (tidak ada pembuatan sebagian)
- [x] Endpoint tambahan: GET /users, GET /payrolls?period=, PATCH /payrolls/{id}/paid, filter lembur per periode
- [x] Recovery middleware (anti panic) + graceful shutdown + health check database
- [x] Timeout server & batas body request

## Phase 7 — Robustness

- [x] Unique constraint (employee_id, period) + mapping 23505 ke 409 (anti race condition)
- [x] Basis upah BPJS sesuai PP 46/2015 (gaji pokok + tunjangan tetap)
- [x] Konsistensi basis tanggal THR (endpoint vs payroll run)
- [x] Update karyawan + tunjangan dalam satu transaksi
- [x] Terverifikasi: uji paralel payroll run, THR, validasi, RBAC (smoke test PostgreSQL asli)

## Phase 8 — Audit Bug Lanjutan

- [x] Preload karyawan di riwayat slip gaji
- [x] Hapus karyawan ber-payroll: RESTRICT + 409 (riwayat aman), mapping 23503 & 23001
- [x] TER-info PTKP invalid -> 400, register employee_id invalid -> 400
- [x] OptionalAuth lenient untuk endpoint publik
- [x] Pesan error payroll run tanpa karyawan aktif
- [x] 5 test service dengan fake repository (total 16 test)

## Phase 9 — Seeder Data Demo

- [x] `SeedDemo` idempotent: 4 karyawan + user, lembur, payroll bulan lalu, user finance
- [x] Dikontrol env `SEED_DEMO` (default false) + `DEMO_USER_PASS`
- [x] Test idempotency dengan fake repository
- [x] Mode `SEED_ONLY` + target `make seed` (migrasi + seeder lalu keluar)
- [x] Urutan pembuatan role deterministik (anti map acak)
- [x] Data terseed & terverifikasi di PostgreSQL lokal

## Kriteria Selesai

- [x] Aplikasi berjalan, migrasi otomatis & seed RBAC sukses
- [x] Semua endpoint mengembalikan JSON konsisten
- [x] Perhitungan sesuai regulasi Indonesia (TER, BPJS, lembur, THR)
- [x] Dokumentasi lengkap di README

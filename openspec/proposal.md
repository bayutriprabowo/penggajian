# Proposal: Sistem Penggajian (Payroll) Indonesia

## Ringkasan

Membangun aplikasi backend penggajian lengkap berbasis **Go (Golang)** dengan input/output **JSON**,
database **PostgreSQL** via **GORM**, dilengkapi **RBAC (Role-Based Access Control)**,
menggunakan pola **interface** untuk repository & service, serta **DTO** untuk komunikasi API.

Aplikasi menghitung seluruh komponen gaji sesuai regulasi Indonesia:

- Gaji pokok, tunjangan tetap & tidak tetap
- Upah lembur (PP 35/2021) — rumus 1/173
- Potongan & iuran BPJS Kesehatan (5%, cap Rp 12.000.000)
- Iuran BPJS Ketenagakerjaan: JHT (5,7%), JKK (0,24%–1,74%), JKM (0,3%), JP (3%, cap)
- PPh 21 bulanan metode **TER (Tarif Efektif Rata-rata)** — PP 58/2023 / PMK 168/2023
- THR (Permenaker 6/2016) — proporsional & penuh
- Rekap pajak tahunan (PPh 21 setahun, biaya jabatan, PTKP, tarif Pasal 17)

## Tujuan (Goals)

1. Backend REST API penggajian yang dapat dijalankan lokal dengan mudah (PostgreSQL lokal + `go run ./cmd/server`).
2. Semua respons API dalam format JSON dengan envelope konsisten.
3. RBAC: `admin`, `hr`, `finance`, `employee` dengan permission granular.
4. Kode modular: repository & service berbasis interface, mudah ditest (mockable).
5. Dokumentasi lengkap (README) dengan langkah & contoh penggunaan API.

## Non-goals

- Tidak menyediakan frontend/UI (backend only, JSON).
- Tidak menangani pemotongan pajak non-PPh 21 (PPN, PPh 23, dll).
- Tidak mengirim slip gaji via email.
- Tidak menangani multi-perusahaan / multi-entitas.

## Fitur Utama

| Kode | Fitur | Deskripsi |
|------|-------|-----------|
| F01 | Autentikasi | Register & login (JWT), password bcrypt |
| F02 | RBAC | Role & permission, middleware proteksi endpoint |
| F03 | Master Karyawan | CRUD karyawan + komponen gaji (tunjangan, BPJS, PTKP) |
| F04 | Lembur | Input jam lembur, hitung upah lembur PP 35/2021 |
| F05 | Payroll | Generate payroll per periode (bulan), slip gaji detail |
| F06 | PPh 21 TER | Kategori TER A/B/C berdasarkan status PTKP, tabel lengkap |
| F07 | THR | Kalkulasi THR penuh & proporsional + pajaknya |
| F08 | Rekap Tahunan | PPh 21 setahun (Pasal 17), untuk rekonsiliasi |

## Tech Stack

- Go 1.22+ (stdlib `net/http` router, Go 1.22 method pattern)
- GORM + driver PostgreSQL
- JWT (`golang-jwt/jwt/v5`), bcrypt (`x/crypto`)
- `gorm.io/datatypes` untuk detail slip gaji (JSONB)

## Estimasi

- Phase 1 (pondasi & autentikasi): F01–F03
- Phase 2 (engine perhitungan): F04–F08
- Phase 3 (dokumentasi & QA): README, build, contoh API

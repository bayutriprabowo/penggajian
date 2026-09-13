# Archive: Sistem Penggajian (Payroll) Indonesia

## Keputusan yang Diambil

| # | Keputusan | Alasan | Alternatif yang Ditolak |
|---|-----------|--------|--------------------------|
| 1 | Router stdlib `net/http` (Go 1.22 pattern) | Minim dependensi, cukup untuk API JSON | chi/gin/echo (butuh dep tambahan) |
| 2 | Metode PPh 21 bulanan: TER (PP 58/2023) | Metode resmi berlaku sejak Jan 2024 | Tarif progresif lama / gross-up (tidak lagi berlaku) |
| 3 | Detail slip disimpan JSONB | Fleksibel untuk komponen baru | Tabel relasional detail (rigid) |
| 4 | Engine perhitungan sebagai pure function terpisah | Mudah diuji, tidak tergantung DB | Logika dicampur di service |
| 5 | Repository & service berbasis interface | Testable (mock), sesuai pola Go idiomatik | Implementasi konkret langsung |
| 6 | DTO terpisah dari model | Mencegah bocornya field sensitif (password) ke JSON | Return model langsung |
| 7 | Seeder RBAC saat startup (idempotent) | Setup sekali jalan, dev-friendly | Migrasi SQL manual |

## Perubahan Selama Pengembangan

- Awalnya direncanakan memakai `gorilla/mux`; diganti stdlib `net/http` agar `go mod tidy` lebih ringan.
- Cap JP (pensiun) dan BPJS Kesehatan dibuat configurable via env karena nilai berubah antar-tahun.
- Endpoint `PATCH /payrolls/{id}/approve` dan `PATCH /payrolls/{id}/paid` untuk alur finance.
- Docker dihapus dari rencana (keputusan user) — setup memakai PostgreSQL lokal + `make db-create`.

## Perbaikan (Iterasi Penyempurnaan)

| Masalah | Perbaikan |
|---|---|
| Register publik bisa memilih role admin (eskalasi hak) | Register tanpa `user:write` dipaksa role employee |
| `MaxBytesReader` dengan writer nil (risiko panic) | Pass ResponseWriter asli |
| Update karyawan tanpa `base_salary` mengubah gaji jadi 0 | Hanya set bila `base_salary > 0` |
| Payroll run bisa gagal di tengah (partial write) | Pre-check konflik seluruh karyawan sebelum insert |
| Karyawan (role employee) tidak bisa akses profil sendiri | Tambah `user:read` ke role employee |
| GORM memetakan `PPh21` ke kolom `p_ph21` | Tag `gorm:"column:pph21"` eksplisit |
| Endpoint `/users/me` tidak menampilkan role | Tambah `role_name` di respons |
| Tanpa recovery middleware & graceful shutdown | Ditambahkan + timeout server + health check DB |

## Perbaikan (Iterasi Kedua — Bug & Robustness)

| Masalah | Perbaikan |
|---|---|
| Payroll duplikat saat run paralel (race condition) | Unique index `(employee_id, period)` + map SQLSTATE 23505 ke 409 |
| Basis upah BPJS keliru (ikut tunjangan tidak tetap) | Basis = gaji pokok + tunjangan tetap (PP 46/2015) |
| THR endpoint vs payroll run beda basis tanggal | Endpoint THR memakai akhir periode (konsisten) |
| Update karyawan + ganti tunjangan tidak atomik | Satu transaksi `UpdateWithAllowances` |
| Duplikat NIK/username saat race menghasilkan 500 | Semua unique violation diterjemahkan ke 409 |
| Konfigurasi & DTO mati (`BPJS_HEALTH_TOTAL`, `annual_bonus`) | Dihapus |

## Perbaikan (Iterasi Ketiga — Audit Bug Lanjutan)

| Masalah | Perbaikan |
|---|---|
| Riwayat slip gaji tanpa nama/NIK karyawan | Preload `Employee` di `ListByEmployee` |
| Hapus karyawan ber-payroll menghapus riwayat slip (CASCADE) | FK diubah ke RESTRICT + mapping 409 (termasuk SQLSTATE 23001) |
| `GET /tax/ter-info` dengan PTKP invalid menghasilkan 500 | Dipetakan ke 400 |
| Register dengan `employee_id` tidak ada menghasilkan 500 (23503) | FK violation dipetakan ke 400 |
| Register publik dengan token kedaluwarsa ditolak 401 | OptionalAuth bersifat lenient (lanjut anonim) |
| Payroll run tanpa karyawan aktif: pesan 404 menyesatkan | Diubah 400 dengan pesan jelas |
| JWT secret default tanpa peringatan | Log peringatan saat startup |
| Belum ada test level service | 5 test service dengan fake repository (auth, employee, payroll, THR) |

## Fitur Ditambahkan Setelah Rilis Awal

| Fitur | Keterangan |
|---|---|
| Seeder data demo | `SEED_DEMO=true` membuat 4 karyawan + user login, user finance, lembur, payroll bulan lalu; idempotent (dilewati bila karyawan sudah ada), password via `DEMO_USER_PASS` |

## Hal yang Belum Dikerjakan (Backlog)

- Unit test komprehensif untuk engine perhitungan (contoh kasus dari PP 58/2023 lampiran).
- Export slip gaji PDF.
- Multi-perusahaan, multi-cabang, penggajian mingguan/jam-jaman.
- Integrasi e-SPT / DJP online.
- PPh 21 untuk tenaga ahli & bukan pegawai (BUKAN pegawai, tarif 50% neto).
- Audit log perubahan data payroll.

## Ringkasan Hasil

- Aplikasi selesai sesuai proposal: autentikasi JWT, RBAC 4 role + 10 permission,
  CRUD karyawan, engine perhitungan lengkap (lembur, BPJS, PPh 21 TER, THR,
  rekap tahunan), API JSON dengan 21 endpoint, terverifikasi `go build`, `go vet`,
  16 test lolos (11 engine + 5 service), serta smoke test end-to-end berulang pada
  PostgreSQL asli (termasuk uji race condition dan pemetaan error DB).

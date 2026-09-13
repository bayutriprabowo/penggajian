# Design: Sistem Penggajian (Payroll) Indonesia

## Arsitektur

```
cmd/server/main.go          # entrypoint
internal/
  config/                   # env configuration
  model/                    # GORM models (PostgreSQL)
  dto/                      # request/response DTO
  repository/               # interface + implementasi GORM
  service/                  # interface + implementasi bisnis logic
  payroll/                  # engine perhitungan murni (pure function)
  handler/                  # HTTP handler (JSON in/out)
  middleware/               # JWT auth + RBAC
  router/                   # route table
```

Alur request: `handler -> service (interface) -> repository (interface) -> PostgreSQL`

## Skema Database

### users
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | bigserial PK | |
| username | varchar(100) unique | |
| email | varchar(100) unique | |
| password | varchar | bcrypt hash |
| role_id | fk -> roles.id | |
| employee_id | fk -> employees.id (nullable) | tautan ke karyawan |
| created_at / updated_at | timestamptz | |

### roles & permissions
- `roles`: id, name, description
- `permissions`: id, name
- `role_permissions`: many2many (role_id, permission_id)

### employees
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | bigserial PK | |
| nik | varchar(50) unique | nomor induk |
| full_name, email, phone, address | varchar/text | |
| position, department | varchar | |
| join_date | date | untuk hitung THR proporsional |
| status | varchar | active / inactive |
| ptkp_status | varchar | TK0, TK1, TK2, TK3, K0, K1, K2, K3 |
| base_salary | numeric | gaji pokok |
| bpjs_health | bool | peserta BPJS Kesehatan |
| bpjs_tk | bool | peserta BPJS Ketenagakerjaan |
| jkk_risk | numeric | kelas risiko JKK (0.24–1.74 %) |

### allowances
id, employee_id (fk), name, type (`fixed`/`non_fixed`), amount

### overtimes
id, employee_id, date, hours, is_holiday, amount

### payrolls
id, employee_id, period (YYYY-MM), basic_salary, allowances, overtime_pay, bonus, thr,
gross_income, employee_deduction, pph21, net_salary, employer_contribution,
details (JSONB detail per komponen), status (`draft`/`approved`/`paid`)

## RBAC

| Permission | admin | hr | finance | employee |
|---|---|---|---|---|
| user:read | x | x | x | x |
| user:list | x | x | | |
| user:write | x | | | |
| employee:read / employee:write | x | x | x(read) | |
| payroll:read | x | x | x | x (milik sendiri) |
| payroll:run | x | x | | |
| payroll:approve | x | | x | |
| overtime:write | x | x | | |
| tax:read | x | x | x | |

Seeder otomatis saat startup: role, permission, dan user `admin` default.

## Engine Perhitungan (internal/payroll — pure functions)

### 1. Lembur (PP 35/2021)
- Upah sejam = `1/173 × (gaji pokok + tunjangan tetap)`
- Hari kerja: jam ke-1 = 1,5× ; jam berikutnya = 2×
- Hari libur (6 hari kerja/minggu): 7 jam pertama 2×, jam ke-8 = 3×, jam 9–10 = 4×

### 2. BPJS Kesehatan (PP 28/2024)
- Total 5% upah: 4% pemberi kerja + 1% pekerja
- Cap upah = Rp 12.000.000/bulan

### 3. BPJS Ketenagakerjaan (PP 46/2015 & perubahannya)
| Program | Pekerja | Perusahaan | Cap |
|---|---|---|---|
| JHT | 2% | 3,7% | tidak ada |
| JKK | - | 0,24%–1,74% (kelas risiko) | tidak ada |
| JKM | - | 0,3% | tidak ada |
| JP | 1% | 2% | upah maks Rp 10.042.300 |

### 4. PPh 21 bulanan — TER (PP 58/2023, berlaku sejak Jan 2024)
- Kategori berdasarkan status PTKP:
  - **A**: TK/0, TK/1, K/0
  - **B**: TK/2, TK/3, K/1, K/2
  - **C**: K/3
- `PPh21 = tarif_TER × penghasilan bruto bulanan`, dibulatkan ke bawah.
- Tabel lengkap A/B/C (38 baris per kategori) dikodekan di `internal/payroll/ter.go`.
- Pajak THR = `TER(bruto + THR) − TER(bruto)`.

### 5. THR (Permenaker 6/2016)
- Masa kerja ≥ 12 bulan: `1 × (gaji pokok + tunjangan tetap)`
- Masa kerja < 12 bulan: `(masa/12) × (gaji pokok + tunjangan tetap)`

### 6. Rekap tahunan (Pasal 17 UU HPP)
- Penghasilan neto = bruto − biaya jabatan (5%, max Rp 6 jt) − iuran pensiun dibayar pekerja
- PKP = neto − PTKP (TK/0 = 54 jt, +4,5 jt per tanggungan/kawin)
- Tarif progresif: 5% (≤60 jt), 15% (≤250 jt), 25% (≤500 jt), 30% (≤5 m), 35% (>5 m)

## Kontrak API

Envelope sukses: `{"status":"success","message":"...","data":{...}}`
Envelope error: `{"status":"error","message":"..."}`

| Method | Path | Permission | Keterangan |
|---|---|---|---|
| POST | /api/v1/auth/register | publik (role employee) / user:write | daftar user |
| POST | /api/v1/auth/login | publik | login -> JWT |
| GET | /api/v1/users/me | user:read | profil user login |
| GET | /api/v1/users | user:list | daftar user (paginasi) |
| POST | /api/v1/employees | employee:write | tambah karyawan |
| GET | /api/v1/employees | employee:read | list karyawan |
| GET | /api/v1/employees/{id} | employee:read | detail |
| PUT | /api/v1/employees/{id} | employee:write | ubah |
| DELETE | /api/v1/employees/{id} | employee:write | hapus |
| POST | /api/v1/overtimes | overtime:write | input lembur |
| GET | /api/v1/overtimes/employee/{id} | employee:read | riwayat lembur (?period=) |
| POST | /api/v1/payrolls/run | payroll:run | generate payroll periode |
| GET | /api/v1/payrolls | payroll:read | list payroll per periode (?period=) |
| GET | /api/v1/payrolls/{id} | payroll:read | slip gaji |
| GET | /api/v1/payrolls/employee/{id} | payroll:read | riwayat slip |
| PATCH | /api/v1/payrolls/{id}/approve | payroll:approve | draft -> approved |
| PATCH | /api/v1/payrolls/{id}/paid | payroll:approve | approved -> paid |
| GET | /api/v1/thr/calculate | payroll:read | kalkulasi THR |
| GET | /api/v1/tax/ter | tax:read | tabel TER |
| GET | /api/v1/tax/annual-recap | tax:read | rekap PPh21 tahunan |

## Keputusan Desain

1. **stdlib `net/http`** (Go 1.22+ pattern routing) — minim dependensi.
2. **JSONB detail slip** — fleksibel menampung komponen baru tanpa migrasi.
3. **Engine terpisah dari service** — pure function mudah diunit-test.
4. **Interface repository/service** — implementasi bisa dimock untuk test.
5. **Rate/limit & cap BPJS dikonfigurasi via env** — mudah menyesuaikan regulasi baru.

# LUMBERA Backend API

LUMBERA adalah backend REST API untuk platform koperasi digital multi-tenant yang berfokus pada pencatatan transaksi koperasi, ledger terverifikasi, offline-first sync, credit scoring, laporan kesehatan koperasi, dan consent-gated data bridge untuk akses pembiayaan formal.

Project ini dibangun dengan Go, Gin, GORM, dan MariaDB. Backend ini menjadi API layer untuk PWA LUMBERA yang digunakan pengurus koperasi dan anggota koperasi.

## Dokumentasi

| Dokumen                                                                                 | Keterangan                                         |
| --------------------------------------------------------------------------------------- | -------------------------------------------------- |
| [Postman API Documentation](https://documenter.getpostman.com/view/33317073/2sBXwsMVk4) | Dokumentasi endpoint API untuk testing via Postman |
| [docs/LUMBERA-PRD.md](docs/LUMBERA-PRD.md)                                              | Product Requirement Document                       |
| [docs/LUMBERA-MCS-Calculation.md](docs/LUMBERA-MCS-Calculation.md)                      | Spesifikasi Member Creditworthiness Score          |
| [docs/LUMBERA-CHS-Calculation.md](docs/LUMBERA-CHS-Calculation.md)                      | Spesifikasi Cooperative Health Score               |

## Fitur Utama

- **Onboarding koperasi dan pengurus**: registrasi pengurus, OTP, PIN, draft onboarding, profil koperasi, dan konfigurasi finansial.
- **Authentication & authorization**: login berbasis PIN/JWT, session management, dan role-based access control.
- **Manajemen anggota**: daftar anggota, tambah anggota, aktivasi anggota, import anggota via template.
- **Transaksi simpan pinjam**: simpanan pokok, wajib, sukarela, pinjaman, angsuran, tarik tunai, dan reversal transaksi.
- **Offline sync API**: queue operasi offline dari frontend dapat disinkronkan melalui `/api/v1/sync/*`.
- **Verifiable ledger**: hash chain per transaksi/mutasi stok dan audit ledger.
- **Toko koperasi**: produk, stok masuk, penyesuaian stok, penjualan, dan mutasi stok.
- **Laporan koperasi**: laporan finansial, dashboard summary, export Excel/PDF, dan Cooperative Health Score.
- **Credit intelligence**: Member Creditworthiness Score, dashboard pinjaman, eligibility, dan pengajuan pinjaman.
- **Consent data bridge**: credit access request, grant/decline/revoke consent untuk mitra pembiayaan.

## Tech Stack

| Komponen              | Teknologi        |
| --------------------- | ---------------- |
| Language              | Go               |
| HTTP Framework        | Gin              |
| ORM                   | GORM             |
| Database              | MariaDB/MySQL    |
| Auth                  | JWT + bcrypt     |
| File/Object Storage   | Supabase Storage |
| WhatsApp/OTP Provider | Fonnte           |
| External Scoring API  | MCS scoring API  |

## Struktur Project

```text
.
├── cmd/app/main.go                  # Entry point aplikasi
├── entity/                          # GORM entities / database models
├── model/                           # Request/response DTO
├── internal/
│   ├── handler/rest/                # HTTP handlers dan route registration
│   ├── repository/                  # Data access layer
│   └── service/                     # Business logic layer
├── pkg/
│   ├── bcrypt/                      # Password/PIN hashing helper
│   ├── config/                      # Environment dan DSN config
│   ├── constant/                    # Domain constants
│   ├── database/mariadb/            # DB connection dan AutoMigrate
│   ├── errors/                      # AppError wrapper
│   ├── identity/                    # NIK hashing/encryption helper
│   ├── jwt/                         # JWT generation/validation
│   ├── middleware/                  # Auth, authorization, CORS
│   ├── response/                    # Standard response envelope
│   ├── supabase/                    # Supabase storage client
│   ├── whatsapp/                    # Fonnte WhatsApp client
│   └── mcsapi/                      # MCS scoring API client
└── docs/                            # Product dan technical docs
```

## Arsitektur Singkat

Backend memakai pola tiga layer:

```text
HTTP Request
    |
    v
Handler / REST
    |
    v
Service
    |
    v
Repository
    |
    v
MariaDB
```

- `internal/handler/rest`: binding request, parsing path/query/body, dan response JSON.
- `internal/service`: validasi bisnis, orchestration, transaction boundary, dan domain rules.
- `internal/repository`: query database via GORM.
- `entity`: struktur tabel database.
- `model`: DTO request/response API.

Dependency wiring dilakukan manual di [cmd/app/main.go](cmd/app/main.go).

## Prasyarat

Pastikan local machine memiliki:

- Go sesuai versi di [go.mod](go.mod)
- MariaDB atau MySQL
- Git
- Postman untuk testing manual API

Opsional:

- Supabase project jika ingin mencoba fitur upload/storage.
- Fonnte token jika ingin mencoba OTP/WhatsApp.
- MCS scoring API URL jika ingin mencoba trigger credit scoring.

## Setup Local

1. Clone repository.

```bash
git clone <repository-url>
cd lumbera-hackathon
```

2. Install dependency Go.

```bash
go mod tidy
```

Atau gunakan Makefile:

```bash
make init
```

3. Siapkan database MariaDB/MySQL.

Contoh:

```sql
CREATE DATABASE lumbera_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. Buat file `.env`.

```bash
cp .env.example .env
```

5. Isi konfigurasi `.env`.

```env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=lumbera_db
DB_USER=root
DB_PASSWORD=your-password

ADDRESS=localhost
PORT=8080

TIME_OUT_LIMIT=10

JWT_SECRET_KEY=a-string-secret-at-least-256-bits-long
JWT_EXP_TIME=1

NIK_HASH_SECRET=change-this-to-a-long-random-secret
NIK_ENCRYPTION_KEY=base64-encoded-32-byte-key

FONNTE_TOKEN=your-token

SUPABASE_URL=your-url
SUPABASE_TOKEN=your-token
SUPABASE_BUCKET=your-bucket

MCS_SCORING_API_URL=https://your-scoring-api.example.com/predict
```

Catatan:

- `NIK_ENCRYPTION_KEY` harus berupa base64 dari key 32 byte.
- `FONNTE_TOKEN`, `SUPABASE_*`, dan `MCS_SCORING_API_URL` dibutuhkan untuk fitur terkait. Endpoint lain tetap bisa digunakan selama fitur tersebut tidak dipanggil.
- Aplikasi menjalankan AutoMigrate saat startup.

## Menjalankan Aplikasi

Jalankan server:

```bash
make run
```

Atau:

```bash
go run cmd/app/main.go
```

Jika `.env` menggunakan:

```env
ADDRESS=localhost
PORT=8080
```

Maka API tersedia di:

```text
http://localhost:8080/api/v1
```

## Database Migration

Migration menggunakan GORM AutoMigrate dan dijalankan otomatis saat aplikasi start.

Command Makefile:

```bash
make migrate
```

Saat ini command tersebut menjalankan entrypoint yang sama dengan server, sehingga aplikasi tetap akan start setelah migrate selesai.

AutoMigrate juga melakukan seed data awal:

- System roles: `SUPERADMIN`, `REGULATOR`, `MITRA`, `ANGGOTA`, `PENGURUS_KOPERASI`
- Partner contoh: `Akseleran`

## Testing

Jalankan compile/test check:

```bash
make test
```

Atau:

```bash
go test ./...
```

Untuk manual API testing, gunakan dokumentasi Postman:

```text
https://documenter.getpostman.com/view/33317073/2sBXwsMVk4
```

Flow umum testing:

1. Jalankan server lokal.
2. Login atau lakukan onboarding sesuai endpoint yang tersedia.
3. Simpan token JWT dari response login.
4. Tambahkan header:

```http
Authorization: Bearer <access_token>
```

5. Akses endpoint yang membutuhkan role sesuai user.

## Offline Sync API

Offline sync dipakai frontend PWA untuk menyimpan transaksi saat device tidak punya koneksi, lalu mengirim operasi tersebut ketika online.

Endpoint utama:

| Endpoint                     | Fungsi                                                           |
| ---------------------------- | ---------------------------------------------------------------- |
| `GET /api/v1/sync/config`    | Mengambil aturan sync seperti max batch dan supported operations |
| `GET /api/v1/sync/bootstrap` | Mengambil cache awal untuk mode offline                          |
| `POST /api/v1/sync/push`     | Mengirim queue operasi offline ke server                         |
| `GET /api/v1/sync/status`    | Mengecek apakah id lokal tertentu sudah tersinkron               |

Mental model:

```text
config    = aturan main sync
bootstrap = data awal/cache dari server untuk dipakai offline
push      = kirim antrean offline ke database server
status    = recovery status berdasarkan client id lokal
```

Panduan detail frontend tersedia di [docs/LUMBERA-OFFLINE-SYNC-FRONTEND.md](docs/LUMBERA-OFFLINE-SYNC-FRONTEND.md).

## Endpoint Groups

Route utama berada di bawah prefix:

```text
/api/v1
```

Group endpoint:

| Group                  | Keterangan                                                               |
| ---------------------- | ------------------------------------------------------------------------ |
| `/onboarding`          | Registrasi pengurus, OTP, setup PIN, draft onboarding, aktivasi koperasi |
| `/auth`                | Login, forgot PIN, logout                                                |
| `/members`             | Manajemen anggota dan import anggota                                     |
| `/transactions`        | Transaksi simpan pinjam dan reversal                                     |
| `/store`               | Produk, stok, mutasi, dan penjualan toko koperasi                        |
| `/sync`                | Offline sync API                                                         |
| `/ledger`              | Audit dan anchor ledger                                                  |
| `/cooperative-members` | Dashboard anggota, buku tabungan, pinjaman, MCS, consent                 |
| `/reports`             | Laporan finansial, CHS, dashboard summary                                |
| `/internal`            | Callback internal seperti MCS scoring callback                           |

## Response Format

Semua response API menggunakan envelope standar:

```json
{
  "status": {
    "code": 200,
    "isSuccess": true
  },
  "message": "success message",
  "data": {}
}
```

Error response:

```json
{
  "status": {
    "code": 400,
    "isSuccess": false
  },
  "message": "error message",
  "data": "error detail"
}
```

## Development Notes

- Jangan mengubah data ledger dengan update/delete langsung. Koreksi transaksi dilakukan via reversal.
- Untuk operasi offline, frontend wajib mengirim idempotency key:
  - `client_transaction_id` untuk transaksi simpan pinjam.
  - `client_reference_id` untuk produk dan mutasi stok.
  - `client_sale_id` untuk penjualan toko.
- Setiap perubahan domain sebaiknya mengikuti alur:

```text
model -> repository -> service -> handler/rest -> route registration
```

- Gunakan dokumentasi di folder `docs/` sebagai sumber konteks product dan perhitungan scoring.

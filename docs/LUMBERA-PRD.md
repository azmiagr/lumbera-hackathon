# PRODUCT REQUIREMENTS DOCUMENT (PRD)

## LUMBERA — Platform Multi-Koperasi Berbasis Verifiable Ledger, Innovative Credit Scoring, dan Consent-Gated Data Bridge

---

## DAFTAR ISI

1. [Problem Statement](#1-problem-statement)
2. [Visi & Misi Produk](#2-visi--misi-produk)
3. [Goals & Success Metrics (OKR)](#3-goals--success-metrics-okr)
4. [Target Pengguna & User Personas](#4-target-pengguna--user-personas)
5. [Jobs-to-be-Done (JTBD)](#5-jobs-to-be-done-jtbd)
6. [User Stories](#6-user-stories)
7. [Arsitektur Fitur (Feature Map)](#7-arsitektur-fitur-feature-map)
8. [Functional Requirements — 6 Pilar LUMBERA](#8-functional-requirements--6-pilar-lumbera)
9. [Non-Functional Requirements](#9-non-functional-requirements)
10. [UX & Design Requirements](#10-ux--design-requirements)
11. [Tech Stack & Arsitektur Sistem](#11-tech-stack--arsitektur-sistem)
12. [Integrasi Eksternal](#12-integrasi-eksternal)
13. [Business Requirements & Revenue Model](#13-business-requirements--revenue-model)
14. [Compliance & Regulatory Requirements](#14-compliance--regulatory-requirements)
15. [Prioritisasi Fitur (RICE Framework)](#15-prioritisasi-fitur-rice-framework)
16. [Roadmap & Milestones](#16-roadmap--milestones)
17. [Risk Register](#17-risk-register)
18. [Deliverables & Output Terukur](#18-deliverables--output-terukur)
19. [Out of Scope](#19-out-of-scope)
20. [Glossarium](#20-glossarium)
21. [Referensi Regulasi](#21-referensi-regulasi)

---

## 1. PROBLEM STATEMENT

### 1.1 Konteks Masalah

Berdasarkan studi kasus pada **Koperasi Viva** (sebagai representasi mayoritas koperasi desa di Indonesia), ditemukan tiga akar permasalahan struktural yang membatasi potensi koperasi:

| #   | Akar Masalah                              | Manifestasi Lapangan                                                                                                      | Dampak                                                                |
| --- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| 1   | **Arsitektur single-tenant**              | Sistem pencatatan tidak mampu melayani karakteristik koperasi yang heterogen (simpan pinjam, komoditas, peternakan, dll.) | Data tercampur lintas koperasi, error pelaporan, konflik antar unit   |
| 2   | **Tata kelola data tidak siap integrasi** | Laporan keuangan koperasi tidak memenuhi standar kredibilitas mitra pembiayaan                                            | Koperasi tertolak saat mengajukan kredit ke fintech/bank (unbankable) |
| 3   | **Rendahnya literasi digital pengurus**   | Pengurus senior tidak mampu mengoperasikan sistem berbasis spreadsheet/desktop; sinyal terbatas di daerah 3T              | Laporan terlambat, data tidak akurat, adopsi digital gagal            |

### 1.2 Analisis 5-Whys (Studi Kasus Koperasi Viva)

```
Mengapa koperasi tidak bisa mendapat pembiayaan formal?
  → Karena mitra pembiayaan tidak mempercayai catatan keuangan koperasi.

Mengapa catatan keuangan tidak terpercaya?
  → Karena laporan dibuat manual di spreadsheet, rentan manipulasi dan kesalahan.

Mengapa laporan tidak terverifikasi?
  → Karena tidak ada sistem audit trail dan mekanisme hash anti-tamper.

Mengapa tidak ada sistem digital yang andal?
  → Karena sistem yang ada tidak offline-first, tidak ramah pengurus senior, dan tidak multi-tenant.

Mengapa sistem yang ada tidak memadai?
  → Karena dibangun untuk UMKM perkotaan, bukan untuk koperasi desa yang heterogen.
```

### 1.3 Ukuran Masalah

- **222.462 unit koperasi aktif** di Indonesia (Kemenkop UKM, 2024; BPS, 2025)
- **~155.000 unit (70%)** adalah koperasi simpan pinjam yang membutuhkan digitalisasi
- **Gap pembiayaan UMKM**: Realisasi porsi UMKM di fintech lending baru **36,53%** (April 2025) dari target 40–50% OJK 2025–2026 (Antara News, 2025)
- **Potensi pasar open finance Indonesia**: USD 2 miliar TAM (Katadata Insight Center & Finantier, 2022)

---

## 2. VISI & MISI PRODUK

### 2.1 Visi

> **"Menjadi lapisan kepercayaan data (trust layer) pertama untuk ekosistem koperasi digital Indonesia, membuka akses pembiayaan formal bagi jutaan anggota masyarakat desa yang selama ini unbankable."**

### 2.2 Misi

LUMBERA membangun platform manajemen multi-koperasi yang:

1. **Mencatat** transaksi harian secara akurat, baik online maupun offline, dengan integritas data yang terjamin.
2. **Memvalidasi** catatan operasional melalui verifiable ledger berbasis hash chain sehingga layak dipercaya oleh mitra pembiayaan.
3. **Menilai** kelayakan kredit anggota dan kesehatan koperasi menggunakan AI scoring dua tingkat yang transparan dan patuh regulasi.
4. **Menghubungkan** koperasi dengan ekosistem pembiayaan formal melalui Open API berstandar SNAP Bank Indonesia, dengan kendali penuh di tangan anggota (consent-gated).

### 2.3 Nama & Konsep

"LUMBERA" mengadopsi prinsip kerja **lumbung tradisional** — pengelolaan aset komunal yang berlandaskan validitas dan transparansi pencatatan — dan mengimplementasikannya ke dalam sistem digital modern.

---

## 3. GOALS & SUCCESS METRICS (OKR)

### 3.1 OKR Tahun 1 (MVP + Pilot)

**Objective 1: Membuktikan validitas teknis platform pada 5 koperasi heterogen**

| Key Result                                                                           | Target           | Metode Pengukuran                    |
| ------------------------------------------------------------------------------------ | ---------------- | ------------------------------------ |
| KR1.1: Onboarding 5 koperasi pilot selesai dalam 6 bulan                             | 100%             | Hitungan koperasi aktif di platform  |
| KR1.2: Offline sync berhasil di area sinyal <1 bar                                   | 95% success rate | Log sinkronisasi Background Sync API |
| KR1.3: Zero hash collision pada verifiable ledger dalam 6 bulan                      | 0 collision      | Audit log SHA-256 Merkle tree        |
| KR1.4: 80% pengurus koperasi pilot dapat mengoperasikan sistem mandiri dalam 3 bulan | ≥80%             | Tes operasional post-training        |

**Objective 2: Memvalidasi credit scoring sebagai pintu akses pembiayaan**

| Key Result                                                                      | Target         | Metode Pengukuran              |
| ------------------------------------------------------------------------------- | -------------- | ------------------------------ |
| KR2.1: Member Creditworthiness Score (MCS) dikalkulasi untuk ≥90% anggota aktif | ≥90%           | Dashboard CCI analytics        |
| KR2.2: Minimal 1 penyaluran kredit riil melalui Open Bridge terealisasi         | ≥1 kasus       | Audit log transaksi SNAP BI    |
| KR2.3: Gini coefficient model CCI ≥0.55 pada data pilot                         | ≥0.55          | Model evaluation MLflow        |
| KR2.4: NPL rate koperasi pilot turun ≥20% vs baseline sebelum adopsi            | ≥20% penurunan | Data perbandingan NPL koperasi |

**Objective 3: Membangun fondasi keberlanjutan bisnis**

| Key Result                                                         | Target        | Metode Pengukuran                            |
| ------------------------------------------------------------------ | ------------- | -------------------------------------------- |
| KR3.1: Minimal 1 kemitraan aktif dengan lembaga pembiayaan ter-OJK | ≥1 mitra      | Dokumen perjanjian kerja sama resmi          |
| KR3.2: Open-source SDK ledger terpublikasi di GitHub               | 1 repo publik | GitHub repository metrics                    |
| KR3.3: Whitepaper metodologi CCI selesai untuk pengajuan OJK       | 1 dokumen     | Konfirmasi OJK Pemeringkat Kredit Alternatif |

### 3.2 OKR Tahun 3–5 (Scale)

| Metric                            | Tahun 3  | Tahun 5       |
| --------------------------------- | -------- | ------------- |
| Koperasi teronboarding            | 200      | 1.550         |
| Anggota terlayani                 | 123.000  | 954.250       |
| Penyaluran kredit via Open Bridge | Rp 200 M | Rp 1,5 T      |
| Penurunan NPL rata-rata           | 25%      | 40%           |
| Revenue                           | Rp 15 M  | Rp 25 M/bulan |
| NPV (discount rate 12%)           | —        | Rp 5,48 M     |
| ROI                               | —        | 290%          |

---

## 4. TARGET PENGGUNA & USER PERSONAS

### 4.1 User Segments

LUMBERA melayani 4 segmen pengguna utama:

| Segmen                | Deskripsi                                           | Jumlah Estimasi            |
| --------------------- | --------------------------------------------------- | -------------------------- |
| **Pengurus Koperasi** | Admin/ketua yang mencatat transaksi harian          | ~222.462 koperasi          |
| **Anggota Koperasi**  | Peminjam/penabung yang menggunakan layanan koperasi | ~40 juta orang             |
| **Mitra Pembiayaan**  | Fintech lending, BPR, BPD yang menyalurkan kredit   | ~102 fintech terdaftar OJK |
| **Regulator**         | OJK, Kemenkop UKM yang memantau kesehatan koperasi  | 2 lembaga utama            |

---

### 4.2 Persona 1 — "Pak Asep" (Pengurus Koperasi Senior)

**Nama:** Pak Asep Suherman
**Archetype:** `offline_first_operator`
**Usia:** 54 tahun
**Lokasi:** Desa Padiwangi, Karawang, Jawa Barat
**Jabatan:** Bendahara Koperasi Padiwangi (15 tahun pengalaman)
**Device:** Android entry-level (RAM 2GB), sinyal 2G/3G tidak stabil
**Tech Proficiency:** 2/10 (bisa WhatsApp dan foto, belum pernah pakai aplikasi akuntansi)

**Quote:**

> _"Kalau aplikasinya ribet kayak Excel, saya mending balik ke buku tulis. Yang penting datanya aman dan laporan bisa saya buat sebelum RAT."_

**Pekerjaan Utama (Customer Jobs):**

- Mencatat setoran Bu Lestari dan anggota lain setiap hari
- Membuat laporan keuangan bulanan untuk rapat pengurus
- Menyimpan bukti transaksi agar tidak hilang
- Mengajukan laporan tahunan ke Kemenkop/OJK

**Pains (Frustrations):**

- Takut salah ketik di Excel, tidak ada "undo" yang mudah
- Harus ke kota (4 jam perjalanan) untuk mencetak laporan saat sinyal jelek
- Laporan sering tidak selesai tepat waktu karena rekonsiliasi manual butuh 3 hari
- Pernah kehilangan data 6 bulan karena laptop rusak kena banjir
- Tidak tahu cara mengajukan kredit ke bank untuk modal koperasi

**Gains yang Diharapkan:**

- Laporan otomatis yang bisa digenerate dalam 30 menit
- Bisa input transaksi dari HP meski tanpa sinyal
- Data aman tersimpan di cloud, tidak hilang meski HP rusak
- Koperasi bisa dapat modal dari bank/fintech

**Skenario Penggunaan:**

> Pak Asep login ke LUMBERA PWA pagi hari. Ia memilih ikon "Setoran" (gambar celengan) dan mengetik nama Bu Lestari. Sistem menampilkan profil Bu Lestari otomatis. Pak Asep input Rp 500.000, tekan tombol besar "Simpan". Muncul notifikasi "Data tersimpan di perangkat, sinkron otomatis saat online". Sore hari saat HP mendapat sinyal, data tersinkron ke ledger tanpa Pak Asep perlu melakukan apapun.

**Design Implications:**

- Ikon besar bergambar (bukan teks) untuk setiap fungsi utama
- Alur maksimal 3 tap untuk transaksi paling umum
- Konfirmasi audio/haptic saat data berhasil tersimpan
- Mode offline harus default, bukan exception

---

### 4.3 Persona 2 — "Bu Lestari" (Anggota Koperasi / Peminjam)

**Nama:** Bu Siti Lestari
**Archetype:** `unbanked_borrower`
**Usia:** 38 tahun
**Lokasi:** Desa Padiwangi, Karawang, Jawa Barat
**Pekerjaan:** Petani padi, anggota Koperasi Padiwangi selama 8 tahun
**Device:** Smartphone basic, Android 8
**Tech Proficiency:** 3/10 (bisa TikTok, WhatsApp, tapi belum pernah dapat layanan kredit formal)
**Status Keuangan:** No-file/thin-file — tidak punya rekening bank, tidak pernah dapat kredit formal

**Quote:**

> _"Kalau bisa dapat pinjaman untuk beli pupuk dari bank, panen saya bisa dua kali lebih besar. Tapi kata bank, saya tidak ada datanya."_

**Customer Jobs:**

- Menyetor tabungan rutin setiap minggu ke koperasi
- Mengajukan pinjaman untuk modal tanam (rata-rata Rp 3–5 juta/siklus)
- Melihat saldo dan riwayat transaksi kapan saja
- Mendapatkan akses kredit yang lebih besar dari fintech/bank

**Pains:**

- Tidak tahu berapa saldo tabungannya tanpa bertanya ke Pak Asep
- Proses pinjam koperasi memakan 1 minggu untuk validasi manual
- Pernah ditolak pinjaman bank karena "tidak ada riwayat kredit"
- Tidak tahu apa itu credit score atau bagaimana cara meningkatkannya

**Gains:**

- Bisa cek saldo sendiri dari HP kapan saja
- Proses pengajuan pinjaman cepat (hasil scoring dalam hitungan menit)
- Memiliki "Data Passport" — identitas kredit digital yang portabel
- Bisa mengakses kredit dari fintech/bank dengan bunga lebih rendah

**Design Implications:**

- Tampilan saldo dan histori transaksi sangat sederhana (gaya buku tabungan)
- Notifikasi push saat pinjaman disetujui atau jatuh tempo mendekat
- Penjelasan MCS dalam bahasa sederhana (bukan skor angka, tapi "Profil kredit Anda: Baik")
- Consent dialog harus jelas: data apa yang dibagikan, ke siapa, untuk apa

---

### 4.4 Persona 3 — "Mbak Rini" (Analis Kredit, Mitra Pembiayaan)

**Nama:** Rini Puspitasari
**Archetype:** `fintech_credit_analyst`
**Usia:** 29 tahun
**Lokasi:** Jakarta Selatan
**Jabatan:** Credit Analyst, Akseleran (Platform P2P Lending OJK)
**Device:** MacBook Pro + iPhone 14
**Tech Proficiency:** 9/10 (berpengalaman dengan API integration, credit model)

**Quote:**

> _"Kita mau ekspansi ke segmen koperasi desa, tapi datanya selalu tidak terstruktur dan tidak bisa diverifikasi. Kalau ada API yang langsung kasih skor terverifikasi, proses approval kita bisa 10x lebih cepat."_

**Customer Jobs:**

- Mengevaluasi kelayakan kredit pemohon dari koperasi desa
- Memvalidasi keaslian data keuangan yang diajukan koperasi
- Menyalurkan kredit sesuai profil risiko yang dapat dipertanggungjawabkan
- Melaporkan portofolio penyaluran kepada OJK

**Pains:**

- Data dari koperasi tidak terstandar: ada yang pakai Excel, ada yang manual
- Tidak bisa memverifikasi apakah laporan keuangan koperasi asli atau dimanipulasi
- Proses due diligence koperasi memakan 2–4 minggu karena verifikasi manual
- Tidak ada skor kredit alternatif yang valid untuk segmen no-file/thin-file

**Gains:**

- API tunggal yang menghasilkan credit score terverifikasi ledger dalam <5 detik
- Bukti audit trail yang immutable sebagai dasar keputusan kredit
- Dashboard monitoring portofolio koperasi secara real-time
- Compliance otomatis terhadap POJK 29/2024 dan POJK 40/2024

**Design Implications:**

- Open API dengan dokumentasi lengkap (OpenAPI 3.0 spec)
- Dashboard analytics portofolio dengan filter risiko, sektor, wilayah
- Webhook untuk notifikasi perubahan skor kredit secara real-time
- Audit log yang dapat diunduh untuk keperluan pelaporan OJK

---

### 4.5 Persona 4 — "Pak Budi" (Inspektur OJK / Regulator)

**Nama:** Budi Santoso
**Archetype:** `regulatory_inspector`
**Usia:** 45 tahun
**Lokasi:** Jakarta
**Jabatan:** Kepala Bidang Pengawasan Fintech, OJK
**Tech Proficiency:** 6/10

**Quote:**

> _"Kami butuh visibilitas real-time kondisi ekosistem koperasi. Sekarang datanya tersebar dan tidak terstandardisasi, kita tidak bisa deteksi masalah sebelum terlambat."_

**Customer Jobs:**

- Memantau kesehatan koperasi secara agregat di seluruh Indonesia
- Memverifikasi kepatuhan koperasi terhadap regulasi yang berlaku
- Mendeteksi early warning indicators sebelum terjadi gagal bayar massal
- Mengkoordinasikan data dengan Kemenkop UKM

**Gains:**

- Regulator Dashboard dengan metrik Cooperative Health Score (CHS) secara agregat
- Distribusi kelas kredit (grade AA–D) per wilayah dan sektor
- Sistem early warning otomatis jika ada koperasi dengan CHS turun drastis
- Ekspor data terstandar untuk pelaporan internal OJK

---

## 5. JOBS-TO-BE-DONE (JTBD)

### Framework JTBD per Persona

| Persona              | When...                                           | I want to...                                     | So I can...                                              |
| -------------------- | ------------------------------------------------- | ------------------------------------------------ | -------------------------------------------------------- |
| Pak Asep (Pengurus)  | mencatat setoran anggota di lapangan tanpa sinyal | input transaksi offline yang tersinkron otomatis | memastikan tidak ada data yang hilang                    |
| Pak Asep (Pengurus)  | menyusun laporan keuangan RAT                     | generate laporan otomatis sesuai template        | menghemat 3 hari kerja menjadi 30 menit                  |
| Bu Lestari (Anggota) | ingin tahu saldo tabungan saya                    | cek saldo dari HP kapan saja                     | tidak harus datang ke kantor koperasi                    |
| Bu Lestari (Anggota) | butuh modal untuk tanam musim ini                 | mengajukan pinjaman dengan proses cepat          | mendapat dana sebelum musim tanam berakhir               |
| Mbak Rini (Analis)   | menerima pengajuan kredit dari koperasi           | verifikasi keaslian data dalam <5 detik          | mempercepat proses approval dari 2 minggu menjadi 1 hari |
| Pak Budi (OJK)       | memantau kondisi koperasi secara nasional         | akses dashboard agregat real-time                | mendeteksi risiko sistemik lebih awal                    |

---

## 6. USER STORIES

### 6.1 Epic 1: Pencatatan Transaksi Offline-First

| ID     | User Story                                                                                                                                                           | Acceptance Criteria                                                                                                                                                                                     | Prioritas   | Estimasi           |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | ------------------ |
| US-001 | Sebagai **Pengurus**, saya ingin **mencatat setoran anggota tanpa koneksi internet** sehingga **tidak ada transaksi yang tertunda saat sinyal jelek**                | (1) Form input transaksi tersedia dalam mode offline. (2) Data tersimpan di IndexedDB lokal. (3) Notifikasi "Tersimpan offline" muncul. (4) Sinkronisasi otomatis dalam <30 detik saat koneksi tersedia | Must Have   | M (5 sprint point) |
| US-002 | Sebagai **Pengurus**, saya ingin **melihat antrian transaksi offline yang belum tersinkron** sehingga **saya tahu status data saya**                                 | (1) Badge jumlah transaksi pending terlihat di home. (2) Daftar transaksi pending dapat dibuka. (3) Status sync (pending/synced/failed) terlihat per transaksi                                          | Must Have   | S (3 sprint point) |
| US-003 | Sebagai **Pengurus**, saya ingin **mencari nama anggota dengan autocomplete** sehingga **tidak salah input identitas**                                               | (1) Autocomplete muncul setelah 2 karakter. (2) Foto profil anggota ditampilkan untuk konfirmasi visual. (3) Jika anggota baru, ada opsi "Daftarkan Anggota Baru"                                       | Must Have   | S                  |
| US-004 | Sebagai **Pengurus**, saya ingin **menggunakan input suara untuk nominal transaksi** sehingga **pengurus senior yang tidak terbiasa keyboard tetap bisa input data** | (1) Tombol microphone tersedia di field nominal. (2) Konversi suara ke angka dengan akurasi ≥85% untuk bahasa Indonesia. (3) Preview angka sebelum dikonfirmasi                                         | Should Have | M                  |
| US-005 | Sebagai **Pengurus**, saya ingin **input mutasi stok komoditas (beras, dll.) secara offline** sehingga **catatan gudang selalu terkini**                             | (1) Modul stok tersedia di semua template koperasi tipe Pangan Bulky. (2) FIFO calculation otomatis. (3) Alert saat stok <20% dari minimum                                                              | Should Have | M                  |

### 6.2 Epic 2: Verifiable Ledger & Audit Trail

| ID     | User Story                                                                                                                                           | Acceptance Criteria                                                                                                                                                                  | Prioritas   | Estimasi             |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------- | -------------------- |
| US-010 | Sebagai **Pengurus**, saya ingin **setiap transaksi yang tersimpan menghasilkan hash unik** sehingga **tidak ada yang bisa mengubah data diam-diam** | (1) Setiap transaksi menghasilkan SHA-256 hash. (2) Hash setiap transaksi di-chain ke hash transaksi sebelumnya (Merkle tree). (3) Sistem menolak jika terdeteksi tampering          | Must Have   | L (8 sprint point)   |
| US-011 | Sebagai **Mitra Pembiayaan**, saya ingin **memverifikasi keaslian ledger koperasi** sehingga **saya bisa yakin data tidak dimanipulasi**             | (1) API endpoint /ledger/verify tersedia. (2) Response berisi: root hash, timestamp anchor, blockchain tx ID. (3) Waktu verifikasi <2 detik                                          | Must Have   | L                    |
| US-012 | Sebagai **Pengurus**, saya ingin **melihat audit trail lengkap setiap transaksi** sehingga **saya bisa melacak siapa mengubah apa dan kapan**        | (1) Setiap perubahan data memiliki log: user_id, timestamp, before_value, after_value, hash. (2) Audit trail tidak dapat dihapus. (3) Dapat diekspor sebagai PDF/CSV                 | Must Have   | M                    |
| US-013 | Sebagai **OJK**, saya ingin **menerima anchor periodik ledger koperasi** sehingga **ada bukti independen integritas data**                           | (1) Anchor ke permissioned blockchain (Hyperledger Fabric) setiap 24 jam. (2) Konfirmasi anchor tersimpan di sistem. (3) Anchor dapat diverifikasi via blockchain explorer regulator | Should Have | XL (13 sprint point) |

### 6.3 Epic 3: Cooperative Credit Intelligence (CCI)

| ID     | User Story                                                                                                                                                                     | Acceptance Criteria                                                                                                                                                                                                                                 | Prioritas   | Estimasi |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------- |
| US-020 | Sebagai **Anggota**, saya ingin **melihat skor kredit saya (MCS)** sehingga **saya tahu seberapa besar pinjaman yang bisa saya dapatkan**                                      | (1) MCS ditampilkan dalam skala 300–850 dengan label deskriptif (Sangat Baik/Baik/Cukup/Perlu Perhatian). (2) Breakdown 5C ditampilkan (Character, Capacity, Capital, Conditions, Collateral). (3) Skor diperbarui otomatis tiap ada transaksi baru | Must Have   | L        |
| US-021 | Sebagai **Mitra Pembiayaan**, saya ingin **mengakses MCS anggota via API** sehingga **keputusan kredit dapat diautomatisasi**                                                  | (1) API endpoint /cci/member/{member_id}/score tersedia. (2) Response berisi: skor, grade, tanggal update, faktor pengurang, SHAP explanation. (3) Waktu response <3 detik. (4) Rate limit: 1.000 request/hari per API key                          | Must Have   | L        |
| US-022 | Sebagai **Pengurus**, saya ingin **melihat Cooperative Health Score (CHS)** sehingga **saya bisa memantau kesehatan koperasi secara keseluruhan**                              | (1) CHS ditampilkan dalam skala 0–100 dengan grade AA/A/B/C/D. (2) Trend 6 bulan terakhir ditampilkan dalam grafik. (3) Alert otomatis jika CHS turun >10 poin dalam 30 hari                                                                        | Must Have   | M        |
| US-023 | Sebagai **Anggota**, saya ingin **memahami mengapa skor kredit saya seperti itu** sehingga **saya tahu apa yang harus diperbaiki**                                             | (1) Penjelasan dalam bahasa Indonesia sederhana (bukan istilah teknis). (2) Top 3 faktor positif dan top 3 faktor negatif ditampilkan. (3) Rekomendasi actionable ("Setor tepat waktu selama 3 bulan untuk meningkatkan skor 15 poin")              | Should Have | M        |
| US-024 | Sebagai **Sistem** (background job), saya ingin **mendeteksi anomali transaksi yang mengindikasikan fraud** sehingga **pengurus dapat diperingatkan sebelum kerugian terjadi** | (1) Model TensorFlow fraud detection berjalan setiap batch harian. (2) Alert email/push dikirim ke pengurus dalam <1 jam setelah deteksi. (3) Anomali disimpan di audit log untuk investigasi                                                       | Should Have | XL       |

### 6.4 Epic 4: Consent-Gated Open Bridge

| ID     | User Story                                                                                                                                                        | Acceptance Criteria                                                                                                                                                                                                                                                                   | Prioritas   | Estimasi |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------- |
| US-030 | Sebagai **Anggota**, saya ingin **memberikan persetujuan eksplisit sebelum data saya dibagikan ke mitra** sehingga **saya memiliki kendali penuh atas data saya** | (1) Dialog consent menampilkan: nama mitra, data yang dibagikan, durasi consent, tujuan penggunaan. (2) Consent harus klik "Setuju" aktif (bukan pre-checked). (3) Consent dapat dicabut kapan saja dari menu "Privasi Saya". (4) Log consent (waktu, IP, device) tersimpan immutable | Must Have   | M        |
| US-031 | Sebagai **Mitra Pembiayaan**, saya ingin **menerima data kredit koperasi via API SNAP BI** sehingga **integrasi dengan sistem kami mudah dan terstandar**         | (1) API mengikuti PADG No. 23/15/2021 SNAP BI. (2) Authentication via OAuth 2.0 + OpenID Connect. (3) Dokumentasi OpenAPI 3.0 tersedia di developer portal. (4) Sandbox environment tersedia untuk testing                                                                            | Must Have   | XL       |
| US-032 | Sebagai **Anggota**, saya ingin **melihat "Data Passport" saya** sehingga **saya tahu data apa saja yang tersimpan dan siapa saja yang pernah mengaksesnya**      | (1) Data Passport menampilkan: profil kredit lengkap, riwayat consent, log akses per mitra. (2) Ekspor Data Passport sebagai PDF. (3) Sesuai UU 27/2022 tentang Pelindungan Data Pribadi Pasal 9 & 20                                                                                 | Must Have   | M        |
| US-033 | Sebagai **Pengurus**, saya ingin **mencairkan kredit melalui Open Bridge ke Virtual Account anggota** sehingga **prosesnya otomatis tanpa proses manual**         | (1) Setelah mitra setujui, dana masuk ke Virtual Account anggota dalam <1 jam. (2) Transaksi pencairan tercatat di ledger. (3) Notifikasi push dikirim ke anggota                                                                                                                     | Should Have | XL       |

### 6.5 Epic 5: Adaptive Reporting Schema & Multi-Tenancy

| ID     | User Story                                                                                                                                                               | Acceptance Criteria                                                                                                                                                                                            | Prioritas   | Estimasi |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------- |
| US-040 | Sebagai **Pengurus**, saya ingin **memilih template laporan yang sesuai jenis koperasi saya** sehingga **laporan mudah dimengerti tanpa membutuhkan keahlian akuntansi** | (1) 6 template tersedia: KSP, Pangan Bulky, Cold-Chain Perishable, Toko Gerai, Utility Subscription, Peternakan. (2) Template ditentukan saat onboarding. (3) Glosarium istilah dalam bahasa pengurus tersedia | Must Have   | L        |
| US-041 | Sebagai **Pengurus**, saya ingin **generate laporan keuangan bulanan dalam 30 menit** sehingga **tidak perlu 3 hari seperti spreadsheet manual**                         | (1) Laporan Neraca, Laba Rugi, dan Arus Kas di-generate otomatis. (2) Format laporan sesuai standar Kemenkop UKM. (3) Ekspor sebagai PDF dan Excel                                                             | Must Have   | M        |
| US-042 | Sebagai **Admin Platform** (LUMBERA), saya ingin **setiap koperasi terisolasi datanya secara teknis** sehingga **tidak ada risiko data bocor antar koperasi**            | (1) Setiap query ke database mengandung filter tenant_id (first-class citizen). (2) Middleware menolak query tanpa tenant context (HTTP 403). (3) Row-Level Security PostgreSQL aktif di semua tabel           | Must Have   | M        |
| US-043 | Sebagai **Koordinator Multi-cabang**, saya ingin **melihat konsolidasi laporan dari semua cabang** sehingga **saya punya gambaran menyeluruh tanpa rekonsiliasi manual** | (1) Dashboard agregat multi-cabang dengan filter per cabang. (2) Rekonsiliasi transaksi antar-cabang otomatis. (3) Alert jika ada ketidakcocokan saldo antar-cabang                                            | Should Have | L        |

### 6.6 Epic 6: Onboarding & Regulator Dashboard

| ID     | User Story                                                                                                                | Acceptance Criteria                                                                                                                                                    | Prioritas   | Estimasi |
| ------ | ------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------- |
| US-050 | Sebagai **Pengurus baru**, saya ingin **onboarding koperasi selesai dalam 1 hari** sehingga **bisa langsung operasional** | (1) Wizard onboarding 7 langkah dengan estimasi waktu per langkah. (2) Import data anggota dari Excel/CSV. (3) Training interaktif in-app (tutorial gamified)          | Must Have   | M        |
| US-051 | Sebagai **OJK**, saya ingin **melihat dashboard agregat kesehatan koperasi** sehingga **bisa deteksi masalah lebih awal** | (1) Dashboard menampilkan: distribusi CHS, peta sebaran koperasi, early warning alerts. (2) Filter per wilayah, sektor, ukuran koperasi. (3) Data diperbarui real-time | Should Have | L        |

---

## 7. ARSITEKTUR FITUR (FEATURE MAP)

```
LUMBERA Platform
│
├── CORE FEATURES (Must Have — MVP)
│   ├── P1. Tenant Context Engine
│   │   ├── Multi-tenant isolation (tenant_id first-class citizen)
│   │   ├── Row-Level Security PostgreSQL
│   │   └── Middleware tenant validation
│   │
│   ├── P2. Adaptive Reporting Schema (ARS)
│   │   ├── Template: KSP (Koperasi Simpan Pinjam)
│   │   ├── Template: Pangan Bulky (Beras, Jagung)
│   │   ├── Template: Cold-Chain Perishable
│   │   ├── Template: Toko Gerai
│   │   ├── Template: Utility Subscription
│   │   ├── Template: Peternakan
│   │   └── Glosarium bahasa pengurus (auto-render)
│   │
│   ├── P3. Offline-First PWA
│   │   ├── Service Worker (Workbox)
│   │   ├── IndexedDB (Dexie.js) — local storage
│   │   ├── Background Sync API — auto-sync
│   │   ├── Push Notifications
│   │   └── Voice Input untuk nominal transaksi
│   │
│   ├── P4. Verifiable Ledger
│   │   ├── Hash chain SHA-256 per transaksi
│   │   ├── Merkle tree accumulation
│   │   ├── Anchor periodik ke blockchain (Hyperledger Fabric)
│   │   ├── API verifikasi ledger untuk mitra
│   │   └── Audit log immutable
│   │
│   ├── P5. Cooperative Credit Intelligence (CCI)
│   │   ├── Member Creditworthiness Score (MCS) — skala 300-850
│   │   │   ├── Character (35%)
│   │   │   ├── Capacity (30%)
│   │   │   ├── Capital (15%)
│   │   │   ├── Conditions (12%)
│   │   │   └── Collateral (8%)
│   │   ├── Cooperative Health Score (CHS) — skala 0-100, grade AA-D
│   │   ├── SHAP explainability
│   │   ├── Fraud anomaly detection (TensorFlow)
│   │   └── Model versioning (MLflow)
│   │
│   └── P6. Consent-Gated Open Bridge
│       ├── Consent Manager (granular, revocable)
│       ├── Data Passport (anggota)
│       ├── Open API SNAP BI (PADG 23/15/2021)
│       ├── OAuth 2.0 + OpenID Connect
│       └── Disbursement via Virtual Account
│
├── SUPPORT FEATURES (Should Have — Post-MVP)
│   ├── Regulator Dashboard (OJK, Kemenkop)
│   ├── Multi-cabang consolidation
│   ├── Fraud alert system
│   ├── Predictive insights (stok, cashflow)
│   └── Supply chain traceability
│
└── FUTURE FEATURES (Nice to Have — Roadmap)
    ├── Trust Network (reputasi koperasi lintas platform)
    ├── Sustainability certification integration
    ├── Expansion: BUMDes, UMKM, Kelompok Tani
    └── WhatsApp bot untuk input transaksi
```

---

## 8. FUNCTIONAL REQUIREMENTS — 6 PILAR LUMBERA

### 8.1 Pilar 1: Tenant Context Engine (P1)

**Tujuan:** Mengisolasi data setiap koperasi secara teknis sehingga tidak ada risiko data tercampur atau kebocoran lintas koperasi.

**Requirement Teknis:**

| ID        | Requirement                                            | Detail                                                                                                                                 | Prioritas |
| --------- | ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| P1-FR-001 | **Tenant Isolation via tenant_id**                     | Setiap record di semua tabel memiliki kolom `tenant_id` (UUID). Tidak ada data yang dapat diakses tanpa tenant_id yang valid.          | MUST      |
| P1-FR-002 | **Row-Level Security (RLS) PostgreSQL**                | RLS policy aktif di semua tabel utama. Query tanpa tenant context otomatis menghasilkan empty result set, bukan error.                 | MUST      |
| P1-FR-003 | **Middleware Tenant Validation**                       | Setiap API request harus menyertakan tenant context di header atau JWT claim. Request tanpa tenant context mendapat response HTTP 403. | MUST      |
| P1-FR-004 | **Tenant Onboarding Wizard**                           | Admin baru dapat membuat tenant koperasi baru dalam <15 menit via wizard 7 langkah.                                                    | MUST      |
| P1-FR-005 | **Tenant-level Configuration**                         | Setiap tenant dapat mengkonfigurasi: nama koperasi, tipe koperasi (untuk ARS), logo, tahun fiskal, mata uang                           | SHOULD    |
| P1-FR-006 | **Cross-tenant Data Sharing** (khusus untuk Regulator) | Data agregat CHS dapat di-export ke Regulator Dashboard tanpa mengekspos data individual anggota (anonymized).                         | SHOULD    |

**Kriteria Keberhasilan:**

- Zero data leakage lintas tenant dalam 6 bulan operasional pilot (verifikasi via penetration test)
- 100% query ke database mengandung filter tenant_id (verifikasi via static code analysis)

---

### 8.2 Pilar 2: Adaptive Reporting Schema (ARS)

**Tujuan:** Menyediakan template laporan yang dapat dipahami oleh pengurus non-akuntan, disesuaikan dengan jenis komoditas/usaha koperasi.

**Requirement Teknis:**

| ID        | Requirement                    | Detail                                                                                                                                                                                                   | Prioritas    |
| --------- | ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| P2-FR-001 | **6 Template Komoditas**       | Template tersedia: (1) KSP, (2) Pangan Bulky, (3) Cold-Chain Perishable, (4) Toko Gerai, (5) Utility Subscription, (6) Peternakan. Setiap template memiliki field, formula, dan terminologi yang sesuai. | MUST         |
| P2-FR-002 | **Auto-render Glosarium**      | Setiap field dalam template memiliki tooltip/label dalam bahasa pengurus (bukan istilah akuntansi). Contoh: "Modal Usaha" bukan "Ekuitas".                                                               | MUST         |
| P2-FR-003 | **Laporan Otomatis**           | Generate laporan Neraca, Laba Rugi, Arus Kas secara otomatis berdasarkan transaksi yang sudah terinput. Waktu generate <30 detik.                                                                        | MUST         |
| P2-FR-004 | **Format Ekspor**              | Laporan dapat diekspor sebagai PDF (siap cetak) dan Excel (.xlsx, dapat diedit).                                                                                                                         | MUST         |
| P2-FR-005 | **Compliance Format Kemenkop** | Format laporan sesuai standar Permenkop UKM No. 8 Tahun 2023 tentang Usaha Simpan Pinjam Koperasi.                                                                                                       | MUST         |
| P2-FR-006 | **Template Custom**            | Admin platform dapat membuat template baru via drag-and-drop schema builder.                                                                                                                             | NICE TO HAVE |

**Acceptance Criteria:**

- Laporan yang sama yang sebelumnya membutuhkan 3 hari kini selesai dalam 30 menit (verifikasi via time-study dengan pengurus Koperasi Padiwangi)
- 80% pengurus pilot menyatakan laporan "mudah dipahami" dalam survei pasca-training (Likert scale ≥4/5)

---

### 8.3 Pilar 3: Offline-First PWA

**Tujuan:** Memastikan pengurus di daerah 3T dengan sinyal tidak stabil dapat mencatat transaksi tanpa gangguan, dengan sinkronisasi otomatis saat koneksi tersedia.

**Requirement Teknis:**

| ID        | Requirement                  | Detail                                                                                                                                                              | Prioritas |
| --------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| P3-FR-001 | **Offline Capability**       | Semua fungsi transaksi utama (simpanan, pinjaman, angsuran, mutasi stok) dapat dilakukan tanpa koneksi internet. Menggunakan IndexedDB via Dexie.js.                | MUST      |
| P3-FR-002 | **Background Sync**          | Transaksi yang disimpan offline otomatis tersinkron ke server dalam <30 detik setelah koneksi tersedia. Menggunakan Background Sync API (Mozilla MDN, 2024).        | MUST      |
| P3-FR-003 | **Conflict Resolution**      | Jika ada konflik data (dua perangkat input data yang sama secara offline), sistem menggunakan strategi Last-Write-Wins dengan notifikasi ke pengurus.               | MUST      |
| P3-FR-004 | **PWA Install Prompt**       | Pengguna dapat menginstall LUMBERA sebagai app dari browser (add to home screen) tanpa perlu download dari Play Store.                                              | MUST      |
| P3-FR-005 | **Offline Status Indicator** | Indikator visual yang jelas menunjukkan status: Online/Offline/Syncing. Badge jumlah transaksi pending ditampilkan.                                                 | MUST      |
| P3-FR-006 | **Data Storage Limit**       | IndexedDB menggunakan max 50MB per device. Transaksi >90 hari di-archive ke cloud. Warning muncul saat storage >80%.                                                | SHOULD    |
| P3-FR-007 | **Voice Input**              | Input nominal transaksi menggunakan speech-to-text (Google Speech API) dengan akurasi ≥85% untuk Bahasa Indonesia. Tombol microphone tersedia di semua field angka. | SHOULD    |
| P3-FR-008 | **Low-bandwidth Mode**       | Mode khusus untuk koneksi 2G: kompresi gambar, minimalkan payload API, text-only UI.                                                                                | SHOULD    |

**Device Support:**

- Android 7.0+ (API level 24+)
- iOS 14+
- Browser: Chrome 80+, Firefox 78+, Safari 14+
- Minimum RAM: 1GB
- Minimum storage: 100MB

---

### 8.4 Pilar 4: Verifiable Ledger

**Tujuan:** Membangun fondasi kepercayaan data yang tidak dapat dimanipulasi, sehingga catatan operasional koperasi memiliki kredibilitas setara dengan laporan keuangan yang diaudit.

**Requirement Teknis:**

| ID        | Requirement             | Detail                                                                                                                                                                  | Prioritas |
| --------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| P4-FR-001 | **Hash per Transaksi**  | Setiap transaksi yang disimpan menghasilkan SHA-256 hash yang mencakup: (prev_hash, transaction_data, timestamp, user_id).                                              | MUST      |
| P4-FR-002 | **Merkle Chain**        | Hash setiap transaksi di-link ke hash transaksi sebelumnya, membentuk Merkle chain. Perusakan satu record akan memutus chain dan terdeteksi otomatis.                   | MUST      |
| P4-FR-003 | **Tamper Detection**    | Background job berjalan setiap 6 jam untuk memverifikasi integritas seluruh chain. Alert dikirim ke admin platform jika ada kerusakan chain.                            | MUST      |
| P4-FR-004 | **Immutable Audit Log** | Semua perubahan data (UPDATE, DELETE) dilarang. Koreksi dilakukan via transaksi reversal dengan referensi ke transaksi asal.                                            | MUST      |
| P4-FR-005 | **Blockchain Anchor**   | Root hash Merkle tree di-anchor ke permissioned blockchain (Hyperledger Fabric) setiap 24 jam. Anchor tersimpan dengan timestamp dan block height.                      | MUST      |
| P4-FR-006 | **Verification API**    | Endpoint `GET /api/v1/ledger/{cooperative_id}/verify` mengembalikan: status valid/invalid, root hash, last anchor timestamp, blockchain tx ID. Waktu response <2 detik. | MUST      |
| P4-FR-007 | **Periodic Anchoring**  | Selain anchor harian otomatis, pengurus dapat memicu anchor manual untuk momen penting (sebelum pengajuan kredit).                                                      | SHOULD    |
| P4-FR-008 | **Export Ledger Proof** | Pengurus dapat mengunduh "Ledger Proof Certificate" sebagai PDF yang berisi merkle root dan blockchain proof, untuk diserahkan ke mitra pembiayaan.                     | SHOULD    |

---

### 8.5 Pilar 5: Cooperative Credit Intelligence (CCI)

**Tujuan:** Menghasilkan penilaian kredit yang akurat, transparan, dan patuh regulasi untuk anggota koperasi yang selama ini unbankable (no-file/thin-file).

#### 8.5.1 Member Creditworthiness Score (MCS)

**Model:**

- Algoritma: XGBoost (sebagai baseline, sesuai standar industri — TransUnion, 2022)
- Skala: 300 – 850 (mengadopsi konvensi FICO untuk interoperabilitas)
- Explainability: SHAP values wajib untuk setiap prediksi (sesuai prinsip transparansi POJK 29/2024)

**Bobot 5C:**

| Komponen             | Bobot | Fitur Data (Input)                                                                         |
| -------------------- | ----- | ------------------------------------------------------------------------------------------ |
| **Character (35%)**  | 35%   | Konsistensi setoran, frekuensi transaksi, riwayat kehadiran RAT, lama keanggotaan          |
| **Capacity (30%)**   | 30%   | Rasio pinjaman/pendapatan, riwayat angsuran (tepat waktu vs terlambat), frekuensi pinjaman |
| **Capital (15%)**    | 15%   | Jumlah simpanan pokok/wajib, total aset yang dilaporkan, saldo rata-rata 3 bulan terakhir  |
| **Conditions (12%)** | 12%   | Sektor usaha anggota, kondisi harga komoditas, lokasi geografis, seasonality usaha         |
| **Collateral (8%)**  | 8%    | Jaminan yang dilaporkan (tanah, kendaraan, dll.), nilai jaminan relatif terhadap pinjaman  |

**Grade MCS:**

| Skor    | Grade | Interpretasi    | Akses Pembiayaan                    |
| ------- | ----- | --------------- | ----------------------------------- |
| 750–850 | AA    | Sangat Baik     | Prioritas, bunga terendah           |
| 680–749 | A     | Baik            | Layak, bunga standar                |
| 580–679 | B     | Cukup           | Kondisional, perlu jaminan tambahan |
| 480–579 | C     | Perlu Perhatian | Terbatas, review manual             |
| 300–479 | D     | Buruk           | Ditolak, perlu restrukturisasi      |

#### 8.5.2 Cooperative Health Score (CHS)

**Model:**

- Skala: 0 – 100
- Grade: AA, A, B, C, D

**Komponen CHS:**

| Komponen              | Bobot | Indikator                                                       |
| --------------------- | ----- | --------------------------------------------------------------- |
| Kesehatan Keuangan    | 35%   | NPL rate, rasio kecukupan modal, ROA, likuiditas                |
| Aktivitas Operasional | 25%   | Volume transaksi aktif, pertumbuhan anggota, utilisasi pinjaman |
| Kualitas Data         | 20%   | Completeness rate data, frekuensi update, konsistensi ledger    |
| Kepatuhan Pelaporan   | 20%   | Ketepatan waktu laporan, kelengkapan dokumen regulasi           |

**Requirement Teknis CCI:**

| ID        | Requirement                   | Detail                                                                                                     | Prioritas |
| --------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------- | --------- |
| P5-FR-001 | **MCS Real-time Calculation** | MCS diperbarui dalam <5 menit setelah setiap transaksi baru.                                               | MUST      |
| P5-FR-002 | **CHS Batch Calculation**     | CHS dikalkulasi ulang setiap malam (batch job Apache Airflow, pukul 01:00 WIB).                            | MUST      |
| P5-FR-003 | **SHAP Explanation**          | Setiap MCS dilengkapi SHAP values yang diterjemahkan menjadi kalimat bahasa Indonesia sederhana.           | MUST      |
| P5-FR-004 | **Model Registry**            | Semua model CCI diregistrasi di MLflow dengan metadata: versi, tanggal training, metrik akurasi.           | MUST      |
| P5-FR-005 | **Model Evaluation Metrics**  | Model dievaluasi menggunakan: Gini ≥0.55, AUC-ROC ≥0.75, KS Statistic ≥0.30, PSI <0.25 (stabilitas model). | MUST      |
| P5-FR-006 | **Model Recalibration**       | Model di-retrain setiap kuartal atau saat PSI >0.25 (population shift).                                    | SHOULD    |
| P5-FR-007 | **Fraud Anomaly Detection**   | Model TensorFlow berjalan setiap batch harian. Alert ke pengurus jika confidence fraud >0.85.              | SHOULD    |
| P5-FR-008 | **POJK 29/2024 Compliance**   | Model terdokumentasi sesuai kerangka Pemeringkat Kredit Alternatif POJK 29/2024.                           | MUST      |

---

### 8.6 Pilar 6: Consent-Gated Open Bridge

**Tujuan:** Menghubungkan koperasi dengan ekosistem pembiayaan formal secara aman, dengan anggota sebagai pemilik data yang berdaulat.

**Requirement Teknis:**

| ID        | Requirement                      | Detail                                                                                                                                                                               | Prioritas |
| --------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------- |
| P6-FR-001 | **Granular Consent**             | Anggota dapat memberikan consent untuk: (a) data tertentu saja, (b) mitra tertentu saja, (c) durasi tertentu. Bukan "accept all".                                                    | MUST      |
| P6-FR-002 | **Revocable Consent**            | Consent dapat dicabut kapan saja. Setelah dicabut, mitra tidak dapat melakukan query baru. Data historis yang sudah diakses tidak dapat di-delete dari sistem mitra (sesuai UU PDP). | MUST      |
| P6-FR-003 | **Consent Audit Log**            | Setiap aksi consent (diberikan, dicabut, diakses) dicatat dengan: user_id, timestamp, IP address, device fingerprint.                                                                | MUST      |
| P6-FR-004 | **SNAP BI API**                  | Open API mengikuti PADG No. 23/15/2021 SNAP Bank Indonesia untuk standardisasi.                                                                                                      | MUST      |
| P6-FR-005 | **OAuth 2.0 + OIDC**             | Authentication menggunakan OAuth 2.0 Authorization Code Flow + PKCE dan OpenID Connect.                                                                                              | MUST      |
| P6-FR-006 | **API Rate Limiting**            | Default: 1.000 request/hari per API key. Enterprise: 10.000 request/hari. Burst: 100 request/menit.                                                                                  | MUST      |
| P6-FR-007 | **Data Passport UI**             | Anggota dapat mengakses "Data Passport" — tampilan lengkap semua data tersimpan dan siapa yang pernah mengaksesnya.                                                                  | MUST      |
| P6-FR-008 | **Developer Portal**             | Dokumentasi API (OpenAPI 3.0), sandbox environment, dan getting started guide tersedia publik.                                                                                       | SHOULD    |
| P6-FR-009 | **Webhook Notifications**        | Mitra dapat mendaftarkan webhook untuk notifikasi: skor kredit berubah, consent dicabut, transaksi baru.                                                                             | SHOULD    |
| P6-FR-010 | **Virtual Account Disbursement** | Dana dari mitra dapat langsung dicairkan ke Virtual Account anggota setelah consent diberikan dan kredit disetujui. Proses disbursement <1 jam.                                      | SHOULD    |

---

## 9. NON-FUNCTIONAL REQUIREMENTS

### 9.1 Performance

| Metric                      | Target                                      | Measurement Method              |
| --------------------------- | ------------------------------------------- | ------------------------------- |
| API Response Time (p95)     | <500ms untuk 95% request                    | Prometheus + Grafana monitoring |
| API Response Time (p99)     | <2 detik untuk 99% request                  | APM tool                        |
| Page Load Time (mobile, 3G) | <3 detik First Contentful Paint             | Lighthouse score ≥70            |
| Offline Sync Delay          | <30 detik setelah koneksi tersedia          | Automated sync log analysis     |
| CCI Score Calculation       | <5 menit per anggota setelah transaksi baru | MLflow tracking                 |
| Ledger Verification         | <2 detik per request                        | API monitoring                  |
| System Availability         | 99.5% uptime per bulan                      | Status page monitoring          |
| Concurrent Users            | Support 5.000 concurrent users              | Load testing (k6)               |

### 9.2 Security

| Requirement                    | Detail                                                                    |
| ------------------------------ | ------------------------------------------------------------------------- |
| **Data Encryption at Rest**    | AES-256 untuk semua data sensitif di database                             |
| **Data Encryption in Transit** | TLS 1.3 untuk semua komunikasi API                                        |
| **Password Hashing**           | bcrypt dengan salt factor ≥12                                             |
| **JWT Security**               | Access token: 15 menit. Refresh token: 30 hari. Rotation on use.          |
| **SQL Injection Prevention**   | Prepared statements wajib. ORM Gorm dengan parameterized queries.         |
| **CSRF Protection**            | CSRF token wajib untuk semua form yang melakukan state change             |
| **Rate Limiting**              | Auth endpoints: 5 request/menit per IP. Lockout 15 menit setelah 5 gagal. |
| **Security Audit**             | Penetration testing sebelum MVP launch dan setiap 6 bulan                 |
| **OWASP Top 10**               | Semua vulnerability OWASP Top 10 harus dimitigasi sebelum go-live         |

### 9.3 Data Privacy & Compliance

| Requirement                | Detail                                                                                                                       |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **UU 27/2022 (UU PDP)**    | Data retention policy: data transaksi max 7 tahun. Right to access dan right to erasure (kecuali data ledger yang immutable) |
| **POJK 29/2024**           | Model CCI terdokumentasi dan dapat diaudit oleh OJK                                                                          |
| **POJK 40/2024 Pasal 150** | Credit scoring berbasis data relevan, transparan, dan dapat dijelaskan                                                       |
| **PADG 23/15/2021**        | API mengikuti standar SNAP Bank Indonesia                                                                                    |
| **Inpres No. 9/2025**      | Sistem mendukung digitalisasi Koperasi Desa Merah Putih                                                                      |
| **Data Minimization**      | Hanya mengumpulkan data yang dibutuhkan untuk scoring (tidak ada data tidak relevan)                                         |
| **Consent Documentation**  | Setiap consent terdokumentasi dan dapat dibuktikan (timestamped, signed)                                                     |

### 9.4 Scalability

| Scenario                    | Requirement                                              |
| --------------------------- | -------------------------------------------------------- |
| **Year 1 (50 koperasi)**    | Database single node, horizontal scaling ready           |
| **Year 3 (200 koperasi)**   | Read replicas untuk database, CDN untuk static assets    |
| **Year 5 (1.550 koperasi)** | Sharding database per region, multi-region deployment    |
| **Data Volume**             | Sistem mampu menangani 10 juta transaksi per tahun       |
| **Storage**                 | Auto-scaling cloud storage (awalnya min 1TB, auto-scale) |

### 9.5 Reliability & Disaster Recovery

| Requirement                        | Detail                                                     |
| ---------------------------------- | ---------------------------------------------------------- |
| **RTO (Recovery Time Objective)**  | <4 jam untuk pemulihan penuh setelah kegagalan             |
| **RPO (Recovery Point Objective)** | <1 jam (maximum data yang boleh hilang)                    |
| **Database Backup**                | Full backup harian, incremental backup setiap jam          |
| **Backup Retention**               | 30 hari untuk backup reguler, 7 tahun untuk ledger archive |
| **Geographic Redundancy**          | Primary: Jakarta. Secondary: Surabaya. Failover otomatis.  |

---

## 10. UX & DESIGN REQUIREMENTS

### 10.1 Design Principles

1. **Mobile-First Rural Design** — Dirancang untuk smartphone entry-level pengurus desa, bukan laptop kantor.
2. **Icon-Based Navigation** — Setiap fungsi utama direpresentasikan dengan ikon bergambar jelas, bukan teks semata.
3. **3-Tap Maximum** — Transaksi yang paling sering dilakukan harus bisa diselesaikan dalam maksimal 3 tap.
4. **Error Prevention > Error Recovery** — Sistem harus mencegah kesalahan terjadi (konfirmasi, preview) daripada hanya memberikan pesan error.
5. **Accessible Language** — Semua teks UI dalam Bahasa Indonesia yang sederhana dan familiar bagi pengurus senior non-IT.
6. **Offline-Aware UI** — UI selalu menampilkan status koneksi dan memberikan feedback yang tepat untuk mode offline.

### 10.2 Accessibility Requirements

| Requirement           | Standard                                                          |
| --------------------- | ----------------------------------------------------------------- |
| **Font Size**         | Minimum 16px untuk body text, 20px untuk label input              |
| **Touch Target Size** | Minimum 44×44dp untuk semua elemen interaktif (Apple HIG)         |
| **Color Contrast**    | WCAG AA compliant (4.5:1 untuk normal text, 3:1 untuk large text) |
| **Color Blind Safe**  | Tidak menggunakan warna sebagai satu-satunya indikator status     |
| **Screen Reader**     | ARIA labels untuk semua elemen interaktif                         |
| **Dark Mode**         | Opsional untuk mengurangi konsumsi baterai                        |

### 10.3 User Interface Specifications

**Home Screen — Pengurus:**

- Grid ikon 2×2 untuk 4 fungsi utama: Simpanan, Pinjaman, Stok, Laporan
- Widget status sync (online/offline/pending)
- Notifikasi badge untuk tugas tertunda
- Shortcut ke transaksi terakhir (recent transactions)

**Transaksi Flow:**

1. Tap ikon jenis transaksi
2. Cari/pilih anggota (autocomplete + foto profil)
3. Input nominal (keyboard numerik besar + voice option)
4. Preview ringkasan transaksi
5. Konfirmasi (tombol besar "SIMPAN")
6. Feedback sukses (animasi + suara)

**Member Scorecard (untuk Anggota):**

- Gauge chart untuk MCS (warna hijau–merah)
- Label Grade yang mudah dimengerti ("Profil Kredit Anda: BAIK")
- 3 faktor positif + 3 faktor negatif dalam bahasa sederhana
- Saran perbaikan yang actionable

**Consent Dialog:**

- Nama mitra pembiayaan dengan logo
- Checklist data yang akan dibagikan
- Durasi consent (dropdown: 7 hari / 30 hari / 90 hari / custom)
- Tujuan penggunaan data
- Link ke kebijakan privasi mitra
- Tombol "Setuju" (hijau) dan "Tolak" (abu-abu, sama besar)

### 10.4 Journey Map — Pengurus Koperasi (Pak Asep)

| Phase           | Action                                           | Touchpoint                     | Emotion                | Pain Points                      | Opportunities                                   |
| --------------- | ------------------------------------------------ | ------------------------------ | ---------------------- | -------------------------------- | ----------------------------------------------- |
| **Awareness**   | Mendengar tentang LUMBERA dari penyuluh Kemenkop | Sosialisasi tatap muka, brosur | Skeptis tapi penasaran | "Apa bedanya dengan Excel?"      | Demo live dengan data koperasi sendiri          |
| **Onboarding**  | Download PWA, daftar koperasi                    | LUMBERA PWA (mobile)           | Cemas, takut salah     | Form onboarding terlalu panjang  | Wizard langkah demi langkah, bisa dijeda        |
| **Adoption**    | Coba input transaksi pertama                     | LUMBERA PWA (offline)          | Lega, surprised        | "Kok bisa input tanpa internet?" | Tutorial interaktif in-app, progress bar        |
| **Regular Use** | Input transaksi harian, generate laporan         | LUMBERA PWA                    | Percaya diri, efisien  | Khawatir data hilang             | Backup confirmation visual setiap hari          |
| **Advocacy**    | Rekomendasikan ke pengurus koperasi lain         | Word of mouth, WhatsApp group  | Bangga, puas           | —                                | Program referral, sertifikat "Koperasi Digital" |

### 10.5 Journey Map — Anggota Koperasi (Bu Lestari)

| Phase              | Action                                               | Touchpoint              | Emotion               | Pain Points                   | Opportunities                                         |
| ------------------ | ---------------------------------------------------- | ----------------------- | --------------------- | ----------------------------- | ----------------------------------------------------- |
| **Awareness**      | Diberitahu Pak Asep bahwa koperasi pakai sistem baru | Langsung/tatap muka     | Netral                | "Untuk saya ada manfaatnya?"  | Notifikasi SMS "Saldo Anda bisa dicek di LUMBERA"     |
| **First Use**      | Cek saldo tabungan pertama kali                      | LUMBERA PWA             | Senang, surprised     | Tidak tahu cara login         | Login via nomor HP + OTP (tanpa perlu ingat password) |
| **Credit Journey** | Ajukan pinjaman, lihat skor MCS                      | LUMBERA PWA             | Khawatir tentang skor | "Skor saya bagus atau jelek?" | Label grade yang jelas + penjelasan sederhana         |
| **Consent Moment** | Diminta consent untuk kirim data ke Akseleran        | Consent dialog PWA      | Ragu, takut           | "Data saya aman?"             | Trust indicators: logo OJK, nama mitra terdaftar      |
| **Disbursement**   | Terima notifikasi dana cair ke Virtual Account       | SMS + Push notification | Sangat senang, lega   | —                             | Konfirmasi disbursement + rincian kredit yang jelas   |

---

## 11. TECH STACK & ARSITEKTUR SISTEM

### 11.1 Tech Stack Detail

| Layer                 | Technology                                  | Versi             | Justifikasi                                               |
| --------------------- | ------------------------------------------- | ----------------- | --------------------------------------------------------- |
| **Frontend PWA**      | Next.js                                     | 14.x              | SSR native, PWA support, excellent DX                     |
| **Frontend Language** | TypeScript                                  | 5.x               | Type safety untuk reduce runtime errors                   |
| **UI Framework**      | TailwindCSS + shadcn/ui                     | Latest            | Rapid responsive UI development                           |
| **Offline Storage**   | Dexie.js (IndexedDB wrapper)                | 3.x               | Mature offline-first solution (Mozilla MDN, 2024)         |
| **Service Worker**    | Workbox                                     | 7.x               | Standar industri untuk PWA caching & sync                 |
| **Backend API**       | Golang + Gin                                | Go 1.22 / Gin 1.9 | High performance, low memory, ideal untuk API koperasi    |
| **Database**          | MariaDB                                     | 10.11 LTS         | Stable, open-source, battle-tested transactional workload |
| **Cache**             | Redis                                       | 7.x               | Session store, rate limiting, leaderboard CHS             |
| **Message Queue**     | RabbitMQ                                    | 3.x               | Async processing untuk background jobs                    |
| **Ledger Hash**       | Custom SHA-256 Merkle Tree                  | —                 | Off-chain first, hindari biaya gas publik blockchain      |
| **Blockchain Anchor** | Hyperledger Fabric                          | 2.5 LTS           | Permissioned, ideal untuk konsorsium koperasi-bank        |
| **AI/ML**             | Python + scikit-learn + XGBoost             | Python 3.11       | Baseline industri credit scoring (TransUnion, 2022)       |
| **AI Explainability** | SHAP                                        | 0.44              | Wajib untuk compliance POJK 29/2024                       |
| **Fraud Detection**   | TensorFlow                                  | 2.16              | LSTM untuk anomali detection time-series transaksi        |
| **Model Registry**    | MLflow                                      | 2.x               | Model versioning & recalibration tracking                 |
| **Orchestration**     | Apache Airflow                              | 2.9               | Schedulter batch jobs (CHS, audit, anchoring)             |
| **Container**         | Docker + Kubernetes                         | K8s 1.29          | Horizontal scaling, blue-green deployment                 |
| **Cloud**             | Google Cloud Platform (GCP)                 | —                 | GKE, Cloud SQL, Cloud Storage, Cloud CDN                  |
| **Auth**              | OAuth 2.0 + OpenID Connect                  | RFC 6749          | Standard auth untuk Open Bridge                           |
| **API Standard**      | REST (OpenAPI 3.0) + SNAP BI                | PADG 23/15/2021   | Standardisasi untuk mitra pembiayaan                      |
| **Monitoring**        | Prometheus + Grafana                        | Latest            | Real-time metrics dashboard                               |
| **Logging**           | ELK Stack (Elasticsearch, Logstash, Kibana) | 8.x               | Centralized logging & audit trail                         |
| **CI/CD**             | GitHub Actions                              | —                 | Automated testing & deployment pipeline                   |

### 11.2 Arsitektur Sistem (5 Lapisan)

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENT / PENGGUNA                            │
│  [Pengurus Koperasi] [Anggota] [Mitra Pembiayaan] [Regulator]   │
│         [Web / Mobile PWA] [API Client] [Regulator Dashboard]   │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│                    EDGE / CDN LAYER                              │
│     Cloud CDN (Google Cloud CDN) — Static assets, caching       │
│     DDoS Protection + WAF (Web Application Firewall)            │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│               FRONTEND APPLICATION LAYER                         │
│  Next.js 14 (SSR + PWA)                                         │
│  ├── Service Worker (Workbox) — Cache, offline sync             │
│  ├── IndexedDB (Dexie.js) — Local transaction storage           │
│  ├── Background Sync API — Auto-sync when online                │
│  └── Adaptive Reporting Schema — Template rendering             │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│                APPLICATION / API LAYER                           │
│  Load Balancer (Google Cloud Load Balancing)                    │
│  │                                                               │
│  ├── Backend API (Golang + Gin)                                  │
│  │   ├── Tenant Context Engine (middleware)                      │
│  │   ├── Consent Manager & Policy Engine                         │
│  │   └── Open API Gateway (SNAP BI)                             │
│  │                                                               │
│  ├── AI Service Layer (Python FastAPI)                           │
│  │   ├── CCI Scoring Service (XGBoost + SHAP)                   │
│  │   └── Fraud Detection Service (TensorFlow)                   │
│  │                                                               │
│  └── Sync & Worker Layer                                         │
│      ├── Background Sync Jobs (Apache Airflow)                   │
│      ├── Verifiable Ledger Service                              │
│      └── Blockchain Anchoring Worker                            │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│                    DATA LAYER                                     │
│  ├── MariaDB (Primary) — Transactional data, multi-tenant       │
│  ├── MariaDB Read Replica — Analytics, reporting queries        │
│  ├── Redis — Session, cache, rate limiting                       │
│  ├── Elasticsearch — Audit log search, full-text member search   │
│  └── Cloud Storage — Ledger archive, reports, exports           │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│              EXTERNAL ECOSYSTEM (MITRA EKSTERNAL)                │
│  ├── Hyperledger Fabric (Blockchain Anchor)                      │
│  ├── Fintech Lending (via SNAP BI Open API)                      │
│  ├── BPR/BFI (via SNAP BI Open API)                              │
│  ├── Bank Indonesia (SNAP Standard compliance)                   │
│  ├── OJK (Regulator Dashboard, pelaporan)                        │
│  └── Kemenkop UKM (Data reporting, statistik nasional)           │
└─────────────────────────────────────────────────────────────────┘
```

### 11.3 Database Schema (Core Tables)

```sql
-- Tenant Management
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cooperative_name VARCHAR(255) NOT NULL,
    cooperative_type ENUM('KSP','PANGAN_BULKY','COLD_CHAIN','TOKO_GERAI','UTILITY','PETERNAKAN'),
    registration_number VARCHAR(50) UNIQUE,
    province_code VARCHAR(10),
    city_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Members (with RLS)
CREATE TABLE members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    full_name VARCHAR(255) NOT NULL,
    nik VARCHAR(16) UNIQUE,  -- Nomor Induk Kependudukan (encrypted)
    phone_number VARCHAR(15),
    joined_date DATE,
    member_status ENUM('ACTIVE','INACTIVE','SUSPENDED'),
    current_mcs_score INTEGER,
    mcs_grade ENUM('AA','A','B','C','D'),
    last_score_updated_at TIMESTAMP
);
-- RLS Policy
ALTER TABLE members ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON members FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Transactions (Ledger)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    member_id UUID NOT NULL REFERENCES members(id),
    transaction_type ENUM('SIMPANAN_POKOK','SIMPANAN_WAJIB','SIMPANAN_SUKARELA','PINJAMAN','ANGSURAN','MUTASI_STOK'),
    amount DECIMAL(15,2) NOT NULL,
    description TEXT,
    officer_id UUID NOT NULL,  -- Pengurus yang mencatat
    recorded_at TIMESTAMP NOT NULL,  -- Waktu input (bisa offline)
    synced_at TIMESTAMP,  -- Waktu sync ke server
    prev_hash VARCHAR(64) NOT NULL,  -- Hash transaksi sebelumnya
    current_hash VARCHAR(64) UNIQUE NOT NULL,  -- SHA-256 hash record ini
    is_offline_created BOOLEAN DEFAULT FALSE
);

-- Consent Records
CREATE TABLE consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    member_id UUID NOT NULL REFERENCES members(id),
    partner_id UUID NOT NULL REFERENCES partners(id),
    data_scope JSONB NOT NULL,  -- Data apa yang di-consent
    purpose TEXT NOT NULL,
    duration_days INTEGER NOT NULL,
    granted_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    consent_signature VARCHAR(512)  -- Digital signature anggota
);
```

---

## 12. INTEGRASI EKSTERNAL

### 12.1 Hyperledger Fabric (Blockchain Anchor)

| Aspek                | Detail                                                                |
| -------------------- | --------------------------------------------------------------------- |
| **Versi**            | Hyperledger Fabric 2.5 LTS                                            |
| **Channel**          | Dedicated channel per consortium (LUMBERA-Koperasi-Network)           |
| **Chaincode**        | AnchorRecord (Go) — menerima rootHash dan menyimpan dengan timestamp  |
| **Anchor Frequency** | Otomatis setiap 24 jam (Apache Airflow trigger)                       |
| **Verification**     | API endpoint publik untuk verifikasi anchor hash                      |
| **Fallback**         | Jika Hyperledger unavailable, anchor ke Polygon Edge (EVM-compatible) |

### 12.2 SNAP Bank Indonesia (Open API)

| Aspek               | Detail                                                           |
| ------------------- | ---------------------------------------------------------------- |
| **Standard**        | PADG No. 23/15/2021 tentang Standar Nasional Open API Pembayaran |
| **Auth Flow**       | OAuth 2.0 Authorization Code + PKCE                              |
| **Token**           | Access Token (15 menit) + Refresh Token (30 hari)                |
| **API Format**      | REST, JSON, UTF-8 encoding                                       |
| **Versioning**      | URL-based (/v1/, /v2/)                                           |
| **Endpoints Utama** | /credit-score, /cooperative-health, /consent, /disbursement      |
| **Sandbox**         | https://sandbox.lumbera.id/api/v1/                               |
| **Production**      | https://api.lumbera.id/api/v1/                                   |

### 12.3 Mitra Pembiayaan yang Ditargetkan

| Mitra                  | Tipe            | Status | Prioritas Integrasi          |
| ---------------------- | --------------- | ------ | ---------------------------- |
| Akseleran              | P2P Lending OJK | Target | Prioritas 1 (MoU Target MVP) |
| KoinWorks              | P2P Lending OJK | Target | Prioritas 2                  |
| Bank Sahabat Sampoerna | BPR Nasional    | Target | Prioritas 3                  |
| BPR/BPD Daerah         | BPR Regional    | Target | Prioritas 4                  |
| Investree              | P2P Lending OJK | Target | Prioritas 5                  |

---

## 13. BUSINESS REQUIREMENTS & REVENUE MODEL

### 13.1 Revenue Streams

**Stream 1: Subscription Tier**

| Tier           | Target                               | Harga/bulan  | Fitur                                                                  |
| -------------- | ------------------------------------ | ------------ | ---------------------------------------------------------------------- |
| **Free**       | Koperasi <100 anggota                | Rp 0         | Pencatatan transaksi, laporan dasar, max 2 pengguna                    |
| **Pro**        | Koperasi 100–500 anggota             | Rp 500.000   | Semua fitur + Verifiable Ledger + CCI scoring                          |
| **Enterprise** | Koperasi >500 anggota / multi-cabang | Rp 2.000.000 | Semua fitur Pro + Open Bridge + Regulator Dashboard + Priority support |

**Stream 2: Success Fee Penyaluran Kredit**

- Fee: **1,5% – 2,5%** dari total nilai pinjaman yang berhasil disalurkan via Open Bridge
- Minimum transaksi: Rp 1 juta
- Penagihan: Real-time saat disbursement terkonfirmasi

**Stream 3: Credit-Scoring-as-a-Service**

- Tarif: **Rp 5.000 – Rp 15.000** per query skor kredit
- Target: Lembaga fintech dan perbankan yang membutuhkan data alternatif
- Compliance: Sesuai amanat POJK No. 29 Tahun 2024

### 13.2 Proyeksi Revenue 5 Tahun

| Tahun | Koperasi  | Revenue (Rp M) | OpEx (Rp M) | Net Cash Flow |
| ----- | --------- | -------------- | ----------- | ------------- |
| Y0    | 0         | 0              | 5,0         | -5,0          |
| Y1    | 5 (pilot) | 0,5            | 3,0         | -2,5          |
| Y2    | 50        | 4,0            | 2,0         | 2,0           |
| Y3    | 200       | 15,0           | 7,0         | 8,0           |
| Y4    | 700       | 5,0            | 10,0        | -5,0\*        |
| Y5    | 1.550     | 25,0           | 10,0        | 15,0          |

\*Y4 negatif karena ekspansi infrastruktur dan tim

**Financial Metrics:**

- NPV (12% discount rate): **Rp 5,48 miliar**
- ROI 5 tahun: **290%**
- Payback Period: **4,03 tahun**
- IRR: **24–26%**

### 13.3 Cost Structure

| Kategori                                | Alokasi Investasi Awal (Y0) | %        |
| --------------------------------------- | --------------------------- | -------- |
| Product Development (engineering)       | Rp 2,5 M                    | 50%      |
| Infrastructure (cloud, blockchain node) | Rp 1,0 M                    | 20%      |
| AI/ML Model Development                 | Rp 0,75 M                   | 15%      |
| Business Development & Partnership      | Rp 0,5 M                    | 10%      |
| Legal & Compliance                      | Rp 0,25 M                   | 5%       |
| **Total**                               | **Rp 5,0 M**                | **100%** |

---

## 14. COMPLIANCE & REGULATORY REQUIREMENTS

| Regulasi                                             | Implikasi Produk                                                  | Tim Responsible       | Deadline Compliance             |
| ---------------------------------------------------- | ----------------------------------------------------------------- | --------------------- | ------------------------------- |
| **POJK No. 29/2024** (Pemeringkat Kredit Alternatif) | Model CCI harus terdokumentasi, dapat diaudit, memiliki Gini ≥0.4 | AI/ML Team            | Sebelum launch Open Bridge      |
| **POJK No. 40/2024 Pasal 150**                       | Credit scoring berbasis data relevan, transparansi model          | AI/ML Team            | Sebelum launch Open Bridge      |
| **UU No. 27/2022 (UU PDP)**                          | Consent framework, data retention, right to access/erasure        | Backend + Legal       | MVP Launch                      |
| **PADG No. 23/15/2021 (SNAP BI)**                    | API standardisasi Open Bridge                                     | Backend Team          | Sebelum integrasi mitra pertama |
| **Inpres No. 9/2025 (Kopdes Merah Putih)**           | Platform mendukung digitalisasi koperasi desa                     | Product Team          | Ongoing                         |
| **Permenkop UKM No. 8/2023**                         | Format laporan keuangan sesuai standar koperasi simpan pinjam     | Reporting Module Team | MVP Launch                      |

### 14.1 Compliance Checklist (Pre-Launch)

- [ ] POJK 29/2024: Whitepaper CCI methodology selesai dan direvisi oleh legal advisor
- [ ] POJK 29/2024: Pengajuan registrasi Pemeringkat Kredit Alternatif ke OJK
- [ ] UU PDP: Privacy Policy dan Terms of Service disetujui oleh Data Protection Officer (DPO)
- [ ] UU PDP: Consent management system diaudit oleh pihak ketiga independen
- [ ] SNAP BI: API lulus compliance check Bank Indonesia
- [ ] Permenkop UKM No. 8/2023: Laporan yang dihasilkan divalidasi oleh akuntan publik
- [ ] Penetration testing selesai, zero critical vulnerability
- [ ] ISO 27001 Information Security assessment (target Year 2)

---

## 15. PRIORITISASI FITUR (RICE FRAMEWORK)

### Formula RICE

`RICE Score = (Reach × Impact × Confidence) / Effort`

**Scoring Rubric:**

- **Reach**: Jumlah pengguna terpengaruh per kuartal
- **Impact**: Massive(3), High(2), Medium(1), Low(0.5), Minimal(0.25)
- **Confidence**: High(1.0), Medium(0.8), Low(0.5)
- **Effort**: XS(1), S(3), M(5), L(8), XL(13) person-sprint

### RICE Scoring Tabel

| Fitur                                  | Reach                     | Impact      | Confidence   | Effort  | RICE Score | Prioritas        |
| -------------------------------------- | ------------------------- | ----------- | ------------ | ------- | ---------- | ---------------- |
| Offline transaction input (PWA)        | 5.000                     | Massive (3) | High (1.0)   | M (5)   | 3.000      | **Must Have**    |
| Verifiable Ledger (hash chain)         | 155.000 (koperasi target) | Massive (3) | High (1.0)   | L (8)   | 58.125     | **Must Have**    |
| Multi-tenant isolation (RLS)           | 5.000                     | Massive (3) | High (1.0)   | M (5)   | 3.000      | **Must Have**    |
| MCS AI scoring + SHAP                  | 5.000                     | Massive (3) | Medium (0.8) | L (8)   | 1.500      | **Must Have**    |
| Consent Manager + Data Passport        | 5.000                     | High (2)    | High (1.0)   | M (5)   | 2.000      | **Must Have**    |
| Open API SNAP BI                       | 50 (mitra)                | Massive (3) | Medium (0.8) | XL (13) | 9.23       | **Must Have**    |
| Adaptive Reporting Schema (6 template) | 5.000                     | High (2)    | High (1.0)   | L (8)   | 1.250      | **Must Have**    |
| CHS (Cooperative Health Score)         | 1.000 (pengurus)          | High (2)    | Medium (0.8) | M (5)   | 320        | **Must Have**    |
| Blockchain anchor (Hyperledger Fabric) | 155.000                   | Medium (1)  | Medium (0.8) | XL (13) | 9.538      | **Should Have**  |
| Voice Input untuk nominal              | 5.000                     | Medium (1)  | Medium (0.8) | S (3)   | 1.333      | **Should Have**  |
| Fraud anomaly detection                | 1.000                     | High (2)    | Low (0.5)    | XL (13) | 76.9       | **Should Have**  |
| Regulator Dashboard                    | 10 (OJK)                  | High (2)    | Medium (0.8) | L (8)   | 2          | **Should Have**  |
| Virtual Account Disbursement           | 5.000                     | Massive (3) | Low (0.5)    | XL (13) | 576.9      | **Should Have**  |
| Supply chain traceability              | 500                       | Low (0.5)   | Low (0.5)    | XL (13) | 9.6        | **Nice to Have** |
| WhatsApp bot integration               | 40.000                    | Medium (1)  | Low (0.5)    | XL (13) | 1.538      | **Nice to Have** |
| Trust Network lintas platform          | 155.000                   | High (2)    | Low (0.5)    | XL (13) | 11.923     | **Nice to Have** |

### Roadmap Sprint Priority (MVP — 6 Sprint)

| Sprint     | Fokus                                                 | Story Points | Deliverable                                 |
| ---------- | ----------------------------------------------------- | ------------ | ------------------------------------------- |
| Sprint 1–2 | Foundation: Tenant isolation, Auth, Offline PWA dasar | 40 SP        | Multi-tenant app dengan offline capability  |
| Sprint 3–4 | Core Transactions + Verifiable Ledger + ARS           | 45 SP        | Transaksi + hash chain + 2 template laporan |
| Sprint 5   | CCI (MCS + CHS) + Consent Manager                     | 35 SP        | Credit scoring live + consent dialog        |
| Sprint 6   | Open Bridge (SNAP BI) + Pilot Launch Prep             | 40 SP        | API mitra + demo end-to-end                 |

---

## 16. ROADMAP & MILESTONES

### 16.1 Phase 0: Foundation (Bulan 1–2)

| Milestone                                              | Target   | Acceptance Criteria                                                            |
| ------------------------------------------------------ | -------- | ------------------------------------------------------------------------------ |
| **M0.1** Development environment setup                 | Minggu 1 | Kubernetes cluster, CI/CD pipeline, database, monitoring stack siap            |
| **M0.2** Multi-tenant architecture selesai             | Minggu 3 | RLS PostgreSQL aktif, middleware tenant validation berjalan                    |
| **M0.3** Offline-First PWA basic                       | Minggu 5 | Service Worker + IndexedDB + Background Sync berfungsi                         |
| **M0.4** Authentication system                         | Minggu 6 | OAuth 2.0 + OIDC, JWT, rate limiting aktif                                     |
| **M0.5** ARS — 2 template pertama (KSP + Pangan Bulky) | Minggu 8 | Template KSP dan Pangan Bulky dapat digunakan untuk input dan generate laporan |

### 16.2 Phase 1: MVP (Bulan 3–4)

| Milestone                     | Target        | Acceptance Criteria                                                    |
| ----------------------------- | ------------- | ---------------------------------------------------------------------- |
| **M1.1** Verifiable Ledger v1 | Minggu 10     | SHA-256 hash chain aktif, tamper detection berjalan                    |
| **M1.2** CCI — MCS Scoring    | Minggu 12     | Model XGBoost trained, SHAP explanation aktif, API berfungsi           |
| **M1.3** Consent Manager v1   | Minggu 13     | Granular consent dialog, revocation, audit log aktif                   |
| **M1.4** MVP Internal Demo    | Akhir Bulan 4 | End-to-end demo: offline input → ledger → scoring → consent → Open API |

### 16.3 Phase 2: Pilot (Bulan 5–6)

| Milestone                               | Target        | Acceptance Criteria                                                                 |
| --------------------------------------- | ------------- | ----------------------------------------------------------------------------------- |
| **M2.1** Onboarding 5 koperasi pilot    | Minggu 17     | Koperasi Padiwangi, Melati Jaya, Sumber Makmur, Tirta Bersama, Harapan Baru onboard |
| **M2.2** Open Bridge API live           | Minggu 18     | Sandbox SNAP BI aktif, 1 mitra terhubung untuk testing                              |
| **M2.3** Blockchain Anchor aktif        | Minggu 20     | Anchor harian ke Hyperledger Fabric berjalan                                        |
| **M2.4** MoU dengan ≥1 mitra pembiayaan | Akhir Bulan 6 | Dokumen MoU ditandatangani                                                          |
| **M2.5** 1 kasus penyaluran kredit riil | Akhir Bulan 6 | Dana terverifikasi cair ke Virtual Account anggota                                  |

### 16.4 Phase 3: Scale (Bulan 7–12, Year 1)

| Milestone                          | Target   | Acceptance Criteria                             |
| ---------------------------------- | -------- | ----------------------------------------------- |
| **M3.1** CHS + Regulator Dashboard | Bulan 8  | Dashboard OJK live dengan data pilot            |
| **M3.2** Open-Source SDK Ledger    | Bulan 9  | GitHub repo publik, dokumentasi lengkap         |
| **M3.3** Whitepaper CCI            | Bulan 10 | Dokumen ilmiah selesai, pengajuan ke OJK        |
| **M3.4** Onboarding 50 koperasi    | Bulan 12 | 50 koperasi aktif, revenue positif              |
| **M3.5** ISO 27001 assessment      | Bulan 12 | Assessment selesai, gap analysis terdokumentasi |

### 16.5 Long-Term Roadmap (Year 2–5)

| Periode          | Focus                                                         | Target                                                 |
| ---------------- | ------------------------------------------------------------- | ------------------------------------------------------ |
| Y2 (Bulan 13–24) | Ekspansi geografis, penambahan template, integrasi mitra baru | 200 koperasi, 5 mitra pembiayaan, revenue positif      |
| Y3 (Bulan 25–36) | Platform maturity, produk lending koperasi lintas wilayah     | 500 koperasi, model CCI v2, Trust Network beta         |
| Y4 (Bulan 37–48) | Ekspansi ke BUMDes dan UMKM, integrasi perbankan besar        | 1.000 koperasi + BUMDes, 1 bank nasional terintegrasi  |
| Y5 (Bulan 49–60) | Target SOM: 1.550 koperasi, NPV target tercapai               | 1.550 koperasi, Rp 1,5 T penyaluran, IPO-ready metrics |

---

## 17. RISK REGISTER

### 17.1 Risk Matrix

| ID    | Risk                                                                         | Likelihood | Impact    | Severity     | Mitigation Strategy                                                                                           | Owner           |
| ----- | ---------------------------------------------------------------------------- | ---------- | --------- | ------------ | ------------------------------------------------------------------------------------------------------------- | --------------- |
| R-001 | **Adopsi lambat oleh pengurus senior** (resistensi digital)                  | High       | High      | **CRITICAL** | Program pendampingan tatap muka, gamified training, incentive koperasi early adopter                          | Product Team    |
| R-002 | **Regulasi berubah** (POJK direvisi, persyaratan kredit berubah)             | Medium     | High      | **HIGH**     | Arsitektur modular untuk adaptasi cepat, legal retainer, aktif di forum OJK                                   | Legal + Product |
| R-003 | **Kegagalan sinkronisasi offline** (data hilang saat sync conflict)          | Medium     | High      | **HIGH**     | Unit testing ekstensif, conflict resolution strategy terdokumentasi, backup lokal 30 hari                     | Engineering     |
| R-004 | **Keamanan data bocor** (breach database koperasi)                           | Low        | Very High | **HIGH**     | Penetration testing rutin, enkripsi end-to-end, SOC 2 compliance roadmap, bug bounty program                  | Security Team   |
| R-005 | **Model CCI tidak akurat** (Gini <0.45, false positive tinggi)               | Medium     | High      | **HIGH**     | Backtesting dengan data historis, human-in-the-loop review untuk keputusan kredit besar, model monitoring PSI | AI/ML Team      |
| R-006 | **Hyperledger Fabric unavailable** (downtime blockchain)                     | Low        | Medium    | **MEDIUM**   | Fallback ke Polygon Edge, local buffering anchor, SLA Hyperledger ≥99.5%                                      | Engineering     |
| R-007 | **Mitra pembiayaan tidak mau integrasi** (API tidak kompatibel)              | Medium     | High      | **HIGH**     | Mulai dengan SNAP BI yang sudah standar, bangun proof-of-concept bersama 1 mitra awal                         | Business Dev    |
| R-008 | **Persaingan dari Telkom/BUMN** (DigiKoperasi diperkuat dengan fitur serupa) | Medium     | Medium    | **MEDIUM**   | Fokus pada trust layer (ledger) yang tidak ada di kompetitor, partnership dengan Kemenkop untuk preferensi    | Strategy        |
| R-009 | **Keterbatasan perangkat anggota** (HP RAM rendah tidak bisa jalankan PWA)   | High       | Medium    | **MEDIUM**   | Lite mode untuk HP entry-level, minimum requirement jelas, testing di perangkat target                        | Engineering     |
| R-010 | **Kekurangan dana** sebelum Payback Period (Year 4)                          | Low        | Very High | **HIGH**     | Bridge funding dari VC/BUMN, grant program Kemenkop, revenue diversification                                  | Finance         |

### 17.2 Contingency Plan untuk Risk Critical

**R-001: Adopsi Lambat**

- Plan A: Pelatihan tatap muka intensif per batch 10 koperasi
- Plan B: Rekrut "Koperasi Champion" — pengurus muda yang menjadi duta digital di komunitas
- Plan C: Simplifikasi UI radikal (fitur dikurangi, ikon diperbesar, alur diperpendek)

**R-005: Model CCI Tidak Akurat**

- Plan A: Hybrid model — AI scoring sebagai rekomendasi, keputusan final tetap oleh pengurus koperasi
- Plan B: Kalibrasi ulang model setiap kuartal dengan data baru dari pilot
- Plan C: Partnership dengan perusahaan credit bureau (PEFINDO/CRIF) untuk data enrichment

---

## 18. DELIVERABLES & OUTPUT TERUKUR

LUMBERA menetapkan 7 deliverable strategis yang dapat diverifikasi secara objektif:

| #      | Deliverable                    | Deskripsi                                                                                                         | Tanggal Target           | Success Metric                                                            |
| ------ | ------------------------------ | ----------------------------------------------------------------------------------------------------------------- | ------------------------ | ------------------------------------------------------------------------- |
| **D1** | **MVP Demo Working**           | Purwarupa sistem end-to-end: offline input → ledger anchoring → CCI scoring → consent → disbursement via Open API | Akhir Sprint 6 (Bulan 4) | Demo berhasil tanpa bug blocker di hadapan juri                           |
| **D2** | **Pilot 5 Koperasi Heterogen** | Implementasi pada Koperasi Padiwangi, Melati Jaya, Sumber Makmur, Tirta Bersama, Harapan Baru                     | Bulan 6                  | 5 koperasi aktif, data real mengalir di platform                          |
| **D3** | **≥1 Kemitraan Aktif**         | MoU resmi dengan mitra pembiayaan ter-OJK + 1 use case penyaluran kredit tereksekusi riil                         | Bulan 6                  | MoU ditandatangani, bukti disbursement tersedia                           |
| **D4** | **Open-Source SDK Ledger**     | Kode sumber SDK verifiable ledger dipublikasikan di GitHub                                                        | Bulan 9                  | GitHub repo publik, ≥10 stars dalam 3 bulan                               |
| **D5** | **Whitepaper Metodologi CCI**  | Dokumen ilmiah: kerangka analisis, skema pembobotan, metrik validasi (Gini, AUC-ROC, KS, PSI), referensi regulasi | Bulan 10                 | Dokumen siap untuk pengajuan ke OJK sebagai Pemeringkat Kredit Alternatif |
| **D6** | **Regulator Dashboard**        | Dashboard real-time untuk OJK dan Kemenkop: CHS agregat, distribusi grade, early warning                          | Bulan 8                  | OJK dapat akses dan melihat data 5 koperasi pilot                         |
| **D7** | **Playbook Onboarding**        | Buku panduan onboarding koperasi baru: organizational readiness assessment, training guide, go-live checklist     | Bulan 9                  | Koperasi ke-6 dapat onboard tanpa bantuan tim teknis inti                 |

---

## 19. OUT OF SCOPE

Fitur dan fungsi berikut **secara eksplisit tidak termasuk** dalam scope MVP dan Year 1:

| Item                                                      | Alasan Tidak Dimasukkan                                                          |
| --------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Mobile native app (iOS/Android)                           | PWA sudah cukup untuk target pengguna; native app akan dipertimbangkan di Year 2 |
| Peer-to-peer lending antar anggota koperasi               | Membutuhkan lisensi OJK tersendiri (POJK P2P), di luar scope MVP                 |
| Integrasi dengan perbankan besar (BRI, Mandiri, BNI)      | Proses integrasi kompleks dan panjang; fokus ke fintech/BPR dulu                 |
| Multi-currency support                                    | Semua koperasi pilot beroperasi dalam IDR                                        |
| Layanan asuransi anggota                                  | Di luar domain koperasi simpan pinjam                                            |
| Real-time market price feed (harga komoditas)             | Dapat dipertimbangkan di Phase 2 sebagai data input Conditions (CCI)             |
| Fitur chat/komunikasi anggota dalam platform              | Menambah kompleksitas tanpa nilai langsung pada credit scoring                   |
| Blockchain publik (Ethereum mainnet)                      | Biaya gas tidak terjangkau; menggunakan Hyperledger Fabric (permissioned)        |
| Pengelolaan payroll karyawan koperasi                     | Di luar fokus inti: transaksi anggota koperasi                                   |
| Ekspansi ke BUMDes dan UMKM                               | Target setelah koperasi sebagai studi kasus utama terbukti berhasil (Year 3+)    |
| Internasionalisasi (multi-bahasa selain Bahasa Indonesia) | Produk dirancang spesifik untuk pasar Indonesia                                  |

---

## 20. GLOSSARIUM

| Istilah                                   | Definisi                                                                                             |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| **ARS (Adaptive Reporting Schema)**       | Sistem template laporan modular yang otomatis menyesuaikan dengan jenis komoditas/usaha koperasi     |
| **CCI (Cooperative Credit Intelligence)** | Fitur AI scoring dua tingkat: Member Creditworthiness Score (MCS) dan Cooperative Health Score (CHS) |
| **CHS (Cooperative Health Score)**        | Skor kesehatan koperasi secara keseluruhan, skala 0-100 dengan grade AA-D                            |
| **Consent Bridge**                        | Mekanisme yang memungkinkan transfer data kredit ke mitra dengan persetujuan anggota                 |
| **Data Passport**                         | Dashboard bagi anggota untuk melihat semua data tersimpan dan riwayat akses                          |
| **Gini Coefficient**                      | Metrik statistik untuk mengukur kekuatan diskriminasi model credit scoring (target ≥0.55)            |
| **Merkle Tree**                           | Struktur data kriptografis yang digunakan untuk membangun verifiable ledger                          |
| **MCS (Member Creditworthiness Score)**   | Skor kredit individual anggota koperasi, skala 300-850 berdasarkan kerangka 5C                       |
| **NPL (Non-Performing Loan)**             | Rasio pinjaman macet terhadap total pinjaman yang disalurkan                                         |
| **Offline-First**                         | Pendekatan desain sistem di mana aplikasi berfungsi penuh tanpa koneksi internet                     |
| **Permissioned Blockchain**               | Blockchain yang hanya dapat diakses oleh pihak yang mendapat izin (berbeda dengan blockchain publik) |
| **PWA (Progressive Web App)**             | Aplikasi web yang dapat berfungsi seperti aplikasi native, termasuk offline capability               |
| **RLS (Row-Level Security)**              | Fitur database untuk membatasi akses data pada level baris (record) berdasarkan kebijakan            |
| **SHAP (SHapley Additive exPlanations)**  | Metode untuk menjelaskan output model AI secara individual, menunjukkan kontribusi setiap fitur      |
| **SNAP BI**                               | Standar Nasional Open API Pembayaran Bank Indonesia (PADG No. 23/15/2021)                            |
| **Tenant**                                | Satu koperasi yang menggunakan platform LUMBERA, terisolasi datanya dari koperasi lain               |
| **Trust Layer**                           | Lapisan kepercayaan data yang menjamin keaslian dan integritas catatan koperasi                      |
| **Unbanked/Unbankable**                   | Individu atau organisasi yang tidak memiliki akses ke layanan perbankan formal                       |
| **Verifiable Ledger**                     | Buku besar digital yang setiap entri-nya dapat diverifikasi keasliannya secara kriptografis          |
| **XGBoost**                               | Algoritma machine learning berbasis gradient boosting, digunakan sebagai baseline model CCI          |

---

## 21. REFERENSI REGULASI

| No. | Regulasi            | Nomor                      | Tentang                                                 | Implikasi LUMBERA                                              |
| --- | ------------------- | -------------------------- | ------------------------------------------------------- | -------------------------------------------------------------- |
| 1   | Instruksi Presiden  | Inpres No. 9/2025          | Koperasi Desa Merah Putih                               | Platform mendukung program digitalisasi koperasi desa nasional |
| 2   | Peraturan OJK       | POJK No. 29/2024           | Pemeringkat Kredit Alternatif                           | Kerangka model CCI harus patuh, pengajuan registrasi ke OJK    |
| 3   | Peraturan OJK       | POJK No. 40/2024 Pasal 150 | Layanan Pendanaan Bersama Berbasis TI (Fintech Lending) | Credit scoring wajib transparan dan berbasis data relevan      |
| 4   | Undang-Undang       | UU No. 27/2022             | Pelindungan Data Pribadi                                | Consent framework, data retention, hak subjek data             |
| 5   | PADG Bank Indonesia | PADG No. 23/15/2021        | Standar Nasional Open API Pembayaran (SNAP)             | Standard Open Bridge API LUMBERA                               |
| 6   | Peraturan Menteri   | Permenkop UKM No. 8/2023   | Usaha Simpan Pinjam Koperasi                            | Format laporan keuangan harus comply                           |
| 7   | OJK Roadmap         | LPBBTI 2023–2028           | Pengembangan Fintech Lending                            | Target 40-50% porsi UMKM adalah konteks peluang bisnis LUMBERA |

---

## APPROVAL & SIGN-OFF

| Role              | Nama                         | Tanda Tangan       | Tanggal          |
| ----------------- | ---------------------------- | ------------------ | ---------------- |
| Product Lead      | Khaizuran Alvaro             | ******\_\_\_****** | **_ / _** / 2026 |
| Engineering Lead  | Azmi Al Ghifari Rahman       | ******\_\_\_****** | **_ / _** / 2026 |
| AI/ML Lead        | Muhammad Irza Dzulhika       | ******\_\_\_****** | **_ / _** / 2026 |
| Business Dev Lead | Muhammad Alfi Tsani Ramadhan | ******\_\_\_****** | **_ / _** / 2026 |

---

**Catatan Versi:**

| Versi | Tanggal    | Perubahan                | Author             |
| ----- | ---------- | ------------------------ | ------------------ |
| v0.1  | 1 Jun 2026 | Draft awal               | Tim SABI SAMA KITA |
| v1.0  | 8 Jun 2026 | Versi final untuk review | Tim SABI SAMA KITA |

---

_Dokumen ini bersifat confidential dan hanya untuk keperluan internal Tim SABI SAMA KITA dalam rangka kompetisi TechnoScape 2026. Seluruh data proyeksi finansial bersifat estimasi berdasarkan asumsi yang didokumentasikan dalam proposal._

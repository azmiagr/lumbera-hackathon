# Spesifikasi Perhitungan Cooperative Health Score (CHS)

## LUMBERA — Modul Laporan Kesehatan Koperasi

**Versi:** 1.0  
**Status:** Draft untuk Implementasi MVP  
**Tujuan:** Mendefinisikan rumus, sumber data, pembobotan, dan mekanisme perhitungan Cooperative Health Score pada platform LUMBERA.

---

## 1. Gambaran Umum

Cooperative Health Score atau **CHS** adalah skor komposit dengan rentang **0–100** yang digunakan untuk memberikan gambaran ringkas mengenai kesehatan koperasi.

CHS terdiri dari empat dimensi utama:

| Dimensi | Bobot |
|---|---:|
| Keuangan | 35% |
| Operasional | 25% |
| Kualitas Data | 20% |
| Kepatuhan | 20% |
| **Total** | **100%** |

Rumus umum CHS:

```text
CHS = (Skor Keuangan × 35%)
    + (Skor Operasional × 25%)
    + (Skor Kualitas Data × 20%)
    + (Skor Kepatuhan × 20%)
```

Atau secara matematis:

```text
CHS = (F × 0,35) + (O × 0,25) + (D × 0,20) + (K × 0,20)
```

Keterangan:

- `F` = Financial Score atau Skor Keuangan
- `O` = Operational Score atau Skor Operasional
- `D` = Data Quality Score atau Skor Kualitas Data
- `K` = Compliance Score atau Skor Kepatuhan

---

## 2. Klasifikasi Nilai CHS

| Rentang CHS | Grade | Kategori | Interpretasi |
|---:|:---:|---|---|
| 85–100 | AA | Sangat Sehat | Kondisi keuangan, operasional, data, dan kepatuhan sangat baik |
| 75–84,99 | A | Sehat | Kondisi koperasi baik, tetapi masih terdapat beberapa area perbaikan |
| 65–74,99 | B | Cukup Sehat | Koperasi dapat beroperasi, tetapi memiliki beberapa risiko yang perlu dipantau |
| 50–64,99 | C | Kurang Sehat | Terdapat risiko material yang membutuhkan tindakan perbaikan |
| 0–49,99 | D | Tidak Sehat | Koperasi menghadapi risiko tinggi dan membutuhkan intervensi segera |

> Rentang grade masih bersifat konfigurasi awal MVP dan perlu divalidasi bersama ahli koperasi atau regulator.

---

# 3. Perhitungan Skor Keuangan

## 3.1 Bobot Indikator Keuangan

Skor Keuangan dihitung dari empat indikator:

| Indikator | Bobot Internal |
|---|---:|
| Non-Performing Loan Rate | 40% |
| Rasio Kecukupan Modal | 25% |
| Return on Assets | 20% |
| Likuiditas | 15% |
| **Total** | **100%** |

Rumus:

```text
Financial Score =
    (Skor NPL × 40%)
  + (Skor Kecukupan Modal × 25%)
  + (Skor ROA × 20%)
  + (Skor Likuiditas × 15%)
```

Atau:

```text
F = (SNPL × 0,40)
  + (SCAR × 0,25)
  + (SROA × 0,20)
  + (SLIQ × 0,15)
```

---

## 3.2 Non-Performing Loan Rate

NPL mengukur persentase sisa pokok pinjaman bermasalah dibandingkan seluruh sisa pokok pinjaman aktif.

Rumus:

```text
NPL Rate =
    Saldo Pokok Pinjaman Bermasalah
    -------------------------------- × 100%
    Total Saldo Pokok Pinjaman Aktif
```

Definisi pinjaman bermasalah untuk MVP:

- Angsuran terlambat lebih dari batas yang ditentukan koperasi; atau
- Status pinjaman telah diklasifikasikan sebagai `NON_PERFORMING`.

Contoh:

```text
Total saldo pinjaman aktif       = Rp100.000.000
Saldo pinjaman bermasalah        = Rp4.000.000

NPL Rate = 4.000.000 / 100.000.000 × 100%
         = 4%
```

Konversi NPL menjadi skor:

| NPL Rate | Skor |
|---:|---:|
| ≤ 2% | 100 |
| > 2% sampai 5% | 85 |
| > 5% sampai 8% | 65 |
| > 8% sampai 12% | 40 |
| > 12% | 20 |

Karena NPL merupakan indikator negatif, nilai NPL yang lebih rendah menghasilkan skor yang lebih tinggi.

---

## 3.3 Rasio Kecukupan Modal

Untuk MVP, rasio kecukupan modal dihitung dengan membandingkan modal sendiri dengan total aset.

Rumus:

```text
Capital Adequacy Ratio =
    Modal Sendiri
    ------------- × 100%
    Total Aset
```

Komponen modal sendiri dapat terdiri dari:

- Simpanan pokok
- Simpanan wajib yang dikategorikan sebagai modal
- Modal penyertaan
- Cadangan koperasi
- SHU ditahan

Contoh:

```text
Modal sendiri = Rp180.000.000
Total aset    = Rp1.000.000.000

CAR = 180.000.000 / 1.000.000.000 × 100%
    = 18%
```

Konversi rasio kecukupan modal menjadi skor:

| Rasio Modal | Skor |
|---:|---:|
| ≥ 20% | 100 |
| 15% sampai <20% | 85 |
| 10% sampai <15% | 70 |
| 5% sampai <10% | 45 |
| < 5% | 20 |

> Pada implementasi lanjutan, penyebut dapat diganti dari total aset menjadi aset tertimbang menurut risiko apabila data dan metodologi tersebut telah tersedia.

---

## 3.4 Return on Assets

ROA mengukur kemampuan koperasi menghasilkan SHU atau laba bersih dari aset yang dikelola.

Rumus:

```text
ROA =
    SHU atau Laba Bersih
    -------------------- × 100%
    Rata-Rata Total Aset
```

Rata-rata total aset:

```text
Rata-Rata Total Aset =
    Aset Awal Periode + Aset Akhir Periode
    ---------------------------------------
                     2
```

Contoh:

```text
SHU bersih             = Rp40.000.000
Aset awal periode      = Rp900.000.000
Aset akhir periode     = Rp1.100.000.000
Rata-rata total aset   = Rp1.000.000.000

ROA = 40.000.000 / 1.000.000.000 × 100%
    = 4%
```

Konversi ROA menjadi skor:

| ROA | Skor |
|---:|---:|
| ≥ 5% | 100 |
| 3% sampai <5% | 85 |
| 1% sampai <3% | 70 |
| 0% sampai <1% | 50 |
| < 0% | 20 |

---

## 3.5 Likuiditas

Likuiditas menunjukkan kemampuan koperasi memenuhi kewajiban jangka pendek.

Untuk MVP digunakan Current Ratio.

Rumus:

```text
Current Ratio =
    Aset Lancar
    -----------------
    Kewajiban Lancar
```

Contoh:

```text
Aset lancar        = Rp300.000.000
Kewajiban lancar   = Rp200.000.000

Current Ratio = 300.000.000 / 200.000.000
              = 1,5
```

Konversi Current Ratio menjadi skor:

| Current Ratio | Skor |
|---:|---:|
| 1,50 sampai 2,50 | 100 |
| 1,20 sampai <1,50 | 85 |
| 1,00 sampai <1,20 | 70 |
| 0,80 sampai <1,00 | 45 |
| < 0,80 | 20 |
| > 2,50 sampai 3,00 | 90 |
| > 3,00 | 75 |

Nilai likuiditas terlalu rendah menunjukkan ketidakmampuan memenuhi kewajiban. Nilai yang terlalu tinggi dapat menunjukkan kas atau aset lancar belum dimanfaatkan secara produktif.

---

## 3.6 Contoh Perhitungan Skor Keuangan

Misalnya diperoleh:

| Indikator | Nilai Aktual | Skor |
|---|---:|---:|
| NPL | 4% | 85 |
| Rasio kecukupan modal | 18% | 85 |
| ROA | 4% | 85 |
| Current Ratio | 1,40 | 85 |

Perhitungan:

```text
Financial Score =
    (85 × 0,40)
  + (85 × 0,25)
  + (85 × 0,20)
  + (85 × 0,15)

Financial Score =
    34
  + 21,25
  + 17
  + 12,75

Financial Score = 85
```

Apabila Skor Likuiditas bernilai 95, maka:

```text
Financial Score =
    (85 × 0,40)
  + (85 × 0,25)
  + (85 × 0,20)
  + (95 × 0,15)

Financial Score =
    34 + 21,25 + 17 + 14,25

Financial Score = 86,5
```

Nilai tampilan dapat dibulatkan menjadi `87`.

---

# 4. Perhitungan Skor Operasional

## 4.1 Bobot Indikator Operasional

Untuk MVP, skor operasional dapat dihitung menggunakan:

| Indikator | Bobot |
|---|---:|
| Tingkat pembayaran angsuran tepat waktu | 35% |
| Keaktifan anggota | 25% |
| Pertumbuhan transaksi | 20% |
| Efisiensi operasional | 20% |
| **Total** | **100%** |

Rumus:

```text
Operational Score =
    (Skor Pembayaran Tepat Waktu × 35%)
  + (Skor Keaktifan Anggota × 25%)
  + (Skor Pertumbuhan Transaksi × 20%)
  + (Skor Efisiensi Operasional × 20%)
```

---

## 4.2 Tingkat Pembayaran Tepat Waktu

Rumus:

```text
On-Time Payment Rate =
    Jumlah Angsuran Dibayar Tepat Waktu
    ----------------------------------- × 100%
    Total Angsuran Jatuh Tempo
```

Konversi:

| On-Time Payment Rate | Skor |
|---:|---:|
| ≥ 95% | 100 |
| 90% sampai <95% | 85 |
| 80% sampai <90% | 70 |
| 70% sampai <80% | 50 |
| < 70% | 25 |

---

## 4.3 Keaktifan Anggota

Anggota aktif adalah anggota yang melakukan minimal satu transaksi atau aktivitas yang diakui dalam periode pengukuran.

Rumus:

```text
Active Member Rate =
    Jumlah Anggota Aktif
    -------------------- × 100%
    Total Anggota
```

Konversi:

| Active Member Rate | Skor |
|---:|---:|
| ≥ 80% | 100 |
| 65% sampai <80% | 85 |
| 50% sampai <65% | 70 |
| 35% sampai <50% | 50 |
| < 35% | 25 |

---

## 4.4 Pertumbuhan Transaksi

Rumus:

```text
Transaction Growth =
    Jumlah Transaksi Periode Saat Ini - Jumlah Transaksi Periode Sebelumnya
    ----------------------------------------------------------------------- × 100%
                       Jumlah Transaksi Periode Sebelumnya
```

Konversi:

| Pertumbuhan Transaksi | Skor |
|---:|---:|
| ≥ 15% | 100 |
| 5% sampai <15% | 85 |
| 0% sampai <5% | 70 |
| -10% sampai <0% | 50 |
| < -10% | 25 |

---

## 4.5 Efisiensi Operasional

Untuk MVP, efisiensi operasional dihitung dengan rasio biaya operasional terhadap pendapatan operasional.

Rumus:

```text
Operational Expense Ratio =
    Biaya Operasional
    --------------------- × 100%
    Pendapatan Operasional
```

Karena rasio ini bersifat negatif, semakin rendah nilainya semakin baik.

| Rasio Biaya Operasional | Skor |
|---:|---:|
| ≤ 60% | 100 |
| > 60% sampai 75% | 85 |
| > 75% sampai 90% | 70 |
| > 90% sampai 100% | 50 |
| > 100% | 25 |

---

# 5. Perhitungan Skor Kualitas Data

## 5.1 Bobot Indikator Kualitas Data

| Indikator | Bobot |
|---|---:|
| Kelengkapan data | 35% |
| Ketepatan waktu sinkronisasi | 25% |
| Konsistensi data | 25% |
| Validitas ledger | 15% |
| **Total** | **100%** |

Rumus:

```text
Data Quality Score =
    (Skor Kelengkapan × 35%)
  + (Skor Sinkronisasi × 25%)
  + (Skor Konsistensi × 25%)
  + (Skor Validitas Ledger × 15%)
```

---

## 5.2 Kelengkapan Data

Rumus:

```text
Data Completeness Rate =
    Jumlah Field Wajib yang Terisi
    ------------------------------ × 100%
    Total Field Wajib
```

Konversi:

| Kelengkapan | Skor |
|---:|---:|
| ≥ 98% | 100 |
| 95% sampai <98% | 85 |
| 90% sampai <95% | 70 |
| 80% sampai <90% | 50 |
| < 80% | 25 |

---

## 5.3 Ketepatan Waktu Sinkronisasi

Rumus:

```text
Sync Timeliness Rate =
    Jumlah Transaksi Tersinkron Sesuai SLA
    -------------------------------------- × 100%
    Total Transaksi yang Harus Disinkronkan
```

Contoh SLA MVP:

- Transaksi online: maksimum 30 detik
- Transaksi offline: maksimum 30 detik setelah koneksi kembali tersedia

Konversi:

| Sync Timeliness Rate | Skor |
|---:|---:|
| ≥ 98% | 100 |
| 95% sampai <98% | 85 |
| 90% sampai <95% | 70 |
| 80% sampai <90% | 50 |
| < 80% | 25 |

---

## 5.4 Konsistensi Data

Rumus:

```text
Data Consistency Rate =
    Record Tanpa Konflik atau Duplikasi
    ----------------------------------- × 100%
    Total Record yang Diperiksa
```

Konversi:

| Konsistensi | Skor |
|---:|---:|
| ≥ 99% | 100 |
| 97% sampai <99% | 85 |
| 94% sampai <97% | 70 |
| 90% sampai <94% | 50 |
| < 90% | 25 |

---

## 5.5 Validitas Ledger

Rumus:

```text
Ledger Verification Rate =
    Jumlah Record dengan Hash Valid
    ------------------------------- × 100%
    Total Record yang Diverifikasi
```

Konversi:

| Ledger Verification Rate | Skor |
|---:|---:|
| 100% | 100 |
| ≥ 99,9% dan <100% | 85 |
| ≥ 99% dan <99,9% | 60 |
| < 99% | 20 |

Apabila ditemukan tampering yang terkonfirmasi, sistem dapat menerapkan aturan pengurang tambahan atau menetapkan status risiko tinggi.

---

# 6. Perhitungan Skor Kepatuhan

## 6.1 Bobot Indikator Kepatuhan

| Indikator | Bobot |
|---|---:|
| Ketepatan laporan berkala | 35% |
| Kelengkapan dokumen wajib | 25% |
| Pelaksanaan RAT | 20% |
| Audit trail dan consent | 20% |
| **Total** | **100%** |

Rumus:

```text
Compliance Score =
    (Skor Ketepatan Laporan × 35%)
  + (Skor Kelengkapan Dokumen × 25%)
  + (Skor RAT × 20%)
  + (Skor Audit dan Consent × 20%)
```

---

## 6.2 Ketepatan Laporan Berkala

Rumus:

```text
On-Time Reporting Rate =
    Jumlah Laporan Disampaikan Tepat Waktu
    -------------------------------------- × 100%
    Total Laporan yang Wajib Disampaikan
```

Konversi:

| Ketepatan Laporan | Skor |
|---:|---:|
| 100% | 100 |
| 90% sampai <100% | 85 |
| 75% sampai <90% | 65 |
| 50% sampai <75% | 40 |
| < 50% | 20 |

---

## 6.3 Kelengkapan Dokumen Wajib

Rumus:

```text
Document Completeness Rate =
    Jumlah Dokumen Wajib yang Valid
    ------------------------------- × 100%
    Total Dokumen Wajib
```

Konversi:

| Kelengkapan Dokumen | Skor |
|---:|---:|
| 100% | 100 |
| 90% sampai <100% | 85 |
| 75% sampai <90% | 65 |
| 50% sampai <75% | 40 |
| < 50% | 20 |

---

## 6.4 Pelaksanaan RAT

Contoh aturan skor:

| Kondisi RAT | Skor |
|---|---:|
| Dilaksanakan tepat waktu dan dokumen lengkap | 100 |
| Dilaksanakan terlambat ≤30 hari | 80 |
| Dilaksanakan terlambat >30 hari | 60 |
| Belum dilaksanakan | 20 |

---

## 6.5 Audit Trail dan Consent

Contoh indikator:

- Persentase aksi penting yang memiliki audit log
- Persentase akses data mitra yang memiliki consent aktif
- Tidak adanya akses data tanpa izin
- Tidak adanya penghapusan audit trail

Rumus gabungan sederhana:

```text
Audit and Consent Score =
    (Audit Log Coverage × 50%)
  + (Valid Consent Coverage × 50%)
```

---

# 7. Contoh Perhitungan CHS Keseluruhan

Misalnya:

| Dimensi | Skor | Bobot | Nilai Tertimbang |
|---|---:|---:|---:|
| Keuangan | 87 | 35% | 30,45 |
| Operasional | 75 | 25% | 18,75 |
| Kualitas Data | 80 | 20% | 16,00 |
| Kepatuhan | 72 | 20% | 14,40 |
| **Total** |  | **100%** | **79,60** |

Perhitungan:

```text
CHS =
    (87 × 0,35)
  + (75 × 0,25)
  + (80 × 0,20)
  + (72 × 0,20)

CHS =
    30,45
  + 18,75
  + 16
  + 14,40

CHS = 79,60
```

Hasil tampilan:

```text
CHS = 80/100
Grade = A
Kategori = Sehat
```

> Dengan angka dimensi di atas, nilai CHS bukan 87. Nilai 87 hanya merupakan Skor Keuangan.

---

# 8. Sumber Data Perhitungan

## 8.1 Skor Keuangan

| Indikator | Sumber Data |
|---|---|
| NPL | Pinjaman, jadwal angsuran, pembayaran, tunggakan |
| Rasio kecukupan modal | Neraca dan saldo akun modal |
| ROA | Laporan SHU/Laba Rugi dan Neraca |
| Likuiditas | Aset lancar dan kewajiban lancar pada Neraca |

## 8.2 Skor Operasional

| Indikator | Sumber Data |
|---|---|
| Pembayaran tepat waktu | Jadwal angsuran dan transaksi pembayaran |
| Keaktifan anggota | Aktivitas transaksi anggota |
| Pertumbuhan transaksi | Rekap transaksi per periode |
| Efisiensi operasional | Biaya dan pendapatan operasional |

## 8.3 Skor Kualitas Data

| Indikator | Sumber Data |
|---|---|
| Kelengkapan data | Profil anggota, transaksi, pinjaman, laporan |
| Ketepatan sinkronisasi | `recorded_at`, `synced_at`, status sinkronisasi |
| Konsistensi data | Hasil validasi, deduplikasi, dan rekonsiliasi |
| Validitas ledger | `prev_hash`, `current_hash`, hasil verifikasi hash |

## 8.4 Skor Kepatuhan

| Indikator | Sumber Data |
|---|---|
| Ketepatan laporan | Histori generate dan submit laporan |
| Kelengkapan dokumen | Dokumen legal dan dokumen pelaporan |
| Pelaksanaan RAT | Data kegiatan RAT dan dokumen pendukung |
| Audit dan consent | Audit log, consent, serta access log |

---

# 9. Alur Proses Sistem

```text
Transaksi, pinjaman, angsuran, dan data operasional
                         ↓
              Jurnal dan saldo akun
                         ↓
     Neraca, Laba Rugi/SHU, dan Arus Kas
                         ↓
       Perhitungan rasio setiap indikator
                         ↓
      Konversi nilai rasio menjadi skor 0–100
                         ↓
       Perhitungan skor setiap dimensi CHS
                         ↓
             Perhitungan total CHS
                         ↓
          Penentuan grade dan rekomendasi
                         ↓
          Penyimpanan histori hasil skor
```

---

# 10. Aturan Implementasi

## 10.1 Periode Perhitungan

Perhitungan CHS dapat dilakukan:

- Secara bulanan
- Setelah laporan bulanan ditutup
- Ketika terdapat perubahan data material
- Secara manual oleh pengurus yang memiliki hak akses

Periode default MVP:

```text
Perhitungan otomatis: setiap tanggal 1 pukul 01.00 WIB
Data yang digunakan: periode bulan sebelumnya
```

---

## 10.2 Snapshot Hasil

Hasil perhitungan harus disimpan sebagai snapshot agar perubahan aturan di masa depan tidak mengubah hasil historis.

Data minimal yang disimpan:

```text
id
tenant_id
period_start
period_end
financial_score
operational_score
data_quality_score
compliance_score
chs_score
grade
calculation_version
calculated_at
```

Detail setiap indikator juga perlu disimpan:

```text
indicator_code
raw_value
normalized_score
weight
weighted_score
source_period
```

---

## 10.3 Versioning Rumus

Setiap perubahan rumus, threshold, atau bobot wajib memiliki versi.

Contoh:

```text
CHS_MVP_V1
CHS_MVP_V2
CHS_REGULATORY_V1
```

Hasil historis harus tetap merujuk pada versi rumus yang digunakan saat perhitungan dilakukan.

---

## 10.4 Penanganan Data Tidak Lengkap

Apabila data indikator tidak tersedia, sistem tidak boleh langsung memberikan skor nol tanpa penjelasan.

Status perhitungan:

| Status | Kondisi |
|---|---|
| `COMPLETE` | Seluruh indikator tersedia |
| `PARTIAL` | Sebagian indikator tidak tersedia |
| `INSUFFICIENT_DATA` | Data tidak cukup untuk menghasilkan CHS yang valid |
| `FAILED` | Terjadi kesalahan proses |

Pilihan MVP untuk status `PARTIAL`:

1. Tetap menghitung menggunakan indikator yang tersedia dengan normalisasi ulang bobot; atau
2. Tidak menerbitkan grade resmi dan hanya menampilkan skor sementara.

Rekomendasi: gunakan opsi kedua agar skor tidak menyesatkan.

---

## 10.5 Pembulatan

- Perhitungan internal menggunakan minimal 4 angka desimal.
- Nilai yang disimpan dapat menggunakan 2 angka desimal.
- Nilai pada antarmuka dibulatkan ke bilangan bulat terdekat.
- Grade ditentukan berdasarkan nilai sebelum pembulatan tampilan.

Contoh:

```text
Nilai internal = 84,60
Tampilan       = 85
Grade          = A
```

---

# 11. Contoh Pseudocode

```go
type DimensionScore struct {
    Score  float64
    Weight float64
}

func CalculateCHS(
    financial float64,
    operational float64,
    dataQuality float64,
    compliance float64,
) float64 {
    chs := (financial * 0.35) +
        (operational * 0.25) +
        (dataQuality * 0.20) +
        (compliance * 0.20)

    return math.Round(chs*100) / 100
}

func DetermineGrade(score float64) string {
    switch {
    case score >= 85:
        return "AA"
    case score >= 75:
        return "A"
    case score >= 65:
        return "B"
    case score >= 50:
        return "C"
    default:
        return "D"
    }
}
```

Contoh perhitungan Financial Score:

```go
func CalculateFinancialScore(
    nplScore float64,
    capitalScore float64,
    roaScore float64,
    liquidityScore float64,
) float64 {
    score := (nplScore * 0.40) +
        (capitalScore * 0.25) +
        (roaScore * 0.20) +
        (liquidityScore * 0.15)

    return math.Round(score*100) / 100
}
```

---

# 12. Catatan Validasi

Seluruh bobot dan threshold pada dokumen ini merupakan rancangan awal untuk kebutuhan MVP dan demonstrasi sistem.

Sebelum digunakan untuk keputusan pembiayaan nyata, metodologi harus:

1. Diverifikasi oleh ahli akuntansi atau kesehatan koperasi.
2. Disesuaikan dengan regulasi dan pedoman penilaian koperasi yang berlaku.
3. Diuji menggunakan data historis koperasi.
4. Dikalibrasi agar tidak menghasilkan skor yang bias.
5. Memiliki dokumentasi metodologi dan audit trail perubahan formula.

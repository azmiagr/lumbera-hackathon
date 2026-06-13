# Spesifikasi Perhitungan Member Creditworthiness Score (MCS)

## LUMBERA — Modul Cooperative Credit Intelligence

**Versi:** 1.0  
**Status:** Draft untuk Implementasi MVP  
**Tujuan:** Mendefinisikan metode, indikator, bobot, sumber data, dan mekanisme perhitungan Member Creditworthiness Score pada platform LUMBERA.

---

## 1. Gambaran Umum

Member Creditworthiness Score atau **MCS** adalah skor kelayakan kredit anggota koperasi dengan rentang **300–850**.

MCS digunakan untuk:

- Memberikan gambaran risiko kredit anggota.
- Membantu pengurus mengevaluasi pengajuan pinjaman.
- Menyediakan penilaian alternatif bagi anggota no-file dan thin-file.
- Menjadi input bagi mitra pembiayaan melalui Open Bridge.
- Memberikan rekomendasi perbaikan kepada anggota.

Pada tahap MVP, MCS dihitung menggunakan pendekatan **rule-based berbobot berdasarkan prinsip 5C**:

1. Character
2. Capacity
3. Capital
4. Conditions
5. Collateral

Pada tahap lanjutan, hasil rule-based dapat diganti atau dikombinasikan dengan model machine learning seperti XGBoost.

---

# 2. Struktur Bobot 5C

| Komponen | Bobot |
|---|---:|
| Character | 35% |
| Capacity | 30% |
| Capital | 15% |
| Conditions | 12% |
| Collateral | 8% |
| **Total** | **100%** |

Rumus skor gabungan 5C:

```text
Five-C Score =
    (Character × 35%)
  + (Capacity × 30%)
  + (Capital × 15%)
  + (Conditions × 12%)
  + (Collateral × 8%)
```

Atau:

```text
S5C =
    (CH × 0,35)
  + (CP × 0,30)
  + (CA × 0,15)
  + (CO × 0,12)
  + (CL × 0,08)
```

Keterangan:

- `CH` = Character Score
- `CP` = Capacity Score
- `CA` = Capital Score
- `CO` = Conditions Score
- `CL` = Collateral Score

Seluruh komponen memiliki rentang nilai **0–100**.

---

# 3. Konversi Skor 5C ke MCS

Setelah memperoleh skor gabungan 5C dalam skala 0–100, nilai dikonversi menjadi skala MCS 300–850.

Rumus:

```text
MCS = 300 + (Five-C Score × 5,5)
```

Bentuk ekuivalen:

```text
MCS = 300 + ((Five-C Score / 100) × 550)
```

Contoh:

```text
Five-C Score = 73

MCS = 300 + (73 × 5,5)
    = 300 + 401,5
    = 701,5
```

Nilai tampilan dibulatkan menjadi:

```text
MCS = 702
```

---

# 4. Klasifikasi Grade MCS

| Rentang MCS | Grade | Label | Interpretasi |
|---:|:---:|---|---|
| 750–850 | AA | Sangat Baik | Risiko kredit sangat rendah |
| 680–749 | A | Baik | Risiko kredit rendah |
| 580–679 | B | Cukup | Risiko kredit moderat |
| 480–579 | C | Perlu Perhatian | Risiko kredit cukup tinggi |
| 300–479 | D | Buruk | Risiko kredit tinggi |

> Rentang grade merupakan rancangan awal MVP dan perlu dikalibrasi menggunakan data historis kredit koperasi.

---

# 5. Perhitungan Character

## 5.1 Tujuan

Character menilai perilaku, kedisiplinan, dan konsistensi anggota selama menjadi bagian dari koperasi.

## 5.2 Bobot Internal Character

| Indikator | Bobot |
|---|---:|
| Konsistensi setoran | 40% |
| Kedisiplinan transaksi | 25% |
| Kehadiran RAT | 15% |
| Lama keanggotaan | 20% |
| **Total** | **100%** |

Rumus:

```text
Character Score =
    (Skor Konsistensi Setoran × 40%)
  + (Skor Kedisiplinan Transaksi × 25%)
  + (Skor Kehadiran RAT × 15%)
  + (Skor Lama Keanggotaan × 20%)
```

---

## 5.3 Konsistensi Setoran

Rumus:

```text
Saving Consistency Rate =
    Jumlah Bulan Melakukan Setoran
    ------------------------------ × 100%
    Jumlah Bulan yang Diwajibkan
```

Konversi nilai:

| Konsistensi Setoran | Skor |
|---:|---:|
| ≥ 95% | 100 |
| 85% sampai <95% | 85 |
| 70% sampai <85% | 70 |
| 50% sampai <70% | 50 |
| < 50% | 25 |

Contoh:

```text
Bulan wajib setoran = 12
Bulan dengan setoran = 10

Saving Consistency Rate = 10 / 12 × 100%
                        = 83,33%

Skor = 70
```

---

## 5.4 Kedisiplinan Transaksi

Kedisiplinan transaksi menilai apakah anggota melakukan transaksi sesuai jadwal dan tidak memiliki pola transaksi bermasalah.

Contoh indikator:

- Tidak sering terlambat melakukan setoran wajib.
- Tidak memiliki transaksi reversal akibat kesalahan berulang.
- Tidak memiliki transaksi yang terindikasi fraud.
- Memiliki frekuensi aktivitas yang wajar.

Konversi sederhana untuk MVP:

| Kondisi | Skor |
|---|---:|
| Tidak ada pelanggaran atau keterlambatan | 100 |
| Pelanggaran ringan ≤5% dari transaksi | 85 |
| Pelanggaran 5–10% | 70 |
| Pelanggaran 10–20% | 50 |
| Pelanggaran >20% | 25 |

---

## 5.5 Kehadiran RAT

Rumus:

```text
RAT Attendance Rate =
    Jumlah RAT yang Dihadiri
    ------------------------ × 100%
    Jumlah RAT yang Seharusnya Dihadiri
```

Konversi:

| Kehadiran RAT | Skor |
|---:|---:|
| 100% | 100 |
| 75% sampai <100% | 85 |
| 50% sampai <75% | 70 |
| 25% sampai <50% | 50 |
| < 25% | 25 |

Apabila anggota belum pernah memiliki kesempatan mengikuti RAT, indikator dapat diberi status `NOT_AVAILABLE`.

---

## 5.6 Lama Keanggotaan

| Lama Keanggotaan | Skor |
|---|---:|
| ≥ 60 bulan | 100 |
| 36–59 bulan | 85 |
| 24–35 bulan | 70 |
| 12–23 bulan | 55 |
| 6–11 bulan | 40 |
| < 6 bulan | 25 |

---

# 6. Perhitungan Capacity

## 6.1 Tujuan

Capacity menilai kemampuan anggota untuk memenuhi kewajiban angsuran berdasarkan arus kas dan riwayat pembayaran.

## 6.2 Bobot Internal Capacity

| Indikator | Bobot |
|---|---:|
| Ketepatan pembayaran | 50% |
| Debt Service Ratio | 35% |
| Frekuensi atau beban pinjaman | 15% |
| **Total** | **100%** |

Rumus:

```text
Capacity Score =
    (Skor Ketepatan Pembayaran × 50%)
  + (Skor DSR × 35%)
  + (Skor Frekuensi Pinjaman × 15%)
```

---

## 6.3 Ketepatan Pembayaran

Rumus:

```text
Payment Timeliness Rate =
    Angsuran Dibayar Tepat Waktu
    ---------------------------- × 100%
    Total Angsuran Jatuh Tempo
```

Konversi:

| Ketepatan Pembayaran | Skor |
|---:|---:|
| ≥ 95% | 100 |
| 90% sampai <95% | 90 |
| 80% sampai <90% | 75 |
| 70% sampai <80% | 55 |
| < 70% | 25 |

Apabila anggota belum memiliki histori angsuran, indikator ini dapat diberi status `INSUFFICIENT_HISTORY`.

---

## 6.4 Debt Service Ratio

Debt Service Ratio mengukur proporsi pendapatan bulanan yang digunakan untuk membayar seluruh angsuran.

Rumus:

```text
DSR =
    Total Angsuran Bulanan
    ---------------------- × 100%
    Pendapatan Bulanan
```

Konversi:

| DSR | Skor |
|---:|---:|
| ≤ 20% | 100 |
| > 20% sampai 30% | 85 |
| > 30% sampai 40% | 70 |
| > 40% sampai 50% | 45 |
| > 50% | 20 |

Apabila data pendapatan tidak tersedia, sistem dapat menggunakan proxy:

```text
Loan-to-Savings Ratio =
    Outstanding Pinjaman
    ---------------------
    Total Simpanan
```

Proxy ini hanya digunakan untuk MVP dan harus ditandai pada hasil penilaian.

---

## 6.5 Frekuensi atau Beban Pinjaman

Contoh aturan MVP:

| Kondisi | Skor |
|---|---:|
| Tidak ada pinjaman aktif atau hanya 1 pinjaman terkendali | 100 |
| 2 pinjaman aktif dengan pembayaran lancar | 85 |
| 3 pinjaman aktif | 65 |
| >3 pinjaman aktif | 40 |
| Memiliki restrukturisasi atau tunggakan aktif | 20 |

---

# 7. Perhitungan Capital

## 7.1 Tujuan

Capital menilai kekuatan dana, simpanan, aset, dan kemampuan anggota menanggung sebagian risiko pinjaman.

## 7.2 Bobot Internal Capital

| Indikator | Bobot |
|---|---:|
| Rasio simpanan terhadap pinjaman | 45% |
| Stabilitas saldo simpanan | 30% |
| Aset yang dilaporkan | 25% |
| **Total** | **100%** |

Rumus:

```text
Capital Score =
    (Skor Savings-to-Loan Ratio × 45%)
  + (Skor Stabilitas Saldo × 30%)
  + (Skor Aset × 25%)
```

---

## 7.3 Savings-to-Loan Ratio

Rumus:

```text
Savings-to-Loan Ratio =
    Total Simpanan
    --------------------- × 100%
    Outstanding Pinjaman
```

Konversi:

| Savings-to-Loan Ratio | Skor |
|---:|---:|
| ≥ 50% | 100 |
| 35% sampai <50% | 85 |
| 20% sampai <35% | 70 |
| 10% sampai <20% | 50 |
| < 10% | 25 |

Apabila anggota tidak memiliki pinjaman aktif:

- Pembagian tidak dilakukan.
- Skor Capital dihitung dari saldo simpanan, stabilitas saldo, dan aset.
- Status indikator disimpan sebagai `NO_ACTIVE_LOAN`.

---

## 7.4 Stabilitas Saldo Simpanan

Stabilitas saldo dapat dihitung dari rata-rata saldo 3 atau 6 bulan.

Contoh pendekatan MVP:

```text
Balance Stability =
    Saldo Rata-Rata Minimum
    ----------------------- × 100%
    Saldo Rata-Rata Periode
```

Atau menggunakan koefisien variasi jika data historis mencukupi.

Konversi sederhana:

| Kondisi Saldo | Skor |
|---|---:|
| Stabil atau meningkat selama ≥6 bulan | 100 |
| Stabil selama 3–5 bulan | 85 |
| Fluktuasi moderat | 70 |
| Sering turun hingga mendekati nol | 45 |
| Tidak memiliki saldo simpanan | 20 |

---

## 7.5 Aset yang Dilaporkan

Contoh konversi berdasarkan rasio aset bersih terhadap pinjaman yang diajukan:

```text
Asset Coverage Ratio =
    Nilai Aset Bersih
    ----------------- × 100%
    Nilai Pinjaman
```

| Asset Coverage Ratio | Skor |
|---:|---:|
| ≥ 200% | 100 |
| 150% sampai <200% | 85 |
| 100% sampai <150% | 70 |
| 50% sampai <100% | 50 |
| < 50% | 25 |

Nilai aset yang dilaporkan harus memiliki timestamp, sumber, dan tingkat verifikasi.

---

# 8. Perhitungan Conditions

## 8.1 Tujuan

Conditions menilai risiko eksternal yang dapat memengaruhi kemampuan anggota membayar pinjaman.

## 8.2 Bobot Internal Conditions

| Indikator | Bobot |
|---|---:|
| Risiko sektor usaha | 40% |
| Risiko geografis | 25% |
| Stabilitas harga atau pendapatan | 20% |
| Kesesuaian musim pinjaman | 15% |
| **Total** | **100%** |

Rumus:

```text
Conditions Score =
    (Skor Risiko Sektor × 40%)
  + (Skor Risiko Geografis × 25%)
  + (Skor Stabilitas Pendapatan × 20%)
  + (Skor Kesesuaian Musim × 15%)
```

---

## 8.3 Risiko Sektor Usaha

Contoh klasifikasi:

| Risiko Sektor | Skor |
|---|---:|
| Sangat rendah | 100 |
| Rendah | 85 |
| Moderat | 70 |
| Tinggi | 45 |
| Sangat tinggi | 20 |

Penilaian sektor pada MVP dapat ditentukan oleh konfigurasi koperasi.

---

## 8.4 Risiko Geografis

Faktor yang dapat digunakan:

- Risiko banjir
- Risiko kekeringan
- Akses pasar
- Ketersediaan infrastruktur
- Stabilitas konektivitas
- Kerawanan bencana

Contoh klasifikasi:

| Risiko Geografis | Skor |
|---|---:|
| Sangat rendah | 100 |
| Rendah | 85 |
| Moderat | 70 |
| Tinggi | 45 |
| Sangat tinggi | 20 |

---

## 8.5 Stabilitas Harga atau Pendapatan

Contoh klasifikasi:

| Kondisi | Skor |
|---|---:|
| Pendapatan sangat stabil | 100 |
| Pendapatan stabil | 85 |
| Fluktuasi moderat | 70 |
| Fluktuasi tinggi | 45 |
| Tidak ada data atau sangat tidak stabil | 25 |

---

## 8.6 Kesesuaian Musim Pinjaman

Untuk pinjaman produktif, waktu pencairan harus sesuai dengan siklus usaha.

Contoh:

| Kondisi | Skor |
|---|---:|
| Sangat sesuai dengan siklus usaha | 100 |
| Sesuai | 85 |
| Cukup sesuai | 70 |
| Kurang sesuai | 45 |
| Tidak sesuai | 20 |

---

# 9. Perhitungan Collateral

## 9.1 Tujuan

Collateral menilai kecukupan nilai jaminan terhadap outstanding atau nilai pinjaman.

## 9.2 Loan-to-Value Ratio

Rumus:

```text
LTV =
    Outstanding Pinjaman
    --------------------- × 100%
    Nilai Jaminan
```

Konversi:

| LTV | Skor |
|---:|---:|
| ≤ 50% | 100 |
| > 50% sampai 70% | 85 |
| > 70% sampai 85% | 70 |
| > 85% sampai 100% | 50 |
| > 100% | 25 |
| Tidak ada jaminan | 20 |

Karena bobot Collateral hanya 8%, anggota tanpa jaminan tidak otomatis memperoleh MCS buruk.

Untuk pinjaman tanpa agunan, sistem dapat menggunakan indikator alternatif seperti:

- Simpanan yang diblokir.
- Penjamin kelompok.
- Riwayat pembayaran.
- Nilai hasil panen atau invoice.
- Aset produktif yang belum diikat secara formal.

---

# 10. Contoh Perhitungan Lengkap

Misalnya anggota memiliki skor:

| Komponen | Skor | Bobot | Nilai Tertimbang |
|---|---:|---:|---:|
| Character | 82 | 35% | 28,70 |
| Capacity | 75 | 30% | 22,50 |
| Capital | 68 | 15% | 10,20 |
| Conditions | 70 | 12% | 8,40 |
| Collateral | 40 | 8% | 3,20 |
| **Total Five-C Score** |  | **100%** | **73,00** |

Perhitungan:

```text
Five-C Score =
    (82 × 0,35)
  + (75 × 0,30)
  + (68 × 0,15)
  + (70 × 0,12)
  + (40 × 0,08)

Five-C Score =
    28,70
  + 22,50
  + 10,20
  + 8,40
  + 3,20

Five-C Score = 73
```

Konversi ke MCS:

```text
MCS = 300 + (73 × 5,5)
    = 701,5
```

Hasil:

```text
MCS   = 702
Grade = A
Label = Baik
```

---

# 11. Sumber Data

| Komponen | Indikator | Sumber Data |
|---|---|---|
| Character | Konsistensi setoran | Transaksi simpanan wajib dan sukarela |
| Character | Kedisiplinan transaksi | Histori transaksi, reversal, fraud alert |
| Character | Kehadiran RAT | Data kehadiran kegiatan RAT |
| Character | Lama keanggotaan | `members.join_date` |
| Capacity | Ketepatan pembayaran | Jadwal angsuran dan transaksi pembayaran |
| Capacity | DSR | Profil pendapatan dan total angsuran |
| Capacity | Beban pinjaman | Pinjaman aktif dan histori pinjaman |
| Capital | Total simpanan | Saldo simpanan anggota |
| Capital | Stabilitas saldo | Histori saldo 3–6 bulan |
| Capital | Aset | Profil aset anggota |
| Conditions | Sektor usaha | Profil pekerjaan atau usaha |
| Conditions | Risiko geografis | Lokasi dan konfigurasi risiko wilayah |
| Conditions | Stabilitas pendapatan | Histori pendapatan atau transaksi usaha |
| Conditions | Siklus usaha | Profil musim dan tujuan pinjaman |
| Collateral | Nilai jaminan | Data collateral dan hasil valuasi |

---

# 12. Syarat Minimum Perhitungan

MCS tidak boleh langsung dianggap valid ketika data anggota masih sangat terbatas.

Syarat minimum MVP:

- Anggota telah memiliki minimal 3 transaksi.
- Profil dasar anggota telah lengkap.
- Lama keanggotaan diketahui.
- Terdapat minimal satu sumber data untuk Character.
- Terdapat minimal satu sumber data untuk Capacity.

Status hasil:

| Status | Kondisi |
|---|---|
| `COMPLETE` | Seluruh komponen 5C memiliki data memadai |
| `PARTIAL` | Sebagian indikator menggunakan proxy |
| `INSUFFICIENT_DATA` | Data belum cukup untuk menghasilkan skor |
| `FAILED` | Proses perhitungan mengalami kesalahan |

MCS dengan status `PARTIAL` harus menampilkan pemberitahuan bahwa skor masih bersifat sementara.

---

# 13. Penanganan Data Tidak Tersedia

Sistem tidak boleh otomatis memberikan skor nol untuk data yang tidak tersedia.

Setiap indikator harus memiliki status:

```text
AVAILABLE
NOT_AVAILABLE
INSUFFICIENT_HISTORY
NOT_APPLICABLE
USING_PROXY
```

Rekomendasi MVP:

- Jika data penting tidak tersedia, jangan menerbitkan grade final.
- Tampilkan MCS sementara dengan label `Data Belum Lengkap`.
- Simpan informasi indikator mana yang menggunakan proxy.
- Jangan menyamakan `NOT_AVAILABLE` dengan performa buruk.

---

# 14. Waktu Perhitungan Ulang

MCS dapat dihitung ulang ketika:

- Transaksi anggota baru berhasil disinkronkan.
- Angsuran dibayar.
- Angsuran melewati jatuh tempo.
- Pinjaman baru disetujui atau dicairkan.
- Data pendapatan diperbarui.
- Data jaminan diperbarui.
- Terdapat perubahan signifikan pada profil risiko.
- Background job bulanan dijalankan.

Untuk MVP, perhitungan dapat dilakukan:

```text
Real-time trigger:
- Setelah transaksi finansial anggota

Scheduled trigger:
- Rekalkulasi bulanan setiap tanggal 1 pukul 02.00 WIB
```

---

# 15. Snapshot dan Histori MCS

Setiap hasil perhitungan harus disimpan sebagai snapshot.

Data minimal:

```text
id
tenant_id
member_id
score_value
grade
character_score
capacity_score
capital_score
conditions_score
collateral_score
calculation_status
calculation_version
calculated_at
triggering_transaction_id
```

Detail indikator:

```text
component_code
indicator_code
raw_value
normalized_score
weight
weighted_score
data_status
source_period
source_reference
```

Histori tidak boleh dihitung ulang menggunakan rumus versi terbaru tanpa membuat snapshot baru.

---

# 16. Versioning Rumus

Setiap perubahan terhadap bobot, threshold, atau metode normalisasi harus memiliki versi.

Contoh:

```text
MCS_RULE_V1
MCS_RULE_V2
MCS_XGBOOST_V1
MCS_HYBRID_V1
```

Hasil historis harus menyimpan versi rumus atau model yang digunakan.

---

# 17. Explainability

Setiap hasil MCS harus memberikan penjelasan yang mudah dipahami anggota.

Contoh:

```text
Faktor Positif:
- Anda membayar 11 dari 12 angsuran tepat waktu.
- Setoran wajib dilakukan secara konsisten selama 10 bulan.
- Saldo simpanan stabil selama 6 bulan.

Faktor yang Perlu Diperbaiki:
- Rasio angsuran terhadap pendapatan mencapai 42%.
- Nilai jaminan lebih rendah dibandingkan nilai pinjaman.
- Data pendapatan belum diperbarui selama 4 bulan.
```

Untuk rule-based, penjelasan berasal dari indikator dengan weighted score terbesar dan terkecil.

Untuk machine learning, penjelasan dapat dihasilkan menggunakan SHAP.

---

# 18. Integrasi dengan XGBoost

## 18.1 Tahap MVP

```text
Data anggota
     ↓
Normalisasi indikator 5C
     ↓
Rule-based weighted score
     ↓
Konversi ke MCS 300–850
     ↓
Penjelasan berbasis aturan
```

## 18.2 Tahap Machine Learning

```text
Data historis anggota
     ↓
Feature engineering
     ↓
Model XGBoost
     ↓
Probability of Default
     ↓
Kalibrasi probabilitas
     ↓
Konversi ke MCS 300–850
     ↓
SHAP explanation
```

Output probabilitas model tidak boleh langsung dikonversi dengan rumus sederhana tanpa proses kalibrasi.

Contoh rumus yang **tidak direkomendasikan**:

```text
MCS = 300 + ((1 - Probability of Default) × 550)
```

Alasannya:

- Distribusi probabilitas model belum tentu linear.
- Skor dapat terlalu optimistis atau terlalu konservatif.
- Rentang skor tidak mencerminkan observed default rate.
- Model perlu dikalibrasi menggunakan data historis.

---

# 19. Pendekatan Hybrid

Pada tahap transisi, sistem dapat menggabungkan rule-based score dan machine learning.

Contoh:

```text
Hybrid Score =
    (Rule-Based Score × 40%)
  + (ML-Calibrated Score × 60%)
```

Penggunaan pendekatan hybrid harus memenuhi syarat:

- Model telah melewati validasi minimum.
- Output ML sudah dikalibrasi.
- Kontribusi rule-based dan ML terdokumentasi.
- Hasil memiliki explanation.
- Versi model dan versi formula tersimpan.

---

# 20. Pembulatan

- Perhitungan internal menggunakan minimal 4 angka desimal.
- Five-C Score disimpan dengan 2 angka desimal.
- MCS disimpan sebagai angka desimal atau integer sesuai kebutuhan.
- Nilai antarmuka dibulatkan ke integer terdekat.
- Grade ditentukan berdasarkan nilai MCS sebelum pembulatan tampilan.

Contoh:

```text
MCS internal = 679,60
Tampilan     = 680
Grade        = B
```

Apabila ingin menghindari perbedaan tampilan dan grade, sistem dapat menampilkan satu angka desimal.

---

# 21. Contoh Pseudocode Go

```go
package scoring

import "math"

type FiveCScores struct {
    Character  float64
    Capacity   float64
    Capital    float64
    Conditions float64
    Collateral float64
}

type MCSResult struct {
    FiveCScore float64
    MCS        float64
    Grade      string
    Label      string
}

func CalculateFiveCScore(scores FiveCScores) float64 {
    score := (scores.Character * 0.35) +
        (scores.Capacity * 0.30) +
        (scores.Capital * 0.15) +
        (scores.Conditions * 0.12) +
        (scores.Collateral * 0.08)

    return math.Round(score*100) / 100
}

func ConvertToMCS(fiveCScore float64) float64 {
    mcs := 300 + (fiveCScore * 5.5)
    return math.Round(mcs*100) / 100
}

func DetermineMCSGrade(mcs float64) (string, string) {
    switch {
    case mcs >= 750:
        return "AA", "Sangat Baik"
    case mcs >= 680:
        return "A", "Baik"
    case mcs >= 580:
        return "B", "Cukup"
    case mcs >= 480:
        return "C", "Perlu Perhatian"
    default:
        return "D", "Buruk"
    }
}

func CalculateMCS(scores FiveCScores) MCSResult {
    fiveCScore := CalculateFiveCScore(scores)
    mcs := ConvertToMCS(fiveCScore)
    grade, label := DetermineMCSGrade(mcs)

    return MCSResult{
        FiveCScore: fiveCScore,
        MCS:        mcs,
        Grade:      grade,
        Label:      label,
    }
}
```

---

# 22. Contoh Struktur Database

```sql
CREATE TABLE member_credit_scores (
    id                  VARCHAR(36) PRIMARY KEY,
    tenant_id           VARCHAR(36) NOT NULL,
    member_id           VARCHAR(36) NOT NULL,
    score_value         DECIMAL(6,2) NOT NULL,
    grade               ENUM('AA','A','B','C','D') NOT NULL,
    character_score     DECIMAL(5,2),
    capacity_score      DECIMAL(5,2),
    capital_score       DECIMAL(5,2),
    conditions_score    DECIMAL(5,2),
    collateral_score    DECIMAL(5,2),
    calculation_status  ENUM(
        'COMPLETE',
        'PARTIAL',
        'INSUFFICIENT_DATA',
        'FAILED'
    ) NOT NULL,
    calculation_version VARCHAR(50) NOT NULL,
    model_version       VARCHAR(50),
    explanation         JSON,
    triggering_tx_id    VARCHAR(36),
    calculated_at       TIMESTAMP NOT NULL,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Detail indikator:

```sql
CREATE TABLE member_score_indicators (
    id                  VARCHAR(36) PRIMARY KEY,
    credit_score_id     VARCHAR(36) NOT NULL,
    component_code      ENUM(
        'CHARACTER',
        'CAPACITY',
        'CAPITAL',
        'CONDITIONS',
        'COLLATERAL'
    ) NOT NULL,
    indicator_code      VARCHAR(100) NOT NULL,
    raw_value           DECIMAL(15,4),
    normalized_score    DECIMAL(5,2),
    weight              DECIMAL(5,4) NOT NULL,
    weighted_score      DECIMAL(5,2),
    data_status         ENUM(
        'AVAILABLE',
        'NOT_AVAILABLE',
        'INSUFFICIENT_HISTORY',
        'NOT_APPLICABLE',
        'USING_PROXY'
    ) NOT NULL,
    source_period_start DATE,
    source_period_end   DATE,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

# 23. Validasi dan Tata Kelola Model

Sebelum MCS digunakan untuk keputusan pembiayaan nyata, metodologi harus:

1. Diuji menggunakan data historis koperasi.
2. Divalidasi oleh ahli risiko kredit.
3. Dikalibrasi terhadap default rate aktual.
4. Dievaluasi terhadap potensi bias.
5. Memiliki dokumentasi model dan formula.
6. Memiliki audit trail perubahan.
7. Menyediakan mekanisme banding atau koreksi data.
8. Tidak menggunakan data sensitif yang tidak relevan.
9. Tidak menjadikan MCS sebagai satu-satunya dasar keputusan kredit.
10. Mematuhi regulasi penilaian kredit dan pelindungan data.

---

# 24. Catatan Implementasi MVP

Untuk kebutuhan hackathon, pendekatan yang direkomendasikan:

```text
Tahap 1:
Gunakan rule-based 5C.

Tahap 2:
Simpan seluruh nilai indikator dan hasil normalisasi.

Tahap 3:
Tampilkan MCS, grade, faktor positif, dan faktor perbaikan.

Tahap 4:
Gunakan dataset sintetis untuk demo XGBoost apabila diperlukan.

Tahap 5:
Jelaskan bahwa model produksi memerlukan data historis dan kalibrasi.
```

Bobot dan threshold pada dokumen ini merupakan rancangan awal, bukan standar resmi penilaian kredit koperasi.

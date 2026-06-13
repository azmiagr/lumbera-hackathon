package service

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ISavingsBookService interface {
	GetSavingsBook(req model.GetSavingsBookRequest) (*model.SavingsBookResponse, error)
	ExportSavingsBookXLSX(req model.ExportSavingsBookRequest) ([]byte, string, error)
	ExportSavingsBookPDF(req model.ExportSavingsBookRequest) ([]byte, string, error)
}

type SavingsBookService struct {
	deps serviceDependency
}

func NewSavingsBookService(deps serviceDependency) ISavingsBookService {
	return &SavingsBookService{deps: deps}
}

func (s *SavingsBookService) GetSavingsBook(req model.GetSavingsBookRequest) (*model.SavingsBookResponse, error) {
	if err := validateSavingsBookAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodStart, periodEnd, periodKey, err := parseSavingsBookPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	filterType, err := normalizeSavingsBookType(req.Type)
	if err != nil {
		return nil, err
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	profile, err := s.deps.repository.SavingsBookRepository.GetProfile(tx, req.UserID, req.CooperativeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil profil anggota")
	}

	summary, err := s.deps.repository.SavingsBookRepository.GetSummary(tx, req.CooperativeID, profile.MemberID, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan buku tabungan")
	}

	items, total, err := s.deps.repository.SavingsBookRepository.ListItems(tx, req.CooperativeID, profile.MemberID, periodStart, periodEnd, filterType, req.Page, req.Limit)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil buku tabungan")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil buku tabungan")
	}

	for i := range items {
		enrichSavingsBookItem(&items[i])
	}

	return &model.SavingsBookResponse{
		Profile: *profile,
		Period:  periodKey,
		Summary: *summary,
		Items:   items,
		Page:    req.Page,
		Limit:   req.Limit,
		Total:   total,
	}, nil
}

func (s *SavingsBookService) ExportSavingsBookXLSX(req model.ExportSavingsBookRequest) ([]byte, string, error) {
	result, err := s.GetSavingsBook(model.GetSavingsBookRequest{
		AuthContext: req.AuthContext,
		Period:      req.Period,
		Type:        req.Type,
		Page:        1,
		Limit:       1000,
	})
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	defer file.Close()

	sheet := "Buku Tabungan"
	_ = file.SetSheetName("Sheet1", sheet)

	if err := writeSavingsBookSheet(file, sheet, result); err != nil {
		return nil, "", err
	}

	buffer := bytes.NewBuffer(nil)
	if err := file.Write(buffer); err != nil {
		return nil, "", appErrors.InternalServer("gagal menulis file buku tabungan")
	}

	fileName := fmt.Sprintf("buku-tabungan-%s-%s.xlsx", result.Profile.MemberNumber, result.Period)
	return buffer.Bytes(), fileName, nil
}

func (s *SavingsBookService) ExportSavingsBookPDF(req model.ExportSavingsBookRequest) ([]byte, string, error) {
	result, err := s.GetSavingsBook(model.GetSavingsBookRequest{
		AuthContext: req.AuthContext,
		Period:      req.Period,
		Type:        req.Type,
		Page:        1,
		Limit:       1000,
	})
	if err != nil {
		return nil, "", err
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.AddPage()

	writeSavingsBookPDF(pdf, result)

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, "", appErrors.InternalServer("gagal menulis file PDF buku tabungan")
	}

	fileName := fmt.Sprintf("buku-tabungan-%s-%s.pdf", result.Profile.MemberNumber, result.Period)
	return buffer.Bytes(), fileName, nil
}

func writeSavingsBookPDF(pdf *gofpdf.Fpdf, result *model.SavingsBookResponse) {
	tealR, tealG, tealB := 15, 118, 110

	pdf.SetFont("Arial", "", 26)
	pdf.SetTextColor(tealR, tealG, tealB)
	pdf.CellFormat(0, 16, "BUKU TABUNGAN ANGGOTA - LUMBERA", "", 1, "C", false, 0, "")
	pdf.Ln(12)

	labelX := 24.0
	valueX := 94.0
	y := pdf.GetY()

	pdf.SetFont("Arial", "", 15)
	pdf.SetTextColor(0, 0, 0)

	summary := []struct {
		Label   string
		Value   string
		R, G, B int
	}{
		{"Nama", ": " + result.Profile.FullName, 0, 0, 0},
		{"No. Anggota", ": " + result.Profile.MemberNumber, 0, 0, 0},
		{"Total Saldo", ": " + formatSavingsBookRupiah(result.Summary.TotalBalance), tealR, tealG, tealB},
		{"Total Pemasukan", ": " + formatSavingsBookRupiah(result.Summary.TotalIncome), 0, 128, 0},
		{"Total Pengeluaran", ": " + formatSavingsBookRupiah(result.Summary.TotalExpense), 220, 38, 38},
	}

	for _, row := range summary {
		pdf.SetXY(labelX, y)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(65, 9, row.Label, "", 0, "L", false, 0, "")

		pdf.SetXY(valueX, y)
		pdf.SetTextColor(row.R, row.G, row.B)
		pdf.CellFormat(100, 9, row.Value, "", 1, "L", false, 0, "")

		y += 10
	}

	pdf.Ln(24)

	headers := []string{"Tanggal", "Jenis Transaksi", "Pencatat", "Pemasukan", "Pengeluaran"}
	widths := []float64{50, 58, 50, 50, 50}

	pdf.SetFont("Arial", "", 14)
	pdf.SetFillColor(tealR, tealG, tealB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.45)

	for i, header := range headers {
		pdf.CellFormat(widths[i], 16, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 13)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(255, 255, 255)

	for _, item := range result.Items {
		income := "-"
		expense := "-"
		if item.Direction == "IN" {
			income = formatSavingsBookRupiah(item.IncomeAmount)
		}
		if item.Direction == "OUT" {
			expense = formatSavingsBookRupiah(item.ExpenseAmount)
		}

		values := []string{
			formatIndonesianDate(item.RecordedAt),
			truncatePDFText(item.TransactionTypeLabel, 22),
			truncatePDFText(item.RecorderName, 20),
			income,
			expense,
		}

		aligns := []string{"L", "L", "L", "R", "R"}

		for i, value := range values {
			pdf.CellFormat(widths[i], 16, value, "1", 0, aligns[i], false, 0, "")
		}
		pdf.Ln(-1)
	}
}

func truncatePDFText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func validateSavingsBookAccess(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}
	if auth.RoleCode != constants.RoleCodeAnggota {
		return appErrors.Forbidden("hanya anggota yang dapat melihat buku tabungan")
	}
	return nil
}

func parseSavingsBookPeriod(value string) (time.Time, time.Time, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = time.Now().Format("2006-01")
	}

	month, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, time.Time{}, "", appErrors.BadRequest("periode harus berformat YYYY-MM")
	}

	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end, start.Format("2006-01"), nil
}

func normalizeSavingsBookType(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return model.SavingsBookTypeAll, nil
	}
	switch value {
	case model.SavingsBookTypeAll, model.SavingsBookTypeIncome, model.SavingsBookTypeExpense:
		return value, nil
	default:
		return "", appErrors.BadRequest("type harus SEMUA, PEMASUKAN, atau PENGELUARAN")
	}
}

func enrichSavingsBookItem(item *model.SavingsBookItem) {
	item.TransactionTypeLabel = getTransactionTypeLabel(item.TransactionType)
	item.RecorderName = strings.TrimSpace(item.RecorderName)
	if item.RecorderName == "" {
		item.RecorderName = "Otomatis"
	}

	switch item.TransactionType {
	case constants.TransactionTypeCashWithdrawal, constants.TransactionTypeInstallment:
		item.Direction = "OUT"
		item.ExpenseAmount = item.Amount
	default:
		item.Direction = "IN"
		item.IncomeAmount = item.Amount
	}
}

func writeSavingsBookSheet(file *excelize.File, sheet string, result *model.SavingsBookResponse) error {
	titleStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 20, Color: "0F766E"},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	labelStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	valueStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 12, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF", Family: "Times New Roman"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "C9C9C9", Style: 1},
			{Type: "right", Color: "C9C9C9", Style: 1},
			{Type: "top", Color: "C9C9C9", Style: 1},
			{Type: "bottom", Color: "C9C9C9", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	bodyStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 12, Family: "Times New Roman"},
		Border: []excelize.Border{
			{Type: "left", Color: "C9C9C9", Style: 1},
			{Type: "right", Color: "C9C9C9", Style: 1},
			{Type: "top", Color: "C9C9C9", Style: 1},
			{Type: "bottom", Color: "C9C9C9", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	amountStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 12, Family: "Times New Roman"},
		Border: []excelize.Border{
			{Type: "left", Color: "C9C9C9", Style: 1},
			{Type: "right", Color: "C9C9C9", Style: 1},
			{Type: "top", Color: "C9C9C9", Style: 1},
			{Type: "bottom", Color: "C9C9C9", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	_ = file.MergeCell(sheet, "A1", "E1")
	_ = file.SetCellValue(sheet, "A1", "BUKU TABUNGAN ANGGOTA - LUMBERA")
	_ = file.SetCellStyle(sheet, "A1", "E1", titleStyle)
	_ = file.SetRowHeight(sheet, 1, 28)

	summaryRows := []struct {
		Row   int
		Label string
		Value string
	}{
		{3, "Nama", ": " + result.Profile.FullName},
		{4, "No. Anggota", ": " + result.Profile.MemberNumber},
		{5, "Total Saldo", ": " + formatSavingsBookRupiah(result.Summary.TotalBalance)},
		{6, "Total Pemasukan", ": " + formatSavingsBookRupiah(result.Summary.TotalIncome)},
		{7, "Total Pengeluaran", ": " + formatSavingsBookRupiah(result.Summary.TotalExpense)},
	}

	for _, item := range summaryRows {
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", item.Row), item.Label)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", item.Row), item.Value)
		_ = file.SetCellStyle(sheet, fmt.Sprintf("A%d", item.Row), fmt.Sprintf("A%d", item.Row), labelStyle)
		_ = file.SetCellStyle(sheet, fmt.Sprintf("B%d", item.Row), fmt.Sprintf("B%d", item.Row), valueStyle)
	}

	headers := []string{"Tanggal", "Jenis Transaksi", "Pencatat", "Pemasukan", "Pengeluaran"}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 10)
		_ = file.SetCellValue(sheet, cell, header)
	}
	_ = file.SetCellStyle(sheet, "A10", "E10", headerStyle)
	_ = file.SetRowHeight(sheet, 10, 20)

	for index, item := range result.Items {
		row := index + 11

		income := "-"
		expense := "-"
		if item.Direction == "IN" {
			income = formatSavingsBookRupiah(item.IncomeAmount)
		}
		if item.Direction == "OUT" {
			expense = formatSavingsBookRupiah(item.ExpenseAmount)
		}

		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", row), formatIndonesianDate(item.RecordedAt))
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.TransactionTypeLabel)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.RecorderName)
		_ = file.SetCellValue(sheet, fmt.Sprintf("D%d", row), income)
		_ = file.SetCellValue(sheet, fmt.Sprintf("E%d", row), expense)

		_ = file.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), bodyStyle)
		_ = file.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("E%d", row), amountStyle)
		_ = file.SetRowHeight(sheet, row, 18)
	}

	_ = file.SetColWidth(sheet, "A", "A", 18)
	_ = file.SetColWidth(sheet, "B", "B", 30)
	_ = file.SetColWidth(sheet, "C", "C", 22)
	_ = file.SetColWidth(sheet, "D", "E", 18)

	return nil
}

func formatIndonesianDate(value time.Time) string {
	months := []string{
		"Januari",
		"Februari",
		"Maret",
		"April",
		"Mei",
		"Juni",
		"Juli",
		"Agustus",
		"September",
		"Oktober",
		"November",
		"Desember",
	}

	month := months[int(value.Month())-1]
	return fmt.Sprintf("%d %s %d", value.Day(), month, value.Year())
}

func formatSavingsBookRupiah(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	raw := fmt.Sprintf("%d", value)
	parts := make([]string, 0)

	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)

	return sign + "Rp " + strings.Join(parts, ".")
}

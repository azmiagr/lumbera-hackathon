package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type IReportService interface {
	GetFinancialReport(req model.FinancialReportRequest) (*model.FinancialReportResponse, error)
	ExportFinancialReportXLSX(req model.FinancialReportRequest) ([]byte, string, error)
}

type ReportService struct {
	deps serviceDependency
}

func NewReportService(deps serviceDependency) IReportService {
	return &ReportService{deps: deps}
}

func (s *ReportService) GetFinancialReport(req model.FinancialReportRequest) (*model.FinancialReportResponse, error) {
	if err := validateReportAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodEnd, err := parseReportPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	columns := buildReportColumns(periodEnd)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	monthlyBalances := make(map[string][]repository.AccountBalanceRow)
	cumulativeBalances := make(map[string][]repository.AccountBalanceRow)

	for _, column := range columns {
		monthStart, monthEnd := monthRange(column.Key)
		cumulativeStart := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

		monthlyRows, err := s.deps.repository.AccountingRepository.GetAccountBalances(
			tx,
			req.CooperativeID,
			monthStart,
			monthEnd,
		)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mengambil saldo laporan")
		}

		cumulativeRows, err := s.deps.repository.AccountingRepository.GetAccountBalances(
			tx,
			req.CooperativeID,
			cumulativeStart,
			monthEnd,
		)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mengambil saldo neraca")
		}

		monthlyBalances[column.Key] = monthlyRows
		cumulativeBalances[column.Key] = cumulativeRows
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil laporan keuangan")
	}

	return &model.FinancialReportResponse{
		PeriodColumns:   columns,
		BalanceSheet:    buildBalanceSheetRows(columns, cumulativeBalances),
		IncomeStatement: buildIncomeStatementRows(columns, monthlyBalances),
		CashFlow:        buildCashFlowRows(columns, monthlyBalances),
	}, nil
}

func (s *ReportService) ExportFinancialReportXLSX(req model.FinancialReportRequest) ([]byte, string, error) {
	report, err := s.GetFinancialReport(req)
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	defer file.Close()

	_ = file.SetSheetName("Sheet1", "Neraca")

	if err := writeReportSheet(file, "Neraca", report.PeriodColumns, report.BalanceSheet); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat sheet neraca")
	}

	if _, err := file.NewSheet("Laba Rugi"); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat sheet laba rugi")
	}
	if err := writeReportSheet(file, "Laba Rugi", report.PeriodColumns, report.IncomeStatement); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat sheet laba rugi")
	}

	if _, err := file.NewSheet("Arus Kas"); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat sheet arus kas")
	}
	if err := writeReportSheet(file, "Arus Kas", report.PeriodColumns, report.CashFlow); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat sheet arus kas")
	}

	file.SetActiveSheet(0)

	buffer := bytes.NewBuffer(nil)
	if err := file.Write(buffer); err != nil {
		return nil, "", appErrors.InternalServer("gagal menulis file laporan")
	}

	period := report.PeriodColumns[len(report.PeriodColumns)-1].Key
	fileName := fmt.Sprintf("laporan-keuangan-%s.xlsx", period)

	return buffer.Bytes(), fileName, nil
}

func validateReportAccess(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}

	if auth.RoleCode != constants.RoleCodePengurusKoperasi {
		return appErrors.Forbidden("hanya pengurus koperasi yang dapat melihat laporan")
	}

	return nil
}

func parseReportPeriod(period string) (time.Time, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	parsed, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, appErrors.BadRequest("format periode harus YYYY-MM")
	}

	return parsed, nil
}

func buildReportColumns(periodEnd time.Time) []model.ReportColumn {
	columns := make([]model.ReportColumn, 0, 3)

	for offset := 2; offset >= 0; offset-- {
		month := periodEnd.AddDate(0, -offset, 0)
		key := month.Format("2006-01")
		columns = append(columns, model.ReportColumn{
			Key:   key,
			Label: formatIndonesianMonth(month),
		})
	}

	return columns
}

func monthRange(periodKey string) (time.Time, time.Time) {
	start, _ := time.Parse("2006-01", periodKey)
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}

func formatIndonesianMonth(value time.Time) string {
	months := []string{
		"JANUARI", "FEBRUARI", "MARET", "APRIL", "MEI", "JUNI",
		"JULI", "AGUSTUS", "SEPTEMBER", "OKTOBER", "NOVEMBER", "DESEMBER",
	}

	return fmt.Sprintf("%s %d", months[int(value.Month())-1], value.Year())
}

func buildBalanceSheetRows(columns []model.ReportColumn, balances map[string][]repository.AccountBalanceRow) []model.ReportRow {
	rows := []model.ReportRow{
		{Section: "AKTIVA", Label: "Uang Kas", Values: valuesByAccount(columns, balances, constants.AccountCodeCash)},
		{Section: "AKTIVA", Label: "Simpanan di Bank", Values: valuesByAccount(columns, balances, constants.AccountCodeBank)},
		{Section: "AKTIVA", Label: "Piutang Pinjaman", Values: valuesByAccount(columns, balances, constants.AccountCodeLoanReceivable)},
	}

	totalAsset := sumRows(columns, rows)
	rows = append(rows,
		model.ReportRow{Section: "AKTIVA", Label: "Total Aktiva Lancar", Values: totalAsset, IsTotal: true},
		model.ReportRow{Section: "AKTIVA", Label: "TOTAL AKTIVA", Values: totalAsset, IsTotal: true},
	)

	liabilityEquityRows := []model.ReportRow{
		{Section: "KEWAJIBAN & MODAL", Label: "Simpanan Anggota", Values: valuesByAccount(columns, balances, constants.AccountCodeMemberSavings)},
		{Section: "KEWAJIBAN & MODAL", Label: "Modal Koperasi", Values: valuesByAccount(columns, balances, constants.AccountCodePrincipalSavings)},
	}

	totalLiabilityEquity := sumRows(columns, liabilityEquityRows)
	rows = append(rows, liabilityEquityRows...)
	rows = append(rows, model.ReportRow{
		Section: "KEWAJIBAN & MODAL",
		Label:   "TOTAL",
		Values:  totalLiabilityEquity,
		IsTotal: true,
	})

	return rows
}

func buildIncomeStatementRows(columns []model.ReportColumn, balances map[string][]repository.AccountBalanceRow) []model.ReportRow {
	revenueRows := []model.ReportRow{
		{Section: "PENDAPATAN USAHA", Label: "Pendapatan Bunga Pinjaman", Values: valuesByAccount(columns, balances, constants.AccountCodeLoanInterest)},
		{Section: "PENDAPATAN USAHA", Label: "Pendapatan Administrasi", Values: valuesByAccount(columns, balances, constants.AccountCodeAdminRevenue)},
	}

	expenseRows := []model.ReportRow{
		{Section: "BEBAN USAHA", Label: "Beban Operasional", Values: valuesByAccount(columns, balances, constants.AccountCodeOperatingExpense)},
		{Section: "BEBAN USAHA", Label: "Beban Gaji Pengurus", Values: valuesByAccount(columns, balances, constants.AccountCodeSalaryExpense)},
		{Section: "BEBAN USAHA", Label: "Beban Penyusutan", Values: valuesByAccount(columns, balances, constants.AccountCodeDepreciation)},
		{Section: "BEBAN USAHA", Label: "Beban ATK & Umum", Values: valuesByAccount(columns, balances, constants.AccountCodeOfficeSupplies)},
	}

	totalRevenue := sumRows(columns, revenueRows)
	totalExpense := sumRows(columns, expenseRows)
	netIncome := subtractValues(columns, totalRevenue, totalExpense)

	rows := append([]model.ReportRow{}, revenueRows...)
	rows = append(rows, model.ReportRow{Section: "PENDAPATAN USAHA", Label: "Total Pendapatan", Values: totalRevenue, IsTotal: true})
	rows = append(rows, expenseRows...)
	rows = append(rows, model.ReportRow{Section: "BEBAN USAHA", Label: "TOTAL", Values: totalExpense, IsTotal: true})
	rows = append(rows, model.ReportRow{Section: "LABA RUGI", Label: "Laba Bersih", Values: netIncome, IsTotal: true})

	return rows
}

func buildCashFlowRows(columns []model.ReportColumn, balances map[string][]repository.AccountBalanceRow) []model.ReportRow {
	operatingRows := []model.ReportRow{
		{Section: "AKTIVITAS OPERASI", Label: "Penerimaan Simpanan", Values: sumAccounts(columns, balances, []string{
			constants.AccountCodeMemberSavings,
			constants.AccountCodePrincipalSavings,
		})},
		{Section: "AKTIVITAS OPERASI", Label: "Penerimaan Angsuran", Values: valuesByAccount(columns, balances, constants.AccountCodeLoanReceivable)},
		{Section: "AKTIVITAS OPERASI", Label: "Beban Operasional", Values: negativeValues(valuesByAccount(columns, balances, constants.AccountCodeOperatingExpense))},
	}

	investingRows := []model.ReportRow{
		{Section: "AKTIVITAS INVESTASI", Label: "Pembelian Inventaris", Values: negativeValues(valuesByAccount(columns, balances, constants.AccountCodeInventoryPurchase))},
	}

	financingRows := []model.ReportRow{
		{Section: "AKTIVITAS PENDANAAN", Label: "Penambahan Modal", Values: valuesByAccount(columns, balances, constants.AccountCodePrincipalSavings)},
	}

	rows := append([]model.ReportRow{}, operatingRows...)
	rows = append(rows, model.ReportRow{Section: "AKTIVITAS OPERASI", Label: "Bersih Operasi", Values: sumRows(columns, operatingRows), IsTotal: true})
	rows = append(rows, investingRows...)
	rows = append(rows, model.ReportRow{Section: "AKTIVITAS INVESTASI", Label: "Bersih Investasi", Values: sumRows(columns, investingRows), IsTotal: true})
	rows = append(rows, financingRows...)
	rows = append(rows, model.ReportRow{Section: "AKTIVITAS PENDANAAN", Label: "Bersih Pendanaan", Values: sumRows(columns, financingRows), IsTotal: true})

	return rows
}

func valuesByAccount(columns []model.ReportColumn, balances map[string][]repository.AccountBalanceRow, accountCode string) map[string]int64 {
	values := emptyValues(columns)

	for _, column := range columns {
		for _, row := range balances[column.Key] {
			if row.AccountCode == accountCode {
				values[column.Key] = accountBalance(row)
				break
			}
		}
	}

	return values
}

func sumAccounts(columns []model.ReportColumn, balances map[string][]repository.AccountBalanceRow, accountCodes []string) map[string]int64 {
	values := emptyValues(columns)

	for _, accountCode := range accountCodes {
		accountValues := valuesByAccount(columns, balances, accountCode)
		for _, column := range columns {
			values[column.Key] += accountValues[column.Key]
		}
	}

	return values
}

func accountBalance(row repository.AccountBalanceRow) int64 {
	if row.NormalBalance == constants.NormalBalanceDebit {
		return row.TotalDebit - row.TotalCredit
	}

	return row.TotalCredit - row.TotalDebit
}

func emptyValues(columns []model.ReportColumn) map[string]int64 {
	values := make(map[string]int64, len(columns))
	for _, column := range columns {
		values[column.Key] = 0
	}
	return values
}

func sumRows(columns []model.ReportColumn, rows []model.ReportRow) map[string]int64 {
	values := emptyValues(columns)
	for _, row := range rows {
		for _, column := range columns {
			values[column.Key] += row.Values[column.Key]
		}
	}
	return values
}

func subtractValues(columns []model.ReportColumn, left map[string]int64, right map[string]int64) map[string]int64 {
	values := emptyValues(columns)
	for _, column := range columns {
		values[column.Key] = left[column.Key] - right[column.Key]
	}
	return values
}

func negativeValues(values map[string]int64) map[string]int64 {
	result := make(map[string]int64, len(values))
	for key, value := range values {
		if value > 0 {
			result[key] = -value
		} else {
			result[key] = value
		}
	}
	return result
}

func writeReportSheet(file *excelize.File, sheetName string, columns []model.ReportColumn, rows []model.ReportRow) error {
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1},
	})
	if err != nil {
		return err
	}

	totalStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "0F766E"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E0F2F1"}, Pattern: 1},
	})
	if err != nil {
		return err
	}

	negativeStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "DC2626"},
	})
	if err != nil {
		return err
	}

	if err := file.SetCellValue(sheetName, "A1", strings.ToUpper(sheetName)); err != nil {
		return err
	}

	if err := file.SetCellValue(sheetName, "A3", "KETERANGAN"); err != nil {
		return err
	}
	if err := file.SetCellStyle(sheetName, "A3", "A3", headerStyle); err != nil {
		return err
	}

	for index, column := range columns {
		cell, _ := excelize.CoordinatesToCellName(index+2, 3)
		if err := file.SetCellValue(sheetName, cell, column.Label); err != nil {
			return err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return err
		}
	}

	rowNumber := 4
	currentSection := ""

	for _, row := range rows {
		if row.Section != "" && row.Section != currentSection {
			currentSection = row.Section
			sectionCell := fmt.Sprintf("A%d", rowNumber)
			if err := file.SetCellValue(sheetName, sectionCell, currentSection); err != nil {
				return err
			}

			endCell, _ := excelize.CoordinatesToCellName(len(columns)+1, rowNumber)
			if err := file.SetCellStyle(sheetName, sectionCell, endCell, headerStyle); err != nil {
				return err
			}
			rowNumber++
		}

		labelCell := fmt.Sprintf("A%d", rowNumber)
		if err := file.SetCellValue(sheetName, labelCell, row.Label); err != nil {
			return err
		}

		for index, column := range columns {
			cell, _ := excelize.CoordinatesToCellName(index+2, rowNumber)
			value := row.Values[column.Key]
			if err := file.SetCellValue(sheetName, cell, value); err != nil {
				return err
			}
			if value < 0 {
				if err := file.SetCellStyle(sheetName, cell, cell, negativeStyle); err != nil {
					return err
				}
			}
		}

		if row.IsTotal {
			endCell, _ := excelize.CoordinatesToCellName(len(columns)+1, rowNumber)
			if err := file.SetCellStyle(sheetName, labelCell, endCell, totalStyle); err != nil {
				return err
			}
		}

		rowNumber++
	}

	_ = file.SetColWidth(sheetName, "A", "A", 30)
	_ = file.SetColWidth(sheetName, "B", "D", 18)

	return nil
}

package service

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type IReportService interface {
	GetFinancialReport(req model.FinancialReportRequest) (*model.FinancialReportResponse, error)
	ExportFinancialReportXLSX(req model.FinancialReportRequest) ([]byte, string, error)
	GetCooperativeHealthScore(req model.CooperativeHealthScoreRequest) (*model.CooperativeHealthScoreResponse, error)
	GetDashboardSummary(req model.DashboardSummaryRequest) (*model.DashboardSummaryResponse, error)
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

func (s *ReportService) GetCooperativeHealthScore(req model.CooperativeHealthScoreRequest) (*model.CooperativeHealthScoreResponse, error) {
	if err := validateReportAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodEndMonth, err := parseReportPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	periodKey := periodEndMonth.Format("2006-01")
	periodStart, periodEnd := monthRange(periodKey)
	previousStart := periodStart.AddDate(0, -1, 0)
	previousEnd := periodStart.Add(-time.Nanosecond)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	cumulativeStart := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	cumulativeRows, err := s.deps.repository.AccountingRepository.GetAccountBalances(tx, req.CooperativeID, cumulativeStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil saldo CHS")
	}

	monthlyRows, err := s.deps.repository.AccountingRepository.GetAccountBalances(tx, req.CooperativeID, periodStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil saldo CHS")
	}

	previousCumulativeRows, err := s.deps.repository.AccountingRepository.GetAccountBalances(tx, req.CooperativeID, cumulativeStart, previousEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil saldo CHS")
	}

	financial, err := s.buildFinancialCHSDimension(tx, req.CooperativeID, periodEnd, cumulativeRows, monthlyRows, previousCumulativeRows)
	if err != nil {
		return nil, err
	}

	operational, err := s.buildOperationalCHSDimension(tx, req.CooperativeID, periodStart, periodEnd, previousStart, previousEnd, monthlyRows)
	if err != nil {
		return nil, err
	}

	dataQuality, err := s.buildDataQualityCHSDimension(tx, req.CooperativeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	compliance := buildComplianceCHSDimension()

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghitung CHS")
	}

	dimensions := []model.CHSDimensionScore{financial, operational, dataQuality, compliance}
	chsScore, status := calculateCHSTotal(dimensions)
	displayScore := int(math.Round(chsScore))
	grade, category := determineCHSGrade(chsScore)
	if status == "INSUFFICIENT_DATA" {
		grade = ""
		category = "Data Tidak Cukup"
		displayScore = 0
	}

	return &model.CooperativeHealthScoreResponse{
		Period:       periodKey,
		Status:       status,
		CHSScore:     round2(chsScore),
		DisplayScore: displayScore,
		Grade:        grade,
		Category:     category,
		Dimensions:   dimensions,
	}, nil
}

func (s *ReportService) GetDashboardSummary(req model.DashboardSummaryRequest) (*model.DashboardSummaryResponse, error) {
	if err := validateReportAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodEnd, err := parseReportPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	period := periodEnd.Format("2006-01")

	chs, err := s.GetCooperativeHealthScore(model.CooperativeHealthScoreRequest{
		AuthContext: req.AuthContext,
		Period:      period,
	})
	if err != nil {
		return nil, err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	activeMembers, err := s.deps.repository.MemberRepository.CountActiveMembersByCooperative(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghitung anggota aktif")
	}

	registeredMembers, err := s.deps.repository.MemberRepository.CountMembersByCooperative(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghitung anggota terdaftar")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan dashboard")
	}

	return &model.DashboardSummaryResponse{
		Period:      period,
		PeriodLabel: formatIndonesianMonth(periodEnd),
		CHS: model.DashboardCHSSummary{
			Score:        chs.CHSScore,
			DisplayScore: chs.DisplayScore,
			Grade:        chs.Grade,
			Category:     chs.Category,
			Status:       chs.Status,
		},
		Members: model.DashboardMemberSummary{
			Active:     activeMembers,
			Registered: registeredMembers,
		},
	}, nil
}

func (s *ReportService) buildFinancialCHSDimension(tx *gorm.DB, cooperativeID uuid.UUID, periodEnd time.Time, cumulativeRows, monthlyRows, previousCumulativeRows []repository.AccountBalanceRow) (model.CHSDimensionScore, error) {
	loanRisk, err := s.deps.repository.CHSRepository.GetLoanRiskMetrics(tx, cooperativeID, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung NPL")
	}

	indicators := []model.CHSIndicatorScore{
		unavailableCHSIndicator("NPL", "Non-Performing Loan Rate", 0.40, "belum ada pinjaman aktif"),
		unavailableCHSIndicator("CAR", "Rasio Kecukupan Modal", 0.25, "data aset/modal belum tersedia"),
		unavailableCHSIndicator("ROA", "Return on Assets", 0.20, "data laba/aset belum tersedia"),
		unavailableCHSIndicator("LIQUIDITY", "Likuiditas", 0.15, "data aset/kewajiban belum tersedia"),
	}

	if loanRisk.TotalRemainingPrincipal > 0 {
		nplRate := percent(loanRisk.BadRemainingPrincipal, loanRisk.TotalRemainingPrincipal)
		indicators[0] = availableCHSIndicator("NPL", "Non-Performing Loan Rate", nplRate, scoreNPL(nplRate), 0.40)
	}

	totalAssets := sumBalancesByAccountType(cumulativeRows, constants.AccountTypeAsset)
	totalEquity := sumBalancesByAccountType(cumulativeRows, constants.AccountTypeEquity)
	if totalAssets > 0 {
		car := percent(totalEquity, totalAssets)
		indicators[1] = availableCHSIndicator("CAR", "Rasio Kecukupan Modal", car, scoreCapitalAdequacy(car), 0.25)
	}

	periodNetIncome := sumBalancesByAccountType(monthlyRows, constants.AccountTypeRevenue) - sumBalancesByAccountType(monthlyRows, constants.AccountTypeExpense)
	assetStart := sumBalancesByAccountType(previousCumulativeRows, constants.AccountTypeAsset)
	assetEnd := totalAssets
	averageAssets := float64(assetStart+assetEnd) / 2
	if averageAssets > 0 {
		roa := (float64(periodNetIncome) / averageAssets) * 100
		indicators[2] = availableCHSIndicator("ROA", "Return on Assets", roa, scoreROA(roa), 0.20)
	}

	totalLiabilities := sumBalancesByAccountType(cumulativeRows, constants.AccountTypeLiability)
	if totalLiabilities > 0 {
		currentRatio := float64(totalAssets) / float64(totalLiabilities)
		indicators[3] = availableCHSIndicator("LIQUIDITY", "Likuiditas", currentRatio, scoreLiquidity(currentRatio), 0.15)
	}

	return buildCHSDimension("FINANCIAL", "Keuangan", 0.35, indicators), nil
}

func (s *ReportService) buildOperationalCHSDimension(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd, previousStart, previousEnd time.Time, monthlyRows []repository.AccountBalanceRow) (model.CHSDimensionScore, error) {
	indicators := []model.CHSIndicatorScore{
		unavailableCHSIndicator("ON_TIME_PAYMENT", "Pembayaran Tepat Waktu", 0.35, "belum ada jadwal angsuran jatuh tempo"),
		unavailableCHSIndicator("ACTIVE_MEMBER", "Keaktifan Anggota", 0.25, "belum ada anggota aktif"),
		unavailableCHSIndicator("TRANSACTION_GROWTH", "Pertumbuhan Transaksi", 0.20, "belum ada transaksi periode pembanding"),
		unavailableCHSIndicator("OPERATIONAL_EFFICIENCY", "Efisiensi Operasional", 0.20, "data pendapatan/biaya belum tersedia"),
	}

	payment, err := s.deps.repository.CHSRepository.GetOnTimePaymentMetrics(tx, cooperativeID, periodStart, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung pembayaran tepat waktu")
	}
	if payment.TotalDue > 0 {
		onTimeRate := percent(payment.OnTime, payment.TotalDue)
		indicators[0] = availableCHSIndicator("ON_TIME_PAYMENT", "Pembayaran Tepat Waktu", onTimeRate, scoreOnTimePayment(onTimeRate), 0.35)
	}

	memberActivity, err := s.deps.repository.CHSRepository.GetMemberActivityMetrics(tx, cooperativeID, periodStart, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung keaktifan anggota")
	}
	if memberActivity.TotalMembers > 0 {
		activeRate := percent(memberActivity.ActiveMembers, memberActivity.TotalMembers)
		indicators[1] = availableCHSIndicator("ACTIVE_MEMBER", "Keaktifan Anggota", activeRate, scoreActiveMember(activeRate), 0.25)
	}

	growth, err := s.deps.repository.CHSRepository.GetTransactionGrowthMetrics(tx, cooperativeID, periodStart, periodEnd, previousStart, previousEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung pertumbuhan transaksi")
	}
	if growth.PreviousTransactions > 0 {
		growthRate := (float64(growth.CurrentTransactions-growth.PreviousTransactions) / float64(growth.PreviousTransactions)) * 100
		indicators[2] = availableCHSIndicator("TRANSACTION_GROWTH", "Pertumbuhan Transaksi", growthRate, scoreTransactionGrowth(growthRate), 0.20)
	} else if growth.CurrentTransactions > 0 {
		indicators[2] = availableCHSIndicator("TRANSACTION_GROWTH", "Pertumbuhan Transaksi", 100, 100, 0.20)
	}

	revenue := sumBalancesByAccountType(monthlyRows, constants.AccountTypeRevenue)
	expense := sumBalancesByAccountType(monthlyRows, constants.AccountTypeExpense)
	if revenue > 0 {
		expenseRatio := percent(expense, revenue)
		indicators[3] = availableCHSIndicator("OPERATIONAL_EFFICIENCY", "Efisiensi Operasional", expenseRatio, scoreOperationalEfficiency(expenseRatio), 0.20)
	}

	return buildCHSDimension("OPERATIONAL", "Operasional", 0.25, indicators), nil
}

func (s *ReportService) buildDataQualityCHSDimension(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) (model.CHSDimensionScore, error) {
	indicators := []model.CHSIndicatorScore{
		unavailableCHSIndicator("COMPLETENESS", "Kelengkapan Data", 0.35, "belum ada data anggota"),
		unavailableCHSIndicator("SYNC_TIMELINESS", "Ketepatan Sinkronisasi", 0.25, "belum ada transaksi pada periode"),
		unavailableCHSIndicator("CONSISTENCY", "Konsistensi Data", 0.25, "belum ada record yang diperiksa"),
		unavailableCHSIndicator("LEDGER_VALIDITY", "Validitas Ledger", 0.15, "belum ada transaksi ledger"),
	}

	completeness, err := s.deps.repository.CHSRepository.GetDataCompletenessMetrics(tx, cooperativeID)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung kelengkapan data")
	}
	if completeness.TotalFields > 0 {
		completenessRate := percent(completeness.FilledFields, completeness.TotalFields)
		indicators[0] = availableCHSIndicator("COMPLETENESS", "Kelengkapan Data", completenessRate, scoreCompleteness(completenessRate), 0.35)
	}

	syncMetrics, err := s.deps.repository.CHSRepository.GetSyncTimelinessMetrics(tx, cooperativeID, periodStart, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung sinkronisasi")
	}
	if syncMetrics.TotalTransactions > 0 {
		syncRate := percent(syncMetrics.TimelyTransactions, syncMetrics.TotalTransactions)
		indicators[1] = availableCHSIndicator("SYNC_TIMELINESS", "Ketepatan Sinkronisasi", syncRate, scoreSyncTimeliness(syncRate), 0.25)
	}

	consistency, err := s.deps.repository.CHSRepository.GetConsistencyMetrics(tx, cooperativeID, periodStart, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal menghitung konsistensi data")
	}
	if consistency.TotalRecords > 0 {
		consistencyRate := percent(consistency.ConsistentRecords, consistency.TotalRecords)
		indicators[2] = availableCHSIndicator("CONSISTENCY", "Konsistensi Data", consistencyRate, scoreConsistency(consistencyRate), 0.25)
	}

	ledgerTransactions, err := s.deps.repository.CHSRepository.ListLedgerTransactions(tx, cooperativeID, periodEnd)
	if err != nil {
		return model.CHSDimensionScore{}, appErrors.InternalServer("gagal memverifikasi ledger")
	}
	if len(ledgerTransactions) > 0 {
		ledgerRate := calculateLedgerVerificationRate(ledgerTransactions)
		indicators[3] = availableCHSIndicator("LEDGER_VALIDITY", "Validitas Ledger", ledgerRate, scoreLedgerValidity(ledgerRate), 0.15)
	}

	return buildCHSDimension("DATA_QUALITY", "Kualitas Data", 0.20, indicators), nil
}

func buildComplianceCHSDimension() model.CHSDimensionScore {
	return buildCHSDimension("COMPLIANCE", "Kepatuhan", 0.20, []model.CHSIndicatorScore{
		unavailableCHSIndicator("ON_TIME_REPORTING", "Ketepatan Laporan Berkala", 0.35, "histori submit laporan belum tersedia"),
		unavailableCHSIndicator("DOCUMENT_COMPLETENESS", "Kelengkapan Dokumen Wajib", 0.25, "dokumen wajib belum tersedia"),
		unavailableCHSIndicator("RAT", "Pelaksanaan RAT", 0.20, "data RAT belum tersedia"),
		unavailableCHSIndicator("AUDIT_CONSENT", "Audit Trail dan Consent", 0.20, "audit log dan consent belum tersedia"),
	})
}

func availableCHSIndicator(code, label string, rawValue, score, weight float64) model.CHSIndicatorScore {
	raw := round2(rawValue)
	normalizedScore := round2(score)

	return model.CHSIndicatorScore{
		Code:          code,
		Label:         label,
		RawValue:      &raw,
		Score:         &normalizedScore,
		Weight:        weight,
		WeightedScore: round2(normalizedScore * weight),
		Status:        "AVAILABLE",
	}
}

func unavailableCHSIndicator(code, label string, weight float64, message string) model.CHSIndicatorScore {
	return model.CHSIndicatorScore{
		Code:    code,
		Label:   label,
		Weight:  weight,
		Status:  "UNAVAILABLE",
		Message: message,
	}
}

func buildCHSDimension(code, label string, weight float64, indicators []model.CHSIndicatorScore) model.CHSDimensionScore {
	var availableWeight float64
	var weightedScore float64
	availableCount := 0

	for _, indicator := range indicators {
		if indicator.Score == nil {
			continue
		}

		availableWeight += indicator.Weight
		weightedScore += (*indicator.Score) * indicator.Weight
		availableCount++
	}

	status := "UNAVAILABLE"
	score := 0.0
	if availableCount > 0 {
		score = weightedScore / availableWeight
		status = "PARTIAL"
		if availableCount == len(indicators) {
			status = "COMPLETE"
		}
	}

	return model.CHSDimensionScore{
		Code:       code,
		Label:      label,
		Weight:     weight,
		Score:      round2(score),
		Status:     status,
		Indicators: indicators,
	}
}

func calculateCHSTotal(dimensions []model.CHSDimensionScore) (float64, string) {
	var availableWeight float64
	var weightedScore float64
	availableCount := 0
	completeCount := 0

	for _, dimension := range dimensions {
		if dimension.Status == "UNAVAILABLE" {
			continue
		}

		availableWeight += dimension.Weight
		weightedScore += dimension.Score * dimension.Weight
		availableCount++
		if dimension.Status == "COMPLETE" {
			completeCount++
		}
	}

	if availableCount == 0 || availableWeight == 0 {
		return 0, "INSUFFICIENT_DATA"
	}

	status := "PARTIAL"
	if completeCount == len(dimensions) {
		status = "COMPLETE"
	}

	return weightedScore / availableWeight, status
}

func determineCHSGrade(score float64) (string, string) {
	switch {
	case score >= 85:
		return "AA", "Sangat Sehat"
	case score >= 75:
		return "A", "Sehat"
	case score >= 65:
		return "B", "Cukup Sehat"
	case score >= 50:
		return "C", "Kurang Sehat"
	default:
		return "D", "Tidak Sehat"
	}
}

func percent(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}

	return (float64(numerator) / float64(denominator)) * 100
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func sumBalancesByAccountType(rows []repository.AccountBalanceRow, accountType string) int64 {
	var total int64
	for _, row := range rows {
		if row.AccountType == accountType {
			total += accountBalance(row)
		}
	}

	return total
}

func calculateLedgerVerificationRate(transactions []entity.Transaction) float64 {
	if len(transactions) == 0 {
		return 0
	}

	validRecords := 0
	expectedPrevHash := genesisTransactionHash
	for _, transaction := range transactions {
		if transaction.PrevHash == expectedPrevHash && transaction.CurrentHash == buildTransactionHash(&transaction) {
			validRecords++
		}
		expectedPrevHash = transaction.CurrentHash
	}

	return percent(int64(validRecords), int64(len(transactions)))
}

func scoreNPL(value float64) float64 {
	switch {
	case value <= 2:
		return 100
	case value <= 5:
		return 85
	case value <= 8:
		return 65
	case value <= 12:
		return 40
	default:
		return 20
	}
}

func scoreCapitalAdequacy(value float64) float64 {
	switch {
	case value >= 20:
		return 100
	case value >= 15:
		return 85
	case value >= 10:
		return 70
	case value >= 5:
		return 45
	default:
		return 20
	}
}

func scoreROA(value float64) float64 {
	switch {
	case value >= 5:
		return 100
	case value >= 3:
		return 85
	case value >= 1:
		return 70
	case value >= 0:
		return 50
	default:
		return 20
	}
}

func scoreLiquidity(value float64) float64 {
	switch {
	case value >= 1.5 && value <= 2.5:
		return 100
	case value > 2.5 && value <= 3:
		return 90
	case value > 3:
		return 75
	case value >= 1.2:
		return 85
	case value >= 1:
		return 70
	case value >= 0.8:
		return 45
	default:
		return 20
	}
}

func scoreOnTimePayment(value float64) float64 {
	switch {
	case value >= 95:
		return 100
	case value >= 90:
		return 85
	case value >= 80:
		return 70
	case value >= 70:
		return 50
	default:
		return 25
	}
}

func scoreActiveMember(value float64) float64 {
	switch {
	case value >= 80:
		return 100
	case value >= 65:
		return 85
	case value >= 50:
		return 70
	case value >= 35:
		return 50
	default:
		return 25
	}
}

func scoreTransactionGrowth(value float64) float64 {
	switch {
	case value >= 15:
		return 100
	case value >= 5:
		return 85
	case value >= 0:
		return 70
	case value >= -10:
		return 50
	default:
		return 25
	}
}

func scoreOperationalEfficiency(value float64) float64 {
	switch {
	case value <= 60:
		return 100
	case value <= 75:
		return 85
	case value <= 90:
		return 70
	case value <= 100:
		return 50
	default:
		return 25
	}
}

func scoreCompleteness(value float64) float64 {
	switch {
	case value >= 98:
		return 100
	case value >= 95:
		return 85
	case value >= 90:
		return 70
	case value >= 80:
		return 50
	default:
		return 25
	}
}

func scoreSyncTimeliness(value float64) float64 {
	return scoreCompleteness(value)
}

func scoreConsistency(value float64) float64 {
	switch {
	case value >= 99:
		return 100
	case value >= 97:
		return 85
	case value >= 94:
		return 70
	case value >= 90:
		return 50
	default:
		return 25
	}
}

func scoreLedgerValidity(value float64) float64 {
	switch {
	case value == 100:
		return 100
	case value >= 99.9:
		return 85
	case value >= 99:
		return 60
	default:
		return 20
	}
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

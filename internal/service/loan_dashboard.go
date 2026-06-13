package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILoanDashboardService interface {
	GetLoanDashboard(req model.GetLoanDashboardRequest) (*model.LoanDashboardResponse, error)
}

type LoanDashboardService struct {
	deps serviceDependency
}

func NewLoanDashboardService(deps serviceDependency) ILoanDashboardService {
	return &LoanDashboardService{deps: deps}
}

func (s *LoanDashboardService) GetLoanDashboard(req model.GetLoanDashboardRequest) (*model.LoanDashboardResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}
	if req.RoleCode != constants.RoleCodeAnggota {
		return nil, appErrors.Forbidden("hanya anggota yang dapat melihat dashboard pinjaman")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	profile, err := s.deps.repository.LoanDashboardRepository.GetMemberProfile(tx, req.UserID, req.CooperativeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil profil anggota")
	}

	member, err := s.deps.repository.MemberRepository.GetActiveMember(tx, req.CooperativeID, profile.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil skor anggota")
	}

	snapshot, err := s.deps.repository.MCSRepository.GetLatestScoreSnapshot(tx, req.CooperativeID, profile.MemberID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil snapshot MCS")
	}

	fiveCScores, err := s.deps.repository.MCSRepository.GetLatestFiveCScores(tx, req.CooperativeID, profile.MemberID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil skor 5C")
	}

	loan, err := s.deps.repository.LoanDashboardRepository.GetActiveLoan(tx, req.CooperativeID, profile.MemberID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil pinjaman aktif")
	}

	var activeLoan *model.LoanDashboardActiveLoan
	installments := make([]model.LoanDashboardInstallment, 0)
	installmentMeta := model.LoanDashboardInstallmentMeta{}
	if loan != nil {
		aggregate, err := s.deps.repository.LoanDashboardRepository.GetLoanDashboardAggregate(tx, loan.LoanID)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman")
		}
		activeLoan = mapLoanDashboardActiveLoan(loan, aggregate)

		schedules, totalSchedules, err := s.deps.repository.LoanDashboardRepository.ListInstallmentSchedules(tx, loan.LoanID, 3)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mengambil jadwal angsuran")
		}
		installments = mapLoanDashboardInstallments(schedules)

		displayed := len(installments)
		remaining := int(totalSchedules) - displayed
		if remaining < 0 {
			remaining = 0
		}

		installmentMeta = model.LoanDashboardInstallmentMeta{
			TotalCount:     int(totalSchedules),
			DisplayedCount: displayed,
			RemainingCount: remaining,
			HasMore:        remaining > 0,
		}
	}

	historyRows, err := s.deps.repository.LoanDashboardRepository.ListLoanHistory(tx, req.CooperativeID, profile.MemberID, 3)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil riwayat pinjaman")
	}
	loanHistory := mapLoanDashboardHistory(historyRows)

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil dashboard pinjaman")
	}

	return &model.LoanDashboardResponse{
		MCS:             mapLoanDashboardMCS(member, snapshot, fiveCScores),
		ActiveLoan:      activeLoan,
		Installments:    installments,
		InstallmentMeta: installmentMeta,
		LoanHistory:     loanHistory,
		Actions: model.LoanDashboardActions{
			HistoryEnabled:         true,
			LoanApplicationEnabled: activeLoan == nil,
			CreditAccessEnabled:    true,
		},
	}, nil
}

func mapLoanDashboardMCS(member *entity.Member, snapshot *entity.MCSScoreSnapshot, fiveC *repository.MCSFiveCScoreRow) model.LoanDashboardMCS {
	score := member.CurrentMCSScore
	grade := member.MCSGrade
	label := getMCSGradeLabel(grade)
	explanation := ""

	if snapshot != nil {
		explanation = snapshot.Explanation
	}
	if fiveC == nil {
		fiveC = &repository.MCSFiveCScoreRow{}
	}

	return model.LoanDashboardMCS{
		Score:              score,
		MaxScore:           850,
		Grade:              grade,
		Label:              label,
		ProfileText:        "Profil Kredit Anda: " + label,
		LastScoreUpdatedAt: member.LastScoreUpdatedAt,
		Explanation:        explanation,
		Components: []model.LoanDashboardMCSComponent{
			{Code: "CHARACTER", Label: "Karakter", Score: fiveC.CharacterScore, Weight: 0.35},
			{Code: "CAPACITY", Label: "Kapasitas", Score: fiveC.CapacityScore, Weight: 0.30},
			{Code: "CAPITAL", Label: "Modal", Score: fiveC.CapitalScore, Weight: 0.15},
			{Code: "CONDITIONS", Label: "Kondisi", Score: fiveC.ConditionsScore, Weight: 0.12},
			{Code: "COLLATERAL", Label: "Jaminan", Score: fiveC.CollateralScore, Weight: 0.08},
		},
	}
}

func mapLoanDashboardInstallments(schedules []entity.LoanInstallmentSchedule) []model.LoanDashboardInstallment {
	items := make([]model.LoanDashboardInstallment, 0, len(schedules))
	for _, schedule := range schedules {
		items = append(items, model.LoanDashboardInstallment{
			InstallmentNo: schedule.InstallmentNo,
			DueDate:       schedule.DueDate,
			DueAmount:     schedule.DueAmount,
			PaidAmount:    schedule.PaidAmount,
			Status:        schedule.Status,
			StatusLabel:   getInstallmentStatusLabel(schedule.Status),
		})
	}
	return items
}

func mapLoanDashboardHistory(rows []repository.LoanDashboardHistoryRow) []model.LoanDashboardHistoryItem {
	items := make([]model.LoanDashboardHistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LoanDashboardHistoryItem{
			LoanID:          row.LoanID,
			LoanNumber:      row.LoanNumber,
			PrincipalAmount: row.PrincipalAmount,
			TermMonths:      row.TermMonths,
			Status:          row.Status,
			StatusLabel:     getLoanStatusLabel(row.Status),
			DisbursedAt:     row.DisbursedAt,
			PaidAt:          row.PaidAt,
			Description:     buildLoanHistoryDescription(row),
		})
	}
	return items
}

func getInstallmentStatusLabel(status string) string {
	switch status {
	case "PAID":
		return "Lunas"
	case "PARTIAL":
		return "Sebagian"
	default:
		return "Belum"
	}
}

func getLoanStatusLabel(status string) string {
	switch status {
	case "PAID":
		return "Lunas"
	case "ACTIVE":
		return "Dalam proses"
	case "CANCELLED":
		return "Dibatalkan"
	default:
		return status
	}
}

func buildLoanHistoryDescription(row repository.LoanDashboardHistoryRow) string {
	if row.Status == "PAID" {
		if row.PaidAt != nil {
			return fmt.Sprintf("%d bulan · Lunas %s", row.TermMonths, formatMonthYearIndonesian(*row.PaidAt))
		}
		return fmt.Sprintf("%d bulan · Lunas", row.TermMonths)
	}
	if row.Status == "ACTIVE" {
		return fmt.Sprintf("%d bulan · Cicilan berjalan", row.TermMonths)
	}
	return fmt.Sprintf("%d bulan", row.TermMonths)
}

func formatMonthYearIndonesian(value time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}

	return fmt.Sprintf("%s %d", months[int(value.Month())-1], value.Year())
}

func mapLoanDashboardActiveLoan(loan *entity.LoanAccount, aggregate *repository.LoanDashboardAggregate) *model.LoanDashboardActiveLoan {
	paidPercentage := 0.0
	if loan.TotalPayableAmount > 0 {
		paidPercentage = (float64(aggregate.PaidAmount) / float64(loan.TotalPayableAmount)) * 100
	}
	paidPercentage = math.Round(paidPercentage*100) / 100

	statusText := fmt.Sprintf("%.0f%% terbayar", paidPercentage)
	if aggregate.PaidInstallmentCount == 0 {
		statusText += " · Belum ada angsuran"
	} else {
		statusText += fmt.Sprintf(" · %d angsuran lunas", aggregate.PaidInstallmentCount)
	}

	return &model.LoanDashboardActiveLoan{
		LoanID:                   loan.LoanID,
		LoanNumber:               loan.LoanNumber,
		PrincipalAmount:          loan.PrincipalAmount,
		TotalPayableAmount:       loan.TotalPayableAmount,
		RemainingPayableAmount:   aggregate.RemainingPayableAmount,
		MonthlyInstallmentAmount: loan.MonthlyInstallmentAmount,
		TermMonths:               loan.TermMonths,
		PaidInstallmentCount:     aggregate.PaidInstallmentCount,
		PaidPercentage:           paidPercentage,
		NextDueDate:              aggregate.NextDueDate,
		PaymentStatusText:        statusText,
	}
}

package service

import (
	"errors"

	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMemberDashboardService interface {
	GetDashboard(req model.GetMemberDashboardRequest) (*model.MemberDashboardResponse, error)
}

type MemberDashboardService struct {
	deps serviceDependency
}

func NewMemberDashboardService(deps serviceDependency) IMemberDashboardService {
	return &MemberDashboardService{deps: deps}
}

func (s *MemberDashboardService) GetDashboard(req model.GetMemberDashboardRequest) (*model.MemberDashboardResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}
	if req.RoleCode != constants.RoleCodeAnggota {
		return nil, appErrors.Forbidden("hanya anggota yang dapat melihat dashboard anggota")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	profile, err := s.deps.repository.MemberRepository.GetDashboardProfile(tx, req.UserID, req.CooperativeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil profil anggota")
	}

	savings, err := s.deps.repository.MemberRepository.GetDashboardSavings(tx, req.CooperativeID, profile.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil saldo anggota")
	}

	member, err := s.deps.repository.MemberRepository.GetActiveMember(tx, req.CooperativeID, profile.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil skor MCS anggota")
	}

	transactions, err := s.deps.repository.MemberRepository.ListRecentMemberTransactions(tx, req.CooperativeID, profile.MemberID, req.RecentLimit)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil transaksi terbaru anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil dashboard anggota")
	}

	for i := range transactions {
		enrichMemberDashboardTransaction(&transactions[i])
	}

	return &model.MemberDashboardResponse{
		Profile: *profile,
		Savings: *savings,
		MCS: model.MemberDashboardMCS{
			Score:              member.CurrentMCSScore,
			Grade:              member.MCSGrade,
			Label:              getMCSGradeLabel(member.MCSGrade),
			Status:             getMCSStatus(member.CurrentMCSScore),
			LastScoreUpdatedAt: member.LastScoreUpdatedAt,
		},
		RecentTransactions: transactions,
	}, nil
}

func enrichMemberDashboardTransaction(item *model.MemberDashboardTransaction) {
	item.TransactionTypeLabel = getTransactionTypeLabel(item.TransactionType)

	switch item.TransactionType {
	case constants.TransactionTypeCashWithdrawal, constants.TransactionTypeInstallment:
		item.Direction = "OUT"
		item.SignedAmount = -item.Amount
	default:
		item.Direction = "IN"
		item.SignedAmount = item.Amount
	}
}

func getMCSStatus(score *int) string {
	if score == nil {
		return "INSUFFICIENT_DATA"
	}
	return "COMPLETE"
}

func getMCSGradeLabel(grade string) string {
	switch grade {
	case "AA":
		return "Sangat Baik"
	case "A":
		return "Baik"
	case "B":
		return "Cukup"
	case "C":
		return "Perlu Perhatian"
	case "D":
		return "Buruk"
	default:
		return "Data Belum Lengkap"
	}
}

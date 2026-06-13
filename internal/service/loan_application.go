package service

import (
	"errors"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultLoanApplicationPartnerName = "Akseleran"

type ILoanApplicationService interface {
	GetLoanApplicationEligibility(req model.GetLoanApplicationEligibilityRequest) (*model.LoanApplicationEligibilityResponse, error)
	CreateLoanApplication(req model.CreateLoanApplicationRequest) (*model.LoanApplicationResponse, error)
	GetLoanApplication(req model.GetLoanApplicationRequest) (*model.LoanApplicationResponse, error)
}

type LoanApplicationService struct {
	deps serviceDependency
}

func NewLoanApplicationService(deps serviceDependency) ILoanApplicationService {
	return &LoanApplicationService{deps: deps}
}

func (s *LoanApplicationService) GetLoanApplicationEligibility(req model.GetLoanApplicationEligibilityRequest) (*model.LoanApplicationEligibilityResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}
	if req.RoleCode != constants.RoleCodeAnggota {
		return nil, appErrors.Forbidden("hanya anggota yang dapat melihat kelayakan pengajuan pinjaman")
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
		return nil, appErrors.InternalServer("gagal mengambil data anggota")
	}

	result := &model.LoanApplicationEligibilityResponse{
		MCSScore: member.CurrentMCSScore,
		MCSGrade: member.MCSGrade,
	}

	financialConfig, err := s.deps.repository.FinancialConfigurationRepository.GetFinancialConfiguration(tx, model.GetFinancialConfigurationParam{
		CooperativeID: req.CooperativeID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result.Reason = "konfigurasi pinjaman belum tersedia"
			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil kelayakan pengajuan")
			}
			return result, nil
		}
		return nil, appErrors.InternalServer("gagal mengambil konfigurasi pinjaman")
	}

	result.CreditLimitAmount = financialConfig.MaxLoanAmountPerMember
	result.MaxTermMonths = financialConfig.MaxLoanTermMonths
	result.InterestRateBpsPerMonth = financialConfig.LoanInterestRateBpsPerMonth

	_, err = s.deps.repository.LoanRepository.GetActiveLoanByMemberForUpdate(tx, req.CooperativeID, profile.MemberID)
	if err == nil {
		result.HasActiveLoan = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi pinjaman aktif")
	}

	_, err = s.deps.repository.LoanApplicationRepository.GetActiveLoanApplication(tx, req.CooperativeID, profile.MemberID)
	if err == nil {
		result.HasActiveApplication = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi pengajuan aktif")
	}

	snapshot, err := s.deps.repository.MCSRepository.GetLatestScoreSnapshot(tx, req.CooperativeID, profile.MemberID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil snapshot MCS")
	}

	result.Eligible, result.Reason = buildLoanApplicationEligibilityReason(member, snapshot, financialConfig, result.HasActiveLoan, result.HasActiveApplication)

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil kelayakan pengajuan")
	}

	return result, nil
}

func (s *LoanApplicationService) CreateLoanApplication(req model.CreateLoanApplicationRequest) (*model.LoanApplicationResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}
	if req.RoleCode != constants.RoleCodeAnggota {
		return nil, appErrors.Forbidden("hanya anggota yang dapat mengajukan pinjaman")
	}
	if req.Amount <= 0 {
		return nil, appErrors.BadRequest("nominal pinjaman wajib lebih dari 0")
	}
	if strings.TrimSpace(req.Purpose) == "" {
		return nil, appErrors.BadRequest("tujuan pinjaman wajib diisi")
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
		return nil, appErrors.InternalServer("gagal mengambil data anggota")
	}

	financialConfig, err := s.deps.repository.FinancialConfigurationRepository.GetFinancialConfiguration(tx, model.GetFinancialConfigurationParam{
		CooperativeID: req.CooperativeID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("konfigurasi pinjaman belum tersedia")
		}
		return nil, appErrors.InternalServer("gagal mengambil konfigurasi pinjaman")
	}
	if financialConfig.MaxLoanAmountPerMember <= 0 || financialConfig.MaxLoanTermMonths <= 0 {
		return nil, appErrors.BadRequest("konfigurasi pinjaman belum tersedia")
	}
	if req.TermMonths <= 0 || req.TermMonths > financialConfig.MaxLoanTermMonths {
		return nil, appErrors.BadRequest("tenor pinjaman tidak valid")
	}
	if req.Amount > financialConfig.MaxLoanAmountPerMember {
		return nil, appErrors.BadRequest("nominal pinjaman melebihi batas kredit")
	}

	if _, err := s.deps.repository.LoanRepository.GetActiveLoanByMemberForUpdate(tx, req.CooperativeID, profile.MemberID); err == nil {
		return nil, appErrors.BadRequest("anggota masih memiliki pinjaman aktif")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi pinjaman aktif")
	}

	if _, err := s.deps.repository.LoanApplicationRepository.GetActiveLoanApplication(tx, req.CooperativeID, profile.MemberID); err == nil {
		return nil, appErrors.Conflict("anggota masih memiliki pengajuan aktif")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi pengajuan aktif")
	}

	snapshot, err := s.deps.repository.MCSRepository.GetLatestScoreSnapshot(tx, req.CooperativeID, profile.MemberID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil snapshot MCS")
	}
	if !isLoanApplicationEligible(member, snapshot) {
		return nil, appErrors.BadRequest("skor kredit belum memenuhi syarat pengajuan")
	}

	now := time.Now()
	totalInterest := req.Amount * int64(financialConfig.LoanInterestRateBpsPerMonth) / 10000 * int64(req.TermMonths)
	totalPayable := req.Amount + totalInterest

	application := &entity.LoanApplication{
		ApplicationID:           uuid.New(),
		CooperativeID:           req.CooperativeID,
		MemberID:                profile.MemberID,
		RequestedAmount:         req.Amount,
		Purpose:                 strings.TrimSpace(req.Purpose),
		TermMonths:              req.TermMonths,
		InterestRateBpsPerMonth: financialConfig.LoanInterestRateBpsPerMonth,
		MonthlyInstallment:      ceilDiv(totalPayable, int64(req.TermMonths)),
		TotalInterestAmount:     totalInterest,
		TotalPayableAmount:      totalPayable,
		MCSScore:                member.CurrentMCSScore,
		MCSGrade:                member.MCSGrade,
		CreditLimitAmount:       financialConfig.MaxLoanAmountPerMember,
		Status:                  "CREDIT_VERIFIED",
		PartnerName:             defaultLoanApplicationPartnerName,
		SubmittedAt:             now,
		CreditVerifiedAt:        &now,
	}

	err = s.deps.repository.LoanApplicationRepository.CreateLoanApplication(tx, application)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan pengajuan pinjaman")
	}

	err = s.createDefaultCreditAccessRequest(tx, application)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat permintaan akses kredit")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan pengajuan pinjaman")
	}

	return mapLoanApplicationResponse(application), nil
}

func (s *LoanApplicationService) GetLoanApplication(req model.GetLoanApplicationRequest) (*model.LoanApplicationResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}
	if req.RoleCode != constants.RoleCodeAnggota {
		return nil, appErrors.Forbidden("hanya anggota yang dapat melihat pengajuan pinjaman")
	}
	if req.ApplicationID == uuid.Nil {
		return nil, appErrors.BadRequest("pengajuan pinjaman wajib dipilih")
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

	application, err := s.deps.repository.LoanApplicationRepository.GetLoanApplicationByID(tx, req.CooperativeID, profile.MemberID, req.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("pengajuan pinjaman tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil pengajuan pinjaman")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil pengajuan pinjaman")
	}

	return mapLoanApplicationResponse(application), nil
}

func (s *LoanApplicationService) createDefaultCreditAccessRequest(tx *gorm.DB, application *entity.LoanApplication) error {
	partner, err := s.deps.repository.CreditAccessRepository.GetActivePartnerByName(tx, defaultLoanApplicationPartnerName)
	if err != nil {
		return err
	}

	request := &entity.CreditAccessRequest{
		RequestID:       uuid.New(),
		CooperativeID:   application.CooperativeID,
		MemberID:        application.MemberID,
		PartnerID:       partner.PartnerID,
		ApplicationID:   &application.ApplicationID,
		RequestedAmount: application.RequestedAmount,
		Purpose:         "Untuk menilai kelayakan pinjaman",
		DataScopeJSON:   buildDefaultCreditAccessScope(application.MCSScore, application.MCSGrade),
		Status:          "PENDING",
		RequestedAt:     time.Now(),
	}

	return s.deps.repository.CreditAccessRepository.CreateCreditAccessRequest(tx, request)
}

func buildLoanApplicationEligibilityReason(member *entity.Member, snapshot *entity.MCSScoreSnapshot, financialConfig *entity.FinancialConfiguration, hasActiveLoan bool, hasActiveApplication bool) (bool, string) {
	if financialConfig == nil || financialConfig.MaxLoanAmountPerMember <= 0 || financialConfig.MaxLoanTermMonths <= 0 {
		return false, "konfigurasi pinjaman belum tersedia"
	}
	if hasActiveLoan {
		return false, "anggota masih memiliki pinjaman aktif"
	}
	if hasActiveApplication {
		return false, "anggota masih memiliki pengajuan aktif"
	}
	if !isLoanApplicationEligible(member, snapshot) {
		return false, "skor kredit belum memenuhi syarat pengajuan"
	}

	return true, ""
}

func isLoanApplicationEligible(member *entity.Member, snapshot *entity.MCSScoreSnapshot) bool {
	if snapshot != nil {
		return snapshot.Eligible
	}
	if member == nil {
		return false
	}

	switch strings.ToUpper(strings.TrimSpace(member.MCSGrade)) {
	case "AA", "A", "B":
		return true
	default:
		return false
	}
}

func mapLoanApplicationResponse(application *entity.LoanApplication) *model.LoanApplicationResponse {
	return &model.LoanApplicationResponse{
		ApplicationID:           application.ApplicationID,
		Status:                  application.Status,
		StatusLabel:             getLoanApplicationStatusLabel(application.Status),
		RequestedAmount:         application.RequestedAmount,
		Purpose:                 application.Purpose,
		TermMonths:              application.TermMonths,
		MonthlyInstallment:      application.MonthlyInstallment,
		TotalInterestAmount:     application.TotalInterestAmount,
		TotalPayableAmount:      application.TotalPayableAmount,
		InterestRateBpsPerMonth: application.InterestRateBpsPerMonth,
		MCSScore:                application.MCSScore,
		MCSGrade:                application.MCSGrade,
		CreditLimitAmount:       application.CreditLimitAmount,
		PartnerName:             application.PartnerName,
		SubmittedAt:             application.SubmittedAt,
		Timeline:                buildLoanApplicationTimeline(application),
	}
}

func buildLoanApplicationTimeline(application *entity.LoanApplication) []model.LoanApplicationTimelineItem {
	submittedAt := application.SubmittedAt

	return []model.LoanApplicationTimelineItem{
		{
			Code:        "RECEIVED",
			Label:       "Pengajuan diterima",
			Description: "Pengajuan pinjaman Anda sudah diterima",
			State:       loanApplicationTimelineState(application.Status, "RECEIVED"),
			OccurredAt:  &submittedAt,
		},
		{
			Code:        "CREDIT_VERIFIED",
			Label:       "Skor kredit terverifikasi",
			Description: buildCreditVerifiedDescription(application),
			State:       loanApplicationTimelineState(application.Status, "CREDIT_VERIFIED"),
			OccurredAt:  application.CreditVerifiedAt,
		},
		{
			Code:        "UNDER_REVIEW",
			Label:       "Peninjauan " + application.PartnerName,
			Description: "Estimasi 1-2 jam",
			State:       loanApplicationTimelineState(application.Status, "UNDER_REVIEW"),
			OccurredAt:  application.ReviewedAt,
		},
		{
			Code:        "DISBURSED",
			Label:       "Dana cair ke Virtual Account",
			Description: "Menunggu persetujuan",
			State:       loanApplicationTimelineState(application.Status, "DISBURSED"),
			OccurredAt:  application.DisbursedAt,
		},
	}
}

func loanApplicationTimelineState(currentStatus string, step string) string {
	order := map[string]int{
		"RECEIVED":        1,
		"CREDIT_VERIFIED": 2,
		"UNDER_REVIEW":    3,
		"APPROVED":        3,
		"DISBURSED":       4,
	}

	currentOrder := order[currentStatus]
	stepOrder := order[step]
	if currentOrder >= stepOrder && currentOrder > 0 {
		return "done"
	}
	return "pending"
}

func buildCreditVerifiedDescription(application *entity.LoanApplication) string {
	if strings.TrimSpace(application.MCSGrade) == "" {
		return "Skor kredit anggota telah diverifikasi"
	}
	return "MCS " + application.MCSGrade
}

func getLoanApplicationStatusLabel(status string) string {
	switch status {
	case "RECEIVED":
		return "Pengajuan diterima"
	case "CREDIT_VERIFIED":
		return "Skor kredit terverifikasi"
	case "UNDER_REVIEW":
		return "Dalam peninjauan"
	case "APPROVED":
		return "Disetujui"
	case "REJECTED":
		return "Ditolak"
	case "DISBURSED":
		return "Dana cair"
	case "CANCELLED":
		return "Dibatalkan"
	default:
		return status
	}
}

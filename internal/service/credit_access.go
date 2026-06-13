package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var allowedCreditAccessDurations = []int{7, 14, 30}

type ICreditAccessService interface {
	ListCreditAccessRequests(req model.ListCreditAccessRequestsRequest) (*model.CreditAccessRequestsResponse, error)
	GetCreditAccessRequest(req model.GetCreditAccessRequestRequest) (*model.CreditAccessDetailResponse, error)
	GrantCreditAccess(req model.GrantCreditAccessRequest) (*model.CreditAccessDetailResponse, error)
	DeclineCreditAccess(req model.DeclineCreditAccessRequest) (*model.CreditAccessDetailResponse, error)
	RevokeCreditAccess(req model.RevokeCreditAccessRequest) (*model.CreditAccessDetailResponse, error)
}

type CreditAccessService struct {
	deps serviceDependency
}

func NewCreditAccessService(deps serviceDependency) ICreditAccessService {
	return &CreditAccessService{deps: deps}
}

func (s *CreditAccessService) ListCreditAccessRequests(req model.ListCreditAccessRequestsRequest) (*model.CreditAccessRequestsResponse, error) {
	if err := validateCreditAccessAuth(req.AuthContext); err != nil {
		return nil, err
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

	rows, err := s.deps.repository.CreditAccessRepository.ListCreditAccessRequests(tx, req.CooperativeID, profile.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil permintaan akses kredit")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil permintaan akses kredit")
	}

	return groupCreditAccessRows(rows), nil
}

func (s *CreditAccessService) GetCreditAccessRequest(req model.GetCreditAccessRequestRequest) (*model.CreditAccessDetailResponse, error) {
	if err := validateCreditAccessAuth(req.AuthContext); err != nil {
		return nil, err
	}
	if req.RequestID == uuid.Nil {
		return nil, appErrors.BadRequest("permintaan akses wajib dipilih")
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

	row, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestDetail(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permintaan akses tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil detail akses kredit")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail akses kredit")
	}

	return mapCreditAccessDetail(row), nil
}

func (s *CreditAccessService) GrantCreditAccess(req model.GrantCreditAccessRequest) (*model.CreditAccessDetailResponse, error) {
	if err := validateCreditAccessAuth(req.AuthContext); err != nil {
		return nil, err
	}
	if req.RequestID == uuid.Nil {
		return nil, appErrors.BadRequest("permintaan akses wajib dipilih")
	}
	if !isAllowedCreditAccessDuration(req.DurationDays) {
		return nil, appErrors.BadRequest("durasi akses tidak valid")
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

	accessRequest, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestForUpdate(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permintaan akses tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil permintaan akses")
	}
	if accessRequest.Status != "PENDING" {
		return nil, appErrors.BadRequest("permintaan akses tidak aktif")
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, req.DurationDays)
	accessRequest.Status = "GRANTED"
	accessRequest.GrantedAt = &now
	accessRequest.AccessExpiresAt = &expiresAt

	consent := &entity.MemberDataConsent{
		ConsentID:     uuid.New(),
		RequestID:     accessRequest.RequestID,
		CooperativeID: accessRequest.CooperativeID,
		MemberID:      accessRequest.MemberID,
		PartnerID:     accessRequest.PartnerID,
		DataScopeJSON: accessRequest.DataScopeJSON,
		DurationDays:  req.DurationDays,
		GrantedAt:     now,
		ExpiresAt:     expiresAt,
		IsActive:      true,
	}

	if err := s.deps.repository.CreditAccessRepository.UpdateCreditAccessRequest(tx, accessRequest); err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui permintaan akses")
	}
	if err := s.deps.repository.CreditAccessRepository.CreateMemberDataConsent(tx, consent); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan consent")
	}

	detail, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestDetail(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail akses kredit")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal memberikan akses")
	}

	return mapCreditAccessDetail(detail), nil
}

func (s *CreditAccessService) DeclineCreditAccess(req model.DeclineCreditAccessRequest) (*model.CreditAccessDetailResponse, error) {
	if err := validateCreditAccessAuth(req.AuthContext); err != nil {
		return nil, err
	}
	if req.RequestID == uuid.Nil {
		return nil, appErrors.BadRequest("permintaan akses wajib dipilih")
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

	accessRequest, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestForUpdate(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permintaan akses tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil permintaan akses")
	}
	if accessRequest.Status != "PENDING" {
		return nil, appErrors.BadRequest("permintaan akses tidak aktif")
	}

	now := time.Now()
	accessRequest.Status = "DECLINED"
	accessRequest.DeclinedAt = &now

	if err := s.deps.repository.CreditAccessRepository.UpdateCreditAccessRequest(tx, accessRequest); err != nil {
		return nil, appErrors.InternalServer("gagal menolak permintaan akses")
	}

	detail, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestDetail(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail akses kredit")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal menolak permintaan akses")
	}

	return mapCreditAccessDetail(detail), nil
}

func (s *CreditAccessService) RevokeCreditAccess(req model.RevokeCreditAccessRequest) (*model.CreditAccessDetailResponse, error) {
	if err := validateCreditAccessAuth(req.AuthContext); err != nil {
		return nil, err
	}
	if req.RequestID == uuid.Nil {
		return nil, appErrors.BadRequest("permintaan akses wajib dipilih")
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

	accessRequest, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestForUpdate(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permintaan akses tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil permintaan akses")
	}
	if accessRequest.Status != "GRANTED" {
		return nil, appErrors.BadRequest("akses kredit belum aktif")
	}

	consent, err := s.deps.repository.CreditAccessRepository.GetActiveConsentByRequest(tx, req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("consent aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil consent aktif")
	}

	now := time.Now()
	accessRequest.Status = "REVOKED"
	accessRequest.RevokedAt = &now
	consent.IsActive = false
	consent.RevokedAt = &now

	if err := s.deps.repository.CreditAccessRepository.UpdateCreditAccessRequest(tx, accessRequest); err != nil {
		return nil, appErrors.InternalServer("gagal mencabut akses kredit")
	}
	if err := s.deps.repository.CreditAccessRepository.UpdateMemberDataConsent(tx, consent); err != nil {
		return nil, appErrors.InternalServer("gagal mencabut consent")
	}

	detail, err := s.deps.repository.CreditAccessRepository.GetCreditAccessRequestDetail(tx, req.CooperativeID, profile.MemberID, req.RequestID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail akses kredit")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mencabut akses kredit")
	}

	return mapCreditAccessDetail(detail), nil
}

func validateCreditAccessAuth(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}
	if auth.RoleCode != constants.RoleCodeAnggota {
		return appErrors.Forbidden("hanya anggota yang dapat mengelola akses kredit")
	}
	return nil
}

func groupCreditAccessRows(rows []repository.CreditAccessRequestRow) *model.CreditAccessRequestsResponse {
	response := &model.CreditAccessRequestsResponse{
		Pending: []model.CreditAccessListItem{},
		Active:  []model.CreditAccessListItem{},
		History: []model.CreditAccessListItem{},
	}

	for _, row := range rows {
		item := mapCreditAccessListItem(row)
		switch {
		case row.Status == "PENDING":
			response.Pending = append(response.Pending, item)
		case isActiveCreditAccess(row):
			response.Active = append(response.Active, item)
		default:
			response.History = append(response.History, item)
		}
	}

	return response
}

func mapCreditAccessDetail(row *repository.CreditAccessRequestRow) *model.CreditAccessDetailResponse {
	item := mapCreditAccessListItem(*row)
	return &model.CreditAccessDetailResponse{
		CreditAccessListItem: item,
		DataItems:            parseCreditAccessDataItems(row.DataScopeJSON),
		AllowedDurations:     allowedCreditAccessDurations,
		Actions: model.CreditAccessActions{
			GrantEnabled:   row.Status == "PENDING",
			DeclineEnabled: row.Status == "PENDING",
			RevokeEnabled:  isActiveCreditAccess(*row),
		},
	}
}

func mapCreditAccessListItem(row repository.CreditAccessRequestRow) model.CreditAccessListItem {
	return model.CreditAccessListItem{
		RequestID:       row.RequestID,
		PartnerID:       row.PartnerID,
		PartnerName:     row.PartnerName,
		PartnerType:     row.PartnerType,
		OJKRegistered:   strings.TrimSpace(row.OJKRegistrationNumber) != "",
		MCSGrade:        row.MCSGrade,
		Status:          row.Status,
		StatusLabel:     getCreditAccessStatusLabel(row),
		RequestedAmount: row.RequestedAmount,
		Purpose:         row.Purpose,
		RequestedAt:     row.RequestedAt,
		GrantedAt:       row.GrantedAt,
		ExpiresAt:       row.AccessExpiresAt,
		DurationDays:    row.DurationDays,
	}
}

func getCreditAccessStatusLabel(row repository.CreditAccessRequestRow) string {
	if row.Status == "GRANTED" && row.AccessExpiresAt != nil && row.AccessExpiresAt.Before(time.Now()) {
		return "Tidak Aktif"
	}

	switch row.Status {
	case "PENDING":
		return "Permintaan"
	case "GRANTED":
		return "Aktif"
	case "DECLINED":
		return "Tidak Aktif"
	case "REVOKED":
		return "Tidak Aktif"
	default:
		return row.Status
	}
}

func isActiveCreditAccess(row repository.CreditAccessRequestRow) bool {
	return row.Status == "GRANTED" && (row.AccessExpiresAt == nil || row.AccessExpiresAt.After(time.Now()))
}

func isAllowedCreditAccessDuration(duration int) bool {
	for _, allowed := range allowedCreditAccessDurations {
		if duration == allowed {
			return true
		}
	}
	return false
}

func parseCreditAccessDataItems(raw string) []model.CreditAccessDataItem {
	items := []model.CreditAccessDataItem{}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []model.CreditAccessDataItem{}
	}
	return items
}

func buildDefaultCreditAccessScope(score *int, grade string) string {
	payload := []model.CreditAccessDataItem{
		{Code: "MCS_SCORE", Label: "Skor kredit MCS", Value: formatMCSValue(score, grade), Included: true},
		{Code: "INSTALLMENT_HISTORY", Label: "Riwayat angsuran", Value: "24 bulan", Included: true},
		{Code: "AVERAGE_SAVINGS", Label: "Saldo tabungan rata-rata", Value: "6 bulan", Included: true},
		{Code: "PHONE_NIK", Label: "Nomor HP & NIK", Value: "tidak dibagikan", Included: false},
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "[]"
	}

	return string(bytes)
}

func formatMCSValue(score *int, grade string) string {
	grade = strings.TrimSpace(grade)
	if score == nil {
		if grade == "" {
			return "-"
		}
		return "Grade " + grade
	}
	if grade == "" {
		return fmt.Sprintf("%d", *score)
	}
	return fmt.Sprintf("%d - Grade %s", *score, grade)
}

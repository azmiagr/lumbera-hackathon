package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IProfileService interface {
	GetProfile(req model.GetProfileRequest) (*model.ProfileResponse, error)
}

type ProfileService struct {
	deps serviceDependency
}

func NewProfileService(deps serviceDependency) IProfileService {
	return &ProfileService{deps: deps}
}

func (s *ProfileService) GetProfile(req model.GetProfileRequest) (*model.ProfileResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	switch req.RoleCode {
	case constants.RoleCodePengurusKoperasi:
		profile, err := s.getOfficerProfile(req)
		if err != nil {
			return nil, err
		}
		return &model.ProfileResponse{
			RoleCode: req.RoleCode,
			Profile:  profile,
		}, nil

	case constants.RoleCodeAnggota:
		profile, err := s.getMemberProfile(req)
		if err != nil {
			return nil, err
		}
		return &model.ProfileResponse{
			RoleCode: req.RoleCode,
			Profile:  profile,
		}, nil

	default:
		return nil, appErrors.Forbidden("role tidak didukung untuk profil")
	}
}

func (s *ProfileService) getOfficerProfile(req model.GetProfileRequest) (*model.OfficerProfileResponse, error) {
	tx := s.deps.db.Begin()
	defer tx.Rollback()

	row, err := s.deps.repository.ProfileRepository.GetOfficerProfile(tx, req.UserID, req.CooperativeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("profil pengurus tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil profil pengurus")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil profil pengurus")
	}

	chs := model.ProfileCHSResponse{
		Status: "INSUFFICIENT_DATA",
	}
	period := time.Now().Format("2006-01")

	report, err := NewReportService(s.deps).GetCooperativeHealthScore(model.CooperativeHealthScoreRequest{
		AuthContext: req.AuthContext,
		Period:      period,
	})
	if err == nil {
		chs = model.ProfileCHSResponse{
			Period:       report.Period,
			Status:       report.Status,
			Score:        report.CHSScore,
			DisplayScore: report.DisplayScore,
			Grade:        report.Grade,
			Category:     report.Category,
		}
	}

	return &model.OfficerProfileResponse{
		UserID:        row.UserID,
		FullName:      row.FullName,
		Initials:      buildProfileInitials(row.FullName),
		AvatarURL:     "",
		PhoneNumber:   row.PhoneNumber,
		PositionCode:  row.PositionCode,
		PositionLabel: resolveProfilePositionLabel(row.PositionCode),
		JoinedAt:      row.JoinedAt,
		Cooperative: model.ProfileCooperativeResponse{
			CooperativeID:      row.CooperativeID,
			Name:               row.CooperativeName,
			CooperativeCode:    row.CooperativeCode,
			RegistrationNumber: row.RegistrationNumber,
		},
		CHS: chs,
	}, nil
}

func (s *ProfileService) getMemberProfile(req model.GetProfileRequest) (*model.MemberProfileResponse, error) {
	tx := s.deps.db.Begin()
	defer tx.Rollback()

	row, err := s.deps.repository.ProfileRepository.GetMemberProfile(tx, req.UserID, req.CooperativeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("profil anggota tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil profil anggota")
	}

	completedLoanCount, err := s.deps.repository.ProfileRepository.CountCompletedLoans(tx, req.CooperativeID, row.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman anggota")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil profil anggota")
	}

	joinedYear := 0
	if row.JoinedDate != nil {
		joinedYear = row.JoinedDate.Year()
	}

	return &model.MemberProfileResponse{
		UserID:       row.UserID,
		MemberID:     row.MemberID,
		FullName:     row.FullName,
		Initials:     buildProfileInitials(row.FullName),
		AvatarURL:    "",
		PhoneNumber:  row.PhoneNumber,
		MemberNumber: row.MemberNumber,
		JoinedDate:   row.JoinedDate,
		JoinedYear:   joinedYear,
		Cooperative: model.ProfileCooperativeResponse{
			CooperativeID:   row.CooperativeID,
			Name:            row.CooperativeName,
			CooperativeCode: row.CooperativeCode,
		},
		MCS: model.ProfileMCSResponse{
			Score:              row.CurrentMCSScore,
			Grade:              row.MCSGrade,
			LastScoreUpdatedAt: row.LastScoreUpdatedAt,
		},
		Loan: model.ProfileLoanResponse{
			CompletedCount: completedLoanCount,
			CompletedLabel: fmt.Sprintf("%d selesai", completedLoanCount),
		},
	}, nil
}

func buildProfileInitials(fullName string) string {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return ""
	}

	initials := ""
	for i, part := range parts {
		if i >= 2 {
			break
		}
		initials += strings.ToUpper(string([]rune(part)[0]))
	}

	return initials
}

func resolveProfilePositionLabel(positionCode string) string {
	switch positionCode {
	case constants.PositionCodeChairman:
		return "Ketua"
	case constants.PositionCodeTreasurer:
		return "Bendahara"
	case constants.PositionCodeSecretary:
		return "Sekretaris"
	case constants.PositionCodeStaff:
		return "Staf"
	default:
		return positionCode
	}
}

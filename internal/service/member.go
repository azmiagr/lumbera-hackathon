package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/azmiagr/lumbera-hackathon/pkg/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMemberService interface {
	ListMembers(req model.ListMembersRequest) (*model.ListMembersResponse, error)
	CreateMember(req model.CreateMemberRequest) (*model.CreateMemberResponse, error)
}

type MemberService struct {
	deps serviceDependency
}

func NewMemberService(deps serviceDependency) IMemberService {
	return &MemberService{deps: deps}
}

func (s *MemberService) ListMembers(req model.ListMembersRequest) (*model.ListMembersResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat melihat daftar anggota")
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	grade := strings.ToUpper(strings.TrimSpace(req.Grade))
	if grade != "" && grade != "SEMUA" && !isValidMCSGrade(grade) {
		return nil, appErrors.BadRequest("grade anggota tidak valid")
	}
	req.Grade = grade

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "" && status != "ACTIVE" && status != "INACTIVE" && status != "SUSPENDED" && status != "RESIGNED" {
		return nil, appErrors.BadRequest("status anggota tidak valid")
	}
	req.Status = status

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	items, total, err := s.deps.repository.MemberRepository.ListMembers(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil daftar anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil daftar anggota")
	}

	for i := range items {
		items[i].Initials = buildMemberInitials(items[i].FullName)
		items[i].MembershipYears = calculateMembershipYears(items[i].JoinedDate)
	}

	return &model.ListMembersResponse{
		Items: items,
		Page:  req.Page,
		Limit: req.Limit,
		Total: total,
	}, nil
}

func (s *MemberService) CreateMember(req model.CreateMemberRequest) (*model.CreateMemberResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mendaftarkan anggota")
	}

	fullName := strings.TrimSpace(req.FullName)
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	nik := identity.NormalizeNIK(req.NIK)
	address := strings.TrimSpace(req.Address)

	if fullName == "" || phoneNumber == "" || nik == "" || address == "" || req.JoinedDate == nil {
		return nil, appErrors.BadRequest("data anggota belum lengkap")
	}

	if !isSixteenDigitNIK(nik) {
		return nil, appErrors.BadRequest("NIK harus 16 digit")
	}

	nikHash, err := identity.HashNIK(nik)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash NIK")
	}

	nikEncrypted, err := identity.EncryptNIK(nik)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengenkripsi NIK")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	_, err = s.deps.repository.UserRepository.GetUser(tx, model.GetUserParam{
		PhoneNumber: phoneNumber,
	})
	if err == nil {
		return nil, appErrors.Conflict("nomor handphone sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi nomor handphone")
	}

	_, err = s.deps.repository.UserIdentityRepository.GetUserIdentity(tx, model.GetUserIdentityParam{
		NIKHash: nikHash,
	})
	if err == nil {
		return nil, appErrors.Conflict("NIK sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi NIK")
	}

	role, err := s.deps.repository.RoleRepository.GetRole(tx, model.GetRoleParam{
		Code:      constants.RoleCodeAnggota,
		ScopeType: constants.RoleScopeCooperative,
	})
	if err != nil {
		return nil, appErrors.InternalServer("role anggota belum tersedia")
	}

	memberNumber, err := s.generateMemberNumber(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat nomor anggota")
	}

	userID := uuid.New()
	user := &entity.User{
		UserID:      userID,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
		Status:      "PIN_REQUIRED",
		UserType:    "COOPERATIVE",
	}

	err = s.deps.repository.UserRepository.CreateUser(tx, user)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan user anggota")
	}

	userIdentity := &entity.UserIdentity{
		IdentityID:   uuid.New(),
		UserID:       userID,
		NIKEncrypted: nikEncrypted,
		NIKHash:      nikHash,
		Address:      address,
	}

	err = s.deps.repository.UserIdentityRepository.CreateUserIdentity(tx, userIdentity)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan identitas anggota")
	}

	memberID := uuid.New()
	member := &entity.Member{
		MemberID:      memberID,
		CooperativeID: req.CooperativeID,
		UserID:        userID,
		MemberNumber:  memberNumber,
		JoinedDate:    req.JoinedDate,
		MemberStatus:  "ACTIVE",
		MCSGrade:      "C",
	}

	err = s.deps.repository.MemberRepository.CreateMember(tx, member)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan anggota")
	}

	membership := &entity.UserCooperativeMembership{
		CooperativeMembershipID: uuid.New(),
		UserID:                  userID,
		CooperativeID:           req.CooperativeID,
		MemberID:                &memberID,
		RoleID:                  role.RoleID,
		PositionCode:            constants.PositionCodeStaff,
		Status:                  "ACTIVE",
		JoinedAt:                time.Now(),
	}

	err = s.deps.repository.UserCooperativeMembershipRepository.CreateUserCooperativeMembership(tx, membership)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan membership anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mendaftarkan anggota")
	}

	return &model.CreateMemberResponse{
		UserID:        userID,
		MemberID:      memberID,
		CooperativeID: req.CooperativeID,
		FullName:      fullName,
		PhoneNumber:   phoneNumber,
		MemberNumber:  memberNumber,
		JoinedDate:    req.JoinedDate,
		MemberStatus:  "ACTIVE",
		AccountStatus: "PIN_REQUIRED",
	}, nil
}

func (s *MemberService) generateMemberNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error) {
	total, err := s.deps.repository.MemberRepository.CountMembersByCooperative(tx, cooperativeID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%04d", total+1), nil
}

func isSixteenDigitNIK(nik string) bool {
	if len(nik) != 16 {
		return false
	}

	for _, char := range nik {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func isValidMCSGrade(grade string) bool {
	switch grade {
	case "AA", "A", "B", "C", "D":
		return true
	default:
		return false
	}
}

func buildMemberInitials(fullName string) string {
	words := strings.Fields(fullName)
	if len(words) == 0 {
		return ""
	}

	if len(words) == 1 {
		return strings.ToUpper(firstLetter(words[0]))
	}

	return strings.ToUpper(firstLetter(words[0]) + firstLetter(words[len(words)-1]))
}

func firstLetter(value string) string {
	for _, char := range value {
		return string(char)
	}

	return ""
}

func calculateMembershipYears(joinedDate *time.Time) int {
	if joinedDate == nil {
		return 0
	}

	now := time.Now()
	years := now.Year() - joinedDate.Year()
	if now.YearDay() < joinedDate.YearDay() {
		years--
	}

	if years < 0 {
		return 0
	}

	return years
}

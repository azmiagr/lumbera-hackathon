package service

import (
	"errors"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	memberActivationTokenTTL = 10 * time.Minute
	memberRefreshTokenTTL    = 30 * 24 * time.Hour
)

type IMemberActivationService interface {
	CheckPhone(req model.CheckMemberPhoneRequest) (*model.CheckMemberPhoneResponse, error)
	SetPIN(req model.SetMemberPINRequest) (*model.SetMemberPINResponse, error)
}

type MemberActivationService struct {
	deps serviceDependency
}

func NewMemberActivationService(deps serviceDependency) IMemberActivationService {
	return &MemberActivationService{deps: deps}
}

func (s *MemberActivationService) CheckPhone(req model.CheckMemberPhoneRequest) (*model.CheckMemberPhoneResponse, error) {
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, appErrors.BadRequest("nomor handphone wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	activationCtx, err := s.deps.repository.MemberActivationRepository.GetEligibleMemberActivationContext(tx, model.GetEligibleMemberActivationContextParam{
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("nomor handphone belum terdaftar sebagai anggota")
		}
		return nil, err
	}

	status := "INACTIVE"
	if activationCtx.User.Status == "ACTIVE" {
		status = "ACTIVE"
	}

	activationToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat token aktivasi")
	}

	activationTokenHash, err := s.deps.bcrypt.GenerateFromPassword(activationToken)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash token aktivasi")
	}

	challenge := &entity.MemberActivationChallenge{
		ChallengeID: uuid.New(),
		UserID:      activationCtx.User.UserID,
		TokenHash:   activationTokenHash,
		ExpiresAt:   time.Now().Add(memberActivationTokenTTL),
	}

	err = s.deps.repository.MemberActivationRepository.CreateMemberActivationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan challenge aktivasi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan challenge aktivasi")
	}

	return &model.CheckMemberPhoneResponse{
		ActivationChallengeID: challenge.ChallengeID,
		ActivationToken:       activationToken,
		PhoneNumber:           phoneNumber,
		ExpiresInSeconds:      int(memberActivationTokenTTL.Seconds()),
		Status:                status,
	}, nil
}

func (s *MemberActivationService) SetPIN(req model.SetMemberPINRequest) (*model.SetMemberPINResponse, error) {
	if req.ActivationChallengeID == uuid.Nil {
		return nil, appErrors.BadRequest("activation_challenge_id wajib diisi")
	}

	if strings.TrimSpace(req.ActivationToken) == "" {
		return nil, appErrors.BadRequest("activation_token wajib diisi")
	}

	if !isSixDigitPIN(req.PIN) || !isSixDigitPIN(req.ConfirmPIN) {
		return nil, appErrors.BadRequest("PIN harus 6 digit")
	}

	if req.PIN != req.ConfirmPIN {
		return nil, appErrors.BadRequest("konfirmasi PIN tidak sama")
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		deviceID = uuid.NewString()
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	challenge, err := s.deps.repository.MemberActivationRepository.GetMemberActivationChallenge(tx, model.GetMemberActivationChallengeParam{
		ActivationChallengeID: req.ActivationChallengeID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("challenge aktivasi tidak ditemukan")
		}
		return nil, err
	}

	if challenge.UsedAt != nil {
		return nil, appErrors.BadRequest("challenge aktivasi sudah digunakan")
	}

	if time.Now().After(challenge.ExpiresAt) {
		return nil, appErrors.BadRequest("challenge aktivasi sudah kedaluwarsa")
	}

	err = s.deps.bcrypt.CompareAndHashPassword(challenge.TokenHash, req.ActivationToken)
	if err != nil {
		return nil, appErrors.Unauthorized("activation token tidak valid")
	}

	activationCtx, err := s.deps.repository.MemberActivationRepository.GetEligibleMemberActivationContext(tx, model.GetEligibleMemberActivationContextParam{
		UserID: challenge.UserID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("akun anggota belum tersedia atau sudah aktif")
		}
		return nil, err
	}

	pinHash, err := s.deps.bcrypt.GenerateFromPassword(req.PIN)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash pin")
	}

	userPin := &entity.UserPINCredential{
		PinCredentialID: uuid.New(),
		UserID:          activationCtx.User.UserID,
		PINHash:         pinHash,
		FailedAttempts:  0,
		LastChangedAt:   time.Now(),
	}

	err = s.deps.repository.UserPinRepository.CreateUserPin(tx, userPin)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan pin anggota")
	}

	refreshToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat refresh token")
	}

	refreshTokenHash, err := s.deps.bcrypt.GenerateFromPassword(refreshToken)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash refresh token")
	}

	session := &entity.UserSession{
		SessionID:        uuid.New(),
		UserID:           activationCtx.User.UserID,
		DeviceID:         deviceID,
		RefreshTokenHash: refreshTokenHash,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		ExpiresAt:        time.Now().Add(memberRefreshTokenTTL),
	}

	accessToken, err := s.deps.jwtAuth.GenerateAccessToken(jwt.GenerateAccessTokenInput{
		UserID:        activationCtx.User.UserID,
		CooperativeID: activationCtx.Member.CooperativeID,
		SessionID:     session.SessionID,
		RoleCode:      constants.RoleCodeAnggota,
	})
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat access token")
	}

	err = s.deps.repository.UserSessionRepository.CreateUserSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan session anggota")
	}

	now := time.Now()
	activationCtx.User.Status = "ACTIVE"
	activationCtx.User.LastLoginAt = &now

	err = s.deps.repository.UserRepository.UpdateUser(tx, activationCtx.User)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengaktifkan akun anggota")
	}

	challenge.UsedAt = &now
	err = s.deps.repository.MemberActivationRepository.UpdateMemberActivationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyelesaikan challenge aktivasi")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengaktifkan akun anggota")
	}

	return &model.SetMemberPINResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		UserID:        activationCtx.User.UserID,
		MemberID:      activationCtx.Member.MemberID,
		CooperativeID: activationCtx.Member.CooperativeID,
		RoleID:        activationCtx.Role.RoleID,
	}, nil
}

func isSixDigitPIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}

	for _, char := range pin {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

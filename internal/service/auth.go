package service

import (
	"errors"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const loginRefreshTokenTTL = 30 * 24 * time.Hour

type IAuthService interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
}

type AuthService struct {
	deps serviceDependency
}

func NewAuthService(deps serviceDependency) IAuthService {
	return &AuthService{deps: deps}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, appErrors.BadRequest("nomor handphone wajib diisi")
	}

	if !isSixDigitPIN(req.PIN) {
		return nil, appErrors.BadRequest("PIN harus 6 digit")
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		deviceID = uuid.NewString()
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	loginCtx, err := s.deps.repository.UserRepository.GetCooperativeLoginContext(tx, model.GetCooperativeLoginContextParam{
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("nomor handphone atau PIN tidak valid")
		}
		return nil, err
	}

	userPin, err := s.deps.repository.UserPinRepository.GetUserPinCredential(tx, model.GetUserPINCredentialParam{
		UserID: loginCtx.User.UserID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("nomor handphone atau PIN tidak valid")
		}
		return nil, err
	}

	if userPin.LockedUntil != nil && time.Now().Before(*userPin.LockedUntil) {
		return nil, appErrors.Forbidden("akun terkunci sementara")
	}

	err = s.deps.bcrypt.CompareAndHashPassword(userPin.PINHash, req.PIN)
	if err != nil {
		userPin.FailedAttempts++
		if userPin.FailedAttempts >= 5 {
			lockedUntil := time.Now().Add(15 * time.Minute)
			userPin.LockedUntil = &lockedUntil
		}

		_ = s.deps.repository.UserPinRepository.UpdateUserPin(tx, userPin)
		_ = tx.Commit().Error

		return nil, appErrors.Unauthorized("nomor handphone atau PIN tidak valid")
	}

	userPin.FailedAttempts = 0
	userPin.LockedUntil = nil
	err = s.deps.repository.UserPinRepository.UpdateUserPin(tx, userPin)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui kredensial pin")
	}

	accessToken, err := s.deps.jwtAuth.GenerateAccessToken(jwt.GenerateAccessTokenInput{
		UserID:        loginCtx.User.UserID,
		CooperativeID: loginCtx.Membership.CooperativeID,
		RoleCode:      loginCtx.Role.Code,
	})
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat access token")
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
		UserID:           loginCtx.User.UserID,
		DeviceID:         deviceID,
		RefreshTokenHash: refreshTokenHash,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		ExpiresAt:        time.Now().Add(loginRefreshTokenTTL),
	}

	err = s.deps.repository.UserSessionRepository.CreateUserSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan session")
	}

	now := time.Now()
	loginCtx.User.LastLoginAt = &now
	err = s.deps.repository.UserRepository.UpdateUser(tx, loginCtx.User)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui data login")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal login")
	}

	var memberID *uuid.UUID
	if loginCtx.Member != nil {
		memberID = &loginCtx.Member.MemberID
	}

	return &model.LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		UserID:        loginCtx.User.UserID,
		CooperativeID: loginCtx.Membership.CooperativeID,
		RoleID:        loginCtx.Role.RoleID,
		RoleCode:      loginCtx.Role.Code,
		MemberID:      memberID,
	}, nil
}

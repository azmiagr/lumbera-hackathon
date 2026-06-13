package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	loginRefreshTokenTTL   = 30 * 24 * time.Hour
	forgotPINOTPExpiry     = 5 * time.Minute
	forgotPINResetTokenTTL = 10 * time.Minute
	maxForgotPINOTPAttempt = 5
)

type IAuthService interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	RequestForgotPINOTP(req model.ForgotPINRequestOTPRequest) (*model.ForgotPINRequestOTPResponse, error)
	VerifyForgotPINOTP(req model.ForgotPINVerifyOTPRequest) (*model.ForgotPINVerifyOTPResponse, error)
	SetForgottenPIN(req model.ForgotPINSetPINRequest) (*model.LoginResponse, error)
	Logout(req model.LogoutRequest) (*model.LogoutResponse, error)
	ValidateAuthenticatedSession(auth model.AuthContext) error
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

	if loginCtx.User.Status == "PIN_REQUIRED" {
		return nil, appErrors.Forbidden("akun belum memiliki PIN")
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

	accessToken, err := s.deps.jwtAuth.GenerateAccessToken(jwt.GenerateAccessTokenInput{
		UserID:        loginCtx.User.UserID,
		CooperativeID: loginCtx.Membership.CooperativeID,
		SessionID:     session.SessionID,
		RoleCode:      loginCtx.Role.Code,
	})
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat access token")
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

func (s *AuthService) RequestForgotPINOTP(req model.ForgotPINRequestOTPRequest) (*model.ForgotPINRequestOTPResponse, error) {
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, appErrors.BadRequest("nomor handphone wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	loginCtx, err := s.deps.repository.UserRepository.GetCooperativeLoginContext(tx, model.GetCooperativeLoginContextParam{
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("akun tidak ditemukan")
		}
		return nil, err
	}

	_, err = s.deps.repository.UserPinRepository.GetUserPinCredential(tx, model.GetUserPINCredentialParam{
		UserID: loginCtx.User.UserID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("akun belum memiliki PIN")
		}
		return nil, err
	}

	err = s.deps.repository.PhoneVerificationRepository.DeletePhoneVerificationChallenges(tx, phoneNumber, "PIN_RESET")
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghapus OTP lama")
	}

	otp, err := generateNumericOTP(6)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat kode otp")
	}

	otpHash, err := s.deps.bcrypt.GenerateFromPassword(otp)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash otp")
	}

	challenge := &entity.PhoneVerificationChallenge{
		ChallengeID:  uuid.New(),
		PhoneNumber:  phoneNumber,
		OTPHash:      otpHash,
		Purpose:      "PIN_RESET",
		AttemptCount: 0,
		ExpiresAt:    time.Now().Add(forgotPINOTPExpiry),
	}

	err = s.deps.repository.PhoneVerificationRepository.CreatePhoneVerificationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan challenge reset PIN")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan challenge reset PIN")
	}

	message := fmt.Sprintf("Kode OTP reset PIN LUMBERA Anda adalah *%s*. Berlaku selama 5 menit. Jangan bagikan kode ini kepada siapa pun.", otp)
	err = s.deps.whatsapp.SendMessage(phoneNumber, message)
	if err != nil {
		return nil, err
	}

	return &model.ForgotPINRequestOTPResponse{
		ChallengeID:      challenge.ChallengeID,
		PhoneNumber:      phoneNumber,
		ExpiresInSeconds: int(forgotPINOTPExpiry.Seconds()),
	}, nil
}

func (s *AuthService) VerifyForgotPINOTP(req model.ForgotPINVerifyOTPRequest) (*model.ForgotPINVerifyOTPResponse, error) {
	if req.ChallengeID == uuid.Nil {
		return nil, appErrors.BadRequest("challenge_id wajib diisi")
	}

	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, appErrors.BadRequest("nomor handphone wajib diisi")
	}

	if strings.TrimSpace(req.OTP) == "" {
		return nil, appErrors.BadRequest("otp wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	challenge, err := s.deps.repository.PhoneVerificationRepository.GetPhoneVerificationChallenge(tx, model.GetPhoneVerificationChallengeParam{
		ChallengeID: req.ChallengeID,
		PhoneNumber: phoneNumber,
		Purpose:     "PIN_RESET",
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("OTP tidak ditemukan")
		}
		return nil, err
	}

	if challenge.ConsumedAt != nil {
		return nil, appErrors.BadRequest("challenge reset PIN sudah digunakan")
	}

	if challenge.VerifiedAt != nil {
		return nil, appErrors.BadRequest("OTP sudah digunakan")
	}

	if time.Now().After(challenge.ExpiresAt) {
		return nil, appErrors.BadRequest("OTP sudah kedaluwarsa")
	}

	if challenge.AttemptCount >= maxForgotPINOTPAttempt {
		return nil, appErrors.Forbidden("percobaan OTP sudah melewati batas")
	}

	err = s.deps.bcrypt.CompareAndHashPassword(challenge.OTPHash, req.OTP)
	if err != nil {
		challenge.AttemptCount++
		updateErr := s.deps.repository.PhoneVerificationRepository.UpdatePhoneVerificationChallenge(tx, challenge)
		if updateErr != nil {
			return nil, updateErr
		}

		if commitErr := tx.Commit().Error; commitErr != nil {
			return nil, commitErr
		}

		return nil, appErrors.Unauthorized("OTP tidak valid")
	}

	pinResetToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat token reset PIN")
	}

	pinResetTokenHash, err := s.deps.bcrypt.GenerateFromPassword(pinResetToken)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash token reset PIN")
	}

	now := time.Now()
	tokenExpiresAt := now.Add(forgotPINResetTokenTTL)

	challenge.VerifiedAt = &now
	challenge.VerificationTokenHash = pinResetTokenHash
	challenge.VerificationTokenExpiresAt = &tokenExpiresAt

	err = s.deps.repository.PhoneVerificationRepository.UpdatePhoneVerificationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal memverifikasi OTP")
	}

	return &model.ForgotPINVerifyOTPResponse{
		ChallengeID:      challenge.ChallengeID,
		PinResetToken:    pinResetToken,
		ExpiresInSeconds: int(forgotPINResetTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) SetForgottenPIN(req model.ForgotPINSetPINRequest) (*model.LoginResponse, error) {
	if req.ChallengeID == uuid.Nil {
		return nil, appErrors.BadRequest("challenge_id wajib diisi")
	}

	if strings.TrimSpace(req.PinResetToken) == "" {
		return nil, appErrors.BadRequest("pin_reset_token wajib diisi")
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

	challenge, err := s.deps.repository.PhoneVerificationRepository.GetPhoneVerificationChallenge(tx, model.GetPhoneVerificationChallengeParam{
		ChallengeID: req.ChallengeID,
		Purpose:     "PIN_RESET",
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("challenge reset PIN tidak ditemukan")
		}
		return nil, err
	}

	if challenge.VerifiedAt == nil {
		return nil, appErrors.BadRequest("OTP belum diverifikasi")
	}

	if challenge.ConsumedAt != nil {
		return nil, appErrors.BadRequest("token reset PIN sudah digunakan")
	}

	if challenge.VerificationTokenExpiresAt == nil || time.Now().After(*challenge.VerificationTokenExpiresAt) {
		return nil, appErrors.BadRequest("token reset PIN sudah kedaluwarsa")
	}

	if challenge.VerificationTokenHash == "" {
		return nil, appErrors.BadRequest("token reset PIN tidak tersedia")
	}

	err = s.deps.bcrypt.CompareAndHashPassword(challenge.VerificationTokenHash, req.PinResetToken)
	if err != nil {
		return nil, appErrors.Unauthorized("token reset PIN tidak valid")
	}

	loginCtx, err := s.deps.repository.UserRepository.GetCooperativeLoginContext(tx, model.GetCooperativeLoginContextParam{
		PhoneNumber: challenge.PhoneNumber,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("akun tidak ditemukan")
		}
		return nil, err
	}

	userPin, err := s.deps.repository.UserPinRepository.GetUserPinCredential(tx, model.GetUserPINCredentialParam{
		UserID: loginCtx.User.UserID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("akun belum memiliki PIN")
		}
		return nil, err
	}

	pinHash, err := s.deps.bcrypt.GenerateFromPassword(req.PIN)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash pin")
	}

	now := time.Now()

	userPin.PINHash = pinHash
	userPin.FailedAttempts = 0
	userPin.LockedUntil = nil
	userPin.LastChangedAt = now

	err = s.deps.repository.UserPinRepository.UpdateUserPin(tx, userPin)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui PIN")
	}

	challenge.ConsumedAt = &now
	err = s.deps.repository.PhoneVerificationRepository.UpdatePhoneVerificationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyelesaikan reset PIN")
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

	accessToken, err := s.deps.jwtAuth.GenerateAccessToken(jwt.GenerateAccessTokenInput{
		UserID:        loginCtx.User.UserID,
		CooperativeID: loginCtx.Membership.CooperativeID,
		SessionID:     session.SessionID,
		RoleCode:      loginCtx.Role.Code,
	})
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat access token")
	}

	err = s.deps.repository.UserSessionRepository.CreateUserSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan session")
	}

	loginCtx.User.LastLoginAt = &now
	err = s.deps.repository.UserRepository.UpdateUser(tx, loginCtx.User)
	if err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui data login")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mereset PIN")
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

func (s *AuthService) Logout(req model.LogoutRequest) (*model.LogoutResponse, error) {
	if req.UserID == uuid.Nil || req.SessionID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	now := time.Now()
	session, err := s.deps.repository.UserSessionRepository.GetActiveUserSession(tx, req.UserID, req.SessionID, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("session tidak valid")
		}
		return nil, appErrors.InternalServer("gagal mengambil session")
	}

	session.RevokedAt = &now
	if err := s.deps.repository.UserSessionRepository.UpdateUserSession(tx, session); err != nil {
		return nil, appErrors.InternalServer("gagal logout")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal logout")
	}

	return &model.LogoutResponse{
		SessionID: session.SessionID,
		RevokedAt: now,
	}, nil
}

func (s *AuthService) ValidateAuthenticatedSession(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.SessionID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	_, err := s.deps.repository.UserSessionRepository.GetActiveUserSession(tx, auth.UserID, auth.SessionID, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.Unauthorized("session tidak valid")
		}
		return appErrors.InternalServer("gagal memvalidasi session")
	}

	if err := tx.Commit().Error; err != nil {
		return appErrors.InternalServer("gagal memvalidasi session")
	}

	return nil
}

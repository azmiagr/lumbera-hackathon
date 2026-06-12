package model

import "github.com/google/uuid"

type GetUserPINCredentialParam struct {
	PinCredentialID uuid.UUID `json:"-"`
	UserID          uuid.UUID `json:"-"`
}

type GetUserSessionParam struct {
	SessionID uuid.UUID `json:"-"`
	UserID    uuid.UUID `json:"-"`
	DeviceID  string    `json:"-"`
}

type GetActiveCooperativeOfficerMembershipParam struct {
	UserID uuid.UUID `json:"-"`
}

type GetPhoneVerificationChallengeParam struct {
	ChallengeID uuid.UUID `json:"-"`
	PhoneNumber string    `json:"-"`
	Purpose     string    `json:"-"`
}

type CheckOfficerPhoneRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type CheckOfficerPhoneResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
}

type RequestOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
	Purpose     string `json:"purpose"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
	Purpose     string `json:"purpose"`
	OTP         string `json:"otp"`
}

type SetPINRequest struct {
	UserID uuid.UUID `json:"user_id"`
	PIN    string    `json:"pin"`
}

type LoginOfficerRequest struct {
	PhoneNumber string `json:"phone_number"`
	PIN         string `json:"pin"`
	DeviceID    string `json:"device_id"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
}

type LoginOfficerResponse struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	UserID        uuid.UUID `json:"user_id"`
	CooperativeID uuid.UUID `json:"cooperative_id"`
	RoleID        uuid.UUID `json:"role_id"`
}

type GetCooperativeLoginContextParam struct {
	PhoneNumber string    `json:"-"`
	UserID      uuid.UUID `json:"-"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	PIN         string `json:"pin"`
	DeviceID    string `json:"device_id"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
}

type LoginResponse struct {
	AccessToken   string     `json:"access_token"`
	RefreshToken  string     `json:"refresh_token"`
	UserID        uuid.UUID  `json:"user_id"`
	CooperativeID uuid.UUID  `json:"cooperative_id"`
	RoleID        uuid.UUID  `json:"role_id"`
	RoleCode      string     `json:"role_code"`
	MemberID      *uuid.UUID `json:"member_id"`
}

type ForgotPINRequestOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type ForgotPINRequestOTPResponse struct {
	ChallengeID      uuid.UUID `json:"challenge_id"`
	PhoneNumber      string    `json:"phone_number"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type ForgotPINVerifyOTPRequest struct {
	ChallengeID uuid.UUID `json:"challenge_id"`
	PhoneNumber string    `json:"phone_number"`
	OTP         string    `json:"otp"`
}

type ForgotPINVerifyOTPResponse struct {
	ChallengeID      uuid.UUID `json:"challenge_id"`
	PinResetToken    string    `json:"pin_reset_token"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type ForgotPINSetPINRequest struct {
	ChallengeID   uuid.UUID `json:"challenge_id"`
	PinResetToken string    `json:"pin_reset_token"`
	PIN           string    `json:"pin"`
	ConfirmPIN    string    `json:"confirm_pin"`
	DeviceID      string    `json:"device_id"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
}

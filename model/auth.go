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

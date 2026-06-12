package model

import "github.com/google/uuid"

type GetMemberActivationChallengeParam struct {
	ActivationChallengeID uuid.UUID `json:"-"`
	UserID                uuid.UUID `json:"-"`
}

type GetEligibleMemberActivationContextParam struct {
	UserID      uuid.UUID `json:"-"`
	PhoneNumber string    `json:"-"`
}

type CheckMemberPhoneRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type CheckMemberPhoneResponse struct {
	ActivationChallengeID uuid.UUID `json:"activation_challenge_id"`
	ActivationToken       string    `json:"activation_token"`
	PhoneNumber           string    `json:"phone_number"`
	ExpiresInSeconds      int       `json:"expires_in_seconds"`
}

type SetMemberPINRequest struct {
	ActivationChallengeID uuid.UUID `json:"activation_challenge_id"`
	ActivationToken       string    `json:"activation_token"`
	PIN                   string    `json:"pin"`
	ConfirmPIN            string    `json:"confirm_pin"`
	DeviceID              string    `json:"device_id"`
	IPAddress             string    `json:"ip_address"`
	UserAgent             string    `json:"user_agent"`
}

type SetMemberPINResponse struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	UserID        uuid.UUID `json:"user_id"`
	MemberID      uuid.UUID `json:"member_id"`
	CooperativeID uuid.UUID `json:"cooperative_id"`
	RoleID        uuid.UUID `json:"role_id"`
}

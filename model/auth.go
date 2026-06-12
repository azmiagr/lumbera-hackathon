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

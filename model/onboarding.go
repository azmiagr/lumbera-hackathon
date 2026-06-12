package model

import "github.com/google/uuid"

type GetUserIdentityParam struct {
	IdentityID uuid.UUID `json:"-"`
	UserID     uuid.UUID `json:"-"`
	NIKHash    string    `json:"-"`
}

type GetRoleParam struct {
	RoleID    uuid.UUID `json:"-"`
	Code      string    `json:"-"`
	ScopeType string    `json:"-"`
}

type GetCooperativeParam struct {
	CooperativeID      uuid.UUID `json:"-"`
	CooperativeCode    string    `json:"-"`
	RegistrationNumber string    `json:"-"`
}

type GetFinancialConfigurationParam struct {
	FinancialConfigurationID uuid.UUID `json:"-"`
	CooperativeID            uuid.UUID `json:"-"`
}

type GetUserCooperativeMembershipParam struct {
	CooperativeMembershipID uuid.UUID `json:"-"`
	UserID                  uuid.UUID `json:"-"`
	CooperativeID           uuid.UUID `json:"-"`
	RoleID                  uuid.UUID `json:"-"`
	Status                  string    `json:"-"`
}

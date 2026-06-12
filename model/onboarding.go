package model

import (
	"mime/multipart"

	"github.com/google/uuid"
)

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

type GetOnboardingDraftParam struct {
	OnboardingDraftID uuid.UUID `json:"-"`
	PhoneNumber       string    `json:"-"`
	Status            string    `json:"-"`
}

type StartOfficerRegistrationRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type StartOfficerRegistrationResponse struct {
	OnboardingDraftID uuid.UUID `json:"onboarding_draft_id"`
	PhoneNumber       string    `json:"phone_number"`
	ExpiresInSeconds  int       `json:"expires_in_seconds"`
}

type VerifyOfficerRegistrationOTPRequest struct {
	OnboardingDraftID uuid.UUID `json:"onboarding_draft_id"`
	OTP               string    `json:"otp"`
}

type SetOfficerRegistrationPINRequest struct {
	OnboardingDraftID uuid.UUID `json:"onboarding_draft_id"`
	PIN               string    `json:"pin"`
	ConfirmPIN        string    `json:"confirm_pin"`
}

type SetOfficerRegistrationPINResponse struct {
	OnboardingDraftID uuid.UUID `json:"onboarding_draft_id"`
	OnboardingToken   string    `json:"onboarding_token"`
	NextStep          int       `json:"next_step"`
}

type UpdatePersonalDataRequest struct {
	OnboardingDraftID       uuid.UUID             `form:"-"`
	OnboardingToken         string                `form:"-"`
	KTPFile                 *multipart.FileHeader `form:"ktp_file"`
	FullName                string                `form:"full_name"`
	NIKEncrypted            string                `form:"nik_encrypted"`
	NIKHash                 string                `form:"nik_hash"`
	PositionCode            string                `form:"position_code"`
	ExistingCooperativeCode string                `form:"existing_cooperative_code"`
}

type UpdateCooperativeTypeRequest struct {
	OnboardingDraftID uuid.UUID `json:"-"`
	OnboardingToken   string    `json:"-"`
	CooperativeType   string    `json:"cooperative_type"`
}

type UpdateCooperativeProfileRequest struct {
	OnboardingDraftID  uuid.UUID `json:"-"`
	OnboardingToken    string    `json:"-"`
	CooperativeName    string    `json:"cooperative_name"`
	RegistrationNumber string    `json:"registration_number"`
	Address            string    `json:"address"`
	EstablishedYear    int       `json:"established_year"`
}

type OnboardingStepResponse struct {
	OnboardingDraftID uuid.UUID `json:"onboarding_draft_id"`
	CurrentStep       int       `json:"current_step"`
	NextStep          string    `json:"next_step"`
}

type UpdateFinancialConfigurationRequest struct {
	OnboardingDraftID           uuid.UUID `json:"-"`
	OnboardingToken             string    `json:"-"`
	MaxLoanAmountPerMember      int64     `json:"max_loan_amount_per_member"`
	LoanInterestRateBpsPerMonth int       `json:"loan_interest_rate_bps_per_month"`
	LateFeeRateBpsPerDay        int       `json:"late_fee_rate_bps_per_day"`
	MaxLoanTermMonths           int       `json:"max_loan_term_months"`
	MandatorySavingsPerMonth    int64     `json:"mandatory_savings_per_month"`
}

type UpdateCooperativeBankAccountRequest struct {
	OnboardingDraftID     uuid.UUID `json:"-"`
	OnboardingToken       string    `json:"-"`
	BankName              string    `json:"bank_name"`
	BankAccountNumber     string    `json:"bank_account_number"`
	BankAccountHolderName string    `json:"bank_account_holder_name"`
}

type ActivateOnboardingDraftRequest struct {
	OnboardingDraftID uuid.UUID `json:"-"`
	OnboardingToken   string    `json:"-"`
}

type ActivateOnboardingDraftResponse struct {
	UserID        uuid.UUID `json:"user_id"`
	CooperativeID uuid.UUID `json:"cooperative_id"`
	MembershipID  uuid.UUID `json:"membership_id"`
	NextStep      string    `json:"next_step"`
}

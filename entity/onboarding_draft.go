package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OnboardingDraft struct {
	OnboardingDraftID           uuid.UUID      `json:"onboarding_draft_id" gorm:"type:varchar(36);primaryKey"`
	PhoneNumber                 string         `json:"phone_number" gorm:"type:varchar(20);not null;index"`
	PINHash                     string         `json:"-" gorm:"type:varchar(255)"`
	SessionTokenHash            string         `json:"-" gorm:"type:varchar(255)"`
	CurrentStep                 int            `json:"current_step" gorm:"not null;default:0"`
	Status                      string         `json:"status" gorm:"type:enum('OTP_PENDING','OTP_VERIFIED','PIN_SET','IN_PROGRESS','ACTIVATED','EXPIRED');default:'OTP_PENDING';not null"`
	KTPImageURL                 string         `json:"ktp_image_url" gorm:"type:text"`
	FullName                    string         `json:"full_name" gorm:"type:varchar(255)"`
	NIKEncrypted                string         `json:"-" gorm:"type:varchar(255)"`
	NIKHash                     string         `json:"nik_hash" gorm:"type:varchar(255);index"`
	RoleCode                    string         `json:"role_code" gorm:"type:varchar(100)"`
	PositionCode                string         `json:"position_code" gorm:"type:varchar(50)"`
	ExistingCooperativeCode     string         `json:"existing_cooperative_code" gorm:"type:varchar(50)"`
	CooperativeType             string         `json:"cooperative_type" gorm:"type:varchar(50)"`
	CooperativeName             string         `json:"cooperative_name" gorm:"type:varchar(255)"`
	RegistrationNumber          string         `json:"registration_number" gorm:"type:varchar(50)"`
	Address                     string         `json:"address" gorm:"type:text"`
	EstablishedYear             int            `json:"established_year" gorm:"type:int"`
	MaxLoanAmountPerMember      int64          `json:"max_loan_amount_per_member" gorm:"not null;default:0"`
	LoanInterestRateBpsPerMonth int            `json:"loan_interest_rate_bps_per_month" gorm:"not null;default:0"`
	LateFeeRateBpsPerDay        int            `json:"late_fee_rate_bps_per_day" gorm:"not null;default:0"`
	MaxLoanTermMonths           int            `json:"max_loan_term_months" gorm:"not null;default:0"`
	MandatorySavingsPerMonth    int64          `json:"mandatory_savings_per_month" gorm:"not null;default:0"`
	BankName                    string         `json:"bank_name" gorm:"type:varchar(100)"`
	BankAccountNumber           string         `json:"bank_account_number" gorm:"type:varchar(100)"`
	BankAccountHolderName       string         `json:"bank_account_holder_name" gorm:"type:varchar(255)"`
	ActivatedUserID             *uuid.UUID     `json:"activated_user_id" gorm:"type:varchar(36)"`
	ActivatedCooperativeID      *uuid.UUID     `json:"activated_cooperative_id" gorm:"type:varchar(36)"`
	PhoneVerifiedAt             *time.Time     `json:"phone_verified_at"`
	PINSetAt                    *time.Time     `json:"pin_set_at"`
	ActivatedAt                 *time.Time     `json:"activated_at"`
	ExpiresAt                   time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt                   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

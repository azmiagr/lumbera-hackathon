package entity

import (
	"time"

	"github.com/google/uuid"
)

type FinancialConfiguration struct {
	FinancialConfigurationID    uuid.UUID `json:"financial_configuration_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID               uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	MaxLoanAmountPerMember      int64     `json:"max_loan_amount_per_member" gorm:"not null;default:0"`
	LoanInterestRateBpsPerMonth int       `json:"loan_interest_rate_bps_per_month" gorm:"not null;default:0"`
	LateFeeRateBpsPerDay        int       `json:"late_fee_rate_bps_per_day" gorm:"not null;default:0"`
	MaxLoanTermMonths           int       `json:"max_loan_term_months" gorm:"not null;default:0"`
	MandatorySavingsPerMonth    int64     `json:"mandatory_savings_per_month" gorm:"not null;default:0"`
	CreatedAt                   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

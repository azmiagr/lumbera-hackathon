package entity

import (
	"time"

	"github.com/google/uuid"
)

type LoanApplication struct {
	ApplicationID           uuid.UUID  `json:"application_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID           uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	MemberID                uuid.UUID  `json:"member_id" gorm:"type:varchar(36);not null;index"`
	RequestedAmount         int64      `json:"requested_amount" gorm:"not null"`
	Purpose                 string     `json:"purpose" gorm:"type:varchar(255);not null"`
	TermMonths              int        `json:"term_months" gorm:"not null"`
	InterestRateBpsPerMonth int        `json:"interest_rate_bps_per_month" gorm:"not null"`
	MonthlyInstallment      int64      `json:"monthly_installment" gorm:"not null"`
	TotalInterestAmount     int64      `json:"total_interest_amount" gorm:"not null"`
	TotalPayableAmount      int64      `json:"total_payable_amount" gorm:"not null"`
	MCSScore                *int       `json:"mcs_score"`
	MCSGrade                string     `json:"mcs_grade" gorm:"type:varchar(5)"`
	CreditLimitAmount       int64      `json:"credit_limit_amount" gorm:"not null"`
	Status                  string     `json:"status" gorm:"type:enum('RECEIVED','CREDIT_VERIFIED','UNDER_REVIEW','APPROVED','REJECTED','DISBURSED','CANCELLED');not null;default:'RECEIVED';index"`
	PartnerName             string     `json:"partner_name" gorm:"type:varchar(100);not null;default:'Akseleran'"`
	SubmittedAt             time.Time  `json:"submitted_at" gorm:"not null;index"`
	CreditVerifiedAt        *time.Time `json:"credit_verified_at"`
	ReviewedAt              *time.Time `json:"reviewed_at"`
	DisbursedAt             *time.Time `json:"disbursed_at"`
	CreatedAt               time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	CreditAccessRequests []CreditAccessRequest `json:"credit_access_requests" gorm:"foreignKey:ApplicationID;constraint:onDelete:CASCADE"`
}

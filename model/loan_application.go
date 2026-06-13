package model

import (
	"time"

	"github.com/google/uuid"
)

type GetLoanApplicationEligibilityRequest struct {
	AuthContext
}

type CreateLoanApplicationRequest struct {
	AuthContext
	Amount     int64  `json:"amount"`
	Purpose    string `json:"purpose"`
	TermMonths int    `json:"term_months"`
}

type GetLoanApplicationRequest struct {
	AuthContext
	ApplicationID uuid.UUID `json:"-"`
}

type LoanApplicationEligibilityResponse struct {
	Eligible                bool   `json:"eligible"`
	Reason                  string `json:"reason,omitempty"`
	MCSScore                *int   `json:"mcs_score"`
	MCSGrade                string `json:"mcs_grade"`
	CreditLimitAmount       int64  `json:"credit_limit_amount"`
	MaxTermMonths           int    `json:"max_term_months"`
	InterestRateBpsPerMonth int    `json:"interest_rate_bps_per_month"`
	HasActiveLoan           bool   `json:"has_active_loan"`
	HasActiveApplication    bool   `json:"has_active_application"`
}

type LoanApplicationResponse struct {
	ApplicationID           uuid.UUID                     `json:"application_id"`
	Status                  string                        `json:"status"`
	StatusLabel             string                        `json:"status_label"`
	RequestedAmount         int64                         `json:"requested_amount"`
	Purpose                 string                        `json:"purpose"`
	TermMonths              int                           `json:"term_months"`
	MonthlyInstallment      int64                         `json:"monthly_installment"`
	TotalInterestAmount     int64                         `json:"total_interest_amount"`
	TotalPayableAmount      int64                         `json:"total_payable_amount"`
	InterestRateBpsPerMonth int                           `json:"interest_rate_bps_per_month"`
	MCSScore                *int                          `json:"mcs_score"`
	MCSGrade                string                        `json:"mcs_grade"`
	CreditLimitAmount       int64                         `json:"credit_limit_amount"`
	PartnerName             string                        `json:"partner_name"`
	SubmittedAt             time.Time                     `json:"submitted_at"`
	Timeline                []LoanApplicationTimelineItem `json:"timeline"`
}

type LoanApplicationTimelineItem struct {
	Code        string     `json:"code"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	State       string     `json:"state"` // done | current | pending
	OccurredAt  *time.Time `json:"occurred_at,omitempty"`
}

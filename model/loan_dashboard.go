package model

import (
	"time"

	"github.com/google/uuid"
)

type GetLoanDashboardRequest struct {
	AuthContext
}

type LoanDashboardResponse struct {
	MCS             LoanDashboardMCS             `json:"mcs"`
	ActiveLoan      *LoanDashboardActiveLoan     `json:"active_loan"`
	Installments    []LoanDashboardInstallment   `json:"installments"`
	InstallmentMeta LoanDashboardInstallmentMeta `json:"installment_meta"`
	LoanHistory     []LoanDashboardHistoryItem   `json:"loan_history"`
	Actions         LoanDashboardActions         `json:"actions"`
}

type LoanDashboardMCS struct {
	Score              *int                        `json:"score"`
	MaxScore           int                         `json:"max_score"`
	Grade              string                      `json:"grade"`
	Label              string                      `json:"label"`
	ProfileText        string                      `json:"profile_text"`
	LastScoreUpdatedAt *time.Time                  `json:"last_score_updated_at"`
	Components         []LoanDashboardMCSComponent `json:"components"`
	Explanation        string                      `json:"explanation"`
}

type LoanDashboardMCSComponent struct {
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	Score  *float64 `json:"score"`
	Weight float64  `json:"weight"`
}

type LoanDashboardActiveLoan struct {
	LoanID                   uuid.UUID  `json:"loan_id"`
	LoanNumber               string     `json:"loan_number"`
	PrincipalAmount          int64      `json:"principal_amount"`
	TotalPayableAmount       int64      `json:"total_payable_amount"`
	RemainingPayableAmount   int64      `json:"remaining_payable_amount"`
	MonthlyInstallmentAmount int64      `json:"monthly_installment_amount"`
	TermMonths               int        `json:"term_months"`
	PaidInstallmentCount     int        `json:"paid_installment_count"`
	PaidPercentage           float64    `json:"paid_percentage"`
	NextDueDate              *time.Time `json:"next_due_date"`
	PaymentStatusText        string     `json:"payment_status_text"`
}

type LoanDashboardInstallment struct {
	InstallmentNo int       `json:"installment_no"`
	DueDate       time.Time `json:"due_date"`
	DueAmount     int64     `json:"due_amount"`
	PaidAmount    int64     `json:"paid_amount"`
	Status        string    `json:"status"`
	StatusLabel   string    `json:"status_label"`
}

type LoanDashboardInstallmentMeta struct {
	TotalCount     int  `json:"total_count"`
	DisplayedCount int  `json:"displayed_count"`
	RemainingCount int  `json:"remaining_count"`
	HasMore        bool `json:"has_more"`
}

type LoanDashboardHistoryItem struct {
	LoanID          uuid.UUID  `json:"loan_id"`
	LoanNumber      string     `json:"loan_number"`
	PrincipalAmount int64      `json:"principal_amount"`
	TermMonths      int        `json:"term_months"`
	Status          string     `json:"status"`
	StatusLabel     string     `json:"status_label"`
	DisbursedAt     time.Time  `json:"disbursed_at"`
	PaidAt          *time.Time `json:"paid_at"`
	Description     string     `json:"description"`
}

type LoanDashboardActions struct {
	HistoryEnabled         bool `json:"history_enabled"`
	LoanApplicationEnabled bool `json:"loan_application_enabled"`
	CreditAccessEnabled    bool `json:"credit_access_enabled"`
}

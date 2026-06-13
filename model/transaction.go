package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthContext struct {
	UserID        uuid.UUID
	CooperativeID uuid.UUID
	SessionID     uuid.UUID
	RoleCode      string
}

type CreateSavingsTransactionRequest struct {
	AuthContext
	MemberID            uuid.UUID  `json:"member_id"`
	SavingsType         string     `json:"savings_type"`
	Amount              int64      `json:"amount"`
	Description         string     `json:"description"`
	RecordedAt          *time.Time `json:"recorded_at"`
	IsOfflineCreated    bool       `json:"is_offline_created"`
	ClientTransactionID string     `json:"client_transaction_id"`
}

type TransactionResponse struct {
	TransactionID            uuid.UUID  `json:"transaction_id"`
	CooperativeID            uuid.UUID  `json:"cooperative_id"`
	MemberID                 uuid.UUID  `json:"member_id"`
	MemberName               string     `json:"member_name"`
	MemberNumber             string     `json:"member_number"`
	MemberMCSGrade           string     `json:"member_mcs_grade"`
	OfficerID                uuid.UUID  `json:"officer_id"`
	OfficerName              string     `json:"officer_name"`
	TransactionType          string     `json:"transaction_type"`
	TransactionTypeLabel     string     `json:"transaction_type_label"`
	TransactionGroup         string     `json:"transaction_group"`
	Amount                   int64      `json:"amount"`
	Description              string     `json:"description"`
	RecordedAt               time.Time  `json:"recorded_at"`
	SyncedAt                 *time.Time `json:"synced_at"`
	PrevHash                 string     `json:"prev_hash"`
	CurrentHash              string     `json:"current_hash"`
	HashPreview              string     `json:"hash_preview"`
	IsOfflineCreated         bool       `json:"is_offline_created"`
	ClientTransactionID      string     `json:"client_transaction_id"`
	SyncStatus               string     `json:"sync_status"`
	MemberSavingsBalance     int64      `json:"member_savings_balance"`
	MemberLoanOutstanding    int64      `json:"member_loan_outstanding"`
	LoanID                   *uuid.UUID `json:"loan_id,omitempty"`
	LoanNumber               string     `json:"loan_number,omitempty"`
	MonthlyInstallmentAmount int64      `json:"monthly_installment_amount,omitempty"`
	RemainingPayableAmount   int64      `json:"remaining_payable_amount,omitempty"`
	CurrentMonthDueAmount    int64      `json:"current_month_due_amount,omitempty"`
}

type SearchTransactionMembersRequest struct {
	AuthContext
	Search string `form:"search"`
	Limit  int    `form:"limit"`
}

type TransactionMemberResponse struct {
	MemberID        uuid.UUID `json:"member_id"`
	UserID          uuid.UUID `json:"user_id"`
	FullName        string    `json:"full_name"`
	MemberNumber    string    `json:"member_number"`
	CooperativeID   uuid.UUID `json:"cooperative_id"`
	MCSGrade        string    `json:"mcs_grade"`
	SavingsBalance  int64     `json:"savings_balance"`
	LoanOutstanding int64     `json:"loan_outstanding"`
}

type ListTransactionsRequest struct {
	AuthContext
	Search string `form:"search"`
	Type   string `form:"type"`
	Limit  int    `form:"limit"`
	Page   int    `form:"page"`
}

type TransactionListItemResponse struct {
	TransactionID        uuid.UUID  `json:"transaction_id"`
	CooperativeID        uuid.UUID  `json:"cooperative_id"`
	MemberID             uuid.UUID  `json:"member_id"`
	MemberName           string     `json:"member_name"`
	MemberNumber         string     `json:"member_number"`
	MemberMCSGrade       string     `json:"member_mcs_grade"`
	OfficerID            uuid.UUID  `json:"officer_id"`
	OfficerName          string     `json:"officer_name"`
	TransactionType      string     `json:"transaction_type"`
	TransactionTypeLabel string     `json:"transaction_type_label"`
	TransactionGroup     string     `json:"transaction_group"`
	Amount               int64      `json:"amount"`
	Description          string     `json:"description"`
	RecordedAt           time.Time  `json:"recorded_at"`
	SyncedAt             *time.Time `json:"synced_at"`
	CurrentHash          string     `json:"current_hash"`
	HashPreview          string     `json:"hash_preview"`
	IsOfflineCreated     bool       `json:"is_offline_created"`
	ClientTransactionID  string     `json:"client_transaction_id"`
	SyncStatus           string     `json:"sync_status"`
}

type ListTransactionsResponse struct {
	Items []TransactionListItemResponse `json:"items"`
	Page  int                           `json:"page"`
	Limit int                           `json:"limit"`
	Total int64                         `json:"total"`
}

type CreateLoanTransactionRequest struct {
	AuthContext
	MemberID            uuid.UUID  `json:"member_id"`
	Amount              int64      `json:"amount"`
	Description         string     `json:"description"`
	RecordedAt          *time.Time `json:"recorded_at"`
	IsOfflineCreated    bool       `json:"is_offline_created"`
	ClientTransactionID string     `json:"client_transaction_id"`
}

type TransactionMemberSummaryResponse struct {
	SavingsBalance  int64 `json:"savings_balance"`
	LoanOutstanding int64 `json:"loan_outstanding"`
}

type CreateInstallmentTransactionRequest struct {
	AuthContext
	LoanID              uuid.UUID  `json:"loan_id"`
	Amount              int64      `json:"amount"`
	Description         string     `json:"description"`
	RecordedAt          *time.Time `json:"recorded_at"`
	IsOfflineCreated    bool       `json:"is_offline_created"`
	ClientTransactionID string     `json:"client_transaction_id"`
}

type LoanSummaryResponse struct {
	LoanID                   uuid.UUID `json:"loan_id"`
	LoanNumber               string    `json:"loan_number"`
	Status                   string    `json:"status"`
	PrincipalAmount          int64     `json:"principal_amount"`
	TotalPayableAmount       int64     `json:"total_payable_amount"`
	MonthlyInstallmentAmount int64     `json:"monthly_installment_amount"`
	RemainingPayableAmount   int64     `json:"remaining_payable_amount"`
	CurrentMonthDueAmount    int64     `json:"current_month_due_amount"`
	TermMonths               int       `json:"term_months"`
}

type InstallmentAllocationResponse struct {
	ScheduleID      uuid.UUID `json:"schedule_id"`
	InstallmentNo   int       `json:"installment_no"`
	DueDate         time.Time `json:"due_date"`
	AllocatedAmount int64     `json:"allocated_amount"`
	ScheduleStatus  string    `json:"schedule_status"`
}

type InstallmentTransactionResponse struct {
	TransactionResponse
	Loan        LoanSummaryResponse             `json:"loan"`
	Allocations []InstallmentAllocationResponse `json:"allocations"`
}

type CreateCashWithdrawalTransactionRequest struct {
	AuthContext
	MemberID            uuid.UUID  `json:"member_id"`
	Amount              int64      `json:"amount"`
	Description         string     `json:"description"`
	RecordedAt          *time.Time `json:"recorded_at"`
	IsOfflineCreated    bool       `json:"is_offline_created"`
	ClientTransactionID string     `json:"client_transaction_id"`
}

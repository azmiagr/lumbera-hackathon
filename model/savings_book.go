package model

import (
	"time"

	"github.com/google/uuid"
)

type SavingsBookType string

const (
	SavingsBookTypeAll     = "SEMUA"
	SavingsBookTypeIncome  = "PEMASUKAN"
	SavingsBookTypeExpense = "PENGELUARAN"
)

type GetSavingsBookRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
	Type   string `form:"type"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type ExportSavingsBookRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
	Type   string `form:"type"`
}

type SavingsBookProfile struct {
	MemberID        uuid.UUID `json:"member_id"`
	UserID          uuid.UUID `json:"user_id"`
	FullName        string    `json:"full_name"`
	MemberNumber    string    `json:"member_number"`
	CooperativeID   uuid.UUID `json:"cooperative_id"`
	CooperativeName string    `json:"cooperative_name"`
}

type SavingsBookSummary struct {
	TotalBalance int64 `json:"total_balance"`
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
}

type SavingsBookItem struct {
	TransactionID        uuid.UUID `json:"transaction_id"`
	TransactionType      string    `json:"transaction_type"`
	TransactionTypeLabel string    `json:"transaction_type_label"`
	Direction            string    `json:"direction"`
	Amount               int64     `json:"amount"`
	IncomeAmount         int64     `json:"income_amount"`
	ExpenseAmount        int64     `json:"expense_amount"`
	RecorderName         string    `json:"recorder_name"`
	Description          string    `json:"description"`
	RecordedAt           time.Time `json:"recorded_at"`
}

type SavingsBookResponse struct {
	Profile SavingsBookProfile `json:"profile"`
	Period  string             `json:"period"`
	Summary SavingsBookSummary `json:"summary"`
	Items   []SavingsBookItem  `json:"items"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int64              `json:"total"`
}

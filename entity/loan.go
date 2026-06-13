package entity

import (
	"time"

	"github.com/google/uuid"
)

type LoanAccount struct {
	LoanID                    uuid.UUID `json:"loan_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID             uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_loan_number"`
	MemberID                  uuid.UUID `json:"member_id" gorm:"type:varchar(36);not null;index"`
	DisbursementTransactionID uuid.UUID `json:"disbursement_transaction_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	LoanNumber                string    `json:"loan_number" gorm:"type:varchar(30);not null;uniqueIndex:idx_coop_loan_number"`
	PrincipalAmount           int64     `json:"principal_amount" gorm:"not null"`
	TotalInterestAmount       int64     `json:"total_interest_amount" gorm:"not null;default:0"`
	TotalPayableAmount        int64     `json:"total_payable_amount" gorm:"not null"`
	MonthlyInstallmentAmount  int64     `json:"monthly_installment_amount" gorm:"not null"`
	InterestRateBpsPerMonth   int       `json:"interest_rate_bps_per_month" gorm:"not null;default:0"`
	TermMonths                int       `json:"term_months" gorm:"not null"`
	Status                    string    `json:"status" gorm:"type:enum('ACTIVE','PAID','CANCELLED');not null;default:'ACTIVE';index"`
	DisbursedAt               time.Time `json:"disbursed_at" gorm:"not null;index"`
	CreatedAt                 time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	LoanInstallmentSchedules []LoanInstallmentSchedule `json:"loan_installment_schedules" gorm:"foreignKey:LoanID;constraint:onDelete:CASCADE"`
	LoanPaymentAllocations   []LoanPaymentAllocation   `json:"loan_payment_allocations" gorm:"foreignKey:LoanID;constraint:onDelete:CASCADE"`
}

type LoanInstallmentSchedule struct {
	ScheduleID      uuid.UUID  `json:"schedule_id" gorm:"type:varchar(36);primaryKey"`
	LoanID          uuid.UUID  `json:"loan_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_loan_installment_no"`
	InstallmentNo   int        `json:"installment_no" gorm:"not null;uniqueIndex:idx_loan_installment_no"`
	DueDate         time.Time  `json:"due_date" gorm:"type:date;not null;index"`
	DueAmount       int64      `json:"due_amount" gorm:"not null"`
	PaidAmount      int64      `json:"paid_amount" gorm:"not null;default:0"`
	RemainingAmount int64      `json:"remaining_amount" gorm:"not null"`
	Status          string     `json:"status" gorm:"type:enum('UNPAID','PARTIAL','PAID');not null;default:'UNPAID';index"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type LoanPaymentAllocation struct {
	AllocationID  uuid.UUID `json:"allocation_id" gorm:"type:varchar(36);primaryKey"`
	LoanID        uuid.UUID `json:"loan_id" gorm:"type:varchar(36);not null;index"`
	ScheduleID    uuid.UUID `json:"schedule_id" gorm:"type:varchar(36);not null;index"`
	TransactionID uuid.UUID `json:"transaction_id" gorm:"type:varchar(36);not null;index"`
	Amount        int64     `json:"amount" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

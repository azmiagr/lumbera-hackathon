package entity

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	AccountID       uuid.UUID `json:"account_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID   uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_account_code"`
	Code            string    `json:"code" gorm:"type:varchar(20);not null;uniqueIndex:idx_coop_account_code"`
	Name            string    `json:"name" gorm:"type:varchar(100);not null"`
	AccountType     string    `json:"account_type" gorm:"type:enum('ASSET','LIABILITY','EQUITY','REVENUE','EXPENSE');not null;index"`
	NormalBalance   string    `json:"normal_balance" gorm:"type:enum('DEBIT','CREDIT');not null"`
	CashFlowGroup   string    `json:"cash_flow_group" gorm:"type:enum('OPERATING','INVESTING','FINANCING','NONE');default:'NONE';not null"`
	IsSystemAccount bool      `json:"is_system_account" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	JournalEntryLines []JournalEntryLine `json:"journal_entry_lines" gorm:"foreignKey:AccountID;constraint:onDelete:CASCADE"`
}

type JournalEntry struct {
	JournalEntryID uuid.UUID  `json:"journal_entry_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID  uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	TransactionID  *uuid.UUID `json:"transaction_id" gorm:"type:varchar(36);index"`
	EntryDate      time.Time  `json:"entry_date" gorm:"not null;index"`
	Description    string     `json:"description" gorm:"type:text"`
	SourceType     string     `json:"source_type" gorm:"type:varchar(50);not null"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`

	Lines []JournalEntryLine `json:"lines" gorm:"foreignKey:JournalEntryID;constraint:onDelete:CASCADE"`
}

type JournalEntryLine struct {
	JournalEntryLineID uuid.UUID `json:"journal_entry_line_id" gorm:"type:varchar(36);primaryKey"`
	JournalEntryID     uuid.UUID `json:"journal_entry_id" gorm:"type:varchar(36);not null;index"`
	AccountID          uuid.UUID `json:"account_id" gorm:"type:varchar(36);not null;index"`
	DebitAmount        int64     `json:"debit_amount" gorm:"not null;default:0"`
	CreditAmount       int64     `json:"credit_amount" gorm:"not null;default:0"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
}

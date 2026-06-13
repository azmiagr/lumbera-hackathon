package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionReversal struct {
	TransactionReversalID uuid.UUID `json:"transaction_reversal_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID         uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	OriginalTransactionID uuid.UUID `json:"original_transaction_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	ReversalTransactionID uuid.UUID `json:"reversal_transaction_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	Reason                string    `json:"reason" gorm:"type:text;not null"`
	CreatedBy             uuid.UUID `json:"created_by" gorm:"type:varchar(36);not null;index"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

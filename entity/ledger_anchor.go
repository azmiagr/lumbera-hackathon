package entity

import (
	"time"

	"github.com/google/uuid"
)

type LedgerAnchor struct {
	LedgerAnchorID        uuid.UUID `json:"ledger_anchor_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID         uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	PeriodStart           time.Time `json:"period_start" gorm:"not null;index"`
	PeriodEnd             time.Time `json:"period_end" gorm:"not null;index"`
	MerkleRootHash        string    `json:"merkle_root_hash" gorm:"type:varchar(64);not null"`
	BlockchainNetwork     string    `json:"blockchain_network" gorm:"type:varchar(80);not null"`
	BlockchainBlockNumber int64     `json:"blockchain_block_number" gorm:"not null"`
	BlockchainTxID        string    `json:"blockchain_tx_id" gorm:"type:varchar(100);not null;uniqueIndex"`
	AnchoredAt            time.Time `json:"anchored_at" gorm:"not null;index"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

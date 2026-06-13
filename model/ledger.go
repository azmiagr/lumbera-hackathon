package model

import (
	"time"

	"github.com/google/uuid"
)

type LedgerAuditRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type LedgerAuditResponse struct {
	Period        string                    `json:"period"`
	OverallStatus string                    `json:"overall_status"`
	Anchor        *LedgerAnchorResponse     `json:"anchor,omitempty"`
	Certificate   LedgerCertificateResponse `json:"certificate"`
	Items         []LedgerAuditItemResponse `json:"items"`
	Page          int                       `json:"page"`
	Limit         int                       `json:"limit"`
	Total         int64                     `json:"total"`
}

type LedgerAuditItemResponse struct {
	RecordID      uuid.UUID `json:"record_id"`
	RecordType    string    `json:"record_type"` // TRANSACTION / STOCK_MOVEMENT
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"`
	Amount        int64     `json:"amount"`
	RecordedAt    time.Time `json:"recorded_at"`
	PrevHash      string    `json:"prev_hash"`
	CurrentHash   string    `json:"current_hash"`
	HashPreview   string    `json:"hash_preview"`
	Status        string    `json:"status"` // VALID / INVALID
	InvalidReason string    `json:"invalid_reason,omitempty"`
}

type LedgerAnchorResponse struct {
	Network        string    `json:"network"`
	BlockNumber    int64     `json:"block_number"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	AnchoredAt     time.Time `json:"anchored_at"`
}

type LedgerCertificateResponse struct {
	CooperativeName     string `json:"cooperative_name"`
	PeriodLabel         string `json:"period_label"`
	MerkleRootHash      string `json:"merkle_root_hash"`
	MerkleRootPreview   string `json:"merkle_root_preview"`
	BlockchainTxID      string `json:"blockchain_tx_id"`
	BlockchainTxPreview string `json:"blockchain_tx_preview"`
}

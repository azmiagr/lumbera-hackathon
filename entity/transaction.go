package entity

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TransactionID       uuid.UUID  `json:"transaction_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID       uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_client_transaction"`
	MemberID            uuid.UUID  `json:"member_id" gorm:"type:varchar(36);not null;index"`
	OfficerID           uuid.UUID  `json:"officer_id" gorm:"type:varchar(36);not null;index"`
	TransactionType     string     `json:"transaction_type" gorm:"type:enum('SIMPANAN_POKOK','SIMPANAN_WAJIB','SIMPANAN_SUKARELA','PINJAMAN','ANGSURAN');not null;index"`
	Amount              int64      `json:"amount" gorm:"not null"`
	Description         string     `json:"description" gorm:"type:text"`
	RecordedAt          time.Time  `json:"recorded_at" gorm:"not null;index"`
	SyncedAt            *time.Time `json:"synced_at"`
	PrevHash            string     `json:"prev_hash" gorm:"type:varchar(64);not null"`
	CurrentHash         string     `json:"current_hash" gorm:"type:varchar(64);not null;uniqueIndex"`
	IsOfflineCreated    bool       `json:"is_offline_created" gorm:"default:false"`
	ClientTransactionID string     `json:"client_transaction_id" gorm:"type:varchar(100);uniqueIndex:idx_coop_client_transaction"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`

	JournalEntries []JournalEntry `json:"journal_entries" gorm:"foreignKey:TransactionID;constraint:onDelete:CASCADE"`
}

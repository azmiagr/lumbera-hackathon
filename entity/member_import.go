package entity

import (
	"time"

	"github.com/google/uuid"
)

type MemberImportBatch struct {
	ImportBatchID uuid.UUID  `json:"import_batch_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	UploadedBy    uuid.UUID  `json:"uploaded_by" gorm:"type:varchar(36);not null;index"`
	FileName      string     `json:"file_name" gorm:"type:varchar(255);not null"`
	Status        string     `json:"status" gorm:"type:enum('DRAFT','SUBMITTED','CANCELLED');default:'DRAFT';not null;index"`
	TotalRows     int        `json:"total_rows" gorm:"not null;default:0"`
	SuccessRows   int        `json:"success_rows" gorm:"not null;default:0"`
	ErrorRows     int        `json:"error_rows" gorm:"not null;default:0"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	Rows []MemberImportRow `json:"rows" gorm:"foreignKey:ImportBatchID;constraint:onDelete:CASCADE"`
}

type MemberImportRow struct {
	ImportRowID   uuid.UUID  `json:"import_row_id" gorm:"type:varchar(36);primaryKey"`
	ImportBatchID uuid.UUID  `json:"import_batch_id" gorm:"type:varchar(36);not null;index"`
	RowNumber     int        `json:"row_number" gorm:"not null"`
	FullName      string     `json:"full_name" gorm:"type:varchar(255)"`
	NIKEncrypted  string     `json:"-" gorm:"type:varchar(255)"`
	NIKHash       string     `json:"-" gorm:"type:varchar(255);index"`
	NIKMasked     string     `json:"nik_masked" gorm:"type:varchar(32)"`
	PhoneNumber   string     `json:"phone_number" gorm:"type:varchar(20)"`
	Address       string     `json:"address" gorm:"type:text"`
	JoinedDate    *time.Time `json:"joined_date"`
	Status        string     `json:"status" gorm:"type:enum('VALID','ERROR','DELETED','IMPORTED');not null;index"`
	ErrorMessage  string     `json:"error_message" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

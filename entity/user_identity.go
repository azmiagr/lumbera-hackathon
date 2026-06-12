package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserIdentity struct {
	IdentityID     uuid.UUID  `json:"identity_id" gorm:"type:varchar(36);primaryKey"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	NIKEncrypted   string     `json:"-" gorm:"type:varchar(255);not null"`
	NIKHash        string     `json:"nik_hash" gorm:"type:varchar(255);not null;index"`
	KTPImageURL    string     `json:"ktp_image_url" gorm:"type:text"`
	BirthDate      *time.Time `json:"birth_date" gorm:"type:date"`
	Gender         string     `json:"gender" gorm:"type:enum('MALE','FEMALE','OTHER')"`
	Address        string     `json:"address" gorm:"type:text"`
	Occupation     string     `json:"occupation" gorm:"type:varchar(100)"`
	BusinessSector string     `json:"business_sector" gorm:"type:varchar(100)"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

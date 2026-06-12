package entity

import (
	"time"

	"github.com/google/uuid"
)

type MemberActivationChallenge struct {
	ChallengeID uuid.UUID  `json:"activation_challenge_id" gorm:"type:varchar(36);primaryKey"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;index"`
	TokenHash   string     `json:"-" gorm:"type:varchar(255);not null"`
	ExpiresAt   time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt      *time.Time `json:"used_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

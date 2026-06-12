package entity

import (
	"time"

	"github.com/google/uuid"
)

type PhoneVerificationChallenge struct {
	ChallengeID  uuid.UUID  `json:"challenge_id" gorm:"type:varchar(36);primaryKey"`
	PhoneNumber  string     `json:"phone_number" gorm:"type:varchar(20);not null;index"`
	OTPHash      string     `json:"-" gorm:"type:varchar(255);not null"`
	Purpose      string     `json:"purpose" gorm:"type:enum('REGISTRATION','LOGIN','PIN_RESET');not null"`
	AttemptCount int        `json:"attempt_count" gorm:"default:0"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	VerifiedAt   *time.Time `json:"verified_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserPINCredential struct {
	PinCredentialID uuid.UUID  `json:"pin_credential_id" gorm:"type:varchar(36);primaryKey"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	PINHash         string     `json:"-" gorm:"type:varchar(255);not null"`
	FailedAttempts  int        `json:"failed_attempts" gorm:"default:0"`
	LockedUntil     *time.Time `json:"locked_until"`
	LastChangedAt   time.Time  `json:"last_changed_at" gorm:"autoCreateTime"`
}

type UserSession struct {
	SessionID        uuid.UUID  `json:"session_id" gorm:"type:varchar(36);primaryKey"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;index"`
	DeviceID         string     `json:"device_id" gorm:"type:varchar(100);index"`
	RefreshTokenHash string     `json:"-" gorm:"type:varchar(255);not null"`
	IPAddress        string     `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent        string     `json:"user_agent" gorm:"type:text"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt        *time.Time `json:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

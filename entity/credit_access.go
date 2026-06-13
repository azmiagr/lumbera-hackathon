package entity

import (
	"time"

	"github.com/google/uuid"
)

type CreditAccessRequest struct {
	RequestID       uuid.UUID  `json:"request_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID   uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	MemberID        uuid.UUID  `json:"member_id" gorm:"type:varchar(36);not null;index"`
	PartnerID       uuid.UUID  `json:"partner_id" gorm:"type:varchar(36);not null;index"`
	ApplicationID   *uuid.UUID `json:"loan_application_id" gorm:"type:varchar(36);index"`
	RequestedAmount int64      `json:"requested_amount" gorm:"not null"`
	Purpose         string     `json:"purpose" gorm:"type:varchar(255);not null"`
	DataScopeJSON   string     `json:"data_scope_json" gorm:"type:json;not null"`
	Status          string     `json:"status" gorm:"type:enum('PENDING','GRANTED','DECLINED','REVOKED');not null;default:'PENDING';index"`
	RequestedAt     time.Time  `json:"requested_at" gorm:"not null;index"`
	GrantedAt       *time.Time `json:"granted_at"`
	DeclinedAt      *time.Time `json:"declined_at"`
	RevokedAt       *time.Time `json:"revoked_at"`
	AccessExpiresAt *time.Time `json:"access_expires_at"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	MemberDataConsents []MemberDataConsent `json:"member_data_consents" gorm:"foreignKey:RequestID;constraint:onDelete:CASCADE"`
}

type MemberDataConsent struct {
	ConsentID        uuid.UUID  `json:"consent_id" gorm:"type:varchar(36);primaryKey"`
	RequestID        uuid.UUID  `json:"request_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	CooperativeID    uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	MemberID         uuid.UUID  `json:"member_id" gorm:"type:varchar(36);not null;index"`
	PartnerID        uuid.UUID  `json:"partner_id" gorm:"type:varchar(36);not null;index"`
	DataScopeJSON    string     `json:"data_scope_json" gorm:"type:json;not null"`
	DurationDays     int        `json:"duration_days" gorm:"not null"`
	GrantedAt        time.Time  `json:"granted_at" gorm:"not null"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null;index"`
	RevokedAt        *time.Time `json:"revoked_at"`
	IsActive         bool       `json:"is_active" gorm:"not null;default:true;index"`
	ConsentSignature string     `json:"consent_signature" gorm:"type:varchar(512)"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

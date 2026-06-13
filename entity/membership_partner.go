package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserCooperativeMembership struct {
	CooperativeMembershipID uuid.UUID  `json:"cooperative_membership_id" gorm:"type:varchar(36);primaryKey"`
	UserID                  uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;index"`
	CooperativeID           uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	MemberID                *uuid.UUID `json:"member_id" gorm:"type:varchar(36);index"`
	RoleID                  uuid.UUID  `json:"role_id" gorm:"type:varchar(36);not null;index"`
	PositionCode            string     `json:"position_code" gorm:"type:enum('CHAIRMAN','TREASURER','SECRETARY','STAFF')"`
	OfficerCode             string     `json:"officer_code" gorm:"type:varchar(50)"`
	Status                  string     `json:"status" gorm:"type:enum('ACTIVE','INACTIVE','SUSPENDED');default:'ACTIVE'"`
	JoinedAt                time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt                  *time.Time `json:"left_at"`
	CreatedAt               time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type Partner struct {
	PartnerID              uuid.UUID      `json:"partner_id" gorm:"type:varchar(36);primaryKey"`
	Name                   string         `json:"name" gorm:"type:varchar(255)"`
	PartnerType            string         `json:"partner_type" gorm:"type:enum('FINTECH','BANK','REGULATOR','PAYMENT_GATEWAY','OTHER');not null"`
	RegistrationNumber     string         `json:"registration_number" gorm:"type:varchar(100)"`
	OJKRegistrationNumber  string         `json:"ojk_registration_number" gorm:"type:varchar(100)"`
	ContactPersonName      string         `json:"contact_person_name" gorm:"type:varchar(255)"`
	ContactPersonEmail     string         `json:"contact_person_email" gorm:"type:varchar(255)"`
	ContactPersonPhone     string         `json:"contact_person_phone" gorm:"type:varchar(20)"`
	Status                 string         `json:"status" gorm:"type:enum('ACTIVE','PENDING','REJECTED','SUSPENDED','INACTIVE');default:'PENDING'"`
	APIAccessEnabled       bool           `json:"api_access_enabled" gorm:"default:false"`
	DefaultRateLimitPerDay int            `json:"default_rate_limit_per_day" gorm:"default:1000"`
	CreatedAt              time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	MemberDataConsents   []MemberDataConsent   `json:"member_data_consents" gorm:"foreignKey:PartnerID;constraint:onDelete:CASCADE"`
	CreditAccessRequests []CreditAccessRequest `json:"credit_access_requests" gorm:"foreignKey:PartnerID;constraint:onDelete:CASCADE"`
	PartnerUsers         []PartnerUser         `json:"partner_users" gorm:"foreignKey:PartnerID;constraint:onDelete:CASCADE"`
}

type PartnerUser struct {
	PartnerUserID uuid.UUID `json:"partner_user_id" gorm:"type:varchar(36);primaryKey"`
	PartnerID     uuid.UUID `json:"partner_id" gorm:"type:varchar(36)"`
	UserID        uuid.UUID `json:"user_id" gorm:"type:varchar(36)"`
	RoleID        uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	Status        string    `json:"status" gorm:"type:enum('ACTIVE','INACTIVE','SUSPENDED');default:'ACTIVE'"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

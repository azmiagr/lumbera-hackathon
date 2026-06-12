package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID      uuid.UUID  `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	FullName    string     `json:"full_name" gorm:"type:varchar(255);not null"`
	PhoneNumber string     `json:"phone_number" gorm:"type:varchar(20);uniqueIndex;not null"`
	Status      string     `json:"status" gorm:"type:enum('PHONE_UNVERIFIED','PIN_REQUIRED','ACTIVE','SUSPENDED','INACTIVE');default:'PHONE_UNVERIFIED';not null"`
	UserType    string     `json:"user_type" gorm:"type:enum('PLATFORM','COOPERATIVE','PARTNER','REGULATOR');not null"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	UserIdentity               *UserIdentity               `json:"user_identity" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UserPINCredential          *UserPINCredential          `json:"user_pin_credential" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	Members                    []Member                    `json:"members" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UserSessions               []UserSession               `json:"user_sessions" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	PartnerUsers               []PartnerUser               `json:"partner_users" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UserRoles                  []UserRole                  `json:"user_roles" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UserCooperativeMemberships []UserCooperativeMembership `json:"user_cooperative_memberships" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
}

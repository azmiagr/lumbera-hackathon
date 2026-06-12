package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"type:varchar(36);primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null"`
	Code         string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	ScopeType    string    `json:"scope_type" gorm:"type:enum('PLATFORM','COOPERATIVE','PARTNER','REGULATOR');not null"`
	IsSystemRole bool      `json:"is_system_role" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	PartnerUsers               []PartnerUser               `json:"partner_users" gorm:"foreignKey:RoleID;constraint:onDelete:CASCADE"`
	RolePermissions            []RolePermission            `json:"role_permissions" gorm:"foreignKey:RoleID;constraint:onDelete:CASCADE"`
	UserRoles                  []UserRole                  `json:"user_roles" gorm:"foreignKey:RoleID;constraint:onDelete:CASCADE"`
	UserCooperativeMemberships []UserCooperativeMembership `json:"user_cooperative_memberships" gorm:"foreignKey:RoleID;constraint:onDelete:CASCADE"`
}

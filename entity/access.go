package entity

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:varchar(36);primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null"`
	Code         string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	RolePermissions []RolePermission `json:"role_permissions" gorm:"foreignKey:PermissionID;constraint:onDelete:CASCADE"`
}

type RolePermission struct {
	RolePermissionID uuid.UUID `json:"role_permission_id" gorm:"type:varchar(36);primaryKey"`
	RoleID           uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	PermissionID     uuid.UUID `json:"permission_id" gorm:"type:varchar(36)"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserRole struct {
	UserRoleID uuid.UUID  `json:"user_role_id" gorm:"type:varchar(36);primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;index"`
	RoleID     uuid.UUID  `json:"role_id" gorm:"type:varchar(36);not null;index"`
	AssignedBy *uuid.UUID `json:"assigned_by" gorm:"type:varchar(36);index"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"autoCreateTime"`
	RevokedAt  *time.Time `json:"revoked_at"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
}

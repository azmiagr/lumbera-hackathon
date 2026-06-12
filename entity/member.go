package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Member struct {
	MemberID           uuid.UUID      `json:"member_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID      uuid.UUID      `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_cooperative_member_number"`
	UserID             uuid.UUID      `json:"user_id" gorm:"type:varchar(36);not null;index"`
	MemberNumber       string         `json:"member_number" gorm:"type:varchar(50);not null;uniqueIndex:idx_cooperative_member_number"`
	JoinedDate         *time.Time     `json:"joined_date" gorm:"type:date"`
	MemberStatus       string         `json:"member_status" gorm:"type:enum('ACTIVE','INACTIVE','SUSPENDED','RESIGNED');default:'ACTIVE'"`
	CurrentMCSScore    *int           `json:"current_mcs_score"`
	MCSGrade           string         `json:"mcs_grade" gorm:"type:enum('AA','A','B','C','D')"`
	LastScoreUpdatedAt *time.Time     `json:"last_score_updated_at"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserCooperativeMemberships []UserCooperativeMembership `json:"user_cooperative_memberships" gorm:"foreignKey:MemberID;constraint:onDelete:CASCADE"`
}

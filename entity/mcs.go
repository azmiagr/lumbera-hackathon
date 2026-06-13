package entity

import (
	"time"

	"github.com/google/uuid"
)

type MCSScoreSnapshot struct {
	SnapshotID             uuid.UUID `json:"snapshot_id" gorm:"type:varchar(36);primaryKey"`
	RequestID              uuid.UUID `json:"request_id" gorm:"type:varchar(36);not null;index"`
	CooperativeID          uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index"`
	MemberID               uuid.UUID `json:"member_id" gorm:"type:varchar(36);not null;index"`
	MCSScore               int       `json:"mcs_score" gorm:"not null"`
	MCSGrade               string    `json:"mcs_grade" gorm:"type:enum('AA','A','B','C','D');not null"`
	Eligible               bool      `json:"eligible" gorm:"not null;default:false"`
	EligibilityProbability float64   `json:"eligibility_probability" gorm:"type:decimal(7,6);not null;default:0"`
	ModelVersion           string    `json:"model_version" gorm:"type:varchar(50);not null"`
	CalculationStatus      string    `json:"calculation_status" gorm:"type:enum('COMPLETE','FAILED');not null"`
	Explanation            string    `json:"explanation" gorm:"type:text"`
	CalculatedAt           time.Time `json:"calculated_at" gorm:"not null;index"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
}

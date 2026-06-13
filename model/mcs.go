package model

import (
	"time"

	"github.com/google/uuid"
)

type TriggerMemberMCSRequest struct {
	AuthContext
	MemberID uuid.UUID `json:"-"`
}

type TriggerMemberMCSResponse struct {
	RequestID    uuid.UUID `json:"request_id"`
	MemberID     uuid.UUID `json:"member_id"`
	Status       string    `json:"status"`
	WebhookState string    `json:"webhook_state"`
}

type MCSFeaturePayload struct {
	RequestID      uuid.UUID      `json:"request_id"`
	CooperativeID  uuid.UUID      `json:"cooperative_id"`
	MemberID       uuid.UUID      `json:"member_id"`
	FeatureVersion string         `json:"feature_version"`
	ModelVersion   string         `json:"model_version"`
	Features       map[string]any `json:"features"`
}

type MCSCallbackRequest struct {
	RequestID              uuid.UUID `json:"request_id"`
	MemberID               uuid.UUID `json:"member_id"`
	MCSScore               int       `json:"mcs_score"`
	MCSGrade               string    `json:"mcs_grade"`
	Eligible               bool      `json:"eligible"`
	EligibilityProbability float64   `json:"eligibility_probability"`
	ModelVersion           string    `json:"model_version"`
	CalculationStatus      string    `json:"calculation_status"`
	Explanation            string    `json:"explanation"`
}

type MCSCallbackResponse struct {
	MemberID           uuid.UUID `json:"member_id"`
	MCSScore           int       `json:"mcs_score"`
	MCSGrade           string    `json:"mcs_grade"`
	LastScoreUpdatedAt time.Time `json:"last_score_updated_at"`
}

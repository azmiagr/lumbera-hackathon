package model

import (
	"time"

	"github.com/google/uuid"
)

type ListCreditAccessRequestsRequest struct {
	AuthContext
}

type GetCreditAccessRequestRequest struct {
	AuthContext
	RequestID uuid.UUID `json:"-"`
}

type GrantCreditAccessRequest struct {
	AuthContext
	RequestID    uuid.UUID `json:"-"`
	DurationDays int       `json:"duration_days"`
}

type DeclineCreditAccessRequest struct {
	AuthContext
	RequestID uuid.UUID `json:"-"`
}

type RevokeCreditAccessRequest struct {
	AuthContext
	RequestID uuid.UUID `json:"-"`
}

type CreditAccessRequestsResponse struct {
	Pending []CreditAccessListItem `json:"pending"`
	Active  []CreditAccessListItem `json:"active"`
	History []CreditAccessListItem `json:"history"`
}

type CreditAccessListItem struct {
	RequestID       uuid.UUID  `json:"request_id"`
	PartnerID       uuid.UUID  `json:"partner_id"`
	PartnerName     string     `json:"partner_name"`
	PartnerType     string     `json:"partner_type"`
	OJKRegistered   bool       `json:"ojk_registered"`
	MCSGrade        string     `json:"mcs_grade"`
	Status          string     `json:"status"`
	StatusLabel     string     `json:"status_label"`
	RequestedAmount int64      `json:"requested_amount"`
	Purpose         string     `json:"purpose"`
	RequestedAt     time.Time  `json:"requested_at"`
	GrantedAt       *time.Time `json:"granted_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DurationDays    int        `json:"duration_days"`
}

type CreditAccessDetailResponse struct {
	CreditAccessListItem
	DataItems        []CreditAccessDataItem `json:"data_items"`
	AllowedDurations []int                  `json:"allowed_durations"`
	Actions          CreditAccessActions    `json:"actions"`
}

type CreditAccessDataItem struct {
	Code     string `json:"code"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Included bool   `json:"included"`
}

type CreditAccessActions struct {
	GrantEnabled   bool `json:"grant_enabled"`
	DeclineEnabled bool `json:"decline_enabled"`
	RevokeEnabled  bool `json:"revoke_enabled"`
}

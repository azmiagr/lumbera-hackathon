package model

import (
	"time"

	"github.com/google/uuid"
)

type GetProfileRequest struct {
	AuthContext
}

type ProfileResponse struct {
	RoleCode string      `json:"role_code"`
	Profile  interface{} `json:"profile"`
}

type ProfileCooperativeResponse struct {
	CooperativeID      uuid.UUID `json:"cooperative_id"`
	Name               string    `json:"name"`
	CooperativeCode    string    `json:"cooperative_code"`
	RegistrationNumber string    `json:"registration_number,omitempty"`
}

type OfficerProfileResponse struct {
	UserID        uuid.UUID                  `json:"user_id"`
	FullName      string                     `json:"full_name"`
	Initials      string                     `json:"initials"`
	AvatarURL     string                     `json:"avatar_url"`
	PhoneNumber   string                     `json:"phone_number"`
	PositionCode  string                     `json:"position_code"`
	PositionLabel string                     `json:"position_label"`
	JoinedAt      time.Time                  `json:"joined_at"`
	Cooperative   ProfileCooperativeResponse `json:"cooperative"`
	CHS           ProfileCHSResponse         `json:"chs"`
}

type ProfileCHSResponse struct {
	Period       string  `json:"period"`
	Status       string  `json:"status"`
	Score        float64 `json:"score"`
	DisplayScore int     `json:"display_score"`
	Grade        string  `json:"grade"`
	Category     string  `json:"category"`
}

type MemberProfileResponse struct {
	UserID       uuid.UUID                  `json:"user_id"`
	MemberID     uuid.UUID                  `json:"member_id"`
	FullName     string                     `json:"full_name"`
	Initials     string                     `json:"initials"`
	AvatarURL    string                     `json:"avatar_url"`
	PhoneNumber  string                     `json:"phone_number"`
	MemberNumber string                     `json:"member_number"`
	JoinedDate   *time.Time                 `json:"joined_date"`
	JoinedYear   int                        `json:"joined_year"`
	Cooperative  ProfileCooperativeResponse `json:"cooperative"`
	MCS          ProfileMCSResponse         `json:"mcs"`
	Loan         ProfileLoanResponse        `json:"loan"`
}

type ProfileMCSResponse struct {
	Score              *int       `json:"score"`
	Grade              string     `json:"grade"`
	LastScoreUpdatedAt *time.Time `json:"last_score_updated_at"`
}

type ProfileLoanResponse struct {
	CompletedCount int    `json:"completed_count"`
	CompletedLabel string `json:"completed_label"`
}

type OfficerProfileRow struct {
	UserID             uuid.UUID
	FullName           string
	PhoneNumber        string
	CooperativeID      uuid.UUID
	CooperativeName    string
	CooperativeCode    string
	RegistrationNumber string
	PositionCode       string
	JoinedAt           time.Time
}

type MemberProfileRow struct {
	UserID             uuid.UUID
	MemberID           uuid.UUID
	FullName           string
	PhoneNumber        string
	MemberNumber       string
	JoinedDate         *time.Time
	CurrentMCSScore    *int
	MCSGrade           string
	LastScoreUpdatedAt *time.Time
	CooperativeID      uuid.UUID
	CooperativeName    string
	CooperativeCode    string
}

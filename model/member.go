package model

import (
	"time"

	"github.com/google/uuid"
)

type ListMembersRequest struct {
	AuthContext
	Search string `form:"search"`
	Grade  string `form:"grade"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type MemberListItemResponse struct {
	MemberID        uuid.UUID  `json:"member_id"`
	UserID          uuid.UUID  `json:"user_id"`
	CooperativeID   uuid.UUID  `json:"cooperative_id"`
	FullName        string     `json:"full_name"`
	Initials        string     `json:"initials"`
	MemberNumber    string     `json:"member_number"`
	JoinedDate      *time.Time `json:"joined_date"`
	MembershipYears int        `json:"membership_years"`
	MemberStatus    string     `json:"member_status"`
	CurrentMCSScore *int       `json:"current_mcs_score"`
	MCSGrade        string     `json:"mcs_grade"`
}

type ListMembersResponse struct {
	Items []MemberListItemResponse `json:"items"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
	Total int64                    `json:"total"`
}

type CreateMemberRequest struct {
	AuthContext
	FullName    string     `json:"full_name"`
	NIK         string     `json:"nik"`
	PhoneNumber string     `json:"phone_number"`
	Address     string     `json:"address"`
	JoinedDate  *time.Time `json:"joined_date"`
}

type CreateMemberResponse struct {
	UserID        uuid.UUID  `json:"user_id"`
	MemberID      uuid.UUID  `json:"member_id"`
	CooperativeID uuid.UUID  `json:"cooperative_id"`
	FullName      string     `json:"full_name"`
	PhoneNumber   string     `json:"phone_number"`
	MemberNumber  string     `json:"member_number"`
	JoinedDate    *time.Time `json:"joined_date"`
	MemberStatus  string     `json:"member_status"`
	AccountStatus string     `json:"account_status"`
}

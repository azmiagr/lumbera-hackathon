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

type GetMemberDashboardRequest struct {
	AuthContext
	RecentLimit int `form:"recent_limit"`
}

type MemberDashboardResponse struct {
	Profile            MemberDashboardProfile       `json:"profile"`
	Savings            MemberDashboardSavings       `json:"savings"`
	MCS                MemberDashboardMCS           `json:"mcs"`
	RecentTransactions []MemberDashboardTransaction `json:"recent_transactions"`
}

type MemberDashboardProfile struct {
	MemberID        uuid.UUID `json:"member_id"`
	UserID          uuid.UUID `json:"user_id"`
	FullName        string    `json:"full_name"`
	MemberNumber    string    `json:"member_number"`
	CooperativeID   uuid.UUID `json:"cooperative_id"`
	CooperativeName string    `json:"cooperative_name"`
}

type MemberDashboardSavings struct {
	TotalBalance        int64 `json:"total_balance"`
	PrincipalBalance    int64 `json:"principal_balance"`
	MandatoryBalance    int64 `json:"mandatory_balance"`
	VoluntaryBalance    int64 `json:"voluntary_balance"`
	CashWithdrawalTotal int64 `json:"cash_withdrawal_total"`
}

type MemberDashboardMCS struct {
	Score              *int       `json:"score"`
	Grade              string     `json:"grade"`
	Label              string     `json:"label"`
	Status             string     `json:"status"`
	LastScoreUpdatedAt *time.Time `json:"last_score_updated_at"`
}

type MemberDashboardTransaction struct {
	TransactionID        uuid.UUID `json:"transaction_id"`
	TransactionType      string    `json:"transaction_type"`
	TransactionTypeLabel string    `json:"transaction_type_label"`
	Direction            string    `json:"direction"`
	Amount               int64     `json:"amount"`
	SignedAmount         int64     `json:"signed_amount"`
	Description          string    `json:"description"`
	RecordedAt           time.Time `json:"recorded_at"`
}

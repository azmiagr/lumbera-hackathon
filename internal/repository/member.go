package repository

import (
	"strings"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMemberRepository interface {
	GetActiveMember(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.Member, error)
	SearchTransactionMembers(tx *gorm.DB, req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error)
	ListMembers(tx *gorm.DB, req model.ListMembersRequest) ([]model.MemberListItemResponse, int64, error)
	CountMembersByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error)
	CreateMember(tx *gorm.DB, member *entity.Member) error
	CountActiveMembersByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error)
	GetDashboardProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.MemberDashboardProfile, error)
	GetDashboardSavings(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*model.MemberDashboardSavings, error)
	ListRecentMemberTransactions(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, limit int) ([]model.MemberDashboardTransaction, error)
}

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) IMemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) GetActiveMember(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.Member, error) {
	var member entity.Member

	query := tx.Debug().
		Where("member_id = ?", memberID).
		Where("member_status = ?", "ACTIVE")

	if cooperativeID != uuid.Nil {
		query = query.Where("cooperative_id = ?", cooperativeID)
	}

	err := query.First(&member).Error
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *MemberRepository) SearchTransactionMembers(tx *gorm.DB, req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error) {
	var members []model.TransactionMemberResponse

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	query := tx.Debug().
		Table("members").
		Select(`
			members.member_id,
			members.user_id,
			users.full_name,
			members.member_number,
			members.cooperative_id,
			members.mcs_grade,
			COALESCE(SUM(CASE
				WHEN transactions.transaction_type IN ? THEN transactions.amount
				WHEN transactions.transaction_type = ? THEN -transactions.amount
				ELSE 0
			END), 0) AS savings_balance,
			COALESCE(SUM(CASE
				WHEN transactions.transaction_type = ? THEN transactions.amount
				WHEN transactions.transaction_type = ? THEN -transactions.amount
				ELSE 0
			END), 0) AS loan_outstanding
		`,
			[]string{
				constants.TransactionTypeSavingsPrincipal,
				constants.TransactionTypeSavingsMandatory,
				constants.TransactionTypeSavingsVoluntary,
			},
			constants.TransactionTypeCashWithdrawal,
			constants.TransactionTypeLoan,
			constants.TransactionTypeInstallment,
		).
		Joins("JOIN users ON users.user_id = members.user_id").
		Joins("LEFT JOIN transactions ON transactions.member_id = members.member_id AND transactions.cooperative_id = members.cooperative_id").
		Where("members.cooperative_id = ?", req.CooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Where("users.status IN ?", []string{"ACTIVE", "PIN_REQUIRED"}).
		Group("members.member_id").
		Order("users.full_name ASC").
		Limit(limit)

	search := strings.TrimSpace(req.Search)
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where("(users.full_name LIKE ? OR members.member_number LIKE ?)", keyword, keyword)
	}

	if err := query.Scan(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

func (r *MemberRepository) ListMembers(tx *gorm.DB, req model.ListMembersRequest) ([]model.MemberListItemResponse, int64, error) {
	var members []model.MemberListItemResponse
	var total int64

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	query := tx.Debug().
		Table("members").
		Select(`
			members.member_id,
			members.user_id,
			members.cooperative_id,
			users.full_name,
			members.member_number,
			members.joined_date,
			members.member_status,
			members.current_m_csscore,
			members.mcs_grade
		`).
		Joins("JOIN users ON users.user_id = members.user_id").
		Where("members.cooperative_id = ?", req.CooperativeID).
		Where("users.status IN ?", []string{"ACTIVE", "PIN_REQUIRED"})

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "ACTIVE"
	}
	query = query.Where("members.member_status = ?", status)

	grade := strings.ToUpper(strings.TrimSpace(req.Grade))
	if grade != "" && grade != "SEMUA" {
		query = query.Where("members.mcs_grade = ?", grade)
	}

	search := strings.TrimSpace(req.Search)
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"(users.full_name LIKE ? OR members.member_number LIKE ?)",
			keyword,
			keyword,
		)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("users.full_name ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

func (r *MemberRepository) CreateMember(tx *gorm.DB, member *entity.Member) error {
	err := tx.Debug().Create(member).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberRepository) CountMembersByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error) {
	var total int64
	err := tx.Debug().
		Model(&entity.Member{}).
		Where("cooperative_id = ?", cooperativeID).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *MemberRepository) CountActiveMembersByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error) {
	var total int64
	err := tx.Debug().
		Table("members").
		Joins("JOIN users ON users.user_id = members.user_id").
		Where("members.cooperative_id = ?", cooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Where("users.status IN ?", []string{"ACTIVE", "PIN_REQUIRED"}).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *MemberRepository) GetDashboardProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.MemberDashboardProfile, error) {
	var result model.MemberDashboardProfile

	err := tx.Debug().
		Table("members").
		Select(`
			members.member_id,
			members.user_id,
			users.full_name,
			members.member_number,
			members.cooperative_id,
			cooperatives.name AS cooperative_name
		`).
		Joins("JOIN users ON users.user_id = members.user_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = members.cooperative_id").
		Joins("JOIN user_cooperative_memberships ON user_cooperative_memberships.member_id = members.member_id").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Where("members.user_id = ?", userID).
		Where("members.cooperative_id = ?", cooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("roles.code = ?", constants.RoleCodeAnggota).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *MemberRepository) GetDashboardSavings(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*model.MemberDashboardSavings, error) {
	var result model.MemberDashboardSavings

	err := tx.Debug().
		Table("transactions").
		Select(`
			COALESCE(SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END), 0) AS principal_balance,
			COALESCE(SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END), 0) AS mandatory_balance,
			COALESCE(SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END), 0) AS voluntary_balance,
			COALESCE(SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END), 0) AS cash_withdrawal_total,
			COALESCE(SUM(CASE
				WHEN transaction_type IN ? THEN amount
				WHEN transaction_type = ? THEN -amount
				ELSE 0
			END), 0) AS total_balance
		`,
			constants.TransactionTypeSavingsPrincipal,
			constants.TransactionTypeSavingsMandatory,
			constants.TransactionTypeSavingsVoluntary,
			constants.TransactionTypeCashWithdrawal,
			[]string{
				constants.TransactionTypeSavingsPrincipal,
				constants.TransactionTypeSavingsMandatory,
				constants.TransactionTypeSavingsVoluntary,
			},
			constants.TransactionTypeCashWithdrawal,
		).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *MemberRepository) ListRecentMemberTransactions(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, limit int) ([]model.MemberDashboardTransaction, error) {
	var results []model.MemberDashboardTransaction

	if limit <= 0 || limit > 20 {
		limit = 6
	}

	err := tx.Debug().
		Table("transactions").
		Select(`
			transaction_id,
			transaction_type,
			amount,
			description,
			recorded_at
		`).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Order("recorded_at DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

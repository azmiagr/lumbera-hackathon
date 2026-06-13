package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILoanDashboardRepository interface {
	GetMemberProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.MemberDashboardProfile, error)
	GetActiveLoan(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanAccount, error)
	GetLoanDashboardAggregate(tx *gorm.DB, loanID uuid.UUID) (*LoanDashboardAggregate, error)
	ListInstallmentSchedules(tx *gorm.DB, loanID uuid.UUID, limit int) ([]entity.LoanInstallmentSchedule, int64, error)
	ListLoanHistory(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, limit int) ([]LoanDashboardHistoryRow, error)
}

type LoanDashboardRepository struct {
	db *gorm.DB
}

type LoanDashboardAggregate struct {
	RemainingPayableAmount int64
	PaidInstallmentCount   int
	PaidAmount             int64
	NextDueDate            *time.Time
}

type LoanDashboardHistoryRow struct {
	LoanID          uuid.UUID
	LoanNumber      string
	PrincipalAmount int64
	TermMonths      int
	Status          string
	DisbursedAt     time.Time
	PaidAt          *time.Time
}

func NewLoanDashboardRepository(db *gorm.DB) ILoanDashboardRepository {
	return &LoanDashboardRepository{db: db}
}

func (r *LoanDashboardRepository) GetMemberProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.MemberDashboardProfile, error) {
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

func (r *LoanDashboardRepository) GetActiveLoan(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanAccount, error) {
	var loan entity.LoanAccount
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("status = ?", "ACTIVE").
		Order("disbursed_at DESC").
		First(&loan).Error
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (r *LoanDashboardRepository) GetLoanDashboardAggregate(tx *gorm.DB, loanID uuid.UUID) (*LoanDashboardAggregate, error) {
	var aggregate LoanDashboardAggregate
	err := tx.Debug().
		Table("loan_installment_schedules").
		Select(`
			COALESCE(SUM(remaining_amount), 0) AS remaining_payable_amount,
			COALESCE(SUM(paid_amount), 0) AS paid_amount,
			COALESCE(SUM(CASE WHEN status = 'PAID' THEN 1 ELSE 0 END), 0) AS paid_installment_count,
			MIN(CASE WHEN status IN ('UNPAID','PARTIAL') THEN due_date ELSE NULL END) AS next_due_date
		`).
		Where("loan_id = ?", loanID).
		Scan(&aggregate).Error
	if err != nil {
		return nil, err
	}

	return &aggregate, nil
}

func (r *LoanDashboardRepository) ListInstallmentSchedules(tx *gorm.DB, loanID uuid.UUID, limit int) ([]entity.LoanInstallmentSchedule, int64, error) {
	var schedules []entity.LoanInstallmentSchedule
	var total int64

	base := tx.Debug().
		Model(&entity.LoanInstallmentSchedule{}).
		Where("loan_id = ?", loanID)

	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 3
	}

	err := base.
		Order("installment_no ASC").
		Limit(limit).
		Find(&schedules).Error
	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *LoanDashboardRepository) ListLoanHistory(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, limit int) ([]LoanDashboardHistoryRow, error) {
	var rows []LoanDashboardHistoryRow

	if limit <= 0 {
		limit = 3
	}

	err := tx.Debug().
		Table("loan_accounts").
		Select(`
			loan_accounts.loan_id,
			loan_accounts.loan_number,
			loan_accounts.principal_amount,
			loan_accounts.term_months,
			loan_accounts.status,
			loan_accounts.disbursed_at,
			MAX(loan_installment_schedules.paid_at) AS paid_at
		`).
		Joins("LEFT JOIN loan_installment_schedules ON loan_installment_schedules.loan_id = loan_accounts.loan_id").
		Where("loan_accounts.cooperative_id = ?", cooperativeID).
		Where("loan_accounts.member_id = ?", memberID).
		Group(`
			loan_accounts.loan_id,
			loan_accounts.loan_number,
			loan_accounts.principal_amount,
			loan_accounts.term_months,
			loan_accounts.status,
			loan_accounts.disbursed_at
		`).
		Order("loan_accounts.disbursed_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

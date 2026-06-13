package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ISavingsBookRepository interface {
	GetProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.SavingsBookProfile, error)
	GetSummary(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, periodEnd time.Time) (*model.SavingsBookSummary, error)
	ListItems(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, startDate time.Time, endDate time.Time, filterType string, page int, limit int) ([]model.SavingsBookItem, int64, error)
}

type SavingsBookRepository struct {
	db *gorm.DB
}

func NewSavingsBookRepository(db *gorm.DB) ISavingsBookRepository {
	return &SavingsBookRepository{db: db}
}

func (r *SavingsBookRepository) GetProfile(tx *gorm.DB, userID uuid.UUID, cooperativeID uuid.UUID) (*model.SavingsBookProfile, error) {
	var result model.SavingsBookProfile

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

func (r *SavingsBookRepository) GetSummary(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, periodEnd time.Time) (*model.SavingsBookSummary, error) {
	var result model.SavingsBookSummary

	err := tx.Debug().
		Table("transactions").
		Select(`
			COALESCE(SUM(CASE
				WHEN transaction_type IN ? THEN amount
				WHEN transaction_type IN ? THEN -amount
				ELSE 0
			END), 0) AS total_balance,
			COALESCE(SUM(CASE WHEN transaction_type IN ? THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN transaction_type IN ? THEN amount ELSE 0 END), 0) AS total_expense
		`,
			incomeTransactionTypes(),
			expenseTransactionTypes(),
			incomeTransactionTypes(),
			expenseTransactionTypes(),
		).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("recorded_at <= ?", periodEnd).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *SavingsBookRepository) ListItems(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, startDate time.Time, endDate time.Time, filterType string, page int, limit int) ([]model.SavingsBookItem, int64, error) {
	var items []model.SavingsBookItem
	var total int64

	query := tx.Debug().
		Table("transactions").
		Select(`
			transactions.transaction_id,
			transactions.transaction_type,
			transactions.amount,
			transactions.description,
			transactions.recorded_at,
			COALESCE(officer_users.full_name, '') AS recorder_name
		`).
		Joins("LEFT JOIN users AS officer_users ON officer_users.user_id = transactions.officer_id").
		Where("transactions.cooperative_id = ?", cooperativeID).
		Where("transactions.member_id = ?", memberID).
		Where("transactions.recorded_at BETWEEN ? AND ?", startDate, endDate).
		Where("transactions.transaction_type IN ?", savingsBookTransactionTypes())

	switch filterType {
	case model.SavingsBookTypeIncome:
		query = query.Where("transactions.transaction_type IN ?", incomeTransactionTypes())
	case model.SavingsBookTypeExpense:
		query = query.Where("transactions.transaction_type IN ?", expenseTransactionTypes())
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("transactions.recorded_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func savingsBookTransactionTypes() []string {
	return append(incomeTransactionTypes(), expenseTransactionTypes()...)
}

func incomeTransactionTypes() []string {
	return []string{
		constants.TransactionTypeSavingsPrincipal,
		constants.TransactionTypeSavingsMandatory,
		constants.TransactionTypeSavingsVoluntary,
		constants.TransactionTypeLoan,
	}
}

func expenseTransactionTypes() []string {
	return []string{
		constants.TransactionTypeCashWithdrawal,
		constants.TransactionTypeInstallment,
	}
}

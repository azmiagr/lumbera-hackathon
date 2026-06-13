package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICHSRepository interface {
	GetLoanRiskMetrics(tx *gorm.DB, cooperativeID uuid.UUID, periodEnd time.Time) (*model.CHSLoanRiskMetrics, error)
	GetOnTimePaymentMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSOnTimePaymentMetrics, error)
	GetMemberActivityMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSMemberActivityMetrics, error)
	GetTransactionGrowthMetrics(tx *gorm.DB, cooperativeID uuid.UUID, currentStart, currentEnd, previousStart, previousEnd time.Time) (*model.CHSTransactionGrowthMetrics, error)
	GetDataCompletenessMetrics(tx *gorm.DB, cooperativeID uuid.UUID) (*model.CHSDataCompletenessMetrics, error)
	GetSyncTimelinessMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSSyncTimelinessMetrics, error)
	GetConsistencyMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSConsistencyMetrics, error)
	ListLedgerTransactions(tx *gorm.DB, cooperativeID uuid.UUID, periodEnd time.Time) ([]entity.Transaction, error)
}

type CHSRepository struct {
	db *gorm.DB
}

func NewCHSRepository(db *gorm.DB) ICHSRepository {
	return &CHSRepository{db: db}
}

func (r *CHSRepository) GetLoanRiskMetrics(tx *gorm.DB, cooperativeID uuid.UUID, periodEnd time.Time) (*model.CHSLoanRiskMetrics, error) {
	var result model.CHSLoanRiskMetrics
	err := tx.Debug().
		Table("loan_installment_schedules").
		Select(`
			COALESCE(SUM(loan_installment_schedules.remaining_amount), 0) AS total_remaining_principal,
			COALESCE(SUM(CASE
				WHEN loan_installment_schedules.due_date <= ? AND loan_installment_schedules.remaining_amount > 0
				THEN loan_installment_schedules.remaining_amount
				ELSE 0
			END), 0) AS bad_remaining_principal
		`, periodEnd).
		Joins("JOIN loan_accounts ON loan_accounts.loan_id = loan_installment_schedules.loan_id").
		Where("loan_accounts.cooperative_id = ?", cooperativeID).
		Where("loan_accounts.status = ?", "ACTIVE").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetOnTimePaymentMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSOnTimePaymentMetrics, error) {
	var result model.CHSOnTimePaymentMetrics
	err := tx.Debug().
		Table("loan_installment_schedules").
		Select(`
			COUNT(*) AS total_due,
			COALESCE(SUM(CASE
				WHEN loan_installment_schedules.status = 'PAID'
					AND loan_installment_schedules.paid_at IS NOT NULL
					AND loan_installment_schedules.paid_at <= TIMESTAMP(loan_installment_schedules.due_date, '23:59:59')
				THEN 1
				ELSE 0
			END), 0) AS on_time
		`).
		Joins("JOIN loan_accounts ON loan_accounts.loan_id = loan_installment_schedules.loan_id").
		Where("loan_accounts.cooperative_id = ?", cooperativeID).
		Where("loan_installment_schedules.due_date BETWEEN ? AND ?", startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetMemberActivityMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSMemberActivityMetrics, error) {
	var result model.CHSMemberActivityMetrics

	if err := tx.Debug().
		Table("members").
		Where("cooperative_id = ?", cooperativeID).
		Where("member_status = ?", "ACTIVE").
		Count(&result.TotalMembers).Error; err != nil {
		return nil, err
	}

	if err := tx.Debug().
		Table("transactions").
		Select("COUNT(DISTINCT member_id)").
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&result.ActiveMembers).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetTransactionGrowthMetrics(tx *gorm.DB, cooperativeID uuid.UUID, currentStart, currentEnd, previousStart, previousEnd time.Time) (*model.CHSTransactionGrowthMetrics, error) {
	var result model.CHSTransactionGrowthMetrics

	if err := tx.Debug().
		Table("transactions").
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", currentStart, currentEnd).
		Count(&result.CurrentTransactions).Error; err != nil {
		return nil, err
	}

	if err := tx.Debug().
		Table("transactions").
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", previousStart, previousEnd).
		Count(&result.PreviousTransactions).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetDataCompletenessMetrics(tx *gorm.DB, cooperativeID uuid.UUID) (*model.CHSDataCompletenessMetrics, error) {
	var result model.CHSDataCompletenessMetrics
	err := tx.Debug().
		Table("members").
		Select(`
			COUNT(*) * 6 AS total_fields,
			COALESCE(SUM(
				CASE WHEN users.full_name <> '' THEN 1 ELSE 0 END +
				CASE WHEN users.phone_number <> '' THEN 1 ELSE 0 END +
				CASE WHEN members.member_number <> '' THEN 1 ELSE 0 END +
				CASE WHEN members.joined_date IS NOT NULL THEN 1 ELSE 0 END +
				CASE WHEN user_identities.nik_hash <> '' THEN 1 ELSE 0 END +
				CASE WHEN user_identities.address <> '' THEN 1 ELSE 0 END
			), 0) AS filled_fields
		`).
		Joins("JOIN users ON users.user_id = members.user_id").
		Joins("LEFT JOIN user_identities ON user_identities.user_id = users.user_id").
		Where("members.cooperative_id = ?", cooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetSyncTimelinessMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSSyncTimelinessMetrics, error) {
	var result model.CHSSyncTimelinessMetrics
	err := tx.Debug().
		Table("transactions").
		Select(`
			COUNT(*) AS total_transactions,
			COALESCE(SUM(CASE
				WHEN synced_at IS NOT NULL AND TIMESTAMPDIFF(SECOND, recorded_at, synced_at) <= 30
				THEN 1
				ELSE 0
			END), 0) AS timely_transactions
		`).
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CHSRepository) GetConsistencyMetrics(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) (*model.CHSConsistencyMetrics, error) {
	var result model.CHSConsistencyMetrics

	if err := tx.Debug().
		Table("transactions").
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", startDate, endDate).
		Count(&result.TotalRecords).Error; err != nil {
		return nil, err
	}

	var duplicateRows []struct {
		ClientTransactionID string
		Total               int64
	}
	if err := tx.Debug().
		Table("transactions").
		Select("client_transaction_id, COUNT(*) AS total").
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", startDate, endDate).
		Where("client_transaction_id <> ''").
		Group("client_transaction_id").
		Having("COUNT(*) > 1").
		Scan(&duplicateRows).Error; err != nil {
		return nil, err
	}

	for _, row := range duplicateRows {
		result.DuplicateRecords += row.Total
	}
	result.ConsistentRecords = result.TotalRecords - result.DuplicateRecords
	if result.ConsistentRecords < 0 {
		result.ConsistentRecords = 0
	}

	return &result, nil
}

func (r *CHSRepository) ListLedgerTransactions(tx *gorm.DB, cooperativeID uuid.UUID, periodEnd time.Time) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at <= ?", periodEnd).
		Order("created_at ASC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

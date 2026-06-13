package repository

import (
	"strings"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ITransactionRepository interface {
	CreateTransaction(tx *gorm.DB, transaction *entity.Transaction) error
	GetLatestTransactionForUpdate(tx *gorm.DB, cooperativeID uuid.UUID) (*entity.Transaction, error)
	GetTransactionByClientID(tx *gorm.DB, cooperativeID uuid.UUID, clientTransactionID string) (*entity.Transaction, error)
	GetTransactionDetail(tx *gorm.DB, cooperativeID uuid.UUID, transactionID uuid.UUID) (*model.TransactionListItemResponse, error)
	ListTransactions(tx *gorm.DB, req model.ListTransactionsRequest) ([]model.TransactionListItemResponse, int64, error)
	GetMemberTransactionSummary(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*model.TransactionMemberSummaryResponse, error)
	GetTransactionForUpdate(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.Transaction, error)
	GetReversalByOriginalTransactionID(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.TransactionReversal, error)
	GetReversalByReversalTransactionID(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.TransactionReversal, error)
	CreateTransactionReversal(tx *gorm.DB, reversal *entity.TransactionReversal) error
}

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) ITransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(tx *gorm.DB, transaction *entity.Transaction) error {
	err := tx.Debug().Create(transaction).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *TransactionRepository) GetLatestTransactionForUpdate(tx *gorm.DB, cooperativeID uuid.UUID) (*entity.Transaction, error) {
	var transaction entity.Transaction

	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Order("created_at DESC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *TransactionRepository) GetTransactionByClientID(tx *gorm.DB, cooperativeID uuid.UUID, clientTransactionID string) (*entity.Transaction, error) {
	var transaction entity.Transaction

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("client_transaction_id = ?", clientTransactionID).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *TransactionRepository) GetTransactionDetail(tx *gorm.DB, cooperativeID uuid.UUID, transactionID uuid.UUID) (*model.TransactionListItemResponse, error) {
	var result model.TransactionListItemResponse

	err := baseTransactionListQuery(tx).
		Where("transactions.cooperative_id = ?", cooperativeID).
		Where("transactions.transaction_id = ?", transactionID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *TransactionRepository) ListTransactions(tx *gorm.DB, req model.ListTransactionsRequest) ([]model.TransactionListItemResponse, int64, error) {
	var results []model.TransactionListItemResponse
	var total int64

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	query := baseTransactionListQuery(tx).
		Where("transactions.cooperative_id = ?", req.CooperativeID)

	query = applyTransactionTypeFilter(query, req.Type)

	search := strings.TrimSpace(req.Search)
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"(member_users.full_name LIKE ? OR members.member_number LIKE ? OR transactions.description LIKE ?)",
			keyword,
			keyword,
			keyword,
		)
	}

	countQuery := query.Session(&gorm.Session{})
	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Order("transactions.recorded_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *TransactionRepository) GetMemberTransactionSummary(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*model.TransactionMemberSummaryResponse, error) {
	var result model.TransactionMemberSummaryResponse

	err := tx.Debug().
		Table("transactions").
		Select(`
			COALESCE(SUM(CASE
				WHEN transaction_type IN ? THEN amount
				WHEN transaction_type = ? THEN -amount
				ELSE 0
			END), 0) AS savings_balance,
			COALESCE(SUM(CASE
				WHEN transaction_type = ? THEN amount
				WHEN transaction_type = ? THEN -amount
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
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *TransactionRepository) GetTransactionForUpdate(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Where("transaction_id = ?", transactionID).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetReversalByOriginalTransactionID(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.TransactionReversal, error) {
	var reversal entity.TransactionReversal
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("original_transaction_id = ?", transactionID).
		First(&reversal).Error
	if err != nil {
		return nil, err
	}
	return &reversal, nil
}

func (r *TransactionRepository) GetReversalByReversalTransactionID(tx *gorm.DB, cooperativeID, transactionID uuid.UUID) (*entity.TransactionReversal, error) {
	var reversal entity.TransactionReversal
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("reversal_transaction_id = ?", transactionID).
		First(&reversal).Error
	if err != nil {
		return nil, err
	}
	return &reversal, nil
}

func (r *TransactionRepository) CreateTransactionReversal(tx *gorm.DB, reversal *entity.TransactionReversal) error {
	err := tx.Debug().Create(reversal).Error
	if err != nil {
		return err
	}

	return nil
}

func baseTransactionListQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("transactions").
		Select(`
			transactions.transaction_id,
			transactions.cooperative_id,
			transactions.member_id,
			member_users.full_name AS member_name,
			members.member_number,
			members.mcs_grade AS member_mcs_grade,
			transactions.officer_id,
			officer_users.full_name AS officer_name,
			transactions.transaction_type,
			transactions.amount,
			transactions.description,
			transactions.recorded_at,
			transactions.synced_at,
			transactions.current_hash,
			transactions.is_offline_created,
			transactions.client_transaction_id,
			CASE WHEN original_reversals.transaction_reversal_id IS NULL THEN false ELSE true END AS is_reversed,
			CASE WHEN reversal_reversals.transaction_reversal_id IS NULL THEN false ELSE true END AS is_reversal,
			reversal_reversals.original_transaction_id,
			original_reversals.reversal_transaction_id,
			COALESCE(original_reversals.reason, reversal_reversals.reason, '') AS reversal_reason
		`).
		Joins("JOIN members ON members.member_id = transactions.member_id").
		Joins("JOIN users AS member_users ON member_users.user_id = members.user_id").
		Joins("JOIN users AS officer_users ON officer_users.user_id = transactions.officer_id").
		Joins("LEFT JOIN transaction_reversals AS original_reversals ON original_reversals.original_transaction_id = transactions.transaction_id").
		Joins("LEFT JOIN transaction_reversals AS reversal_reversals ON reversal_reversals.reversal_transaction_id = transactions.transaction_id")
}

func applyTransactionTypeFilter(query *gorm.DB, transactionType string) *gorm.DB {
	switch strings.ToUpper(strings.TrimSpace(transactionType)) {
	case constants.TransactionGroupSavings:
		return query.Where("transactions.transaction_type IN ?", []string{
			constants.TransactionTypeSavingsPrincipal,
			constants.TransactionTypeSavingsMandatory,
			constants.TransactionTypeSavingsVoluntary,
		})
	case constants.TransactionGroupLoan:
		return query.Where("transactions.transaction_type = ?", constants.TransactionTypeLoan)
	case constants.TransactionGroupInstallment:
		return query.Where("transactions.transaction_type = ?", constants.TransactionTypeInstallment)
	case constants.TransactionGroupCashWithdrawal:
		return query.Where("transactions.transaction_type = ?", constants.TransactionTypeCashWithdrawal)
	default:
		return query
	}
}

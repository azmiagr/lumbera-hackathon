package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILedgerRepository interface {
	ListFinancialLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) ([]entity.Transaction, error)
	ListStockLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) ([]entity.StockMovement, error)
	CountLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) (int64, error)
	GetLatestAnchor(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) (*entity.LedgerAnchor, error)
	GetAnchorByRootHash(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time, merkleRootHash string) (*entity.LedgerAnchor, error)
	CreateAnchor(tx *gorm.DB, anchor *entity.LedgerAnchor) error
	GetCooperativeName(tx *gorm.DB, cooperativeID uuid.UUID) (string, error)
}

type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) ILedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) ListFinancialLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) ([]entity.Transaction, error) {
	var transactions []entity.Transaction

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", periodStart, periodEnd).
		Order("created_at ASC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *LedgerRepository) ListStockLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) ([]entity.StockMovement, error) {
	var movements []entity.StockMovement

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", periodStart, periodEnd).
		Order("created_at ASC").
		Find(&movements).Error
	if err != nil {
		return nil, err
	}

	return movements, nil
}

func (r *LedgerRepository) CountLedgerRows(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) (int64, error) {
	var transactionTotal int64
	err := tx.Debug().
		Model(&entity.Transaction{}).
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", periodStart, periodEnd).
		Count(&transactionTotal).Error
	if err != nil {
		return 0, err
	}

	var stockMovementTotal int64
	err = tx.Debug().
		Model(&entity.StockMovement{}).
		Where("cooperative_id = ?", cooperativeID).
		Where("recorded_at BETWEEN ? AND ?", periodStart, periodEnd).
		Count(&stockMovementTotal).Error
	if err != nil {
		return 0, err
	}

	return transactionTotal + stockMovementTotal, nil
}

func (r *LedgerRepository) GetLatestAnchor(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time) (*entity.LedgerAnchor, error) {
	var anchor entity.LedgerAnchor

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("period_start = ?", periodStart).
		Where("period_end = ?", periodEnd).
		Order("anchored_at DESC").
		First(&anchor).Error
	if err != nil {
		return nil, err
	}

	return &anchor, nil
}

func (r *LedgerRepository) GetAnchorByRootHash(tx *gorm.DB, cooperativeID uuid.UUID, periodStart, periodEnd time.Time, merkleRootHash string) (*entity.LedgerAnchor, error) {
	var anchor entity.LedgerAnchor

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("period_start = ?", periodStart).
		Where("period_end = ?", periodEnd).
		Where("merkle_root_hash = ?", merkleRootHash).
		Order("anchored_at DESC").
		First(&anchor).Error
	if err != nil {
		return nil, err
	}

	return &anchor, nil
}

func (r *LedgerRepository) CreateAnchor(tx *gorm.DB, anchor *entity.LedgerAnchor) error {
	err := tx.Debug().Create(anchor).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LedgerRepository) GetCooperativeName(tx *gorm.DB, cooperativeID uuid.UUID) (string, error) {
	var cooperative entity.Cooperative

	err := tx.Debug().
		Select("name").
		Where("cooperative_id = ?", cooperativeID).
		First(&cooperative).Error
	if err != nil {
		return "", err
	}

	return cooperative.Name, nil
}

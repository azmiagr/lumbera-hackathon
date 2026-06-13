package repository

import (
	"strconv"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMCSRepository interface {
	GetLatestTrainingFeatures(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (map[string]any, error)
	CreateScoreSnapshot(tx *gorm.DB, snapshot *entity.MCSScoreSnapshot) error
	UpdateMemberCurrentScore(tx *gorm.DB, memberID uuid.UUID, score int, grade string, updatedAt time.Time) error
}

type MCSRepository struct {
	db *gorm.DB
}

var mcsScoringFeatureColumns = []string{
	"active_loan_count",
	"active_transaction_months_12m",
	"asset_coverage_ratio",
	"asset_data_available",
	"average_days_past_due_12m",
	"average_saving_balance_3m",
	"average_saving_balance_6m",
	"average_sync_delay_seconds_12m",
	"business_sector_code",
	"business_sector_risk_score",
	"capacity_score_rule",
	"capital_score_rule",
	"cash_withdrawal_total_12m",
	"character_score_rule",
	"collateral_data_available",
	"collateral_score_rule",
	"collateral_value",
	"completed_loan_count",
	"conditions_score_rule",
	"occupation_code",
}

func NewMCSRepository(db *gorm.DB) IMCSRepository {
	return &MCSRepository{db: db}
}

func (r *MCSRepository) GetLatestTrainingFeatures(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (map[string]any, error) {
	row := map[string]any{}

	err := tx.Debug().
		Table("mcs_training_samples").
		Select(mcsScoringFeatureColumns).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("sample_status IN ?", []string{"READY", "LABEL_PENDING", "DRAFT"}).
		Order("observation_end_date DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		return nil, err
	}

	return normalizeFeatureMap(row), nil
}

func (r *MCSRepository) CreateScoreSnapshot(tx *gorm.DB, snapshot *entity.MCSScoreSnapshot) error {
	err := tx.Debug().Create(snapshot).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MCSRepository) UpdateMemberCurrentScore(tx *gorm.DB, memberID uuid.UUID, score int, grade string, updatedAt time.Time) error {
	err := tx.Debug().
		Model(&entity.Member{}).
		Where("member_id = ?", memberID).
		Updates(map[string]any{
			"current_m_csscore":     score,
			"mcs_grade":             grade,
			"last_score_updated_at": updatedAt,
		}).Error
	if err != nil {
		return err
	}

	return nil
}

func normalizeFeatureMap(features map[string]any) map[string]any {
	normalized := make(map[string]any, len(features))
	for key, value := range features {
		normalized[key] = normalizeFeatureValue(value)
	}
	return normalized
}

func normalizeFeatureValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		text := string(typed)
		if parsed, err := strconv.ParseFloat(text, 64); err == nil {
			return parsed
		}
		return text
	case time.Time:
		return typed.Format(time.RFC3339)
	default:
		return value
	}
}

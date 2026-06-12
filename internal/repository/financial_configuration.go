package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IFinancialConfigurationRepository interface {
	CreateFinancialConfiguration(tx *gorm.DB, financialConfiguration *entity.FinancialConfiguration) error
	GetFinancialConfiguration(tx *gorm.DB, param model.GetFinancialConfigurationParam) (*entity.FinancialConfiguration, error)
	UpdateFinancialConfiguration(tx *gorm.DB, financialConfiguration *entity.FinancialConfiguration) error
}

type FinancialConfigurationRepository struct {
	db *gorm.DB
}

func NewFinancialConfigurationRepository(db *gorm.DB) IFinancialConfigurationRepository {
	return &FinancialConfigurationRepository{db: db}
}

func (r *FinancialConfigurationRepository) CreateFinancialConfiguration(tx *gorm.DB, financialConfiguration *entity.FinancialConfiguration) error {
	err := tx.Debug().Create(financialConfiguration).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *FinancialConfigurationRepository) GetFinancialConfiguration(tx *gorm.DB, param model.GetFinancialConfigurationParam) (*entity.FinancialConfiguration, error) {
	var financialConfiguration *entity.FinancialConfiguration
	err := tx.Debug().Where(&param).First(&financialConfiguration).Error
	if err != nil {
		return nil, err
	}

	return financialConfiguration, nil
}

func (r *FinancialConfigurationRepository) UpdateFinancialConfiguration(tx *gorm.DB, financialConfiguration *entity.FinancialConfiguration) error {
	err := tx.Debug().Save(financialConfiguration).Error
	if err != nil {
		return err
	}

	return nil
}

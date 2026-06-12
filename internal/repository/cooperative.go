package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type ICooperativeRepository interface {
	CreateCooperative(tx *gorm.DB, cooperative *entity.Cooperative) error
	GetCooperative(tx *gorm.DB, param model.GetCooperativeParam) (*entity.Cooperative, error)
	UpdateCooperative(tx *gorm.DB, cooperative *entity.Cooperative) error
}

type CooperativeRepository struct {
	db *gorm.DB
}

func NewCooperativeRepository(db *gorm.DB) ICooperativeRepository {
	return &CooperativeRepository{db: db}
}

func (r *CooperativeRepository) CreateCooperative(tx *gorm.DB, cooperative *entity.Cooperative) error {
	err := tx.Debug().Create(cooperative).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CooperativeRepository) GetCooperative(tx *gorm.DB, param model.GetCooperativeParam) (*entity.Cooperative, error) {
	var cooperative *entity.Cooperative
	err := tx.Debug().Where(&param).First(&cooperative).Error
	if err != nil {
		return nil, err
	}
	return cooperative, nil
}

func (r *CooperativeRepository) UpdateCooperative(tx *gorm.DB, cooperative *entity.Cooperative) error {
	err := tx.Debug().Save(cooperative).Error
	if err != nil {
		return err
	}
	return nil
}

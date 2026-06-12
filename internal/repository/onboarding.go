package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IOnboardingDraftRepository interface {
	CreateOnboardingDraft(tx *gorm.DB, draft *entity.OnboardingDraft) error
	GetOnboardingDraft(tx *gorm.DB, param model.GetOnboardingDraftParam) (*entity.OnboardingDraft, error)
	UpdateOnboardingDraft(tx *gorm.DB, draft *entity.OnboardingDraft) error
}

type OnboardingDraftRepository struct {
	db *gorm.DB
}

func NewOnboardingDraftRepository(db *gorm.DB) IOnboardingDraftRepository {
	return &OnboardingDraftRepository{db: db}
}

func (r *OnboardingDraftRepository) CreateOnboardingDraft(tx *gorm.DB, draft *entity.OnboardingDraft) error {
	err := tx.Debug().Create(draft).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OnboardingDraftRepository) GetOnboardingDraft(tx *gorm.DB, param model.GetOnboardingDraftParam) (*entity.OnboardingDraft, error) {
	var draft *entity.OnboardingDraft
	err := tx.Debug().Where(&param).First(&draft).Error
	if err != nil {
		return nil, err
	}
	return draft, nil
}

func (r *OnboardingDraftRepository) UpdateOnboardingDraft(tx *gorm.DB, draft *entity.OnboardingDraft) error {
	err := tx.Debug().Save(draft).Error
	if err != nil {
		return err
	}
	return nil
}

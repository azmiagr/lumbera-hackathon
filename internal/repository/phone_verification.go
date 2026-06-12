package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IPhoneVerification interface {
	CreatePhoneVerificationChallenge(tx *gorm.DB, challenge *entity.PhoneVerificationChallenge) error
	GetPhoneVerificationChallenge(tx *gorm.DB, param model.GetPhoneVerificationChallengeParam) (*entity.PhoneVerificationChallenge, error)
	UpdatePhoneVerificationChallenge(tx *gorm.DB, challenge *entity.PhoneVerificationChallenge) error
}

type PhoneVerificationRepository struct {
	db *gorm.DB
}

func NewPhoneVerificationRepository(db *gorm.DB) IPhoneVerification {
	return &PhoneVerificationRepository{db: db}
}

func (r *PhoneVerificationRepository) CreatePhoneVerificationChallenge(tx *gorm.DB, challenge *entity.PhoneVerificationChallenge) error {
	err := tx.Debug().Create(challenge).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PhoneVerificationRepository) GetPhoneVerificationChallenge(tx *gorm.DB, param model.GetPhoneVerificationChallengeParam) (*entity.PhoneVerificationChallenge, error) {
	var challenge *entity.PhoneVerificationChallenge
	err := tx.Debug().Where(&param).First(&challenge).Error
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

func (r *PhoneVerificationRepository) UpdatePhoneVerificationChallenge(tx *gorm.DB, challenge *entity.PhoneVerificationChallenge) error {
	err := tx.Debug().Save(challenge).Error
	if err != nil {
		return err
	}

	return nil
}

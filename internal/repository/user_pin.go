package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IUserPinRepository interface {
	CreateUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error
	GetUserPinCredential(tx *gorm.DB, param model.GetUserPINCredentialParam) (*entity.UserPINCredential, error)
	UpdateUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error
	DeleteUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error
}

type UserPinRepository struct {
	db *gorm.DB
}

func NewUserPinRepository(db *gorm.DB) IUserPinRepository {
	return &UserPinRepository{db: db}
}

func (r *UserPinRepository) CreateUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error {
	err := tx.Debug().Create(userPin).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserPinRepository) GetUserPinCredential(tx *gorm.DB, param model.GetUserPINCredentialParam) (*entity.UserPINCredential, error) {
	var userPin *entity.UserPINCredential
	err := tx.Debug().Where(&param).First(&userPin).Error
	if err != nil {
		return nil, err
	}

	return userPin, nil
}

func (r *UserPinRepository) UpdateUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error {
	err := tx.Debug().Save(userPin).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserPinRepository) DeleteUserPin(tx *gorm.DB, userPin *entity.UserPINCredential) error {
	err := tx.Debug().Delete(userPin).Error
	if err != nil {
		return err
	}

	return nil
}

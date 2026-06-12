package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IUserIdentityRepository interface {
	CreateUserIdentity(tx *gorm.DB, userIdentity *entity.UserIdentity) error
	GetUserIdentity(tx *gorm.DB, param model.GetUserIdentityParam) (*entity.UserIdentity, error)
	UpdateUserIdentity(tx *gorm.DB, userIdentity *entity.UserIdentity) error
}

type UserIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) IUserIdentityRepository {
	return &UserIdentityRepository{db: db}
}

func (r *UserIdentityRepository) CreateUserIdentity(tx *gorm.DB, userIdentity *entity.UserIdentity) error {
	err := tx.Debug().Create(userIdentity).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserIdentityRepository) GetUserIdentity(tx *gorm.DB, param model.GetUserIdentityParam) (*entity.UserIdentity, error) {
	var userIdentity *entity.UserIdentity
	err := tx.Debug().Where(&param).First(&userIdentity).Error
	if err != nil {
		return nil, err
	}

	return userIdentity, nil
}

func (r *UserIdentityRepository) UpdateUserIdentity(tx *gorm.DB, userIdentity *entity.UserIdentity) error {
	err := tx.Debug().Save(userIdentity).Error
	if err != nil {
		return err
	}

	return nil
}

package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IUserSessionRepository interface {
	CreateUserSession(tx *gorm.DB, session *entity.UserSession) error
	GetUserSession(tx *gorm.DB, param model.GetUserSessionParam) (*entity.UserSession, error)
	UpdateUserSession(tx *gorm.DB, session *entity.UserSession) error
}

type UserSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) IUserSessionRepository {
	return &UserSessionRepository{db: db}
}

func (r *UserSessionRepository) CreateUserSession(tx *gorm.DB, session *entity.UserSession) error {
	err := tx.Debug().Create(session).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserSessionRepository) GetUserSession(tx *gorm.DB, param model.GetUserSessionParam) (*entity.UserSession, error) {
	var session *entity.UserSession
	err := tx.Debug().Where(&param).First(&session).Error
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *UserSessionRepository) UpdateUserSession(tx *gorm.DB, session *entity.UserSession) error {
	err := tx.Debug().Save(session).Error
	if err != nil {
		return err
	}

	return nil
}

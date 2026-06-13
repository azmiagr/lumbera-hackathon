package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserSessionRepository interface {
	CreateUserSession(tx *gorm.DB, session *entity.UserSession) error
	GetUserSession(tx *gorm.DB, param model.GetUserSessionParam) (*entity.UserSession, error)
	GetActiveUserSession(tx *gorm.DB, userID uuid.UUID, sessionID uuid.UUID, now time.Time) (*entity.UserSession, error)
	UpdateUserSession(tx *gorm.DB, session *entity.UserSession) error
	ListActiveUserSessions(tx *gorm.DB, userID uuid.UUID, now time.Time) ([]entity.UserSession, error)
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

func (r *UserSessionRepository) GetActiveUserSession(tx *gorm.DB, userID uuid.UUID, sessionID uuid.UUID, now time.Time) (*entity.UserSession, error) {
	var session entity.UserSession
	err := tx.Debug().
		Where("session_id = ?", sessionID).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *UserSessionRepository) UpdateUserSession(tx *gorm.DB, session *entity.UserSession) error {
	err := tx.Debug().Save(session).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserSessionRepository) ListActiveUserSessions(tx *gorm.DB, userID uuid.UUID, now time.Time) ([]entity.UserSession, error) {
	var sessions []entity.UserSession

	err := tx.Debug().
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILoanApplicationRepository interface {
	CreateLoanApplication(tx *gorm.DB, application *entity.LoanApplication) error
	GetLoanApplicationByID(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, applicationID uuid.UUID) (*entity.LoanApplication, error)
	GetActiveLoanApplication(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanApplication, error)
}

type LoanApplicationRepository struct {
	db *gorm.DB
}

func NewLoanApplicationRepository(db *gorm.DB) ILoanApplicationRepository {
	return &LoanApplicationRepository{db: db}
}

func (r *LoanApplicationRepository) CreateLoanApplication(tx *gorm.DB, application *entity.LoanApplication) error {
	err := tx.Debug().Create(application).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanApplicationRepository) GetLoanApplicationByID(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, applicationID uuid.UUID) (*entity.LoanApplication, error) {
	var application entity.LoanApplication
	err := tx.Debug().
		Where("application_id = ?", applicationID).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		First(&application).Error
	if err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *LoanApplicationRepository) GetActiveLoanApplication(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanApplication, error) {
	var application entity.LoanApplication
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("status IN ?", []string{"RECEIVED", "CREDIT_VERIFIED", "UNDER_REVIEW", "APPROVED"}).
		Order("submitted_at DESC").
		First(&application).Error
	if err != nil {
		return nil, err
	}

	return &application, nil
}

package repository

import (
	"fmt"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ILoanRepository interface {
	CreateLoanAccount(tx *gorm.DB, loan *entity.LoanAccount) error
	CreateLoanInstallmentSchedules(tx *gorm.DB, schedules []entity.LoanInstallmentSchedule) error
	GetActiveLoanByMemberForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanAccount, error)
	GetLoanByIDForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, loanID uuid.UUID) (*entity.LoanAccount, error)
	GetLoanByDisbursementTransactionID(tx *gorm.DB, transactionID uuid.UUID) (*entity.LoanAccount, error)
	ListPayableSchedulesForUpdate(tx *gorm.DB, loanID uuid.UUID) ([]entity.LoanInstallmentSchedule, error)
	CreateLoanPaymentAllocation(tx *gorm.DB, allocation *entity.LoanPaymentAllocation) error
	ListPaymentAllocationsByTransaction(tx *gorm.DB, transactionID uuid.UUID) ([]entity.LoanPaymentAllocation, error)
	ListSchedulesByIDs(tx *gorm.DB, scheduleIDs []uuid.UUID) ([]entity.LoanInstallmentSchedule, error)
	UpdateLoanSchedule(tx *gorm.DB, schedule *entity.LoanInstallmentSchedule) error
	UpdateLoanAccount(tx *gorm.DB, loan *entity.LoanAccount) error
	GetLoanSummary(tx *gorm.DB, loanID uuid.UUID, asOf time.Time) (*model.LoanSummaryResponse, error)
	GenerateNextLoanNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error)
}

type LoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) ILoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) CreateLoanAccount(tx *gorm.DB, loan *entity.LoanAccount) error {
	err := tx.Debug().Create(loan).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanRepository) CreateLoanInstallmentSchedules(tx *gorm.DB, schedules []entity.LoanInstallmentSchedule) error {
	if len(schedules) == 0 {
		return nil
	}

	err := tx.Debug().Create(&schedules).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanRepository) GetActiveLoanByMemberForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.LoanAccount, error) {
	var loan entity.LoanAccount
	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("status = ?", "ACTIVE").
		First(&loan).Error
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (r *LoanRepository) GetLoanByIDForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, loanID uuid.UUID) (*entity.LoanAccount, error) {
	var loan entity.LoanAccount
	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Where("loan_id = ?", loanID).
		First(&loan).Error
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (r *LoanRepository) GetLoanByDisbursementTransactionID(tx *gorm.DB, transactionID uuid.UUID) (*entity.LoanAccount, error) {
	var loan entity.LoanAccount
	err := tx.Debug().
		Where("disbursement_transaction_id = ?", transactionID).
		First(&loan).Error
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (r *LoanRepository) ListPayableSchedulesForUpdate(tx *gorm.DB, loanID uuid.UUID) ([]entity.LoanInstallmentSchedule, error) {
	var schedules []entity.LoanInstallmentSchedule
	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("loan_id = ?", loanID).
		Where("status IN ?", []string{"UNPAID", "PARTIAL"}).
		Order("installment_no ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *LoanRepository) CreateLoanPaymentAllocation(tx *gorm.DB, allocation *entity.LoanPaymentAllocation) error {
	err := tx.Debug().Create(allocation).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanRepository) ListPaymentAllocationsByTransaction(tx *gorm.DB, transactionID uuid.UUID) ([]entity.LoanPaymentAllocation, error) {
	var allocations []entity.LoanPaymentAllocation
	err := tx.Debug().
		Where("transaction_id = ?", transactionID).
		Order("created_at ASC").
		Find(&allocations).Error
	if err != nil {
		return nil, err
	}

	return allocations, nil
}

func (r *LoanRepository) ListSchedulesByIDs(tx *gorm.DB, scheduleIDs []uuid.UUID) ([]entity.LoanInstallmentSchedule, error) {
	if len(scheduleIDs) == 0 {
		return []entity.LoanInstallmentSchedule{}, nil
	}

	var schedules []entity.LoanInstallmentSchedule
	err := tx.Debug().
		Where("schedule_id IN ?", scheduleIDs).
		Order("installment_no ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *LoanRepository) UpdateLoanSchedule(tx *gorm.DB, schedule *entity.LoanInstallmentSchedule) error {
	err := tx.Debug().Save(schedule).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanRepository) UpdateLoanAccount(tx *gorm.DB, loan *entity.LoanAccount) error {
	err := tx.Debug().Save(loan).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *LoanRepository) GetLoanSummary(tx *gorm.DB, loanID uuid.UUID, asOf time.Time) (*model.LoanSummaryResponse, error) {
	var loan entity.LoanAccount
	if err := tx.Debug().Where("loan_id = ?", loanID).First(&loan).Error; err != nil {
		return nil, err
	}

	var aggregate struct {
		RemainingPayableAmount int64
		CurrentMonthDueAmount  int64
	}

	monthStart := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, asOf.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	err := tx.Debug().
		Table("loan_installment_schedules").
		Select(`
			COALESCE(SUM(remaining_amount), 0) AS remaining_payable_amount,
			COALESCE(SUM(CASE WHEN due_date BETWEEN ? AND ? THEN remaining_amount ELSE 0 END), 0) AS current_month_due_amount
		`, monthStart, monthEnd).
		Where("loan_id = ?", loanID).
		Scan(&aggregate).Error
	if err != nil {
		return nil, err
	}

	return &model.LoanSummaryResponse{
		LoanID:                   loan.LoanID,
		LoanNumber:               loan.LoanNumber,
		Status:                   loan.Status,
		PrincipalAmount:          loan.PrincipalAmount,
		TotalPayableAmount:       loan.TotalPayableAmount,
		MonthlyInstallmentAmount: loan.MonthlyInstallmentAmount,
		RemainingPayableAmount:   aggregate.RemainingPayableAmount,
		CurrentMonthDueAmount:    aggregate.CurrentMonthDueAmount,
		TermMonths:               loan.TermMonths,
	}, nil
}

func (r *LoanRepository) GenerateNextLoanNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error) {
	var total int64
	err := tx.Debug().
		Model(&entity.LoanAccount{}).
		Where("cooperative_id = ?", cooperativeID).
		Count(&total).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("P-%03d", total+1), nil
}

package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAccountingRepository interface {
	CreateAccount(tx *gorm.DB, account *entity.Account) error
	GetAccountByCode(tx *gorm.DB, cooperativeID uuid.UUID, code string) (*entity.Account, error)
	CreateJournalEntry(tx *gorm.DB, entry *entity.JournalEntry) error
	GetAccountBalances(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) ([]AccountBalanceRow, error)
}

type AccountBalanceRow struct {
	AccountCode   string
	AccountName   string
	AccountType   string
	NormalBalance string
	CashFlowGroup string
	TotalDebit    int64
	TotalCredit   int64
}

type AccountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) IAccountingRepository {
	return &AccountingRepository{db: db}
}

func (r *AccountingRepository) CreateAccount(tx *gorm.DB, account *entity.Account) error {
	err := tx.Debug().Create(account).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AccountingRepository) GetAccountByCode(tx *gorm.DB, cooperativeID uuid.UUID, code string) (*entity.Account, error) {
	var account entity.Account
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("code = ?", code).
		First(&account).Error
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AccountingRepository) CreateJournalEntry(tx *gorm.DB, entry *entity.JournalEntry) error {
	err := tx.Debug().Create(entry).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AccountingRepository) GetAccountBalances(tx *gorm.DB, cooperativeID uuid.UUID, startDate, endDate time.Time) ([]AccountBalanceRow, error) {
	var rows []AccountBalanceRow
	err := tx.Debug().
		Table("accounts").
		Select(`
			accounts.code AS account_code,
			accounts.name AS account_name,
			accounts.account_type,
			accounts.normal_balance,
			accounts.cash_flow_group,
			COALESCE(SUM(journal_entry_lines.debit_amount), 0) AS total_debit,
			COALESCE(SUM(journal_entry_lines.credit_amount), 0) AS total_credit
		`).
		Joins("LEFT JOIN journal_entry_lines ON journal_entry_lines.account_id = accounts.account_id").
		Joins("LEFT JOIN journal_entries ON journal_entries.journal_entry_id = journal_entry_lines.journal_entry_id").
		Where("accounts.cooperative_id = ?", cooperativeID).
		Where("(journal_entries.entry_date IS NULL OR journal_entries.entry_date BETWEEN ? AND ?)", startDate, endDate).
		Group("accounts.account_id").
		Order("accounts.code ASC").
		Scan(&rows).Error
	return rows, err
}

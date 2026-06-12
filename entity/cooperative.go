package entity

import "github.com/google/uuid"

type Cooperative struct {
	CooperativeID         uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);primaryKey"`
	Name                  string    `json:"name" gorm:"type:varchar(255);not null"`
	CooperativeType       string    `json:"cooperative_type" gorm:"type:enum('KSP','PANGAN_BULKY','COLD_CHAIN','TOKO_GERAI','UTILITY','PETERNAKAN');default:'KSP'"`
	RegistrationNumber    string    `json:"registration_number" gorm:"type:varchar(50);uniqueIndex;not null"`
	CooperativeCode       string    `json:"cooperative_code" gorm:"type:varchar(50);uniqueIndex"`
	EstablishedYear       int       `json:"established_year" gorm:"type:int;not null"`
	Status                string    `json:"status" gorm:"type:enum('ACTIVE','SUSPENDED','INACTIVE');default:'ACTIVE'"`
	Address               string    `json:"address" gorm:"type:text;not null"`
	BankName              string    `json:"bank_name" gorm:"type:varchar(100)"`
	BankAccountNumber     string    `json:"bank_account_number" gorm:"type:varchar(100);not null"`
	BankAccountHolderName string    `json:"bank_account_holder_name" gorm:"type:varchar(255);not null"`

	FinancialConfiguration     *FinancialConfiguration     `json:"financial_configuration" gorm:"foreignKey:CooperativeID;constraint:onDelete:CASCADE"`
	UserCooperativeMemberships []UserCooperativeMembership `json:"user_cooperative_memberships" gorm:"foreignKey:CooperativeID;constraint:onDelete:CASCADE"`
	Members                    []Member                    `json:"members" gorm:"foreignKey:CooperativeID;constraint:onDelete:CASCADE"`
	Transactions               []Transaction               `json:"transactions" gorm:"foreignKey:CooperativeID;constraint:onDelete:CASCADE"`
}

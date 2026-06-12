package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository                      IUserRepository
	UserIdentityRepository              IUserIdentityRepository
	RoleRepository                      IRoleRepository
	CooperativeRepository               ICooperativeRepository
	FinancialConfigurationRepository    IFinancialConfigurationRepository
	UserCooperativeMembershipRepository IUserCooperativeMembershipRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:                      NewUserRepository(db),
		UserIdentityRepository:              NewUserIdentityRepository(db),
		RoleRepository:                      NewRoleRepository(db),
		CooperativeRepository:               NewCooperativeRepository(db),
		FinancialConfigurationRepository:    NewFinancialConfigurationRepository(db),
		UserCooperativeMembershipRepository: NewUserCooperativeMembershipRepository(db),
	}
}

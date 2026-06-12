package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository                      IUserRepository
	UserIdentityRepository              IUserIdentityRepository
	RoleRepository                      IRoleRepository
	CooperativeRepository               ICooperativeRepository
	FinancialConfigurationRepository    IFinancialConfigurationRepository
	UserCooperativeMembershipRepository IUserCooperativeMembershipRepository
	UserPinRepository                   IUserPinRepository
	UserSessionRepository               IUserSessionRepository
	PhoneVerificationRepository         IPhoneVerificationRepository
	OnboardingDraftRepository           IOnboardingDraftRepository
	MemberActivationRepository          IMemberActivationRepository
	MemberRepository                    IMemberRepository
	TransactionRepository               ITransactionRepository
	AccountingRepository                IAccountingRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:                      NewUserRepository(db),
		UserIdentityRepository:              NewUserIdentityRepository(db),
		RoleRepository:                      NewRoleRepository(db),
		CooperativeRepository:               NewCooperativeRepository(db),
		FinancialConfigurationRepository:    NewFinancialConfigurationRepository(db),
		UserCooperativeMembershipRepository: NewUserCooperativeMembershipRepository(db),
		UserPinRepository:                   NewUserPinRepository(db),
		UserSessionRepository:               NewUserSessionRepository(db),
		PhoneVerificationRepository:         NewPhoneVerificationRepository(db),
		OnboardingDraftRepository:           NewOnboardingDraftRepository(db),
		MemberActivationRepository:          NewMemberActivationRepository(db),
		MemberRepository:                    NewMemberRepository(db),
		TransactionRepository:               NewTransactionRepository(db),
		AccountingRepository:                NewAccountingRepository(db),
	}
}

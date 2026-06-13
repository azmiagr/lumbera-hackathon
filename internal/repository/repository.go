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
	LoanRepository                      ILoanRepository
	AccountingRepository                IAccountingRepository
	CHSRepository                       ICHSRepository
	MemberImportRepository              IMemberImportRepository
	MCSRepository                       IMCSRepository
	SavingsBookRepository               ISavingsBookRepository
	LoanDashboardRepository             ILoanDashboardRepository
	LoanApplicationRepository           ILoanApplicationRepository
	CreditAccessRepository              ICreditAccessRepository
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
		LoanRepository:                      NewLoanRepository(db),
		AccountingRepository:                NewAccountingRepository(db),
		CHSRepository:                       NewCHSRepository(db),
		MemberImportRepository:              NewMemberImportRepository(db),
		MCSRepository:                       NewMCSRepository(db),
		SavingsBookRepository:               NewSavingsBookRepository(db),
		LoanDashboardRepository:             NewLoanDashboardRepository(db),
		LoanApplicationRepository:           NewLoanApplicationRepository(db),
		CreditAccessRepository:              NewCreditAccessRepository(db),
	}
}

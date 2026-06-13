package service

import (
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/pkg/bcrypt"
	"github.com/azmiagr/lumbera-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/azmiagr/lumbera-hackathon/pkg/n8n"
	"github.com/azmiagr/lumbera-hackathon/pkg/supabase"
	"github.com/azmiagr/lumbera-hackathon/pkg/whatsapp"
	"gorm.io/gorm"
)

type Service struct {
	OfficerRegistrationService IOfficerRegistrationService
	MemberActivationService    IMemberActivationService
	AuthService                IAuthService
	TransactionService         ITransactionService
	MemberService              IMemberService
	ReportService              IReportService
	MCSService                 IMCSService
}

type serviceDependency struct {
	db         *gorm.DB
	repository *repository.Repository
	bcrypt     bcrypt.Interface
	jwtAuth    jwt.Interface
	whatsapp   whatsapp.Interface
	supabase   supabase.Interface
	n8n        n8n.Interface
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, whatsapp whatsapp.Interface, supabase supabase.Interface, n8nClient n8n.Interface) *Service {
	deps := serviceDependency{
		db:         mariadb.Connection,
		repository: repository,
		bcrypt:     bcrypt,
		jwtAuth:    jwtAuth,
		whatsapp:   whatsapp,
		supabase:   supabase,
		n8n:        n8nClient,
	}

	return &Service{
		OfficerRegistrationService: NewOfficerRegistrationService(deps),
		MemberActivationService:    NewMemberActivationService(deps),
		AuthService:                NewAuthService(deps),
		TransactionService:         NewTransactionService(deps),
		MemberService:              NewMemberService(deps),
		ReportService:              NewReportService(deps),
		MCSService:                 NewMCSService(deps),
	}
}

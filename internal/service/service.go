package service

import (
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/pkg/bcrypt"
	"github.com/azmiagr/lumbera-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/azmiagr/lumbera-hackathon/pkg/supabase"
	"github.com/azmiagr/lumbera-hackathon/pkg/whatsapp"
	"gorm.io/gorm"
)

type Service struct {
	OfficerRegistrationService IOfficerRegistrationService
	MemberActivationService    IMemberActivationService
	AuthService                IAuthService
}

type serviceDependency struct {
	db         *gorm.DB
	repository *repository.Repository
	bcrypt     bcrypt.Interface
	jwtAuth    jwt.Interface
	whatsapp   whatsapp.Interface
	supabase   supabase.Interface
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, whatsapp whatsapp.Interface, supabase supabase.Interface) *Service {
	deps := serviceDependency{
		db:         mariadb.Connection,
		repository: repository,
		bcrypt:     bcrypt,
		jwtAuth:    jwtAuth,
		whatsapp:   whatsapp,
		supabase:   supabase,
	}

	return &Service{
		OfficerRegistrationService: NewOfficerRegistrationService(deps),
		MemberActivationService:    NewMemberActivationService(deps),
		AuthService:                NewAuthService(deps),
	}
}

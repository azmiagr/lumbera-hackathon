package mariadb

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.User{},
		&entity.UserIdentity{},
		&entity.Role{},
		&entity.Permission{},
		&entity.Cooperative{},
		&entity.FinancialConfiguration{},
		&entity.Partner{},
		&entity.RolePermission{},
		&entity.UserRole{},
		&entity.Member{},
		&entity.UserCooperativeMembership{},
		&entity.PartnerUser{},
		&entity.PhoneVerificationChallenge{},
		&entity.UserPINCredential{},
		&entity.UserSession{},
		&entity.MemberActivationChallenge{},
		&entity.OnboardingDraft{},
		&entity.Transaction{},
		&entity.LoanAccount{},
		&entity.LoanInstallmentSchedule{},
		&entity.LoanPaymentAllocation{},
		&entity.Account{},
		&entity.JournalEntry{},
		&entity.JournalEntryLine{},
		&entity.MemberImportBatch{},
		&entity.MemberImportRow{},
	)

	if err != nil {
		return err
	}

	return seedRoles(db)
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{
			RoleID:       uuid.New(),
			Name:         "Super Admin",
			Code:         constants.RoleCodeSuperAdmin,
			Description:  "Admin utama platform LUMBERA.",
			ScopeType:    constants.RoleScopePlatform,
			IsSystemRole: true,
		},
		{
			RoleID:       uuid.New(),
			Name:         "Regulator",
			Code:         constants.RoleCodeRegulator,
			Description:  "Pengguna regulator untuk monitoring dan pengawasan.",
			ScopeType:    constants.RoleScopeRegulator,
			IsSystemRole: true,
		},
		{
			RoleID:       uuid.New(),
			Name:         "Mitra",
			Code:         constants.RoleCodeMitra,
			Description:  "Pengguna mitra eksternal seperti fintech, bank, atau lembaga pembiayaan.",
			ScopeType:    constants.RoleScopePartner,
			IsSystemRole: true,
		},
		{
			RoleID:       uuid.New(),
			Name:         "Anggota",
			Code:         constants.RoleCodeAnggota,
			Description:  "Anggota koperasi.",
			ScopeType:    constants.RoleScopeCooperative,
			IsSystemRole: true,
		},
		{
			RoleID:       uuid.New(),
			Name:         "Pengurus Koperasi",
			Code:         constants.RoleCodePengurusKoperasi,
			Description:  "Pengurus koperasi seperti ketua, bendahara, sekretaris, atau staf.",
			ScopeType:    constants.RoleScopeCooperative,
			IsSystemRole: true,
		},
	}

	for _, role := range roles {
		existingRole := entity.Role{}
		err := db.
			Where("code = ?", role.Code).
			Attrs(entity.Role{
				RoleID: role.RoleID,
				Code:   role.Code,
			}).
			Assign(entity.Role{
				Name:         role.Name,
				Description:  role.Description,
				ScopeType:    role.ScopeType,
				IsSystemRole: role.IsSystemRole,
			}).
			FirstOrCreate(&existingRole).Error
		if err != nil {
			return err
		}
	}

	return nil
}

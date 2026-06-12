package mariadb

import (
	"github.com/azmiagr/lumbera-hackathon/entity"

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
	)

	if err != nil {
		return err
	}

	return nil
}

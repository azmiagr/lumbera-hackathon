package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
	CreateUser(tx *gorm.DB, user *entity.User) error
	GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error)
	UpdateUser(tx *gorm.DB, user *entity.User) error
	GetCooperativeLoginContext(tx *gorm.DB, param model.GetCooperativeLoginContextParam) (*CooperativeLoginContext, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

type CooperativeLoginContext struct {
	User       *entity.User
	Membership *entity.UserCooperativeMembership
	Role       *entity.Role
	Member     *entity.Member
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error) {
	var user *entity.User
	err := tx.Debug().Where(&param).First(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Save(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetCooperativeLoginContext(tx *gorm.DB, param model.GetCooperativeLoginContextParam) (*CooperativeLoginContext, error) {
	var user entity.User

	userQuery := tx.Debug().
		Where("status = ?", "ACTIVE").
		Where("user_type = ?", "COOPERATIVE")

	if param.UserID != uuid.Nil {
		userQuery = userQuery.Where("user_id = ?", param.UserID)
	}

	if param.PhoneNumber != "" {
		userQuery = userQuery.Where("phone_number = ?", param.PhoneNumber)
	}

	if err := userQuery.First(&user).Error; err != nil {
		return nil, err
	}

	var membership entity.UserCooperativeMembership
	if err := tx.Debug().
		Table("user_cooperative_memberships").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = user_cooperative_memberships.cooperative_id").
		Where("user_cooperative_memberships.user_id = ?", user.UserID).
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("cooperatives.status = ?", "ACTIVE").
		Where("roles.scope_type = ?", constants.RoleScopeCooperative).
		Where("roles.code IN ?", []string{constants.RoleCodeAnggota, constants.RoleCodePengurusKoperasi}).
		Order("roles.code = '" + constants.RoleCodePengurusKoperasi + "' DESC").
		First(&membership).Error; err != nil {
		return nil, err
	}

	var role entity.Role
	if err := tx.Debug().
		Where("role_id = ?", membership.RoleID).
		Where("scope_type = ?", constants.RoleScopeCooperative).
		Where("code IN ?", []string{constants.RoleCodeAnggota, constants.RoleCodePengurusKoperasi}).
		First(&role).Error; err != nil {
		return nil, err
	}

	var member *entity.Member
	if role.Code == constants.RoleCodeAnggota {
		if membership.MemberID == nil {
			return nil, gorm.ErrRecordNotFound
		}

		var activeMember entity.Member
		if err := tx.Debug().
			Where("member_id = ?", *membership.MemberID).
			Where("user_id = ?", user.UserID).
			Where("cooperative_id = ?", membership.CooperativeID).
			Where("member_status = ?", "ACTIVE").
			First(&activeMember).Error; err != nil {
			return nil, err
		}

		member = &activeMember
	}

	return &CooperativeLoginContext{
		User:       &user,
		Membership: &membership,
		Role:       &role,
		Member:     member,
	}, nil
}

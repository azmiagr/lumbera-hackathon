package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IUserCooperativeMembershipRepository interface {
	CreateUserCooperativeMembership(tx *gorm.DB, membership *entity.UserCooperativeMembership) error
	GetUserCooperativeMembership(tx *gorm.DB, param model.GetUserCooperativeMembershipParam) (*entity.UserCooperativeMembership, error)
	GetActiveCooperativeOfficerMembership(tx *gorm.DB, param model.GetActiveCooperativeOfficerMembershipParam) (*entity.UserCooperativeMembership, error)
	UpdateUserCooperativeMembership(tx *gorm.DB, membership *entity.UserCooperativeMembership) error
}

type UserCooperativeMembershipRepository struct {
	db *gorm.DB
}

func NewUserCooperativeMembershipRepository(db *gorm.DB) IUserCooperativeMembershipRepository {
	return &UserCooperativeMembershipRepository{db: db}
}

func (r *UserCooperativeMembershipRepository) CreateUserCooperativeMembership(tx *gorm.DB, membership *entity.UserCooperativeMembership) error {
	err := tx.Debug().Create(membership).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserCooperativeMembershipRepository) GetUserCooperativeMembership(tx *gorm.DB, param model.GetUserCooperativeMembershipParam) (*entity.UserCooperativeMembership, error) {
	var membership *entity.UserCooperativeMembership
	err := tx.Debug().Where(&param).First(&membership).Error
	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (r *UserCooperativeMembershipRepository) GetActiveCooperativeOfficerMembership(tx *gorm.DB, param model.GetActiveCooperativeOfficerMembershipParam) (*entity.UserCooperativeMembership, error) {
	var membership *entity.UserCooperativeMembership

	err := tx.Debug().
		Table("user_cooperative_memberships").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = user_cooperative_memberships.cooperative_id").
		Where("user_cooperative_memberships.user_id = ?", param.UserID).
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("cooperatives.status = ?", "ACTIVE").
		Where("roles.scope_type = ?", "COOPERATIVE").
		Where("roles.code <> ?", "MEMBER").
		First(&membership).Error
	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (r *UserCooperativeMembershipRepository) UpdateUserCooperativeMembership(tx *gorm.DB, membership *entity.UserCooperativeMembership) error {
	err := tx.Debug().Save(membership).Error
	if err != nil {
		return err
	}

	return nil
}

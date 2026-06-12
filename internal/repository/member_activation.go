package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MemberActivationContext struct {
	User       *entity.User
	Member     *entity.Member
	Membership *entity.UserCooperativeMembership
	Role       *entity.Role
}

type IMemberActivationRepository interface {
	GetEligibleMemberActivationContext(tx *gorm.DB, param model.GetEligibleMemberActivationContextParam) (*MemberActivationContext, error)
	CreateMemberActivationChallenge(tx *gorm.DB, challenge *entity.MemberActivationChallenge) error
	GetMemberActivationChallenge(tx *gorm.DB, param model.GetMemberActivationChallengeParam) (*entity.MemberActivationChallenge, error)
	UpdateMemberActivationChallenge(tx *gorm.DB, challenge *entity.MemberActivationChallenge) error
}

type MemberActivationRepository struct {
	db *gorm.DB
}

func NewMemberActivationRepository(db *gorm.DB) IMemberActivationRepository {
	return &MemberActivationRepository{db: db}
}

func (r *MemberActivationRepository) GetEligibleMemberActivationContext(tx *gorm.DB, param model.GetEligibleMemberActivationContextParam) (*MemberActivationContext, error) {
	var user entity.User
	query := tx.Debug().
		Where("status = ?", "PIN_REQUIRED").
		Where("user_type = ?", "COOPERATIVE")

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}

	if param.PhoneNumber != "" {
		query = query.Where("phone_number = ?", param.PhoneNumber)
	}

	if err := query.First(&user).Error; err != nil {
		return nil, err
	}

	var member entity.Member
	if err := tx.Debug().
		Where("user_id = ?", user.UserID).
		Where("member_status = ?", "ACTIVE").
		First(&member).Error; err != nil {
		return nil, err
	}

	var membership entity.UserCooperativeMembership
	if err := tx.Debug().
		Table("user_cooperative_memberships").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = user_cooperative_memberships.cooperative_id").
		Where("user_cooperative_memberships.user_id = ?", user.UserID).
		Where("user_cooperative_memberships.member_id = ?", member.MemberID).
		Where("user_cooperative_memberships.cooperative_id = ?", member.CooperativeID).
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("cooperatives.status = ?", "ACTIVE").
		Where("roles.scope_type = ?", constants.RoleScopeCooperative).
		Where("roles.code = ?", constants.RoleCodeAnggota).
		First(&membership).Error; err != nil {
		return nil, err
	}

	var role entity.Role
	if err := tx.Debug().
		Where("role_id = ?", membership.RoleID).
		Where("scope_type = ?", constants.RoleScopeCooperative).
		Where("code = ?", constants.RoleCodeAnggota).
		First(&role).Error; err != nil {
		return nil, err
	}

	return &MemberActivationContext{
		User:       &user,
		Member:     &member,
		Membership: &membership,
		Role:       &role,
	}, nil
}

func (r *MemberActivationRepository) CreateMemberActivationChallenge(tx *gorm.DB, challenge *entity.MemberActivationChallenge) error {
	err := tx.Debug().Create(challenge).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberActivationRepository) GetMemberActivationChallenge(tx *gorm.DB, param model.GetMemberActivationChallengeParam) (*entity.MemberActivationChallenge, error) {
	var challenge entity.MemberActivationChallenge
	query := tx.Debug()

	if param.ActivationChallengeID != uuid.Nil {
		query = query.Where("challenge_id = ?", param.ActivationChallengeID)
	}

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}

	err := query.First(&challenge).Error
	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (r *MemberActivationRepository) UpdateMemberActivationChallenge(tx *gorm.DB, challenge *entity.MemberActivationChallenge) error {
	err := tx.Debug().Save(challenge).Error
	if err != nil {
		return err
	}

	return nil
}

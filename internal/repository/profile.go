package repository

import (
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IProfileRepository interface {
	GetOfficerProfile(tx *gorm.DB, userID, cooperativeID uuid.UUID) (*model.OfficerProfileRow, error)
	GetMemberProfile(tx *gorm.DB, userID, cooperativeID uuid.UUID) (*model.MemberProfileRow, error)
	CountCompletedLoans(tx *gorm.DB, cooperativeID, memberID uuid.UUID) (int, error)
}

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) IProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetOfficerProfile(tx *gorm.DB, userID, cooperativeID uuid.UUID) (*model.OfficerProfileRow, error) {
	var result model.OfficerProfileRow

	err := tx.Debug().
		Table("user_cooperative_memberships").
		Select(`
			users.user_id,
			users.full_name,
			users.phone_number,
			cooperatives.cooperative_id,
			cooperatives.name AS cooperative_name,
			cooperatives.cooperative_code,
			cooperatives.registration_number,
			user_cooperative_memberships.position_code,
			user_cooperative_memberships.joined_at
		`).
		Joins("JOIN users ON users.user_id = user_cooperative_memberships.user_id").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = user_cooperative_memberships.cooperative_id").
		Where("users.user_id = ?", userID).
		Where("cooperatives.cooperative_id = ?", cooperativeID).
		Where("users.status IN ?", []string{"ACTIVE", "PIN_REQUIRED"}).
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("roles.code = ?", constants.RoleCodePengurusKoperasi).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *ProfileRepository) GetMemberProfile(tx *gorm.DB, userID, cooperativeID uuid.UUID) (*model.MemberProfileRow, error) {
	var result model.MemberProfileRow

	err := tx.Debug().
		Table("members").
		Select(`
			users.user_id,
			members.member_id,
			users.full_name,
			users.phone_number,
			members.member_number,
			members.joined_date,
			members.current_m_csscore,
			members.mcs_grade,
			members.last_score_updated_at,
			cooperatives.cooperative_id,
			cooperatives.name AS cooperative_name,
			cooperatives.cooperative_code
		`).
		Joins("JOIN users ON users.user_id = members.user_id").
		Joins("JOIN cooperatives ON cooperatives.cooperative_id = members.cooperative_id").
		Joins("JOIN user_cooperative_memberships ON user_cooperative_memberships.member_id = members.member_id").
		Joins("JOIN roles ON roles.role_id = user_cooperative_memberships.role_id").
		Where("users.user_id = ?", userID).
		Where("members.cooperative_id = ?", cooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Where("user_cooperative_memberships.status = ?", "ACTIVE").
		Where("roles.code = ?", constants.RoleCodeAnggota).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *ProfileRepository) CountCompletedLoans(tx *gorm.DB, cooperativeID, memberID uuid.UUID) (int, error) {
	var total int64
	err := tx.Debug().
		Table("loan_accounts").
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		Where("status = ?", "PAID").
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

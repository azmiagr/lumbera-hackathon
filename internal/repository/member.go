package repository

import (
	"strings"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMemberRepository interface {
	GetActiveMember(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.Member, error)
	SearchTransactionMembers(tx *gorm.DB, req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error)
}

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) IMemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) GetActiveMember(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) (*entity.Member, error) {
	var member entity.Member

	err := tx.Debug().
		Where("member_id = ?", memberID).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_status = ?", "ACTIVE").
		First(&member).Error
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *MemberRepository) SearchTransactionMembers(tx *gorm.DB, req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error) {
	var members []model.TransactionMemberResponse

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	query := tx.Debug().
		Table("members").
		Select(`
			members.member_id,
			members.user_id,
			users.full_name,
			members.member_number,
			members.cooperative_id,
			members.mcs_grade
		`).
		Joins("JOIN users ON users.user_id = members.user_id").
		Where("members.cooperative_id = ?", req.CooperativeID).
		Where("members.member_status = ?", "ACTIVE").
		Where("users.status = ?", "ACTIVE").
		Order("users.full_name ASC").
		Limit(limit)

	search := strings.TrimSpace(req.Search)
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"(users.full_name LIKE ? OR members.member_number LIKE ?)",
			keyword,
			keyword,
		)
	}

	err := query.Scan(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

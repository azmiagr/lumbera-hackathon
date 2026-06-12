package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	CreateRole(tx *gorm.DB, role *entity.Role) error
	GetRole(tx *gorm.DB, param model.GetRoleParam) (*entity.Role, error)
	UpdateRole(tx *gorm.DB, role *entity.Role) error
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) CreateRole(tx *gorm.DB, role *entity.Role) error {
	err := tx.Debug().Create(role).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) GetRole(tx *gorm.DB, param model.GetRoleParam) (*entity.Role, error) {
	var role *entity.Role
	err := tx.Debug().Where(&param).First(&role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) UpdateRole(tx *gorm.DB, role *entity.Role) error {
	err := tx.Debug().Save(role).Error
	if err != nil {
		return err
	}
	return nil
}

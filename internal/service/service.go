package service

import (
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/pkg/bcrypt"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
)

type Service struct {
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{}
}

package middleware

import (
	"github.com/azmiagr/lumbera-hackathon/internal/service"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type Interface interface {
	Cors() gin.HandlerFunc
	AuthenticateUser() gin.HandlerFunc
	RequireRole(allowedRoles ...string) gin.HandlerFunc
}

type middleware struct {
	service *service.Service
	jwtAuth jwt.Interface
}

func Init(service *service.Service, jwtAuth jwt.Interface) Interface {
	return &middleware{
		service: service,
		jwtAuth: jwtAuth,
	}
}

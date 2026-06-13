package middleware

import (
	"net/http"
	"strings"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (m *middleware) AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" || token == auth {
			response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		claims, err := m.jwtAuth.ParseAccessToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token", err)
			c.Abort()
			return
		}

		err = m.service.AuthService.ValidateAuthenticatedSession(model.AuthContext{
			UserID:        claims.UserID,
			CooperativeID: claims.CooperativeID,
			SessionID:     claims.SessionID,
			RoleCode:      claims.RoleCode,
		})
		if err != nil {
			response.HandleError(c, err)
			c.Abort()
			return
		}

		c.Set("auth_claims", claims)
		c.Next()
	}
}

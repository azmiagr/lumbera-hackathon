package middleware

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (m *middleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		value, exists := c.Get("auth_claims")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		claims, ok := value.(*jwt.Claims)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "invalid auth context", nil)
			c.Abort()
			return
		}

		if _, ok := allowed[claims.RoleCode]; !ok {
			response.Error(c, http.StatusForbidden, "forbidden", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

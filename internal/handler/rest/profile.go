package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetProfile(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	result, err := r.service.ProfileService.GetProfile(model.GetProfileRequest{
		AuthContext: authContext,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get profile", result)
}

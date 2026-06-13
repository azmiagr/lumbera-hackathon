package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) TriggerMemberMCS(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	memberID, err := uuid.Parse(c.Param("memberID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid member id", err)
		return
	}

	result, err := r.service.MCSService.TriggerMemberMCS(model.TriggerMemberMCSRequest{
		AuthContext: authContext,
		MemberID:    memberID,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusAccepted, "success to trigger MCS scoring", result)
}

func (r *Rest) ApplyMCSCallback(c *gin.Context) {
	var req model.MCSCallbackRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.MCSService.ApplyMCSCallback(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to apply MCS callback", result)
}

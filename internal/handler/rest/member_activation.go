package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) CheckMemberPhone(c *gin.Context) {
	var req model.CheckMemberPhoneRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.MemberActivationService.CheckPhone(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to check member phone", result)
}

func (r *Rest) SetMemberPIN(c *gin.Context) {
	var req model.SetMemberPINRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}

	if req.UserAgent == "" {
		req.UserAgent = c.Request.UserAgent()
	}

	result, err := r.service.MemberActivationService.SetPIN(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to set member pin", result)
}

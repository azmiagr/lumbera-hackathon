package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) Login(c *gin.Context) {
	var req model.LoginRequest
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

	result, err := r.service.AuthService.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to login", result)
}

func (r *Rest) RequestForgotPINOTP(c *gin.Context) {
	var req model.ForgotPINRequestOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.AuthService.RequestForgotPINOTP(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to request forgot PIN OTP", result)
}

func (r *Rest) VerifyForgotPINOTP(c *gin.Context) {
	var req model.ForgotPINVerifyOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.AuthService.VerifyForgotPINOTP(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to verify forgot PIN OTP", result)
}

func (r *Rest) SetForgottenPIN(c *gin.Context) {
	var req model.ForgotPINSetPINRequest
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

	result, err := r.service.AuthService.SetForgottenPIN(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to reset PIN", result)
}

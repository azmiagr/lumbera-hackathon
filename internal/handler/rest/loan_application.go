package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) GetLoanApplicationEligibility(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	result, err := r.service.LoanApplicationService.GetLoanApplicationEligibility(model.GetLoanApplicationEligibilityRequest{
		AuthContext: authContext,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get loan application eligibility", result)
}

func (r *Rest) CreateLoanApplication(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateLoanApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.LoanApplicationService.CreateLoanApplication(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create loan application", result)
}

func (r *Rest) GetLoanApplication(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	applicationID, err := uuid.Parse(c.Param("applicationID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid application id", err)
		return
	}

	result, err := r.service.LoanApplicationService.GetLoanApplication(model.GetLoanApplicationRequest{
		AuthContext:   authContext,
		ApplicationID: applicationID,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get loan application", result)
}

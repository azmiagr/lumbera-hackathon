package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) ListCreditAccessRequests(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	result, err := r.service.CreditAccessService.ListCreditAccessRequests(model.ListCreditAccessRequestsRequest{
		AuthContext: authContext,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get credit access requests", result)
}

func (r *Rest) GetCreditAccessRequest(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	requestID, err := parseCreditAccessRequestID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request id", err)
		return
	}

	result, err := r.service.CreditAccessService.GetCreditAccessRequest(model.GetCreditAccessRequestRequest{
		AuthContext: authContext,
		RequestID:   requestID,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get credit access request", result)
}

func (r *Rest) GrantCreditAccess(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	requestID, err := parseCreditAccessRequestID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request id", err)
		return
	}

	var req model.GrantCreditAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}
	req.AuthContext = authContext
	req.RequestID = requestID

	result, err := r.service.CreditAccessService.GrantCreditAccess(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to grant credit access", result)
}

func (r *Rest) DeclineCreditAccess(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	requestID, err := parseCreditAccessRequestID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request id", err)
		return
	}

	result, err := r.service.CreditAccessService.DeclineCreditAccess(model.DeclineCreditAccessRequest{
		AuthContext: authContext,
		RequestID:   requestID,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to decline credit access", result)
}

func (r *Rest) RevokeCreditAccess(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	requestID, err := parseCreditAccessRequestID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request id", err)
		return
	}

	result, err := r.service.CreditAccessService.RevokeCreditAccess(model.RevokeCreditAccessRequest{
		AuthContext: authContext,
		RequestID:   requestID,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to revoke credit access", result)
}

func parseCreditAccessRequestID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("requestID"))
}

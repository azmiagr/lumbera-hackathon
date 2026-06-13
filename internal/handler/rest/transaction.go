package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) SearchTransactionMembers(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.SearchTransactionMembersRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.SearchTransactionMembers(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to search transaction members", result)
}

func (r *Rest) CreateSavingsTransaction(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateSavingsTransactionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.CreateSavingsTransaction(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create savings transaction", result)
}

func (r *Rest) ListTransactions(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ListTransactionsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.ListTransactions(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get transactions", result)
}

func (r *Rest) CreateLoanTransaction(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateLoanTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.CreateLoanTransaction(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create loan transaction", result)
}

func (r *Rest) CreateInstallmentTransaction(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateInstallmentTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.CreateInstallmentTransaction(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create installment transaction", result)
}

func (r *Rest) CreateCashWithdrawalTransaction(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateCashWithdrawalTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.TransactionService.CreateCashWithdrawalTransaction(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create cash withdrawal transaction", result)
}

func getAuthContext(c *gin.Context) (model.AuthContext, bool) {
	value, exists := c.Get("auth_claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return model.AuthContext{}, false
	}

	claims, ok := value.(*jwt.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid auth context", nil)
		return model.AuthContext{}, false
	}

	return model.AuthContext{
		UserID:        claims.UserID,
		CooperativeID: claims.CooperativeID,
		SessionID:     claims.SessionID,
		RoleCode:      claims.RoleCode,
	}, true
}

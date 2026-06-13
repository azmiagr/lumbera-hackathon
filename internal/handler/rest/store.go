package rest

import (
	"net/http"
	"strings"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) GetStoreDashboard(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	req := model.StoreDashboardRequest{
		AuthContext: authContext,
	}

	result, err := r.service.StoreService.GetStoreDashboard(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get store dashboard", result)
}

func (r *Rest) ListStoreProducts(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ListProductsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.StoreService.ListProducts(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get store products", result)
}

func (r *Rest) CreateStoreProduct(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateProductRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.StoreService.CreateProduct(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create store product", result)
}

func (r *Rest) GetStoreProduct(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id", err)
		return
	}

	req := model.GetProductRequest{
		AuthContext: authContext,
		ProductID:   productID,
	}

	result, err := r.service.StoreService.GetProduct(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get store product", result)
}

func (r *Rest) CreateStockIn(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id", err)
		return
	}

	var req model.CreateStockInRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext
	req.ProductID = productID

	result, err := r.service.StoreService.CreateStockIn(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create stock in", result)
}

func (r *Rest) CreateStockAdjustment(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id", err)
		return
	}

	var req model.CreateStockAdjustmentRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext
	req.ProductID = productID

	result, err := r.service.StoreService.CreateStockAdjustment(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create stock adjustment", result)
}

func (r *Rest) ListStockMovements(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ListStockMovementsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext
	productIDRaw := strings.TrimSpace(req.ProductIDRaw)
	if productIDRaw != "" {
		productID, err := uuid.Parse(productIDRaw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid product id", err)
			return
		}
		req.ProductID = productID
	}

	result, err := r.service.StoreService.ListStockMovements(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get stock movements", result)
}

func (r *Rest) CreateStoreSale(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateStoreSaleRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.StoreService.CreateStoreSale(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create store sale", result)
}

func (r *Rest) GetStoreSale(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	saleID, err := uuid.Parse(c.Param("saleID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid sale id", err)
		return
	}

	req := model.GetStoreSaleRequest{
		AuthContext: authContext,
		StoreSaleID: saleID,
	}

	result, err := r.service.StoreService.GetStoreSale(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get store sale", result)
}

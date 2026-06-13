package rest

import (
	"fmt"
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetSavingsBook(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.GetSavingsBookRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}
	req.AuthContext = authContext

	result, err := r.service.SavingsBookService.GetSavingsBook(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get savings book", result)
}

func (r *Rest) ExportSavingsBookXLSX(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ExportSavingsBookRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}
	req.AuthContext = authContext

	fileBytes, fileName, err := r.service.SavingsBookService.ExportSavingsBookXLSX(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-store")
	c.Data(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		fileBytes,
	)
}

func (r *Rest) ExportSavingsBookPDF(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ExportSavingsBookRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}
	req.AuthContext = authContext

	fileBytes, fileName, err := r.service.SavingsBookService.ExportSavingsBookPDF(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/pdf", fileBytes)
}

package rest

import (
	"fmt"
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetFinancialReport(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.FinancialReportRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.ReportService.GetFinancialReport(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get financial report", result)
}

func (r *Rest) ExportFinancialReportXLSX(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.FinancialReportRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	fileBytes, fileName, err := r.service.ReportService.ExportFinancialReportXLSX(req)
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

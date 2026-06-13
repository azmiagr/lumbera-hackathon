package rest

import (
	"fmt"
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) ListMembers(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.ListMembersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.MemberService.ListMembers(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get members", result)
}

func (r *Rest) CreateMember(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.MemberService.CreateMember(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create member", result)
}

func (r *Rest) UploadMemberImport(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file excel wajib diupload", err)
		return
	}

	req := model.UploadMemberImportRequest{
		AuthContext: authContext,
		File:        file,
	}

	result, err := r.service.MemberService.UploadMemberImport(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to upload member import", result)
}

func (r *Rest) GetMemberImport(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	batchID, err := uuid.Parse(c.Param("batchID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid batch id", err)
		return
	}

	var req model.GetMemberImportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext
	req.ImportBatchID = batchID

	result, err := r.service.MemberService.GetMemberImport(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get member import", result)
}

func (r *Rest) UpdateMemberImportRow(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	batchID, err := uuid.Parse(c.Param("batchID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid batch id", err)
		return
	}

	rowID, err := uuid.Parse(c.Param("rowID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid row id", err)
		return
	}

	var req model.UpdateMemberImportRowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext
	req.ImportBatchID = batchID
	req.ImportRowID = rowID

	result, err := r.service.MemberService.UpdateMemberImportRow(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update member import row", result)
}

func (r *Rest) DeleteMemberImportRow(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	batchID, err := uuid.Parse(c.Param("batchID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid batch id", err)
		return
	}

	rowID, err := uuid.Parse(c.Param("rowID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid row id", err)
		return
	}

	req := model.DeleteMemberImportRowRequest{
		AuthContext:   authContext,
		ImportBatchID: batchID,
		ImportRowID:   rowID,
	}

	if err := r.service.MemberService.DeleteMemberImportRow(req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to delete member import row", nil)
}

func (r *Rest) SubmitMemberImport(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	batchID, err := uuid.Parse(c.Param("batchID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid batch id", err)
		return
	}

	req := model.SubmitMemberImportRequest{
		AuthContext:   authContext,
		ImportBatchID: batchID,
	}

	result, err := r.service.MemberService.SubmitMemberImport(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to submit member import", result)
}

func (r *Rest) DownloadMemberImportTemplate(c *gin.Context) {
	fileBytes, fileName, err := r.service.MemberService.DownloadMemberImportTemplate()
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

func (r *Rest) GetMemberDashboard(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.GetMemberDashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.MemberDashboardService.GetDashboard(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get member dashboard", result)
}

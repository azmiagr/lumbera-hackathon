package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) PushSync(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.SyncPushRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.SyncService.Push(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to push sync operations", result)
}

func (r *Rest) GetSyncConfig(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	result, err := r.service.SyncService.GetConfig(authContext)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get sync config", result)
}

func (r *Rest) GetSyncBootstrap(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	result, err := r.service.SyncService.GetBootstrap(authContext)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get sync bootstrap", result)
}

func (r *Rest) GetSyncStatus(c *gin.Context) {
	authContext, ok := getAuthContext(c)
	if !ok {
		return
	}

	var req model.SyncStatusRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	req.AuthContext = authContext

	result, err := r.service.SyncService.GetStatus(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get sync status", result)
}

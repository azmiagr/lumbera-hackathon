package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
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

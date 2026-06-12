package rest

import (
	"net/http"

	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/azmiagr/lumbera-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const onboardingTokenHeader = "X-Onboarding-Token"

func (r *Rest) StartOfficerRegistration(c *gin.Context) {
	var req model.StartOfficerRegistrationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.OfficerRegistrationService.StartRegistration(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to start officer registration", result)
}

func (r *Rest) VerifyOfficerRegistrationOTP(c *gin.Context) {
	var req model.VerifyOfficerRegistrationOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	err = r.service.OfficerRegistrationService.VerifyOTP(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to verify OTP", nil)
}

func (r *Rest) SetOfficerRegistrationPIN(c *gin.Context) {
	var req model.SetOfficerRegistrationPINRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.OfficerRegistrationService.SetPIN(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to set PIN", result)
}

func (r *Rest) UpdateOnboardingPersonalData(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	file, err := c.FormFile("ktp_file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ktp file is required", err)
		return
	}

	req := model.UpdatePersonalDataRequest{
		OnboardingDraftID:       draftID,
		OnboardingToken:         token,
		KTPFile:                 file,
		FullName:                c.PostForm("full_name"),
		NIKEncrypted:            c.PostForm("nik_encrypted"),
		NIKHash:                 c.PostForm("nik_hash"),
		PositionCode:            c.PostForm("position_code"),
		ExistingCooperativeCode: c.PostForm("existing_cooperative_code"),
	}

	result, err := r.service.OfficerRegistrationService.UpdatePersonalData(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update personal data", result)
}

func (r *Rest) UpdateOnboardingCooperativeType(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	var req model.UpdateCooperativeTypeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.OnboardingDraftID = draftID
	req.OnboardingToken = token

	result, err := r.service.OfficerRegistrationService.UpdateCooperativeType(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update cooperative type", result)
}

func (r *Rest) UpdateOnboardingCooperativeProfile(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	var req model.UpdateCooperativeProfileRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.OnboardingDraftID = draftID
	req.OnboardingToken = token

	result, err := r.service.OfficerRegistrationService.UpdateCooperativeProfile(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update cooperative profile", result)
}

func (r *Rest) UpdateOnboardingFinancialConfiguration(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	var req model.UpdateFinancialConfigurationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.OnboardingDraftID = draftID
	req.OnboardingToken = token

	result, err := r.service.OfficerRegistrationService.UpdateFinancialConfiguration(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update financial configuration", result)
}

func (r *Rest) UpdateOnboardingCooperativeBankAccount(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	var req model.UpdateCooperativeBankAccountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	req.OnboardingDraftID = draftID
	req.OnboardingToken = token

	result, err := r.service.OfficerRegistrationService.UpdateCooperativeBankAccount(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update cooperative bank account", result)
}

func (r *Rest) ActivateOnboardingDraft(c *gin.Context) {
	draftID, token, ok := r.parseOnboardingDraftContext(c)
	if !ok {
		return
	}

	req := model.ActivateOnboardingDraftRequest{
		OnboardingDraftID: draftID,
		OnboardingToken:   token,
	}

	result, err := r.service.OfficerRegistrationService.ActivateOnboardingDraft(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to activate cooperative", result)
}

func (r *Rest) parseOnboardingDraftContext(c *gin.Context) (uuid.UUID, string, bool) {
	draftID, err := uuid.Parse(c.Param("draftID"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid onboarding draft id", err)
		return uuid.Nil, "", false
	}

	token := c.GetHeader(onboardingTokenHeader)
	if token == "" {
		response.Error(c, http.StatusUnauthorized, "onboarding token is required", nil)
		return uuid.Nil, "", false
	}

	return draftID, token, true
}

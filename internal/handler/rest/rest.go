package rest

import (
	"fmt"
	"os"

	"github.com/azmiagr/lumbera-hackathon/internal/service"
	"github.com/azmiagr/lumbera-hackathon/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Rest struct {
	router     *gin.Engine
	service    *service.Service
	middleware middleware.Interface
}

func NewRest(service *service.Service, middleware middleware.Interface) *Rest {
	return &Rest{
		router:     gin.Default(),
		service:    service,
		middleware: middleware,
	}
}

func (r *Rest) MountEndpoint() {
	r.router.Use(r.middleware.Cors())
	baseURL := r.router.Group("/api/v1")

	onboarding := baseURL.Group("/onboarding")
	onboarding.POST("officer/start", r.StartOfficerRegistration)
	onboarding.POST("officer/verify-otp", r.VerifyOfficerRegistrationOTP)
	onboarding.POST("officer/set-pin", r.SetOfficerRegistrationPIN)

	onboarding.GET("drafts/:draftID/state", r.GetOnboardingState)
	onboarding.PATCH("drafts/:draftID/personal-data", r.UpdateOnboardingPersonalData)
	onboarding.PATCH("drafts/:draftID/cooperative-type", r.UpdateOnboardingCooperativeType)
	onboarding.PATCH("drafts/:draftID/cooperative-profile", r.UpdateOnboardingCooperativeProfile)
	onboarding.PATCH("drafts/:draftID/financial-configuration", r.UpdateOnboardingFinancialConfiguration)
	onboarding.PATCH("drafts/:draftID/bank-account", r.UpdateOnboardingCooperativeBankAccount)
	onboarding.POST("drafts/:draftID/activate", r.ActivateOnboardingDraft)

	onboarding.POST("member/check-phone", r.CheckMemberPhone)
	onboarding.POST("member/set-pin", r.SetMemberPIN)

	auth := baseURL.Group("/auth")
	auth.POST("login", r.Login)
}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}

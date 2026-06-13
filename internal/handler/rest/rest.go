package rest

import (
	"fmt"
	"os"

	"github.com/azmiagr/lumbera-hackathon/internal/service"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
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
	auth.POST("forgot-pin/request-otp", r.RequestForgotPINOTP)
	auth.POST("forgot-pin/verify-otp", r.VerifyForgotPINOTP)
	auth.POST("forgot-pin/set-pin", r.SetForgottenPIN)
	auth.POST("logout", r.middleware.AuthenticateUser(), r.Logout)

	transactions := baseURL.Group("/transactions")
	transactions.Use(r.middleware.AuthenticateUser())
	transactions.Use(r.middleware.RequireRole(constants.RoleCodePengurusKoperasi))
	transactions.GET("", r.ListTransactions)
	transactions.GET("/members", r.SearchTransactionMembers)
	transactions.POST("/savings", r.CreateSavingsTransaction)
	transactions.POST("/loans", r.CreateLoanTransaction)
	transactions.POST("/installments", r.CreateInstallmentTransaction)
	transactions.POST("/cash-withdrawals", r.CreateCashWithdrawalTransaction)
	transactions.POST("/:transactionID/reverse", r.ReverseTransaction)

	members := baseURL.Group("/members")
	members.Use(r.middleware.AuthenticateUser())
	members.Use(r.middleware.RequireRole(constants.RoleCodePengurusKoperasi))
	members.GET("", r.ListMembers)
	members.POST("", r.CreateMember)
	members.GET("/imports/template", r.DownloadMemberImportTemplate)
	members.POST("/imports/upload", r.UploadMemberImport)
	members.GET("/imports/:batchID", r.GetMemberImport)
	members.PATCH("/imports/:batchID/rows/:rowID", r.UpdateMemberImportRow)
	members.DELETE("/imports/:batchID/rows/:rowID", r.DeleteMemberImportRow)
	members.POST("/imports/:batchID/submit", r.SubmitMemberImport)

	store := baseURL.Group("/store")
	store.Use(r.middleware.AuthenticateUser())
	store.Use(r.middleware.RequireRole(constants.RoleCodePengurusKoperasi))
	store.GET("/dashboard", r.GetStoreDashboard)
	store.GET("/products", r.ListStoreProducts)
	store.POST("/products", r.CreateStoreProduct)
	store.GET("/products/:productID", r.GetStoreProduct)
	store.POST("/products/:productID/stock-in", r.CreateStockIn)
	store.POST("/products/:productID/adjustments", r.CreateStockAdjustment)
	store.GET("/movements", r.ListStockMovements)
	store.POST("/sales", r.CreateStoreSale)
	store.GET("/sales/:saleID", r.GetStoreSale)

	ledger := baseURL.Group("/ledger")
	ledger.Use(r.middleware.AuthenticateUser())
	ledger.Use(r.middleware.RequireRole(constants.RoleCodePengurusKoperasi))
	ledger.GET("/audit", r.GetLedgerAudit)
	ledger.POST("/anchors", r.AnchorLedger)

	cooperativeMember := baseURL.Group("/cooperative-members")
	cooperativeMember.Use(r.middleware.AuthenticateUser())
	cooperativeMember.GET("/dashboard", r.GetMemberDashboard)
	cooperativeMember.GET("/savings-book", r.GetSavingsBook)
	cooperativeMember.GET("/savings-book/export", r.ExportSavingsBookXLSX)
	cooperativeMember.POST("/:memberID/mcs/run", r.TriggerMemberMCS)
	cooperativeMember.GET("/savings-book/export/pdf", r.ExportSavingsBookPDF)
	cooperativeMember.GET("/loan-dashboard", r.GetLoanDashboard)
	cooperativeMember.GET("/loan-applications/eligibility", r.GetLoanApplicationEligibility)
	cooperativeMember.POST("/loan-applications", r.CreateLoanApplication)
	cooperativeMember.GET("/loan-applications/:applicationID", r.GetLoanApplication)
	cooperativeMember.GET("/credit-access-requests", r.ListCreditAccessRequests)
	cooperativeMember.GET("/credit-access-requests/:requestID", r.GetCreditAccessRequest)
	cooperativeMember.POST("/credit-access-requests/:requestID/grant", r.GrantCreditAccess)
	cooperativeMember.POST("/credit-access-requests/:requestID/decline", r.DeclineCreditAccess)
	cooperativeMember.POST("/credit-access-requests/:requestID/revoke", r.RevokeCreditAccess)

	internal := baseURL.Group("/internal")
	internal.PATCH("/mcs/callback", r.ApplyMCSCallback)

	reports := baseURL.Group("/reports")
	reports.Use(r.middleware.AuthenticateUser())
	reports.Use(r.middleware.RequireRole(constants.RoleCodePengurusKoperasi))
	reports.GET("/financial", r.GetFinancialReport)
	reports.GET("/financial/export", r.ExportFinancialReportXLSX)
	reports.GET("/cooperative-health-score", r.GetCooperativeHealthScore)
	reports.GET("/dashboard-summary", r.GetDashboardSummary)

}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}

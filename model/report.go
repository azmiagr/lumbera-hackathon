package model

type FinancialReportRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
}

type CooperativeHealthScoreRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
}

type ReportColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type ReportRow struct {
	Section string           `json:"section"`
	Label   string           `json:"label"`
	Values  map[string]int64 `json:"values"`
	IsTotal bool             `json:"is_total"`
}

type FinancialReportResponse struct {
	PeriodColumns   []ReportColumn `json:"period_columns"`
	BalanceSheet    []ReportRow    `json:"balance_sheet"`
	IncomeStatement []ReportRow    `json:"income_statement"`
	CashFlow        []ReportRow    `json:"cash_flow"`
}

type CooperativeHealthScoreResponse struct {
	Period       string              `json:"period"`
	Status       string              `json:"status"`
	CHSScore     float64             `json:"chs_score"`
	DisplayScore int                 `json:"display_score"`
	Grade        string              `json:"grade"`
	Category     string              `json:"category"`
	Dimensions   []CHSDimensionScore `json:"dimensions"`
}

type CHSDimensionScore struct {
	Code       string              `json:"code"`
	Label      string              `json:"label"`
	Weight     float64             `json:"weight"`
	Score      float64             `json:"score"`
	Status     string              `json:"status"`
	Indicators []CHSIndicatorScore `json:"indicators"`
}

type CHSIndicatorScore struct {
	Code          string   `json:"code"`
	Label         string   `json:"label"`
	RawValue      *float64 `json:"raw_value"`
	Score         *float64 `json:"score"`
	Weight        float64  `json:"weight"`
	WeightedScore float64  `json:"weighted_score"`
	Status        string   `json:"status"`
	Message       string   `json:"message,omitempty"`
}

type CHSLoanRiskMetrics struct {
	TotalRemainingPrincipal int64
	BadRemainingPrincipal   int64
}

type CHSOnTimePaymentMetrics struct {
	TotalDue int64
	OnTime   int64
}

type CHSMemberActivityMetrics struct {
	TotalMembers  int64
	ActiveMembers int64
}

type CHSTransactionGrowthMetrics struct {
	CurrentTransactions  int64
	PreviousTransactions int64
}

type CHSDataCompletenessMetrics struct {
	TotalFields  int64
	FilledFields int64
}

type CHSSyncTimelinessMetrics struct {
	TotalTransactions  int64
	TimelyTransactions int64
}

type CHSConsistencyMetrics struct {
	TotalRecords      int64
	DuplicateRecords  int64
	ConsistentRecords int64
}

type DashboardSummaryRequest struct {
	AuthContext
	Period string `form:"period"` // YYYY-MM
}

type DashboardSummaryResponse struct {
	Period      string                 `json:"period"`
	PeriodLabel string                 `json:"period_label"`
	CHS         DashboardCHSSummary    `json:"chs"`
	Members     DashboardMemberSummary `json:"members"`
}

type DashboardCHSSummary struct {
	Score        float64 `json:"score"`
	DisplayScore int     `json:"display_score"`
	Grade        string  `json:"grade"`
	Category     string  `json:"category"`
	Status       string  `json:"status"`
}

type DashboardMemberSummary struct {
	Active     int64 `json:"active"`
	Registered int64 `json:"registered"`
}

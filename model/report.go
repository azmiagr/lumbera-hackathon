package model

type FinancialReportRequest struct {
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

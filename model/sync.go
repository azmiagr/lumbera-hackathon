package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	SyncOperationCreateSavingsTransaction     = "CREATE_SAVINGS_TRANSACTION"
	SyncOperationCreateLoanTransaction        = "CREATE_LOAN_TRANSACTION"
	SyncOperationCreateInstallmentTransaction = "CREATE_INSTALLMENT_TRANSACTION"
	SyncOperationCreateCashWithdrawal         = "CREATE_CASH_WITHDRAWAL"
	SyncOperationReverseTransaction           = "REVERSE_TRANSACTION"
	SyncOperationCreateProduct                = "CREATE_PRODUCT"
	SyncOperationCreateStockIn                = "CREATE_STOCK_IN"
	SyncOperationCreateStockAdjustment        = "CREATE_STOCK_ADJUSTMENT"
	SyncOperationCreateStoreSale              = "CREATE_STORE_SALE"

	SyncStatusSynced         = "synced"
	SyncStatusDuplicate      = "duplicate"
	SyncStatusRejected       = "rejected"
	SyncStatusRetryableError = "retryable_error"
)

type SyncPushRequest struct {
	AuthContext
	DeviceID   string              `json:"device_id"`
	BatchID    string              `json:"batch_id"`
	Operations []SyncOperationItem `json:"operations"`
}

type SyncOperationItem struct {
	ClientOperationID string          `json:"client_operation_id"`
	OperationType     string          `json:"operation_type"`
	RecordedAt        *time.Time      `json:"recorded_at"`
	Payload           json.RawMessage `json:"payload"`
}

type SyncPushResponse struct {
	BatchID        string                `json:"batch_id"`
	ServerTime     time.Time             `json:"server_time"`
	Results        []SyncOperationResult `json:"results"`
	NextPullCursor string                `json:"next_pull_cursor"`
}

type SyncOperationResult struct {
	ClientOperationID string      `json:"client_operation_id"`
	OperationType     string      `json:"operation_type"`
	Status            string      `json:"status"`
	EntityType        string      `json:"entity_type,omitempty"`
	ServerID          *uuid.UUID  `json:"server_id,omitempty"`
	ClientReferenceID string      `json:"client_reference_id,omitempty"`
	SyncedAt          *time.Time  `json:"synced_at,omitempty"`
	ErrorCode         string      `json:"error_code,omitempty"`
	Message           string      `json:"message,omitempty"`
	Data              interface{} `json:"data,omitempty"`
}

type SyncStatusRequest struct {
	AuthContext
	ClientTransactionIDs []string `form:"client_transaction_ids[]"`
	ClientReferenceIDs   []string `form:"client_reference_ids[]"`
	ClientSaleIDs        []string `form:"client_sale_ids[]"`
}

type SyncConfigResponse struct {
	MaxBatchSize        int      `json:"max_batch_size"`
	RetryBackoffSeconds []int    `json:"retry_backoff_seconds"`
	SupportedOperations []string `json:"supported_operations"`
}

type SyncBootstrapResponse struct {
	ServerTime   time.Time                     `json:"server_time"`
	Members      []TransactionMemberResponse   `json:"members"`
	Products     []ProductResponse             `json:"products"`
	Transactions []TransactionListItemResponse `json:"transactions"`
	Movements    []StockMovementResponse       `json:"movements"`
	SyncCursor   string                        `json:"sync_cursor"`
}

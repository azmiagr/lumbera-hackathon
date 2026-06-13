package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/model"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const syncMaxBatchSize = 50

type ISyncService interface {
	Push(req model.SyncPushRequest) (*model.SyncPushResponse, error)
	GetConfig(req model.AuthContext) (*model.SyncConfigResponse, error)
	GetBootstrap(req model.AuthContext) (*model.SyncBootstrapResponse, error)
	GetStatus(req model.SyncStatusRequest) ([]model.SyncOperationResult, error)
}

type SyncService struct {
	deps serviceDependency
}

func NewSyncService(deps serviceDependency) ISyncService {
	return &SyncService{deps: deps}
}

func (s *SyncService) Push(req model.SyncPushRequest) (*model.SyncPushResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if len(req.Operations) > syncMaxBatchSize {
		return nil, appErrors.BadRequest("maksimal 50 operasi per batch")
	}

	results := make([]model.SyncOperationResult, 0, len(req.Operations))
	for _, op := range req.Operations {
		results = append(results, s.processOperation(req.AuthContext, op))
	}

	now := time.Now()
	return &model.SyncPushResponse{
		BatchID:        strings.TrimSpace(req.BatchID),
		ServerTime:     now,
		Results:        results,
		NextPullCursor: now.Format(time.RFC3339Nano),
	}, nil
}

func (s *SyncService) GetConfig(req model.AuthContext) (*model.SyncConfigResponse, error) {
	if err := validateStoreAccess(req); err != nil {
		return nil, err
	}

	return &model.SyncConfigResponse{
		MaxBatchSize:        syncMaxBatchSize,
		RetryBackoffSeconds: []int{5, 15, 60, 300},
		SupportedOperations: []string{
			model.SyncOperationCreateSavingsTransaction,
			model.SyncOperationCreateLoanTransaction,
			model.SyncOperationCreateInstallmentTransaction,
			model.SyncOperationCreateCashWithdrawal,
			model.SyncOperationReverseTransaction,
			model.SyncOperationCreateProduct,
			model.SyncOperationCreateStockIn,
			model.SyncOperationCreateStockAdjustment,
			model.SyncOperationCreateStoreSale,
		},
	}, nil
}

func (s *SyncService) GetBootstrap(req model.AuthContext) (*model.SyncBootstrapResponse, error) {
	if err := validateStoreAccess(req); err != nil {
		return nil, err
	}

	transactionService := NewTransactionService(s.deps)
	storeService := NewStoreService(s.deps)

	members, err := transactionService.SearchTransactionMembers(model.SearchTransactionMembersRequest{
		AuthContext: req,
		Limit:       syncMaxBatchSize,
	})
	if err != nil {
		return nil, err
	}

	products, err := storeService.ListProducts(model.ListProductsRequest{
		AuthContext: req,
		Page:        1,
		Limit:       syncMaxBatchSize,
	})
	if err != nil {
		return nil, err
	}

	transactions, err := transactionService.ListTransactions(model.ListTransactionsRequest{
		AuthContext: req,
		Page:        1,
		Limit:       syncMaxBatchSize,
	})
	if err != nil {
		return nil, err
	}

	movements, err := storeService.ListStockMovements(model.ListStockMovementsRequest{
		AuthContext: req,
		Page:        1,
		Limit:       syncMaxBatchSize,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &model.SyncBootstrapResponse{
		ServerTime:   now,
		Members:      members,
		Products:     products.Items,
		Transactions: transactions.Items,
		Movements:    movements.Items,
		SyncCursor:   now.Format(time.RFC3339Nano),
	}, nil
}

func (s *SyncService) GetStatus(req model.SyncStatusRequest) ([]model.SyncOperationResult, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	results := make([]model.SyncOperationResult, 0, len(req.ClientTransactionIDs)+len(req.ClientReferenceIDs)+len(req.ClientSaleIDs))

	for _, rawID := range req.ClientTransactionIDs {
		clientID := strings.TrimSpace(rawID)
		if clientID == "" {
			continue
		}
		result := model.SyncOperationResult{
			Status:            model.SyncStatusRejected,
			EntityType:        "transaction",
			ClientReferenceID: clientID,
			Message:           "transaksi belum ditemukan",
		}

		transaction, err := s.deps.repository.TransactionRepository.GetTransactionByClientID(tx, req.CooperativeID, clientID)
		if err == nil {
			detail, detailErr := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
			if detailErr != nil {
				return nil, appErrors.InternalServer("gagal mengambil status transaksi")
			}
			serverID := transaction.TransactionID
			result.Status = model.SyncStatusSynced
			result.ServerID = &serverID
			result.SyncedAt = transaction.SyncedAt
			result.Message = ""
			result.Data = detail
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal mengambil status transaksi")
		}

		results = append(results, result)
	}

	for _, rawID := range req.ClientReferenceIDs {
		clientID := strings.TrimSpace(rawID)
		if clientID == "" {
			continue
		}
		result := model.SyncOperationResult{
			Status:            model.SyncStatusRejected,
			EntityType:        "stock_movement",
			ClientReferenceID: clientID,
			Message:           "mutasi stok belum ditemukan",
		}

		movement, err := s.deps.repository.StoreRepository.GetStockMovementByClientReferenceID(tx, req.CooperativeID, clientID)
		if err == nil {
			detail, detailErr := s.deps.repository.StoreRepository.GetStockMovementResponseByID(tx, req.CooperativeID, movement.StockMovementID)
			if detailErr != nil {
				return nil, appErrors.InternalServer("gagal mengambil status mutasi stok")
			}
			serverID := movement.StockMovementID
			result.Status = model.SyncStatusSynced
			result.ServerID = &serverID
			result.SyncedAt = movement.SyncedAt
			result.Message = ""
			result.Data = detail
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal mengambil status mutasi stok")
		}

		results = append(results, result)
	}

	for _, rawID := range req.ClientSaleIDs {
		clientID := strings.TrimSpace(rawID)
		if clientID == "" {
			continue
		}
		result := model.SyncOperationResult{
			Status:            model.SyncStatusRejected,
			EntityType:        "store_sale",
			ClientReferenceID: clientID,
			Message:           "penjualan belum ditemukan",
		}

		sale, err := s.deps.repository.StoreRepository.GetSaleByClientID(tx, req.CooperativeID, clientID)
		if err == nil {
			detail, detailErr := s.deps.repository.StoreRepository.GetSaleDetail(tx, req.CooperativeID, sale.StoreSaleID)
			if detailErr != nil {
				return nil, appErrors.InternalServer("gagal mengambil status penjualan")
			}
			serverID := sale.StoreSaleID
			result.Status = model.SyncStatusSynced
			result.ServerID = &serverID
			result.Message = ""
			result.Data = detail
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal mengambil status penjualan")
		}

		results = append(results, result)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil status sinkronisasi")
	}

	return results, nil
}

func (s *SyncService) processOperation(auth model.AuthContext, op model.SyncOperationItem) model.SyncOperationResult {
	result := model.SyncOperationResult{
		ClientOperationID: strings.TrimSpace(op.ClientOperationID),
		OperationType:     strings.ToUpper(strings.TrimSpace(op.OperationType)),
	}

	switch result.OperationType {
	case model.SyncOperationCreateSavingsTransaction:
		var req model.CreateSavingsTransactionRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewTransactionService(s.deps).CreateSavingsTransaction(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedTransactionResult(result, response, response.ClientTransactionID)

	case model.SyncOperationCreateLoanTransaction:
		var req model.CreateLoanTransactionRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewTransactionService(s.deps).CreateLoanTransaction(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedTransactionResult(result, response, response.ClientTransactionID)

	case model.SyncOperationCreateInstallmentTransaction:
		var req model.CreateInstallmentTransactionRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewTransactionService(s.deps).CreateInstallmentTransaction(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedTransactionResult(result, &response.TransactionResponse, response.ClientTransactionID)

	case model.SyncOperationCreateCashWithdrawal:
		var req model.CreateCashWithdrawalTransactionRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewTransactionService(s.deps).CreateCashWithdrawalTransaction(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedTransactionResult(result, response, response.ClientTransactionID)

	case model.SyncOperationReverseTransaction:
		var req syncReverseTransactionRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		reversalReq := model.ReverseTransactionRequest{
			AuthContext:         auth,
			TransactionID:       req.TransactionID,
			Reason:              req.Reason,
			RecordedAt:          chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt),
			IsOfflineCreated:    true,
			ClientTransactionID: req.ClientTransactionID,
		}
		response, err := NewTransactionService(s.deps).ReverseTransaction(reversalReq)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedTransactionResult(result, &response.ReversalTransaction, response.ReversalTransaction.ClientTransactionID)

	case model.SyncOperationCreateProduct:
		var req model.CreateProductRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewStoreService(s.deps).CreateProduct(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		serverID := response.ProductID
		result.Status = model.SyncStatusSynced
		result.EntityType = "product"
		result.ServerID = &serverID
		result.ClientReferenceID = req.ClientReferenceID
		result.Data = response
		return result

	case model.SyncOperationCreateStockIn:
		var req model.CreateStockInRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewStoreService(s.deps).CreateStockIn(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedStockMovementResult(result, response, response.ClientReferenceID)

	case model.SyncOperationCreateStockAdjustment:
		var req model.CreateStockAdjustmentRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		req.IsOfflineCreated = true
		response, err := NewStoreService(s.deps).CreateStockAdjustment(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		return syncedStockMovementResult(result, response, response.ClientReferenceID)

	case model.SyncOperationCreateStoreSale:
		var req model.CreateStoreSaleRequest
		if err := decodeSyncPayload(op.Payload, &req); err != nil {
			return rejectedSyncResult(result, "INVALID_PAYLOAD", err.Error())
		}
		req.AuthContext = auth
		req.RecordedAt = chooseSyncRecordedAt(op.RecordedAt, req.RecordedAt)
		response, err := NewStoreService(s.deps).CreateStoreSale(req)
		if err != nil {
			return failedSyncResult(result, err)
		}
		serverID := response.StoreSaleID
		result.Status = model.SyncStatusSynced
		result.EntityType = "store_sale"
		result.ServerID = &serverID
		result.ClientReferenceID = response.ClientSaleID
		result.Data = response
		return result

	default:
		return rejectedSyncResult(result, "UNSUPPORTED_OPERATION", "jenis operasi sinkronisasi tidak didukung")
	}
}

type syncReverseTransactionRequest struct {
	TransactionID       uuid.UUID  `json:"transaction_id"`
	Reason              string     `json:"reason"`
	RecordedAt          *time.Time `json:"recorded_at"`
	ClientTransactionID string     `json:"client_transaction_id"`
}

func decodeSyncPayload(payload json.RawMessage, target interface{}) error {
	if len(payload) == 0 {
		return errors.New("payload wajib diisi")
	}
	return json.Unmarshal(payload, target)
}

func chooseSyncRecordedAt(operationRecordedAt, payloadRecordedAt *time.Time) *time.Time {
	if operationRecordedAt != nil {
		return operationRecordedAt
	}
	return payloadRecordedAt
}

func syncedTransactionResult(result model.SyncOperationResult, response *model.TransactionResponse, clientReferenceID string) model.SyncOperationResult {
	serverID := response.TransactionID
	result.Status = model.SyncStatusSynced
	result.EntityType = "transaction"
	result.ServerID = &serverID
	result.ClientReferenceID = clientReferenceID
	result.SyncedAt = response.SyncedAt
	result.Data = response
	return result
}

func syncedStockMovementResult(result model.SyncOperationResult, response *model.StockMovementResponse, clientReferenceID string) model.SyncOperationResult {
	serverID := response.StockMovementID
	result.Status = model.SyncStatusSynced
	result.EntityType = "stock_movement"
	result.ServerID = &serverID
	result.ClientReferenceID = clientReferenceID
	result.SyncedAt = response.SyncedAt
	result.Data = response
	return result
}

func rejectedSyncResult(result model.SyncOperationResult, code string, message string) model.SyncOperationResult {
	result.Status = model.SyncStatusRejected
	result.ErrorCode = code
	result.Message = message
	return result
}

func failedSyncResult(result model.SyncOperationResult, err error) model.SyncOperationResult {
	if appErr, ok := err.(*appErrors.AppError); ok {
		result.Message = appErr.Message
		if appErr.Code == http.StatusInternalServerError {
			result.Status = model.SyncStatusRetryableError
			result.ErrorCode = "RETRYABLE_ERROR"
			return result
		}
		result.Status = model.SyncStatusRejected
		result.ErrorCode = "REJECTED"
		return result
	}

	result.Status = model.SyncStatusRetryableError
	result.ErrorCode = "RETRYABLE_ERROR"
	result.Message = err.Error()
	return result
}

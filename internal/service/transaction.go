package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const genesisTransactionHash = "0000000000000000000000000000000000000000000000000000000000000000"

type ITransactionService interface {
	SearchTransactionMembers(req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error)
	CreateSavingsTransaction(req model.CreateSavingsTransactionRequest) (*model.TransactionResponse, error)
	ListTransactions(req model.ListTransactionsRequest) (*model.ListTransactionsResponse, error)
	CreateLoanTransaction(req model.CreateLoanTransactionRequest) (*model.TransactionResponse, error)
}

type TransactionService struct {
	deps serviceDependency
}

func NewTransactionService(deps serviceDependency) ITransactionService {
	return &TransactionService{deps: deps}
}

func (s *TransactionService) SearchTransactionMembers(req model.SearchTransactionMembersRequest) ([]model.TransactionMemberResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mencari anggota transaksi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	members, err := s.deps.repository.MemberRepository.SearchTransactionMembers(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mencari anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mencari anggota")
	}

	return members, nil
}

func (s *TransactionService) CreateSavingsTransaction(req model.CreateSavingsTransactionRequest) (*model.TransactionResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mencatat transaksi")
	}

	if req.MemberID == uuid.Nil {
		return nil, appErrors.BadRequest("anggota wajib dipilih")
	}

	if req.Amount <= 0 {
		return nil, appErrors.BadRequest("nominal transaksi wajib lebih dari 0")
	}

	transactionType, err := mapSavingsType(req.SavingsType)
	if err != nil {
		return nil, err
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	now := time.Now()
	clientTransactionID := strings.TrimSpace(req.ClientTransactionID)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	if clientTransactionID != "" {
		existingTransaction, err := s.deps.repository.TransactionRepository.GetTransactionByClientID(tx, req.CooperativeID, clientTransactionID)
		if err == nil {
			detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, existingTransaction.TransactionID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
			}

			err = tx.Commit().Error
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil transaksi")
			}

			return mapDetailedTransactionResponse(detail, existingTransaction.PrevHash), nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa transaksi duplikat")
		}
	}

	_, err = s.deps.repository.MemberRepository.GetActiveMember(tx, req.CooperativeID, req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal memvalidasi anggota")
	}

	prevHash := genesisTransactionHash
	latestTransaction, err := s.deps.repository.TransactionRepository.GetLatestTransactionForUpdate(tx, req.CooperativeID)
	if err == nil {
		prevHash = latestTransaction.CurrentHash
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil transaksi terakhir")
	}

	transaction := &entity.Transaction{
		TransactionID:       uuid.New(),
		CooperativeID:       req.CooperativeID,
		MemberID:            req.MemberID,
		OfficerID:           req.UserID,
		TransactionType:     transactionType,
		Amount:              req.Amount,
		Description:         strings.TrimSpace(req.Description),
		RecordedAt:          recordedAt,
		SyncedAt:            &now,
		PrevHash:            prevHash,
		IsOfflineCreated:    req.IsOfflineCreated,
		ClientTransactionID: clientTransactionID,
	}

	transaction.CurrentHash = buildTransactionHash(transaction)

	err = s.deps.repository.TransactionRepository.CreateTransaction(tx, transaction)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi")
	}

	detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi")
	}

	return mapDetailedTransactionResponse(detail, transaction.PrevHash), nil
}

func (s *TransactionService) ListTransactions(req model.ListTransactionsRequest) (*model.ListTransactionsResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat melihat transaksi")
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	items, total, err := s.deps.repository.TransactionRepository.ListTransactions(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil transaksi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil transaksi")
	}

	for i := range items {
		enrichTransactionListItem(&items[i])
	}

	return &model.ListTransactionsResponse{
		Items: items,
		Page:  req.Page,
		Limit: req.Limit,
		Total: total,
	}, nil
}

func (s *TransactionService) CreateLoanTransaction(req model.CreateLoanTransactionRequest) (*model.TransactionResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mencatat transaksi")
	}

	if req.MemberID == uuid.Nil {
		return nil, appErrors.BadRequest("anggota wajib dipilih")
	}

	if req.Amount <= 0 {
		return nil, appErrors.BadRequest("nominal pinjaman wajib lebih dari 0")
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	now := time.Now()
	clientTransactionID := strings.TrimSpace(req.ClientTransactionID)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	if clientTransactionID != "" {
		existingTransaction, err := s.deps.repository.TransactionRepository.GetTransactionByClientID(tx, req.CooperativeID, clientTransactionID)
		if err == nil {
			detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, existingTransaction.TransactionID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
			}

			summary, err := s.deps.repository.TransactionRepository.GetMemberTransactionSummary(tx, req.CooperativeID, existingTransaction.MemberID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil ringkasan anggota")
			}

			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil transaksi")
			}

			return mapDetailedTransactionResponseWithSummary(detail, existingTransaction.PrevHash, summary), nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa transaksi duplikat")
		}
	}

	_, err := s.deps.repository.MemberRepository.GetActiveMember(tx, req.CooperativeID, req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal memvalidasi anggota")
	}

	summary, err := s.deps.repository.TransactionRepository.GetMemberTransactionSummary(tx, req.CooperativeID, req.MemberID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan anggota")
	}

	financialConfig, err := s.deps.repository.FinancialConfigurationRepository.GetFinancialConfiguration(tx, model.GetFinancialConfigurationParam{
		CooperativeID: req.CooperativeID,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil konfigurasi pinjaman")
	}
	if financialConfig != nil && financialConfig.MaxLoanAmountPerMember > 0 && summary.LoanOutstanding+req.Amount > financialConfig.MaxLoanAmountPerMember {
		return nil, appErrors.BadRequest("nominal pinjaman melebihi batas pinjaman anggota")
	}

	prevHash := genesisTransactionHash
	latestTransaction, err := s.deps.repository.TransactionRepository.GetLatestTransactionForUpdate(tx, req.CooperativeID)
	if err == nil {
		prevHash = latestTransaction.CurrentHash
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil transaksi terakhir")
	}

	transaction := &entity.Transaction{
		TransactionID:       uuid.New(),
		CooperativeID:       req.CooperativeID,
		MemberID:            req.MemberID,
		OfficerID:           req.UserID,
		TransactionType:     constants.TransactionTypeLoan,
		Amount:              req.Amount,
		Description:         strings.TrimSpace(req.Description),
		RecordedAt:          recordedAt,
		SyncedAt:            &now,
		PrevHash:            prevHash,
		IsOfflineCreated:    req.IsOfflineCreated,
		ClientTransactionID: clientTransactionID,
	}
	transaction.CurrentHash = buildTransactionHash(transaction)

	err = s.deps.repository.TransactionRepository.CreateTransaction(tx, transaction)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi pinjaman")
	}

	detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
	}

	summary.LoanOutstanding += req.Amount

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi pinjaman")
	}

	return mapDetailedTransactionResponseWithSummary(detail, transaction.PrevHash, summary), nil
}

func mapDetailedTransactionResponseWithSummary(item *model.TransactionListItemResponse, prevHash string, summary *model.TransactionMemberSummaryResponse) *model.TransactionResponse {
	response := mapDetailedTransactionResponse(item, prevHash)
	if summary != nil {
		response.MemberSavingsBalance = summary.SavingsBalance
		response.MemberLoanOutstanding = summary.LoanOutstanding
	}
	return response
}

func mapSavingsType(savingsType string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(savingsType)) {
	case "POKOK":
		return constants.TransactionTypeSavingsPrincipal, nil
	case "WAJIB":
		return constants.TransactionTypeSavingsMandatory, nil
	case "SUKARELA":
		return constants.TransactionTypeSavingsVoluntary, nil
	default:
		return "", appErrors.BadRequest("jenis simpanan tidak valid")
	}
}

func buildTransactionHash(transaction *entity.Transaction) string {
	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%d|%s|%s|%t|%s",
		transaction.PrevHash,
		transaction.TransactionID.String(),
		transaction.CooperativeID.String(),
		transaction.MemberID.String(),
		transaction.OfficerID.String(),
		transaction.TransactionType,
		transaction.Amount,
		transaction.Description,
		transaction.RecordedAt.UTC().Format(time.RFC3339Nano),
		transaction.IsOfflineCreated,
		transaction.ClientTransactionID,
	)

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func mapDetailedTransactionResponse(item *model.TransactionListItemResponse, prevHash string) *model.TransactionResponse {
	enrichTransactionListItem(item)

	return &model.TransactionResponse{
		TransactionID:        item.TransactionID,
		CooperativeID:        item.CooperativeID,
		MemberID:             item.MemberID,
		MemberName:           item.MemberName,
		MemberNumber:         item.MemberNumber,
		MemberMCSGrade:       item.MemberMCSGrade,
		OfficerID:            item.OfficerID,
		OfficerName:          item.OfficerName,
		TransactionType:      item.TransactionType,
		TransactionTypeLabel: item.TransactionTypeLabel,
		TransactionGroup:     item.TransactionGroup,
		Amount:               item.Amount,
		Description:          item.Description,
		RecordedAt:           item.RecordedAt,
		SyncedAt:             item.SyncedAt,
		PrevHash:             prevHash,
		CurrentHash:          item.CurrentHash,
		HashPreview:          item.HashPreview,
		IsOfflineCreated:     item.IsOfflineCreated,
		ClientTransactionID:  item.ClientTransactionID,
		SyncStatus:           item.SyncStatus,
	}
}

func enrichTransactionListItem(item *model.TransactionListItemResponse) {
	item.TransactionGroup = getTransactionGroup(item.TransactionType)
	item.TransactionTypeLabel = getTransactionTypeLabel(item.TransactionType)
	item.HashPreview = buildHashPreview(item.CurrentHash)
	item.SyncStatus = constants.SyncStatusSynced
}

func getTransactionGroup(transactionType string) string {
	switch transactionType {
	case constants.TransactionTypeSavingsPrincipal,
		constants.TransactionTypeSavingsMandatory,
		constants.TransactionTypeSavingsVoluntary:
		return constants.TransactionGroupSavings
	case constants.TransactionTypeLoan:
		return constants.TransactionGroupLoan
	case constants.TransactionTypeInstallment:
		return constants.TransactionGroupInstallment
	default:
		return constants.TransactionGroupAll
	}
}

func getTransactionTypeLabel(transactionType string) string {
	switch transactionType {
	case constants.TransactionTypeSavingsPrincipal:
		return "Simpanan Pokok"
	case constants.TransactionTypeSavingsMandatory:
		return "Simpanan Wajib"
	case constants.TransactionTypeSavingsVoluntary:
		return "Simpanan Sukarela"
	case constants.TransactionTypeLoan:
		return "Pinjaman"
	case constants.TransactionTypeInstallment:
		return "Angsuran Pinjaman"
	default:
		return transactionType
	}
}

func buildHashPreview(hash string) string {
	if len(hash) <= 10 {
		return hash
	}

	return "SHA-256: " + hash[:8] + "..."
}

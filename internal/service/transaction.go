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
	CreateInstallmentTransaction(req model.CreateInstallmentTransactionRequest) (*model.InstallmentTransactionResponse, error)
	CreateCashWithdrawalTransaction(req model.CreateCashWithdrawalTransactionRequest) (*model.TransactionResponse, error)
	ReverseTransaction(req model.ReverseTransactionRequest) (*model.TransactionReversalResponse, error)
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
			if existingTransaction.TransactionType != constants.TransactionTypeCashWithdrawal {
				return nil, appErrors.BadRequest("client_transaction_id sudah digunakan untuk transaksi lain")
			}

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

			loan, err := s.deps.repository.LoanRepository.GetLoanByDisbursementTransactionID(tx, existingTransaction.TransactionID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.InternalServer("gagal mengambil pinjaman")
			}

			var loanSummary *model.LoanSummaryResponse
			if loan != nil {
				loanSummary, err = s.deps.repository.LoanRepository.GetLoanSummary(tx, loan.LoanID, recordedAt)
				if err != nil {
					return nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman")
				}
			}

			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil transaksi")
			}

			return mapDetailedTransactionResponseWithSummaryAndLoan(detail, existingTransaction.PrevHash, summary, loanSummary), nil
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

	if financialConfig == nil || financialConfig.MaxLoanTermMonths <= 0 {
		return nil, appErrors.BadRequest("konfigurasi tenor pinjaman belum tersedia")
	}

	if _, err := s.deps.repository.LoanRepository.GetActiveLoanByMemberForUpdate(tx, req.CooperativeID, req.MemberID); err == nil {
		return nil, appErrors.BadRequest("anggota masih memiliki pinjaman aktif")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi pinjaman aktif")
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

	loanNumber, err := s.deps.repository.LoanRepository.GenerateNextLoanNumber(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat nomor pinjaman")
	}

	loan, schedules := buildLoanAccountAndSchedules(req, transaction.TransactionID, loanNumber, financialConfig, recordedAt)

	if err := s.deps.repository.LoanRepository.CreateLoanAccount(tx, loan); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan pinjaman")
	}

	if err := s.deps.repository.LoanRepository.CreateLoanInstallmentSchedules(tx, schedules); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan jadwal angsuran")
	}

	detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
	}

	summary.LoanOutstanding += req.Amount
	loanSummary, err := s.deps.repository.LoanRepository.GetLoanSummary(tx, loan.LoanID, recordedAt)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi pinjaman")
	}

	return mapDetailedTransactionResponseWithSummaryAndLoan(detail, transaction.PrevHash, summary, loanSummary), nil
}

func mapDetailedTransactionResponseWithSummary(item *model.TransactionListItemResponse, prevHash string, summary *model.TransactionMemberSummaryResponse) *model.TransactionResponse {
	response := mapDetailedTransactionResponse(item, prevHash)
	if summary != nil {
		response.MemberSavingsBalance = summary.SavingsBalance
		response.MemberLoanOutstanding = summary.LoanOutstanding
	}
	return response
}

func (s *TransactionService) CreateInstallmentTransaction(req model.CreateInstallmentTransactionRequest) (*model.InstallmentTransactionResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mencatat transaksi")
	}

	if req.LoanID == uuid.Nil {
		return nil, appErrors.BadRequest("pinjaman wajib dipilih")
	}

	if req.Amount <= 0 {
		return nil, appErrors.BadRequest("nominal angsuran wajib lebih dari 0")
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

			loanSummary, allocations, err := s.getInstallmentResponseMetadata(tx, req.CooperativeID, existingTransaction.TransactionID, existingTransaction.RecordedAt)
			if err != nil {
				return nil, err
			}

			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil transaksi")
			}

			response := mapDetailedInstallmentResponse(detail, existingTransaction.PrevHash, loanSummary, allocations)
			return response, nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa transaksi duplikat")
		}
	}

	loan, err := s.deps.repository.LoanRepository.GetLoanByIDForUpdate(tx, req.CooperativeID, req.LoanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("pinjaman tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil pinjaman")
	}

	if loan.Status != "ACTIVE" {
		return nil, appErrors.BadRequest("pinjaman tidak aktif")
	}

	schedules, err := s.deps.repository.LoanRepository.ListPayableSchedulesForUpdate(tx, loan.LoanID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil jadwal angsuran")
	}

	remainingPayable := sumScheduleRemaining(schedules)
	if remainingPayable <= 0 {
		return nil, appErrors.BadRequest("pinjaman sudah lunas")
	}

	if req.Amount > remainingPayable {
		return nil, appErrors.BadRequest("nominal angsuran melebihi sisa tagihan")
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
		MemberID:            loan.MemberID,
		OfficerID:           req.UserID,
		TransactionType:     constants.TransactionTypeInstallment,
		Amount:              req.Amount,
		Description:         strings.TrimSpace(req.Description),
		RecordedAt:          recordedAt,
		SyncedAt:            &now,
		PrevHash:            prevHash,
		IsOfflineCreated:    req.IsOfflineCreated,
		ClientTransactionID: clientTransactionID,
	}
	transaction.CurrentHash = buildTransactionHash(transaction)

	if err := s.deps.repository.TransactionRepository.CreateTransaction(tx, transaction); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi angsuran")
	}

	allocations, updatedSchedules, remainingPayment := allocateInstallmentPayment(schedules, req.Amount, recordedAt)
	if remainingPayment > 0 {
		return nil, appErrors.BadRequest("nominal angsuran melebihi sisa tagihan")
	}

	for i := range allocations {
		allocations[i].TransactionID = transaction.TransactionID
		if err := s.deps.repository.LoanRepository.CreateLoanPaymentAllocation(tx, &allocations[i]); err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan alokasi angsuran")
		}
	}

	for i := range updatedSchedules {
		if err := s.deps.repository.LoanRepository.UpdateLoanSchedule(tx, &updatedSchedules[i]); err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui jadwal angsuran")
		}
	}

	if remainingPayable-req.Amount == 0 {
		loan.Status = "PAID"
		if err := s.deps.repository.LoanRepository.UpdateLoanAccount(tx, loan); err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui status pinjaman")
		}
	}

	detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
	}

	loanSummary, err := s.deps.repository.LoanRepository.GetLoanSummary(tx, loan.LoanID, recordedAt)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman")
	}

	allocationResponses := mapInstallmentAllocationResponses(updatedSchedules, allocations)

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi angsuran")
	}

	return mapDetailedInstallmentResponse(detail, transaction.PrevHash, loanSummary, allocationResponses), nil
}

func (s *TransactionService) CreateCashWithdrawalTransaction(req model.CreateCashWithdrawalTransactionRequest) (*model.TransactionResponse, error) {
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
		return nil, appErrors.BadRequest("nominal tarik tunai wajib lebih dari 0")
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
			if existingTransaction.TransactionType != constants.TransactionTypeSavingsPrincipal &&
				existingTransaction.TransactionType != constants.TransactionTypeSavingsMandatory &&
				existingTransaction.TransactionType != constants.TransactionTypeSavingsVoluntary {
				return nil, appErrors.BadRequest("client_transaction_id sudah digunakan untuk transaksi lain")
			}

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

	if summary.SavingsBalance <= 0 {
		return nil, appErrors.BadRequest("saldo anggota tidak mencukupi")
	}

	if req.Amount > summary.SavingsBalance {
		return nil, appErrors.BadRequest("saldo anggota tidak mencukupi")
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
		TransactionType:     constants.TransactionTypeCashWithdrawal,
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
		return nil, appErrors.InternalServer("gagal menyimpan transaksi tarik tunai")
	}

	detail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, req.CooperativeID, transaction.TransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil detail transaksi")
	}

	summary.SavingsBalance -= req.Amount

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi tarik tunai")
	}

	return mapDetailedTransactionResponseWithSummary(detail, transaction.PrevHash, summary), nil
}

func (s *TransactionService) ReverseTransaction(req model.ReverseTransactionRequest) (*model.TransactionReversalResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat membatalkan transaksi")
	}

	if req.TransactionID == uuid.Nil {
		return nil, appErrors.BadRequest("transaksi wajib dipilih")
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, appErrors.BadRequest("alasan pembatalan wajib diisi")
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	clientTransactionID := strings.TrimSpace(req.ClientTransactionID)
	if clientTransactionID == "" {
		clientTransactionID = "REV-" + req.TransactionID.String()
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	original, err := s.deps.repository.TransactionRepository.GetTransactionForUpdate(tx, req.CooperativeID, req.TransactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("transaksi tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil transaksi")
	}

	existingReversal, err := s.deps.repository.TransactionRepository.GetReversalByOriginalTransactionID(tx, req.CooperativeID, original.TransactionID)
	if err == nil {
		response, err := s.buildTransactionReversalResponse(tx, req.CooperativeID, original.TransactionID, existingReversal.ReversalTransactionID, existingReversal.Reason, original.PrevHash, "")
		if err != nil {
			return nil, err
		}

		if err := tx.Commit().Error; err != nil {
			return nil, appErrors.InternalServer("gagal mengambil pembatalan transaksi")
		}

		return response, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memeriksa pembatalan transaksi")
	}

	_, err = s.deps.repository.TransactionRepository.GetReversalByReversalTransactionID(tx, req.CooperativeID, original.TransactionID)
	if err == nil {
		return nil, appErrors.BadRequest("transaksi pembatalan tidak dapat dibatalkan ulang")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memeriksa transaksi pembatalan")
	}

	if err := s.validateTransactionReversal(tx, original); err != nil {
		return nil, err
	}

	if clientTransactionID != "" {
		existingTransaction, err := s.deps.repository.TransactionRepository.GetTransactionByClientID(tx, req.CooperativeID, clientTransactionID)
		if err == nil && existingTransaction.TransactionID != original.TransactionID {
			return nil, appErrors.BadRequest("client_transaction_id sudah digunakan")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa transaksi duplikat")
		}
	}

	prevHash := genesisTransactionHash
	latestTransaction, err := s.deps.repository.TransactionRepository.GetLatestTransactionForUpdate(tx, req.CooperativeID)
	if err == nil {
		prevHash = latestTransaction.CurrentHash
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil transaksi terakhir")
	}

	now := time.Now()
	reversalTransaction := &entity.Transaction{
		TransactionID:       uuid.New(),
		CooperativeID:       req.CooperativeID,
		MemberID:            original.MemberID,
		OfficerID:           req.UserID,
		TransactionType:     original.TransactionType,
		Amount:              -original.Amount,
		Description:         fmt.Sprintf("Pembatalan transaksi %s: %s", buildHashPreview(original.CurrentHash), reason),
		RecordedAt:          recordedAt,
		SyncedAt:            &now,
		PrevHash:            prevHash,
		IsOfflineCreated:    req.IsOfflineCreated,
		ClientTransactionID: clientTransactionID,
	}
	reversalTransaction.CurrentHash = buildTransactionHash(reversalTransaction)

	err = s.deps.repository.TransactionRepository.CreateTransaction(tx, reversalTransaction)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan transaksi pembatalan")
	}

	err = s.applyTransactionReversalEffects(tx, original, reversalTransaction)
	if err != nil {
		return nil, err
	}

	reversal := &entity.TransactionReversal{
		TransactionReversalID: uuid.New(),
		CooperativeID:         req.CooperativeID,
		OriginalTransactionID: original.TransactionID,
		ReversalTransactionID: reversalTransaction.TransactionID,
		Reason:                reason,
		CreatedBy:             req.UserID,
	}
	err = s.deps.repository.TransactionRepository.CreateTransactionReversal(tx, reversal)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan data pembatalan")
	}

	response, err := s.buildTransactionReversalResponse(tx, req.CooperativeID, original.TransactionID, reversalTransaction.TransactionID, reason, original.PrevHash, reversalTransaction.PrevHash)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal membatalkan transaksi")
	}

	return response, nil
}

func (s *TransactionService) buildTransactionReversalResponse(tx *gorm.DB, cooperativeID, originalTransactionID, reversalTransactionID uuid.UUID, reason, originalPrevHash, reversalPrevHash string) (*model.TransactionReversalResponse, error) {
	originalDetail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, cooperativeID, originalTransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil transaksi asal")
	}

	reversalDetail, err := s.deps.repository.TransactionRepository.GetTransactionDetail(tx, cooperativeID, reversalTransactionID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil transaksi pembatalan")
	}

	if reversalPrevHash == "" {
		reversalTransaction, err := s.deps.repository.TransactionRepository.GetTransactionForUpdate(tx, cooperativeID, reversalTransactionID)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mengambil transaksi pembatalan")
		}
		reversalPrevHash = reversalTransaction.PrevHash
	}

	return &model.TransactionReversalResponse{
		OriginalTransaction: *mapDetailedTransactionResponse(originalDetail, originalPrevHash),
		ReversalTransaction: *mapDetailedTransactionResponse(reversalDetail, reversalPrevHash),
		Reason:              reason,
	}, nil
}

func (s *TransactionService) validateTransactionReversal(tx *gorm.DB, original *entity.Transaction) error {
	switch original.TransactionType {
	case constants.TransactionTypeLoan:
		loan, err := s.deps.repository.LoanRepository.GetLoanByDisbursementTransactionIDForUpdate(tx, original.TransactionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.NotFound("pinjaman tidak ditemukan")
			}
			return appErrors.InternalServer("gagal mengambil pinjaman")
		}

		allocations, err := s.deps.repository.LoanRepository.CountPaymentAllocationsByLoanID(tx, loan.LoanID)
		if err != nil {
			return appErrors.InternalServer("gagal memeriksa angsuran pinjaman")
		}
		if allocations > 0 {
			return appErrors.BadRequest("pinjaman yang sudah memiliki angsuran tidak dapat dibatalkan")
		}
	case constants.TransactionTypeInstallment:
		allocations, err := s.deps.repository.LoanRepository.ListPaymentAllocationsByTransaction(tx, original.TransactionID)
		if err != nil {
			return appErrors.InternalServer("gagal mengambil alokasi angsuran")
		}
		if len(allocations) == 0 {
			return appErrors.BadRequest("alokasi angsuran tidak ditemukan")
		}
	}

	return nil
}

func (s *TransactionService) applyTransactionReversalEffects(tx *gorm.DB, original, reversal *entity.Transaction) error {
	switch original.TransactionType {
	case constants.TransactionTypeLoan:
		loan, err := s.deps.repository.LoanRepository.GetLoanByDisbursementTransactionIDForUpdate(tx, original.TransactionID)
		if err != nil {
			return appErrors.InternalServer("gagal mengambil pinjaman")
		}

		loan.Status = "CANCELLED"
		if err := s.deps.repository.LoanRepository.UpdateLoanAccount(tx, loan); err != nil {
			return appErrors.InternalServer("gagal membatalkan pinjaman")
		}
	case constants.TransactionTypeInstallment:
		if err := s.reverseInstallmentAllocation(tx, original, reversal); err != nil {
			return err
		}
	}

	return nil
}

func (s *TransactionService) reverseInstallmentAllocation(tx *gorm.DB, original, reversal *entity.Transaction) error {
	allocations, err := s.deps.repository.LoanRepository.ListPaymentAllocationsByTransaction(tx, original.TransactionID)
	if err != nil {
		return appErrors.InternalServer("gagal mengambil alokasi angsuran")
	}

	scheduleIDs := make([]uuid.UUID, 0, len(allocations))
	for _, allocation := range allocations {
		scheduleIDs = append(scheduleIDs, allocation.ScheduleID)
	}

	schedules, err := s.deps.repository.LoanRepository.ListSchedulesByIDsForUpdate(tx, scheduleIDs)
	if err != nil {
		return appErrors.InternalServer("gagal mengambil jadwal angsuran")
	}

	scheduleByID := make(map[uuid.UUID]*entity.LoanInstallmentSchedule, len(schedules))
	for i := range schedules {
		scheduleByID[schedules[i].ScheduleID] = &schedules[i]
	}

	reversalAllocations := make([]entity.LoanPaymentAllocation, 0, len(allocations))
	var loanID uuid.UUID

	for _, allocation := range allocations {
		loanID = allocation.LoanID
		schedule := scheduleByID[allocation.ScheduleID]
		if schedule == nil {
			return appErrors.InternalServer("jadwal angsuran tidak ditemukan")
		}

		schedule.PaidAmount -= allocation.Amount
		if schedule.PaidAmount < 0 {
			schedule.PaidAmount = 0
		}
		schedule.RemainingAmount = schedule.DueAmount - schedule.PaidAmount
		schedule.PaidAt = nil
		if schedule.PaidAmount == 0 {
			schedule.Status = "UNPAID"
		} else {
			schedule.Status = "PARTIAL"
		}

		if err := s.deps.repository.LoanRepository.UpdateLoanSchedule(tx, schedule); err != nil {
			return appErrors.InternalServer("gagal memperbarui jadwal angsuran")
		}

		reversalAllocations = append(reversalAllocations, entity.LoanPaymentAllocation{
			AllocationID:  uuid.New(),
			LoanID:        allocation.LoanID,
			ScheduleID:    allocation.ScheduleID,
			TransactionID: reversal.TransactionID,
			Amount:        -allocation.Amount,
		})
	}

	if err := s.deps.repository.LoanRepository.CreateLoanPaymentAllocations(tx, reversalAllocations); err != nil {
		return appErrors.InternalServer("gagal menyimpan alokasi pembatalan angsuran")
	}

	if loanID != uuid.Nil {
		loan, err := s.deps.repository.LoanRepository.GetLoanByIDForUpdate(tx, original.CooperativeID, loanID)
		if err != nil {
			return appErrors.InternalServer("gagal mengambil pinjaman")
		}
		if loan.Status == "PAID" {
			loan.Status = "ACTIVE"
			if err := s.deps.repository.LoanRepository.UpdateLoanAccount(tx, loan); err != nil {
				return appErrors.InternalServer("gagal memperbarui status pinjaman")
			}
		}
	}

	return nil
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
func mapDetailedTransactionResponseWithSummaryAndLoan(item *model.TransactionListItemResponse, prevHash string, memberSummary *model.TransactionMemberSummaryResponse, loanSummary *model.LoanSummaryResponse) *model.TransactionResponse {
	response := mapDetailedTransactionResponseWithSummary(item, prevHash, memberSummary)
	enrichTransactionResponseWithLoan(response, loanSummary)
	return response
}

func buildLoanAccountAndSchedules(req model.CreateLoanTransactionRequest, transactionID uuid.UUID, loanNumber string, financialConfig *entity.FinancialConfiguration, disbursedAt time.Time) (*entity.LoanAccount, []entity.LoanInstallmentSchedule) {
	termMonths := financialConfig.MaxLoanTermMonths
	monthlyInterest := req.Amount * int64(financialConfig.LoanInterestRateBpsPerMonth) / 10000
	totalInterest := monthlyInterest * int64(termMonths)
	totalPayable := req.Amount + totalInterest
	monthlyDue := ceilDiv(totalPayable, int64(termMonths))

	loan := &entity.LoanAccount{
		LoanID:                    uuid.New(),
		CooperativeID:             req.CooperativeID,
		MemberID:                  req.MemberID,
		DisbursementTransactionID: transactionID,
		LoanNumber:                loanNumber,
		PrincipalAmount:           req.Amount,
		TotalInterestAmount:       totalInterest,
		TotalPayableAmount:        totalPayable,
		MonthlyInstallmentAmount:  monthlyDue,
		InterestRateBpsPerMonth:   financialConfig.LoanInterestRateBpsPerMonth,
		TermMonths:                termMonths,
		Status:                    "ACTIVE",
		DisbursedAt:               disbursedAt,
	}

	schedules := make([]entity.LoanInstallmentSchedule, 0, termMonths)
	remainingPayable := totalPayable
	firstDueDate := disbursedAt.AddDate(0, 1, 0)

	for installmentNo := 1; installmentNo <= termMonths; installmentNo++ {
		dueAmount := monthlyDue
		if installmentNo == termMonths || remainingPayable < monthlyDue {
			dueAmount = remainingPayable
		}

		schedules = append(schedules, entity.LoanInstallmentSchedule{
			ScheduleID:      uuid.New(),
			LoanID:          loan.LoanID,
			InstallmentNo:   installmentNo,
			DueDate:         firstDueDate.AddDate(0, installmentNo-1, 0),
			DueAmount:       dueAmount,
			PaidAmount:      0,
			RemainingAmount: dueAmount,
			Status:          "UNPAID",
		})

		remainingPayable -= dueAmount
	}

	return loan, schedules
}

func ceilDiv(value int64, divisor int64) int64 {
	if divisor <= 0 {
		return 0
	}

	return (value + divisor - 1) / divisor
}

func enrichTransactionResponseWithLoan(response *model.TransactionResponse, loanSummary *model.LoanSummaryResponse) {
	if response == nil || loanSummary == nil {
		return
	}

	response.LoanID = &loanSummary.LoanID
	response.LoanNumber = loanSummary.LoanNumber
	response.MonthlyInstallmentAmount = loanSummary.MonthlyInstallmentAmount
	response.RemainingPayableAmount = loanSummary.RemainingPayableAmount
	response.CurrentMonthDueAmount = loanSummary.CurrentMonthDueAmount
}

func allocateInstallmentPayment(schedules []entity.LoanInstallmentSchedule, amount int64, paidAt time.Time) ([]entity.LoanPaymentAllocation, []entity.LoanInstallmentSchedule, int64) {
	remainingPayment := amount
	allocations := make([]entity.LoanPaymentAllocation, 0)
	updatedSchedules := make([]entity.LoanInstallmentSchedule, 0)

	for i := range schedules {
		if remainingPayment <= 0 {
			break
		}

		allocAmount := minInt64(remainingPayment, schedules[i].RemainingAmount)
		if allocAmount <= 0 {
			continue
		}

		schedules[i].PaidAmount += allocAmount
		schedules[i].RemainingAmount -= allocAmount

		if schedules[i].RemainingAmount == 0 {
			schedules[i].Status = "PAID"
			schedules[i].PaidAt = &paidAt
		} else {
			schedules[i].Status = "PARTIAL"
		}

		allocations = append(allocations, entity.LoanPaymentAllocation{
			AllocationID: uuid.New(),
			LoanID:       schedules[i].LoanID,
			ScheduleID:   schedules[i].ScheduleID,
			Amount:       allocAmount,
		})
		updatedSchedules = append(updatedSchedules, schedules[i])
		remainingPayment -= allocAmount
	}

	return allocations, updatedSchedules, remainingPayment
}

func (s *TransactionService) getInstallmentResponseMetadata(tx *gorm.DB, cooperativeID uuid.UUID, transactionID uuid.UUID, asOf time.Time) (*model.LoanSummaryResponse, []model.InstallmentAllocationResponse, error) {
	allocations, err := s.deps.repository.LoanRepository.ListPaymentAllocationsByTransaction(tx, transactionID)
	if err != nil {
		return nil, nil, appErrors.InternalServer("gagal mengambil alokasi angsuran")
	}

	if len(allocations) == 0 {
		return nil, nil, appErrors.InternalServer("alokasi angsuran tidak ditemukan")
	}

	scheduleIDs := make([]uuid.UUID, 0, len(allocations))
	for _, allocation := range allocations {
		scheduleIDs = append(scheduleIDs, allocation.ScheduleID)
	}

	schedules, err := s.deps.repository.LoanRepository.ListSchedulesByIDs(tx, scheduleIDs)
	if err != nil {
		return nil, nil, appErrors.InternalServer("gagal mengambil jadwal angsuran")
	}

	loan, err := s.deps.repository.LoanRepository.GetLoanByIDForUpdate(tx, cooperativeID, allocations[0].LoanID)
	if err != nil {
		return nil, nil, appErrors.InternalServer("gagal mengambil pinjaman")
	}

	loanSummary, err := s.deps.repository.LoanRepository.GetLoanSummary(tx, loan.LoanID, asOf)
	if err != nil {
		return nil, nil, appErrors.InternalServer("gagal mengambil ringkasan pinjaman")
	}

	return loanSummary, mapInstallmentAllocationResponses(schedules, allocations), nil
}

func mapInstallmentAllocationResponses(schedules []entity.LoanInstallmentSchedule, allocations []entity.LoanPaymentAllocation) []model.InstallmentAllocationResponse {
	scheduleByID := make(map[uuid.UUID]entity.LoanInstallmentSchedule, len(schedules))
	for _, schedule := range schedules {
		scheduleByID[schedule.ScheduleID] = schedule
	}

	responses := make([]model.InstallmentAllocationResponse, 0, len(allocations))
	for _, allocation := range allocations {
		schedule := scheduleByID[allocation.ScheduleID]
		responses = append(responses, model.InstallmentAllocationResponse{
			ScheduleID:      allocation.ScheduleID,
			InstallmentNo:   schedule.InstallmentNo,
			DueDate:         schedule.DueDate,
			AllocatedAmount: allocation.Amount,
			ScheduleStatus:  schedule.Status,
		})
	}

	return responses
}

func mapDetailedInstallmentResponse(item *model.TransactionListItemResponse, prevHash string, loanSummary *model.LoanSummaryResponse, allocations []model.InstallmentAllocationResponse) *model.InstallmentTransactionResponse {
	transactionResponse := mapDetailedTransactionResponse(item, prevHash)
	enrichTransactionResponseWithLoan(transactionResponse, loanSummary)

	response := &model.InstallmentTransactionResponse{
		TransactionResponse: *transactionResponse,
		Allocations:         allocations,
	}

	if loanSummary != nil {
		response.Loan = *loanSummary
	}

	return response
}

func sumScheduleRemaining(schedules []entity.LoanInstallmentSchedule) int64 {
	var total int64
	for _, schedule := range schedules {
		total += schedule.RemainingAmount
	}

	return total
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}

	return b
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
		TransactionID:         item.TransactionID,
		CooperativeID:         item.CooperativeID,
		MemberID:              item.MemberID,
		MemberName:            item.MemberName,
		MemberNumber:          item.MemberNumber,
		MemberMCSGrade:        item.MemberMCSGrade,
		OfficerID:             item.OfficerID,
		OfficerName:           item.OfficerName,
		TransactionType:       item.TransactionType,
		TransactionTypeLabel:  item.TransactionTypeLabel,
		TransactionGroup:      item.TransactionGroup,
		Amount:                item.Amount,
		Description:           item.Description,
		RecordedAt:            item.RecordedAt,
		SyncedAt:              item.SyncedAt,
		PrevHash:              prevHash,
		CurrentHash:           item.CurrentHash,
		HashPreview:           item.HashPreview,
		IsOfflineCreated:      item.IsOfflineCreated,
		ClientTransactionID:   item.ClientTransactionID,
		SyncStatus:            item.SyncStatus,
		IsReversed:            item.IsReversed,
		IsReversal:            item.IsReversal,
		OriginalTransactionID: item.OriginalTransactionID,
		ReversalTransactionID: item.ReversalTransactionID,
		ReversalReason:        item.ReversalReason,
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
	case constants.TransactionTypeCashWithdrawal:
		return constants.TransactionGroupCashWithdrawal
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
	case constants.TransactionTypeCashWithdrawal:
		return "Tarik Tunai"
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

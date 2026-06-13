package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ledgerStatusValid   = "VALID"
	ledgerStatusInvalid = "INVALID"

	ledgerRecordTypeTransaction   = "TRANSACTION"
	ledgerRecordTypeStockMovement = "STOCK_MOVEMENT"

	ledgerNetworkHyperledger = "Hyperledger Fabric"
)

type ILedgerService interface {
	GetLedgerAudit(req model.LedgerAuditRequest) (*model.LedgerAuditResponse, error)
	AnchorLedger(req model.LedgerAuditRequest) (*model.LedgerAnchorResponse, error)
}

type LedgerService struct {
	deps serviceDependency
}

func NewLedgerService(deps serviceDependency) ILedgerService {
	return &LedgerService{deps: deps}
}

func (s *LedgerService) GetLedgerAudit(req model.LedgerAuditRequest) (*model.LedgerAuditResponse, error) {
	if err := validateLedgerAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodStart, periodEnd, periodKey, err := resolveLedgerPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	chainStart := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	transactions, err := s.deps.repository.LedgerRepository.ListFinancialLedgerRows(tx, req.CooperativeID, chainStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ledger transaksi")
	}

	stockMovements, err := s.deps.repository.LedgerRepository.ListStockLedgerRows(tx, req.CooperativeID, chainStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ledger stok")
	}

	cooperativeName, err := s.deps.repository.LedgerRepository.GetCooperativeName(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil data koperasi")
	}

	anchor, err := s.deps.repository.LedgerRepository.GetLatestAnchor(tx, req.CooperativeID, periodStart, periodEnd)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil anchor ledger")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil audit ledger")
	}

	financialItems := verifyFinancialLedgerItems(transactions, periodStart, periodEnd)
	stockItems := verifyStockLedgerItems(stockMovements, periodStart, periodEnd)

	items := append(financialItems, stockItems...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].RecordedAt.After(items[j].RecordedAt)
	})

	overallStatus := ledgerStatusValid
	for _, item := range items {
		if item.Status == ledgerStatusInvalid {
			overallStatus = ledgerStatusInvalid
			break
		}
	}

	allHashes := collectLedgerHashes(transactions, stockMovements, periodStart, periodEnd)
	merkleRoot := buildMerkleRoot(allHashes)
	pagedItems := paginateLedgerItems(items, req.Page, req.Limit)

	return &model.LedgerAuditResponse{
		Period:        periodKey,
		OverallStatus: overallStatus,
		Anchor:        mapLedgerAnchorResponse(anchor),
		Certificate: model.LedgerCertificateResponse{
			CooperativeName:     cooperativeName,
			PeriodLabel:         formatLedgerPeriodLabel(periodStart, periodEnd),
			MerkleRootHash:      merkleRoot,
			MerkleRootPreview:   buildLedgerHashPreview(merkleRoot),
			BlockchainTxID:      anchorTxID(anchor),
			BlockchainTxPreview: buildBlockchainTxPreview(anchor),
		},
		Items: pagedItems,
		Page:  req.Page,
		Limit: req.Limit,
		Total: int64(len(items)),
	}, nil
}

func (s *LedgerService) AnchorLedger(req model.LedgerAuditRequest) (*model.LedgerAnchorResponse, error) {
	if err := validateLedgerAccess(req.AuthContext); err != nil {
		return nil, err
	}

	periodStart, periodEnd, _, err := resolveLedgerPeriod(req.Period)
	if err != nil {
		return nil, err
	}

	chainStart := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	transactions, err := s.deps.repository.LedgerRepository.ListFinancialLedgerRows(tx, req.CooperativeID, chainStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ledger transaksi")
	}

	stockMovements, err := s.deps.repository.LedgerRepository.ListStockLedgerRows(tx, req.CooperativeID, chainStart, periodEnd)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil ledger stok")
	}

	hashes := collectLedgerHashes(transactions, stockMovements, periodStart, periodEnd)
	if len(hashes) == 0 {
		return nil, appErrors.BadRequest("belum ada ledger pada periode ini")
	}

	merkleRoot := buildMerkleRoot(hashes)

	existingAnchor, err := s.deps.repository.LedgerRepository.GetAnchorByRootHash(tx, req.CooperativeID, periodStart, periodEnd, merkleRoot)
	if err == nil {
		if err := tx.Commit().Error; err != nil {
			return nil, appErrors.InternalServer("gagal mengambil anchor ledger")
		}
		return mapLedgerAnchorResponse(existingAnchor), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memeriksa anchor ledger")
	}

	now := time.Now()
	anchor := &entity.LedgerAnchor{
		LedgerAnchorID:        uuid.New(),
		CooperativeID:         req.CooperativeID,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
		MerkleRootHash:        merkleRoot,
		BlockchainNetwork:     ledgerNetworkHyperledger,
		BlockchainBlockNumber: buildSyntheticBlockNumber(now, merkleRoot),
		BlockchainTxID:        buildSyntheticBlockchainTxID(merkleRoot),
		AnchoredAt:            now,
	}

	err = s.deps.repository.LedgerRepository.CreateAnchor(tx, anchor)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan anchor ledger")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan anchor ledger")
	}

	return mapLedgerAnchorResponse(anchor), nil
}

func validateLedgerAccess(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}
	if auth.RoleCode != constants.RoleCodePengurusKoperasi {
		return appErrors.Forbidden("hanya pengurus koperasi yang dapat melihat ledger")
	}
	return nil
}

func resolveLedgerPeriod(period string) (time.Time, time.Time, string, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		now := time.Now()
		period = now.Format("2006-01")
	}

	periodStart, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, time.Time{}, "", appErrors.BadRequest("format periode tidak valid")
	}

	periodStart = time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return periodStart, periodEnd, periodStart.Format("2006-01"), nil
}

func verifyFinancialLedgerItems(transactions []entity.Transaction, periodStart, periodEnd time.Time) []model.LedgerAuditItemResponse {
	items := make([]model.LedgerAuditItemResponse, 0)
	expectedPrevHash := genesisTransactionHash

	for _, transaction := range transactions {
		expectedCurrentHash := buildTransactionHash(&transaction)
		status := ledgerStatusValid
		reason := ""

		if transaction.PrevHash != expectedPrevHash {
			status = ledgerStatusInvalid
			reason = "prev_hash tidak sesuai dengan transaksi sebelumnya"
		} else if transaction.CurrentHash != expectedCurrentHash {
			status = ledgerStatusInvalid
			reason = "current_hash tidak sesuai dengan payload transaksi"
		}

		if isInLedgerPeriod(transaction.RecordedAt, periodStart, periodEnd) {
			items = append(items, model.LedgerAuditItemResponse{
				RecordID:      transaction.TransactionID,
				RecordType:    ledgerRecordTypeTransaction,
				Title:         getTransactionTypeLabel(transaction.TransactionType),
				Subtitle:      buildTransactionLedgerSubtitle(transaction),
				Amount:        transaction.Amount,
				RecordedAt:    transaction.RecordedAt,
				PrevHash:      transaction.PrevHash,
				CurrentHash:   transaction.CurrentHash,
				HashPreview:   buildLedgerHashPreview(transaction.CurrentHash),
				Status:        status,
				InvalidReason: reason,
			})
		}

		expectedPrevHash = transaction.CurrentHash
	}

	return items
}

func verifyStockLedgerItems(movements []entity.StockMovement, periodStart, periodEnd time.Time) []model.LedgerAuditItemResponse {
	items := make([]model.LedgerAuditItemResponse, 0)
	expectedPrevHash := constants.GenesisStockHash

	for _, movement := range movements {
		expectedCurrentHash := buildStockMovementHash(&movement)
		status := ledgerStatusValid
		reason := ""

		if movement.PrevHash != expectedPrevHash {
			status = ledgerStatusInvalid
			reason = "prev_hash tidak sesuai dengan mutasi stok sebelumnya"
		} else if movement.CurrentHash != expectedCurrentHash {
			status = ledgerStatusInvalid
			reason = "current_hash tidak sesuai dengan payload mutasi stok"
		}

		if isInLedgerPeriod(movement.RecordedAt, periodStart, periodEnd) {
			items = append(items, model.LedgerAuditItemResponse{
				RecordID:      movement.StockMovementID,
				RecordType:    ledgerRecordTypeStockMovement,
				Title:         getStockMovementLedgerTitle(movement.MovementType),
				Subtitle:      buildStockMovementLedgerSubtitle(movement),
				Amount:        movement.QuantityDelta,
				RecordedAt:    movement.RecordedAt,
				PrevHash:      movement.PrevHash,
				CurrentHash:   movement.CurrentHash,
				HashPreview:   buildLedgerHashPreview(movement.CurrentHash),
				Status:        status,
				InvalidReason: reason,
			})
		}

		expectedPrevHash = movement.CurrentHash
	}

	return items
}

func collectLedgerHashes(transactions []entity.Transaction, movements []entity.StockMovement, periodStart, periodEnd time.Time) []string {
	type hashRow struct {
		hash       string
		recordedAt time.Time
		createdAt  time.Time
	}

	rows := make([]hashRow, 0, len(transactions)+len(movements))

	for _, transaction := range transactions {
		if isInLedgerPeriod(transaction.RecordedAt, periodStart, periodEnd) {
			rows = append(rows, hashRow{
				hash:       transaction.CurrentHash,
				recordedAt: transaction.RecordedAt,
				createdAt:  transaction.CreatedAt,
			})
		}
	}

	for _, movement := range movements {
		if isInLedgerPeriod(movement.RecordedAt, periodStart, periodEnd) {
			rows = append(rows, hashRow{
				hash:       movement.CurrentHash,
				recordedAt: movement.RecordedAt,
				createdAt:  movement.CreatedAt,
			})
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].recordedAt.Equal(rows[j].recordedAt) {
			return rows[i].createdAt.Before(rows[j].createdAt)
		}
		return rows[i].recordedAt.Before(rows[j].recordedAt)
	})

	hashes := make([]string, 0, len(rows))
	for _, row := range rows {
		hashes = append(hashes, row.hash)
	}

	return hashes
}

func buildMerkleRoot(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}

	level := append([]string(nil), hashes...)
	for len(level) > 1 {
		next := make([]string, 0, (len(level)+1)/2)

		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := left
			if i+1 < len(level) {
				right = level[i+1]
			}

			sum := sha256.Sum256([]byte(left + right))
			next = append(next, hex.EncodeToString(sum[:]))
		}

		level = next
	}

	return level[0]
}

func paginateLedgerItems(items []model.LedgerAuditItemResponse, page, limit int) []model.LedgerAuditItemResponse {
	start := (page - 1) * limit
	if start >= len(items) {
		return []model.LedgerAuditItemResponse{}
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}

func isInLedgerPeriod(value, periodStart, periodEnd time.Time) bool {
	return !value.Before(periodStart) && !value.After(periodEnd)
}

func buildTransactionLedgerSubtitle(transaction entity.Transaction) string {
	sign := "+"
	if transaction.TransactionType == constants.TransactionTypeCashWithdrawal {
		sign = "-"
	}
	return fmt.Sprintf("%sRp %d", sign, transaction.Amount)
}

func getStockMovementLedgerTitle(movementType string) string {
	switch movementType {
	case constants.StockMovementProductCreated:
		return "Produk Baru"
	case constants.StockMovementStockIn:
		return "Stok Masuk"
	case constants.StockMovementAdjustment:
		return "Penyesuaian Stok"
	case constants.StockMovementSaleOut:
		return "Penjualan Toko"
	default:
		return movementType
	}
}

func buildStockMovementLedgerSubtitle(movement entity.StockMovement) string {
	sign := "+"
	if movement.QuantityDelta < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%d unit", sign, movement.QuantityDelta)
}

func mapLedgerAnchorResponse(anchor *entity.LedgerAnchor) *model.LedgerAnchorResponse {
	if anchor == nil {
		return &model.LedgerAnchorResponse{}
	}

	return &model.LedgerAnchorResponse{
		Network:        anchor.BlockchainNetwork,
		BlockNumber:    anchor.BlockchainBlockNumber,
		BlockchainTxID: anchor.BlockchainTxID,
		AnchoredAt:     anchor.AnchoredAt,
	}
}

func buildLedgerHashPreview(hash string) string {
	if hash == "" {
		return ""
	}
	if len(hash) <= 16 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-4:]
}

func anchorTxID(anchor *entity.LedgerAnchor) string {
	if anchor == nil {
		return ""
	}
	return anchor.BlockchainTxID
}

func buildBlockchainTxPreview(anchor *entity.LedgerAnchor) string {
	if anchor == nil {
		return ""
	}
	txPreview := anchor.BlockchainTxID
	if len(txPreview) > 6 {
		txPreview = txPreview[:3] + "..."
	}
	return fmt.Sprintf("%sHyperledger #%d", txPreview, anchor.BlockchainBlockNumber)
}

func formatLedgerPeriodLabel(periodStart, periodEnd time.Time) string {
	return fmt.Sprintf("%d-%d %s %d", periodStart.Day(), periodEnd.Day(), ledgerIndonesianMonth(periodStart.Month()), periodStart.Year())
}

func ledgerIndonesianMonth(month time.Month) string {
	names := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	return names[month]
}

func buildSyntheticBlockNumber(now time.Time, merkleRoot string) int64 {
	if merkleRoot == "" {
		return 48000 + now.Unix()%5000
	}

	sum := sha256.Sum256([]byte(merkleRoot))
	return 48000 + int64(sum[0])*16 + int64(sum[1])
}

func buildSyntheticBlockchainTxID(merkleRoot string) string {
	if len(merkleRoot) <= 16 {
		return "b92-" + merkleRoot
	}
	return "b92-" + merkleRoot[:16]
}

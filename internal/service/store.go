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

type IStoreService interface {
	GetStoreDashboard(req model.StoreDashboardRequest) (*model.StoreDashboardResponse, error)
	ListProducts(req model.ListProductsRequest) (*model.ListProductsResponse, error)
	GetProduct(req model.GetProductRequest) (*model.ProductResponse, error)
	CreateProduct(req model.CreateProductRequest) (*model.ProductResponse, error)
	CreateStockIn(req model.CreateStockInRequest) (*model.StockMovementResponse, error)
	CreateStockAdjustment(req model.CreateStockAdjustmentRequest) (*model.StockMovementResponse, error)
	ListStockMovements(req model.ListStockMovementsRequest) (*model.ListStockMovementsResponse, error)
	CreateStoreSale(req model.CreateStoreSaleRequest) (*model.StoreSaleResponse, error)
	GetStoreSale(req model.GetStoreSaleRequest) (*model.StoreSaleResponse, error)
}

type StoreService struct {
	deps serviceDependency
}

func NewStoreService(deps serviceDependency) IStoreService {
	return &StoreService{deps: deps}
}

func (s *StoreService) GetStoreDashboard(req model.StoreDashboardRequest) (*model.StoreDashboardResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	result, err := s.deps.repository.StoreRepository.GetStoreDashboard(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil dashboard toko")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil dashboard toko")
	}

	return result, nil
}

func (s *StoreService) ListProducts(req model.ListProductsRequest) (*model.ListProductsResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	items, total, err := s.deps.repository.StoreRepository.ListProducts(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil katalog produk")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil katalog produk")
	}

	return &model.ListProductsResponse{Items: items, Page: req.Page, Limit: req.Limit, Total: total}, nil
}

func (s *StoreService) GetProduct(req model.GetProductRequest) (*model.ProductResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.ProductID == uuid.Nil {
		return nil, appErrors.BadRequest("produk wajib dipilih")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	product, err := s.deps.repository.StoreRepository.GetProductResponseByID(tx, req.CooperativeID, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("produk tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil produk")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil produk")
	}

	return product, nil
}

func (s *StoreService) CreateProduct(req model.CreateProductRequest) (*model.ProductResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if err := validateProductInput(req); err != nil {
		return nil, err
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}
	now := time.Now()
	clientReferenceID := strings.TrimSpace(req.ClientReferenceID)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	if clientReferenceID != "" {
		existingMovement, err := s.deps.repository.StoreRepository.GetStockMovementByClientReferenceID(tx, req.CooperativeID, clientReferenceID)
		if err == nil {
			if existingMovement.MovementType != constants.StockMovementProductCreated {
				return nil, appErrors.BadRequest("client_reference_id sudah digunakan untuk mutasi lain")
			}

			product, err := s.deps.repository.StoreRepository.GetProductResponseByID(tx, req.CooperativeID, existingMovement.ProductID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil produk")
			}
			product.HashPreview = buildHashPreview(existingMovement.CurrentHash)

			err = tx.Commit().Error
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil produk")
			}

			return product, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa duplikasi produk")
		}
	}

	totalProducts, err := s.deps.repository.StoreRepository.CountProductsByCooperative(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat kode produk")
	}

	product := &entity.StoreProduct{
		ProductID:        uuid.New(),
		CooperativeID:    req.CooperativeID,
		ProductCode:      fmt.Sprintf("ST%03d", totalProducts+1),
		Name:             strings.TrimSpace(req.Name),
		Unit:             strings.TrimSpace(req.Unit),
		Category:         strings.TrimSpace(req.Category),
		CostPrice:        req.CostPrice,
		SalePrice:        req.SalePrice,
		StockQuantity:    req.InitialStockQuantity,
		MinStockQuantity: req.MinStockQuantity,
		IsActive:         true,
		CreatedBy:        req.UserID,
	}

	err = s.deps.repository.StoreRepository.CreateProduct(tx, product)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan produk")
	}

	movement, err := s.buildStockMovement(tx, stockMovementInput{
		CooperativeID:          req.CooperativeID,
		ProductID:              product.ProductID,
		OfficerID:              req.UserID,
		MovementType:           constants.StockMovementProductCreated,
		QuantityDelta:          req.InitialStockQuantity,
		ResultingStockQuantity: product.StockQuantity,
		UnitCost:               req.CostPrice,
		ReferenceType:          "PRODUCT",
		ReferenceID:            &product.ProductID,
		Description:            "Produk baru",
		RecordedAt:             recordedAt,
		SyncedAt:               &now,
		IsOfflineCreated:       req.IsOfflineCreated,
		ClientReferenceID:      clientReferenceID,
	})
	if err != nil {
		return nil, err
	}

	err = s.deps.repository.StoreRepository.CreateStockMovement(tx, movement)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan mutasi stok")
	}

	response, err := s.deps.repository.StoreRepository.GetProductResponseByID(tx, req.CooperativeID, product.ProductID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil produk")
	}
	response.HashPreview = buildHashPreview(movement.CurrentHash)

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan produk")
	}

	return response, nil
}

func (s *StoreService) CreateStockIn(req model.CreateStockInRequest) (*model.StockMovementResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.ProductID == uuid.Nil {
		return nil, appErrors.BadRequest("produk wajib dipilih")
	}
	if req.Quantity <= 0 {
		return nil, appErrors.BadRequest("jumlah stok masuk wajib lebih dari 0")
	}
	if req.UnitCost < 0 {
		return nil, appErrors.BadRequest("harga beli tidak boleh negatif")
	}
	if req.SalePrice < 0 {
		return nil, appErrors.BadRequest("harga jual tidak boleh negatif")
	}

	return s.createSingleProductMovement(singleProductMovementInput{
		AuthContext:       req.AuthContext,
		ProductID:         req.ProductID,
		MovementType:      constants.StockMovementStockIn,
		QuantityDelta:     req.Quantity,
		UnitCost:          req.UnitCost,
		SalePrice:         req.SalePrice,
		Description:       req.Description,
		RecordedAt:        req.RecordedAt,
		IsOfflineCreated:  req.IsOfflineCreated,
		ClientReferenceID: strings.TrimSpace(req.ClientReferenceID),
	})
}

func (s *StoreService) CreateStockAdjustment(req model.CreateStockAdjustmentRequest) (*model.StockMovementResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.ProductID == uuid.Nil {
		return nil, appErrors.BadRequest("produk wajib dipilih")
	}
	if req.QuantityDelta == 0 {
		return nil, appErrors.BadRequest("jumlah penyesuaian tidak boleh 0")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, appErrors.BadRequest("alasan penyesuaian wajib diisi")
	}

	return s.createSingleProductMovement(singleProductMovementInput{
		AuthContext:       req.AuthContext,
		ProductID:         req.ProductID,
		MovementType:      constants.StockMovementAdjustment,
		QuantityDelta:     req.QuantityDelta,
		Description:       req.Description,
		RecordedAt:        req.RecordedAt,
		IsOfflineCreated:  req.IsOfflineCreated,
		ClientReferenceID: strings.TrimSpace(req.ClientReferenceID),
	})
}

func (s *StoreService) ListStockMovements(req model.ListStockMovementsRequest) (*model.ListStockMovementsResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	items, total, err := s.deps.repository.StoreRepository.ListStockMovements(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil mutasi stok")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil mutasi stok")
	}

	return &model.ListStockMovementsResponse{Items: items, Page: req.Page, Limit: req.Limit, Total: total}, nil
}

func (s *StoreService) CreateStoreSale(req model.CreateStoreSaleRequest) (*model.StoreSaleResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if len(req.Items) == 0 {
		return nil, appErrors.BadRequest("item penjualan wajib diisi")
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}
	now := time.Now()
	clientSaleID := strings.TrimSpace(req.ClientSaleID)

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	if clientSaleID != "" {
		existingSale, err := s.deps.repository.StoreRepository.GetSaleByClientID(tx, req.CooperativeID, clientSaleID)
		if err == nil {
			detail, err := s.deps.repository.StoreRepository.GetSaleDetail(tx, req.CooperativeID, existingSale.StoreSaleID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil penjualan")
			}
			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil penjualan")
			}
			return detail, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa duplikasi penjualan")
		}
	}

	saleID := uuid.New()
	saleNumber, err := s.deps.repository.StoreRepository.GenerateSaleNumber(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat nomor penjualan")
	}

	saleItems := make([]entity.StoreSaleItem, 0, len(req.Items))
	products := make(map[uuid.UUID]*entity.StoreProduct, len(req.Items))
	var totalAmount int64
	var totalCost int64

	for _, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return nil, appErrors.BadRequest("produk wajib dipilih")
		}
		if item.Quantity <= 0 {
			return nil, appErrors.BadRequest("jumlah item wajib lebih dari 0")
		}

		product, ok := products[item.ProductID]
		if !ok {
			product, err = s.deps.repository.StoreRepository.GetProductForUpdate(tx, req.CooperativeID, item.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.NotFound("produk tidak ditemukan")
				}
				return nil, appErrors.InternalServer("gagal mengambil produk")
			}
			products[item.ProductID] = product
		}

		if product.StockQuantity < item.Quantity {
			return nil, appErrors.BadRequest("stok produk tidak mencukupi")
		}

		product.StockQuantity -= item.Quantity
		subtotal := calculateStoreSaleSubtotal(item.Quantity, product.SalePrice)
		costSubtotal := calculateStoreSaleSubtotal(item.Quantity, product.CostPrice)

		totalAmount += subtotal
		totalCost += costSubtotal

		saleItems = append(saleItems, entity.StoreSaleItem{
			StoreSaleItemID: uuid.New(),
			StoreSaleID:     saleID,
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			UnitPrice:       product.SalePrice,
			UnitCost:        product.CostPrice,
			Subtotal:        subtotal,
		})
	}

	if req.CashReceived < totalAmount {
		return nil, appErrors.BadRequest("uang diterima tidak mencukupi")
	}

	sale := &entity.StoreSale{
		StoreSaleID:   saleID,
		CooperativeID: req.CooperativeID,
		OfficerID:     req.UserID,
		SaleNumber:    saleNumber,
		TotalAmount:   totalAmount,
		TotalCost:     totalCost,
		CashReceived:  req.CashReceived,
		ChangeAmount:  req.CashReceived - totalAmount,
		Status:        constants.StoreSaleStatusCompleted,
		RecordedAt:    recordedAt,
		ClientSaleID:  clientSaleID,
		Items:         saleItems,
	}

	err = s.deps.repository.StoreRepository.CreateSale(tx, sale)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan penjualan")
	}

	for _, product := range products {
		product.UpdatedBy = &req.UserID
		err := s.deps.repository.StoreRepository.UpdateProduct(tx, product)
		if err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui stok produk")
		}
	}

	for _, item := range saleItems {
		product := products[item.ProductID]
		movementClientID := buildSaleMovementClientReferenceID(clientSaleID, item.ProductID)
		movement, err := s.buildStockMovement(tx, stockMovementInput{
			CooperativeID:          req.CooperativeID,
			ProductID:              item.ProductID,
			OfficerID:              req.UserID,
			MovementType:           constants.StockMovementSaleOut,
			QuantityDelta:          -item.Quantity,
			ResultingStockQuantity: product.StockQuantity,
			UnitCost:               item.UnitPrice,
			ReferenceType:          "STORE_SALE",
			ReferenceID:            &saleID,
			Description:            fmt.Sprintf("Penjualan toko %s", saleNumber),
			RecordedAt:             recordedAt,
			SyncedAt:               &now,
			ClientReferenceID:      movementClientID,
		})
		if err != nil {
			return nil, err
		}
		err = s.deps.repository.StoreRepository.CreateStockMovement(tx, movement)
		if err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan mutasi penjualan")
		}
	}

	detail, err := s.deps.repository.StoreRepository.GetSaleDetail(tx, req.CooperativeID, saleID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil penjualan")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan penjualan")
	}

	return detail, nil
}

func (s *StoreService) GetStoreSale(req model.GetStoreSaleRequest) (*model.StoreSaleResponse, error) {
	if err := validateStoreAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.StoreSaleID == uuid.Nil {
		return nil, appErrors.BadRequest("penjualan wajib dipilih")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	sale, err := s.deps.repository.StoreRepository.GetSaleDetail(tx, req.CooperativeID, req.StoreSaleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("penjualan tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil penjualan")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil penjualan")
	}

	return sale, nil
}

type singleProductMovementInput struct {
	model.AuthContext
	ProductID         uuid.UUID
	MovementType      string
	QuantityDelta     int64
	UnitCost          int64
	SalePrice         int64
	Description       string
	RecordedAt        *time.Time
	IsOfflineCreated  bool
	ClientReferenceID string
}

type stockMovementInput struct {
	CooperativeID          uuid.UUID
	ProductID              uuid.UUID
	OfficerID              uuid.UUID
	MovementType           string
	QuantityDelta          int64
	ResultingStockQuantity int64
	UnitCost               int64
	ReferenceType          string
	ReferenceID            *uuid.UUID
	Description            string
	RecordedAt             time.Time
	SyncedAt               *time.Time
	IsOfflineCreated       bool
	ClientReferenceID      string
}

func (s *StoreService) createSingleProductMovement(input singleProductMovementInput) (*model.StockMovementResponse, error) {
	recordedAt := time.Now()
	if input.RecordedAt != nil {
		recordedAt = *input.RecordedAt
	}
	now := time.Now()

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	if input.ClientReferenceID != "" {
		existingMovement, err := s.deps.repository.StoreRepository.GetStockMovementByClientReferenceID(tx, input.CooperativeID, input.ClientReferenceID)
		if err == nil {
			if existingMovement.MovementType != input.MovementType {
				return nil, appErrors.BadRequest("client_reference_id sudah digunakan untuk mutasi lain")
			}

			response, err := s.deps.repository.StoreRepository.GetStockMovementResponseByID(tx, input.CooperativeID, existingMovement.StockMovementID)
			if err != nil {
				return nil, appErrors.InternalServer("gagal mengambil mutasi stok")
			}

			if err := tx.Commit().Error; err != nil {
				return nil, appErrors.InternalServer("gagal mengambil mutasi stok")
			}

			return response, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memeriksa duplikasi mutasi")
		}
	}

	product, err := s.deps.repository.StoreRepository.GetProductForUpdate(tx, input.CooperativeID, input.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("produk tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil produk")
	}

	resultingStock := product.StockQuantity + input.QuantityDelta
	if resultingStock < 0 {
		return nil, appErrors.BadRequest("stok produk tidak boleh kurang dari 0")
	}

	if input.UnitCost == 0 {
		input.UnitCost = product.CostPrice
	}

	product.StockQuantity = resultingStock
	product.UpdatedBy = &input.UserID
	if input.MovementType == constants.StockMovementStockIn && input.UnitCost > 0 {
		product.CostPrice = input.UnitCost
	}
	if input.MovementType == constants.StockMovementStockIn && input.SalePrice > 0 {
		product.SalePrice = input.SalePrice
	}

	if err := s.deps.repository.StoreRepository.UpdateProduct(tx, product); err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui stok produk")
	}

	movement, err := s.buildStockMovement(tx, stockMovementInput{
		CooperativeID:          input.CooperativeID,
		ProductID:              input.ProductID,
		OfficerID:              input.UserID,
		MovementType:           input.MovementType,
		QuantityDelta:          input.QuantityDelta,
		ResultingStockQuantity: resultingStock,
		UnitCost:               input.UnitCost,
		ReferenceType:          "PRODUCT",
		ReferenceID:            &input.ProductID,
		Description:            strings.TrimSpace(input.Description),
		RecordedAt:             recordedAt,
		SyncedAt:               &now,
		IsOfflineCreated:       input.IsOfflineCreated,
		ClientReferenceID:      input.ClientReferenceID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.deps.repository.StoreRepository.CreateStockMovement(tx, movement); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan mutasi stok")
	}

	response, err := s.deps.repository.StoreRepository.GetStockMovementResponseByID(tx, input.CooperativeID, movement.StockMovementID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil mutasi stok")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan mutasi stok")
	}

	return response, nil
}

func (s *StoreService) buildStockMovement(tx *gorm.DB, input stockMovementInput) (*entity.StockMovement, error) {
	prevHash := constants.GenesisStockHash
	latestMovement, err := s.deps.repository.StoreRepository.GetLatestStockMovementForUpdate(tx, input.CooperativeID)
	if err == nil {
		prevHash = latestMovement.CurrentHash
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal mengambil mutasi stok terakhir")
	}

	movement := &entity.StockMovement{
		StockMovementID:        uuid.New(),
		CooperativeID:          input.CooperativeID,
		ProductID:              input.ProductID,
		OfficerID:              input.OfficerID,
		MovementType:           input.MovementType,
		QuantityDelta:          input.QuantityDelta,
		ResultingStockQuantity: input.ResultingStockQuantity,
		UnitCost:               input.UnitCost,
		ReferenceType:          strings.TrimSpace(input.ReferenceType),
		ReferenceID:            input.ReferenceID,
		Description:            strings.TrimSpace(input.Description),
		RecordedAt:             input.RecordedAt,
		SyncedAt:               input.SyncedAt,
		PrevHash:               prevHash,
		IsOfflineCreated:       input.IsOfflineCreated,
		ClientReferenceID:      input.ClientReferenceID,
	}
	movement.CurrentHash = buildStockMovementHash(movement)

	return movement, nil
}

func validateStoreAccess(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}
	if auth.RoleCode != constants.RoleCodePengurusKoperasi {
		return appErrors.Forbidden("hanya pengurus koperasi yang dapat mengelola stok")
	}
	return nil
}

func validateProductInput(req model.CreateProductRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return appErrors.BadRequest("nama produk wajib diisi")
	}
	if strings.TrimSpace(req.Unit) == "" {
		return appErrors.BadRequest("satuan wajib diisi")
	}
	if strings.TrimSpace(req.Category) == "" {
		return appErrors.BadRequest("kategori wajib diisi")
	}
	if req.CostPrice < 0 {
		return appErrors.BadRequest("hpp tidak boleh negatif")
	}
	if req.SalePrice <= 0 {
		return appErrors.BadRequest("harga jual wajib lebih dari 0")
	}
	if req.MinStockQuantity < 0 {
		return appErrors.BadRequest("stok minimum tidak boleh negatif")
	}
	if req.InitialStockQuantity < 0 {
		return appErrors.BadRequest("stok awal tidak boleh negatif")
	}
	return nil
}

func buildStockMovementHash(movement *entity.StockMovement) string {
	referenceID := ""
	if movement.ReferenceID != nil {
		referenceID = movement.ReferenceID.String()
	}

	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%d|%d|%d|%s|%s|%s|%s|%t|%s",
		movement.PrevHash,
		movement.StockMovementID.String(),
		movement.CooperativeID.String(),
		movement.ProductID.String(),
		movement.OfficerID.String(),
		movement.MovementType,
		movement.QuantityDelta,
		movement.ResultingStockQuantity,
		movement.UnitCost,
		movement.ReferenceType,
		referenceID,
		movement.Description,
		movement.RecordedAt.UTC().Format(time.RFC3339Nano),
		movement.IsOfflineCreated,
		movement.ClientReferenceID,
	)

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func buildSaleMovementClientReferenceID(clientSaleID string, productID uuid.UUID) string {
	if strings.TrimSpace(clientSaleID) == "" {
		return ""
	}

	return fmt.Sprintf("%s:%s", clientSaleID, productID.String())
}

func calculateStoreSaleSubtotal(quantity, unitPrice int64) int64 {
	return quantity * unitPrice / 100
}

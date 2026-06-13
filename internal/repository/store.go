package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IStoreRepository interface {
	CreateProduct(tx *gorm.DB, product *entity.StoreProduct) error
	GetProductByID(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*entity.StoreProduct, error)
	GetProductForUpdate(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*entity.StoreProduct, error)
	GetProductResponseByID(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*model.ProductResponse, error)
	ListProducts(tx *gorm.DB, req model.ListProductsRequest) ([]model.ProductResponse, int64, error)
	CountProductsByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error)
	UpdateProduct(tx *gorm.DB, product *entity.StoreProduct) error
	CreateStockMovement(tx *gorm.DB, movement *entity.StockMovement) error
	GetLatestStockMovementForUpdate(tx *gorm.DB, cooperativeID uuid.UUID) (*entity.StockMovement, error)
	GetStockMovementByClientReferenceID(tx *gorm.DB, cooperativeID uuid.UUID, clientReferenceID string) (*entity.StockMovement, error)
	GetStockMovementResponseByID(tx *gorm.DB, cooperativeID, movementID uuid.UUID) (*model.StockMovementResponse, error)
	ListStockMovements(tx *gorm.DB, req model.ListStockMovementsRequest) ([]model.StockMovementResponse, int64, error)
	GetStoreDashboard(tx *gorm.DB, cooperativeID uuid.UUID) (*model.StoreDashboardResponse, error)
	CreateSale(tx *gorm.DB, sale *entity.StoreSale) error
	GetSaleByClientID(tx *gorm.DB, cooperativeID uuid.UUID, clientSaleID string) (*entity.StoreSale, error)
	GetSaleDetail(tx *gorm.DB, cooperativeID, saleID uuid.UUID) (*model.StoreSaleResponse, error)
	GenerateSaleNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error)
}

type StoreRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) IStoreRepository {
	return &StoreRepository{db: db}
}

func (r *StoreRepository) CreateProduct(tx *gorm.DB, product *entity.StoreProduct) error {
	err := tx.Debug().Create(product).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *StoreRepository) GetProductByID(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*entity.StoreProduct, error) {
	var product entity.StoreProduct

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("product_id = ?", productID).
		First(&product).Error
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *StoreRepository) GetProductForUpdate(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*entity.StoreProduct, error) {
	var product entity.StoreProduct

	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Where("product_id = ?", productID).
		First(&product).Error
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *StoreRepository) GetProductResponseByID(tx *gorm.DB, cooperativeID, productID uuid.UUID) (*model.ProductResponse, error) {
	var result model.ProductResponse

	err := baseProductQuery(tx).
		Where("store_products.cooperative_id = ?", cooperativeID).
		Where("store_products.product_id = ?", productID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	enrichProductResponse(&result)
	return &result, nil
}

func (r *StoreRepository) ListProducts(tx *gorm.DB, req model.ListProductsRequest) ([]model.ProductResponse, int64, error) {
	var results []model.ProductResponse
	var total int64

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	query := baseProductQuery(tx).
		Where("store_products.cooperative_id = ?", req.CooperativeID).
		Where("store_products.is_active = ?", true)

	search := strings.TrimSpace(req.Search)
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"(store_products.name LIKE ? OR store_products.product_code LIKE ? OR store_products.category LIKE ?)",
			keyword,
			keyword,
			keyword,
		)
	}

	switch strings.ToUpper(strings.TrimSpace(req.Status)) {
	case constants.ProductStatusSafe:
		query = query.Where("store_products.stock_quantity > store_products.min_stock_quantity")
	case constants.ProductStatusLow:
		query = query.Where("store_products.stock_quantity <= store_products.min_stock_quantity")
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("store_products.created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range results {
		enrichProductResponse(&results[i])
	}

	return results, total, nil
}

func (r *StoreRepository) CountProductsByCooperative(tx *gorm.DB, cooperativeID uuid.UUID) (int64, error) {
	var total int64

	err := tx.Debug().
		Model(&entity.StoreProduct{}).
		Where("cooperative_id = ?", cooperativeID).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *StoreRepository) UpdateProduct(tx *gorm.DB, product *entity.StoreProduct) error {
	err := tx.Debug().Save(product).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *StoreRepository) CreateStockMovement(tx *gorm.DB, movement *entity.StockMovement) error {
	err := tx.Debug().Create(movement).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *StoreRepository) GetLatestStockMovementForUpdate(tx *gorm.DB, cooperativeID uuid.UUID) (*entity.StockMovement, error) {
	var movement entity.StockMovement

	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cooperative_id = ?", cooperativeID).
		Order("created_at DESC").
		First(&movement).Error
	if err != nil {
		return nil, err
	}

	return &movement, nil
}

func (r *StoreRepository) GetStockMovementByClientReferenceID(tx *gorm.DB, cooperativeID uuid.UUID, clientReferenceID string) (*entity.StockMovement, error) {
	var movement entity.StockMovement

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("client_reference_id = ?", clientReferenceID).
		First(&movement).Error
	if err != nil {
		return nil, err
	}

	return &movement, nil
}

func (r *StoreRepository) GetStockMovementResponseByID(tx *gorm.DB, cooperativeID, movementID uuid.UUID) (*model.StockMovementResponse, error) {
	var result model.StockMovementResponse

	err := baseStockMovementQuery(tx).
		Where("stock_movements.cooperative_id = ?", cooperativeID).
		Where("stock_movements.stock_movement_id = ?", movementID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	enrichStockMovementResponse(&result)
	return &result, nil
}

func (r *StoreRepository) ListStockMovements(tx *gorm.DB, req model.ListStockMovementsRequest) ([]model.StockMovementResponse, int64, error) {
	var results []model.StockMovementResponse
	var total int64

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	query := baseStockMovementQuery(tx).
		Where("stock_movements.cooperative_id = ?", req.CooperativeID)

	if req.ProductID != uuid.Nil {
		query = query.Where("stock_movements.product_id = ?", req.ProductID)
	}

	movementType := strings.ToUpper(strings.TrimSpace(req.Type))
	if movementType != "" {
		query = query.Where("stock_movements.movement_type = ?", movementType)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("stock_movements.recorded_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range results {
		enrichStockMovementResponse(&results[i])
	}

	return results, total, nil
}

func (r *StoreRepository) GetStoreDashboard(tx *gorm.DB, cooperativeID uuid.UUID) (*model.StoreDashboardResponse, error) {
	var summary struct {
		TotalProducts    int64
		SafeProducts     int64
		LowStockProducts int64
	}

	err := tx.Debug().
		Table("store_products").
		Select(`
			COUNT(*) AS total_products,
			COALESCE(SUM(CASE WHEN stock_quantity > min_stock_quantity THEN 1 ELSE 0 END), 0) AS safe_products,
			COALESCE(SUM(CASE WHEN stock_quantity <= min_stock_quantity THEN 1 ELSE 0 END), 0) AS low_stock_products
		`).
		Where("cooperative_id = ?", cooperativeID).
		Where("is_active = ?", true).
		Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	movements, _, err := r.ListStockMovements(tx, model.ListStockMovementsRequest{
		AuthContext: model.AuthContext{CooperativeID: cooperativeID},
		Page:        1,
		Limit:       5,
	})
	if err != nil {
		return nil, err
	}

	return &model.StoreDashboardResponse{
		TotalProducts:    summary.TotalProducts,
		SafeProducts:     summary.SafeProducts,
		LowStockProducts: summary.LowStockProducts,
		LatestMovements:  movements,
	}, nil
}

func (r *StoreRepository) CreateSale(tx *gorm.DB, sale *entity.StoreSale) error {
	err := tx.Debug().Create(sale).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *StoreRepository) GetSaleByClientID(tx *gorm.DB, cooperativeID uuid.UUID, clientSaleID string) (*entity.StoreSale, error) {
	var sale entity.StoreSale

	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("client_sale_id = ?", clientSaleID).
		First(&sale).Error
	if err != nil {
		return nil, err
	}

	return &sale, nil
}

func (r *StoreRepository) GetSaleDetail(tx *gorm.DB, cooperativeID, saleID uuid.UUID) (*model.StoreSaleResponse, error) {
	var saleHeader struct {
		StoreSaleID   uuid.UUID
		CooperativeID uuid.UUID
		OfficerID     uuid.UUID
		OfficerName   string
		SaleNumber    string
		TotalAmount   int64
		TotalCost     int64
		CashReceived  int64
		ChangeAmount  int64
		Status        string
		RecordedAt    time.Time
		ClientSaleID  string
		CreatedAt     time.Time
		CurrentHash   string
	}

	err := tx.Debug().
		Table("store_sales").
		Select(`
			store_sales.store_sale_id,
			store_sales.cooperative_id,
			store_sales.officer_id,
			officer_users.full_name AS officer_name,
			store_sales.sale_number,
			store_sales.total_amount,
			store_sales.total_cost,
			store_sales.cash_received,
			store_sales.change_amount,
			store_sales.status,
			store_sales.recorded_at,
			store_sales.client_sale_id,
			store_sales.created_at,
			(
				SELECT stock_movements.current_hash
				FROM stock_movements
				WHERE stock_movements.reference_type = 'STORE_SALE'
					AND stock_movements.reference_id = store_sales.store_sale_id
				ORDER BY stock_movements.created_at DESC
				LIMIT 1
			) AS current_hash
		`).
		Joins("JOIN users AS officer_users ON officer_users.user_id = store_sales.officer_id").
		Where("store_sales.cooperative_id = ?", cooperativeID).
		Where("store_sales.store_sale_id = ?", saleID).
		Scan(&saleHeader).Error
	if err != nil {
		return nil, err
	}
	if saleHeader.StoreSaleID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	var items []model.StoreSaleItemResponse
	err = tx.Debug().
		Table("store_sale_items").
		Select(`
			store_sale_items.store_sale_item_id,
			store_sale_items.store_sale_id,
			store_sale_items.product_id,
			store_products.product_code,
			store_products.name AS product_name,
			store_products.unit,
			store_sale_items.quantity,
			store_sale_items.unit_price,
			store_sale_items.unit_cost,
			store_sale_items.subtotal
		`).
		Joins("JOIN store_products ON store_products.product_id = store_sale_items.product_id").
		Where("store_sale_items.store_sale_id = ?", saleID).
		Order("store_sale_items.created_at ASC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}

	return &model.StoreSaleResponse{
		StoreSaleID:   saleHeader.StoreSaleID,
		CooperativeID: saleHeader.CooperativeID,
		OfficerID:     saleHeader.OfficerID,
		OfficerName:   saleHeader.OfficerName,
		SaleNumber:    saleHeader.SaleNumber,
		TotalAmount:   saleHeader.TotalAmount,
		TotalCost:     saleHeader.TotalCost,
		GrossProfit:   saleHeader.TotalAmount - saleHeader.TotalCost,
		CashReceived:  saleHeader.CashReceived,
		ChangeAmount:  saleHeader.ChangeAmount,
		Status:        saleHeader.Status,
		RecordedAt:    saleHeader.RecordedAt,
		ClientSaleID:  saleHeader.ClientSaleID,
		CreatedAt:     saleHeader.CreatedAt,
		CurrentHash:   saleHeader.CurrentHash,
		HashPreview:   buildStoreHashPreview(saleHeader.CurrentHash),
		Items:         items,
	}, nil
}

func (r *StoreRepository) GenerateSaleNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error) {
	var total int64

	err := tx.Debug().
		Model(&entity.StoreSale{}).
		Where("cooperative_id = ?", cooperativeID).
		Count(&total).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("INV-%04d", total+1), nil
}

func baseProductQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("store_products").
		Select(`
			store_products.product_id,
			store_products.cooperative_id,
			store_products.product_code,
			store_products.name,
			store_products.unit,
			store_products.category,
			store_products.cost_price,
			store_products.sale_price,
			store_products.stock_quantity,
			store_products.min_stock_quantity,
			store_products.created_at,
			store_products.updated_at
		`)
}

func baseStockMovementQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("stock_movements").
		Select(`
			stock_movements.stock_movement_id,
			stock_movements.cooperative_id,
			stock_movements.product_id,
			store_products.product_code,
			store_products.name AS product_name,
			store_products.unit,
			stock_movements.officer_id,
			officer_users.full_name AS officer_name,
			stock_movements.movement_type,
			stock_movements.quantity_delta,
			stock_movements.resulting_stock_quantity,
			stock_movements.unit_cost,
			store_products.sale_price,
			stock_movements.reference_type,
			stock_movements.reference_id,
			stock_movements.description,
			stock_movements.recorded_at,
			stock_movements.synced_at,
			stock_movements.current_hash,
			stock_movements.is_offline_created,
			stock_movements.client_reference_id
		`).
		Joins("JOIN store_products ON store_products.product_id = stock_movements.product_id").
		Joins("JOIN users AS officer_users ON officer_users.user_id = stock_movements.officer_id")
}

func enrichProductResponse(product *model.ProductResponse) {
	if product.SalePrice > 0 {
		product.MarginPercent = int(((product.SalePrice - product.CostPrice) * 100) / product.SalePrice)
	}

	if product.StockQuantity <= product.MinStockQuantity {
		product.StockStatus = constants.ProductStatusLow
		return
	}

	product.StockStatus = constants.ProductStatusSafe
}

func enrichStockMovementResponse(movement *model.StockMovementResponse) {
	movement.MovementTypeLabel = getStockMovementTypeLabel(movement.MovementType)
	movement.AdjustmentType = getStockAdjustmentType(movement)
	movement.HashPreview = buildStoreHashPreview(movement.CurrentHash)
	movement.SyncStatus = constants.SyncStatusSynced

	if movement.QuantityDelta < 0 {
		movement.StockValueDelta = -movement.QuantityDelta * movement.UnitCost
		return
	}

	movement.StockValueDelta = movement.QuantityDelta * movement.UnitCost
}

func getStockMovementTypeLabel(movementType string) string {
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

func getStockAdjustmentType(movement *model.StockMovementResponse) string {
	if movement.MovementType != constants.StockMovementAdjustment {
		return ""
	}

	if movement.QuantityDelta < 0 {
		return "PENGURANGAN"
	}

	return "PENAMBAHAN"
}

func buildStoreHashPreview(hash string) string {
	if len(hash) <= 12 {
		return hash
	}

	return hash[:12] + "..."
}

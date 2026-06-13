package model

import (
	"time"

	"github.com/google/uuid"
)

type StoreDashboardRequest struct {
	AuthContext
}

type StoreDashboardResponse struct {
	TotalProducts    int64                   `json:"total_products"`
	SafeProducts     int64                   `json:"safe_products"`
	LowStockProducts int64                   `json:"low_stock_products"`
	LatestMovements  []StockMovementResponse `json:"latest_movements"`
}

type ListProductsRequest struct {
	AuthContext

	Search string `form:"search"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type CreateProductRequest struct {
	AuthContext

	Name                 string     `json:"name"`
	Unit                 string     `json:"unit"`
	Category             string     `json:"category"`
	CostPrice            int64      `json:"cost_price"`
	SalePrice            int64      `json:"sale_price"`
	MinStockQuantity     int64      `json:"min_stock_quantity"`
	InitialStockQuantity int64      `json:"initial_stock_quantity"`
	RecordedAt           *time.Time `json:"recorded_at"`
	IsOfflineCreated     bool       `json:"is_offline_created"`
	ClientReferenceID    string     `json:"client_reference_id"`
}

type ProductResponse struct {
	ProductID        uuid.UUID `json:"product_id"`
	ProductCode      string    `json:"product_code"`
	Name             string    `json:"name"`
	Unit             string    `json:"unit"`
	Category         string    `json:"category"`
	CostPrice        int64     `json:"cost_price"`
	SalePrice        int64     `json:"sale_price"`
	StockQuantity    int64     `json:"stock_quantity"`
	MinStockQuantity int64     `json:"min_stock_quantity"`
	MarginPercent    int       `json:"margin_percent"`
	StockStatus      string    `json:"stock_status"`
	HashPreview      string    `json:"hash_preview,omitempty"`
}

type CreateStockInRequest struct {
	AuthContext

	ProductID         uuid.UUID  `json:"product_id"`
	Quantity          int64      `json:"quantity"`
	UnitCost          int64      `json:"unit_cost"`
	SalePrice         int64      `json:"sale_price"`
	Description       string     `json:"description"`
	RecordedAt        *time.Time `json:"recorded_at"`
	IsOfflineCreated  bool       `json:"is_offline_created"`
	ClientReferenceID string     `json:"client_reference_id"`
}

type CreateStockAdjustmentRequest struct {
	AuthContext

	ProductID         uuid.UUID  `json:"product_id"`
	QuantityDelta     int64      `json:"quantity_delta"`
	Description       string     `json:"description"`
	RecordedAt        *time.Time `json:"recorded_at"`
	IsOfflineCreated  bool       `json:"is_offline_created"`
	ClientReferenceID string     `json:"client_reference_id"`
}

type CreateStoreSaleRequest struct {
	AuthContext

	Items        []CreateStoreSaleItemRequest `json:"items"`
	CashReceived int64                        `json:"cash_received"`
	RecordedAt   *time.Time                   `json:"recorded_at"`
	ClientSaleID string                       `json:"client_sale_id"`
}

type CreateStoreSaleItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

type ListProductsResponse struct {
	Items []ProductResponse `json:"items"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int64             `json:"total"`
}

type ListStockMovementsRequest struct {
	AuthContext
	ProductID    uuid.UUID
	ProductIDRaw string `form:"product_id"`
	Type         string `form:"type"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type ListStockMovementsResponse struct {
	Items []StockMovementResponse `json:"items"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
	Total int64                   `json:"total"`
}

type StockMovementResponse struct {
	StockMovementID        uuid.UUID  `json:"stock_movement_id"`
	CooperativeID          uuid.UUID  `json:"cooperative_id"`
	ProductID              uuid.UUID  `json:"product_id"`
	ProductCode            string     `json:"product_code"`
	ProductName            string     `json:"product_name"`
	Unit                   string     `json:"unit"`
	OfficerID              uuid.UUID  `json:"officer_id"`
	OfficerName            string     `json:"officer_name"`
	MovementType           string     `json:"movement_type"`
	MovementTypeLabel      string     `json:"movement_type_label"`
	AdjustmentType         string     `json:"adjustment_type,omitempty"`
	QuantityDelta          int64      `json:"quantity_delta"`
	ResultingStockQuantity int64      `json:"resulting_stock_quantity"`
	UnitCost               int64      `json:"unit_cost"`
	SalePrice              int64      `json:"sale_price"`
	StockValueDelta        int64      `json:"stock_value_delta"`
	ReferenceType          string     `json:"reference_type"`
	ReferenceID            *uuid.UUID `json:"reference_id,omitempty"`
	Description            string     `json:"description"`
	RecordedAt             time.Time  `json:"recorded_at"`
	SyncedAt               *time.Time `json:"synced_at"`
	CurrentHash            string     `json:"current_hash"`
	HashPreview            string     `json:"hash_preview"`
	IsOfflineCreated       bool       `json:"is_offline_created"`
	ClientReferenceID      string     `json:"client_reference_id"`
	SyncStatus             string     `json:"sync_status"`
}

type StoreSaleResponse struct {
	StoreSaleID   uuid.UUID               `json:"store_sale_id"`
	CooperativeID uuid.UUID               `json:"cooperative_id"`
	OfficerID     uuid.UUID               `json:"officer_id"`
	OfficerName   string                  `json:"officer_name"`
	SaleNumber    string                  `json:"sale_number"`
	TotalAmount   int64                   `json:"total_amount"`
	TotalCost     int64                   `json:"total_cost"`
	GrossProfit   int64                   `json:"gross_profit"`
	CashReceived  int64                   `json:"cash_received"`
	ChangeAmount  int64                   `json:"change_amount"`
	Status        string                  `json:"status"`
	RecordedAt    time.Time               `json:"recorded_at"`
	ClientSaleID  string                  `json:"client_sale_id"`
	CreatedAt     time.Time               `json:"created_at"`
	CurrentHash   string                  `json:"current_hash"`
	HashPreview   string                  `json:"hash_preview"`
	Items         []StoreSaleItemResponse `json:"items"`
}

type StoreSaleItemResponse struct {
	StoreSaleItemID uuid.UUID `json:"store_sale_item_id"`
	StoreSaleID     uuid.UUID `json:"store_sale_id"`
	ProductID       uuid.UUID `json:"product_id"`
	ProductCode     string    `json:"product_code"`
	ProductName     string    `json:"product_name"`
	Unit            string    `json:"unit"`
	Quantity        int64     `json:"quantity"`
	UnitPrice       int64     `json:"unit_price"`
	UnitCost        int64     `json:"unit_cost"`
	Subtotal        int64     `json:"subtotal"`
}

type GetProductRequest struct {
	AuthContext
	ProductID uuid.UUID `uri:"productID"`
}

type GetStoreSaleRequest struct {
	AuthContext
	StoreSaleID uuid.UUID `uri:"saleID"`
}

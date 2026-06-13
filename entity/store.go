package entity

import (
	"time"

	"github.com/google/uuid"
)

type StoreProduct struct {
	ProductID        uuid.UUID  `json:"product_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID    uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_product_code"`
	ProductCode      string     `json:"product_code" gorm:"type:varchar(30);not null;uniqueIndex:idx_coop_product_code"`
	Name             string     `json:"name" gorm:"type:varchar(120);not null;index"`
	Unit             string     `json:"unit" gorm:"type:varchar(20);not null"`
	Category         string     `json:"category" gorm:"type:varchar(80);not null;index"`
	CostPrice        int64      `json:"cost_price" gorm:"not null"`
	SalePrice        int64      `json:"sale_price" gorm:"not null"`
	StockQuantity    int64      `json:"stock_quantity" gorm:"not null;default:0"`     // centi-unit: 980 = 9.80 kg
	MinStockQuantity int64      `json:"min_stock_quantity" gorm:"not null;default:0"` // centi-unit
	IsActive         bool       `json:"is_active" gorm:"not null;default:true"`
	CreatedBy        uuid.UUID  `json:"created_by" gorm:"type:varchar(36);not null;index"`
	UpdatedBy        *uuid.UUID `json:"updated_by" gorm:"type:varchar(36);index"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	StoreSaleItems []StoreSaleItem `json:"store_sale_items" gorm:"foreignKey:ProductID;constraint:onDelete:CASCADE"`
	StockMovements []StockMovement `json:"stock_movements" gorm:"foreignKey:ProductID;constraint:onDelete:CASCADE"`
}

type StockMovement struct {
	StockMovementID        uuid.UUID  `json:"stock_movement_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID          uuid.UUID  `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_client_stock_movement"`
	ProductID              uuid.UUID  `json:"product_id" gorm:"type:varchar(36);not null;index"`
	OfficerID              uuid.UUID  `json:"officer_id" gorm:"type:varchar(36);not null;index"`
	MovementType           string     `json:"movement_type" gorm:"type:enum('PRODUCT_CREATED','STOCK_IN','ADJUSTMENT','SALE_OUT');not null;index"`
	QuantityDelta          int64      `json:"quantity_delta" gorm:"not null"`
	ResultingStockQuantity int64      `json:"resulting_stock_quantity" gorm:"not null"`
	UnitCost               int64      `json:"unit_cost" gorm:"not null;default:0"`
	ReferenceType          string     `json:"reference_type" gorm:"type:varchar(40)"`
	ReferenceID            *uuid.UUID `json:"reference_id" gorm:"type:varchar(36);index"`
	Description            string     `json:"description" gorm:"type:text"`
	RecordedAt             time.Time  `json:"recorded_at" gorm:"not null;index"`
	SyncedAt               *time.Time `json:"synced_at"`
	PrevHash               string     `json:"prev_hash" gorm:"type:varchar(64);not null"`
	CurrentHash            string     `json:"current_hash" gorm:"type:varchar(64);not null;uniqueIndex"`
	IsOfflineCreated       bool       `json:"is_offline_created" gorm:"default:false"`
	ClientReferenceID      string     `json:"client_reference_id" gorm:"type:varchar(100);uniqueIndex:idx_coop_client_stock_movement"`
	CreatedAt              time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

type StoreSale struct {
	StoreSaleID   uuid.UUID `json:"store_sale_id" gorm:"type:varchar(36);primaryKey"`
	CooperativeID uuid.UUID `json:"cooperative_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_coop_client_sale"`
	OfficerID     uuid.UUID `json:"officer_id" gorm:"type:varchar(36);not null;index"`
	SaleNumber    string    `json:"sale_number" gorm:"type:varchar(40);not null;index"`
	TotalAmount   int64     `json:"total_amount" gorm:"not null"`
	TotalCost     int64     `json:"total_cost" gorm:"not null"`
	CashReceived  int64     `json:"cash_received" gorm:"not null;default:0"`
	ChangeAmount  int64     `json:"change_amount" gorm:"not null;default:0"`
	Status        string    `json:"status" gorm:"type:enum('COMPLETED');not null;default:'COMPLETED'"`
	RecordedAt    time.Time `json:"recorded_at" gorm:"not null;index"`
	ClientSaleID  string    `json:"client_sale_id" gorm:"type:varchar(100);uniqueIndex:idx_coop_client_sale"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`

	Items []StoreSaleItem `json:"items" gorm:"foreignKey:StoreSaleID;constraint:onDelete:CASCADE"`
}

type StoreSaleItem struct {
	StoreSaleItemID uuid.UUID `json:"store_sale_item_id" gorm:"type:varchar(36);primaryKey"`
	StoreSaleID     uuid.UUID `json:"store_sale_id" gorm:"type:varchar(36);not null;index"`
	ProductID       uuid.UUID `json:"product_id" gorm:"type:varchar(36);not null;index"`
	Quantity        int64     `json:"quantity" gorm:"not null"`
	UnitPrice       int64     `json:"unit_price" gorm:"not null"`
	UnitCost        int64     `json:"unit_cost" gorm:"not null"`
	Subtotal        int64     `json:"subtotal" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

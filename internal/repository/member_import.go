package repository

import (
	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMemberImportRepository interface {
	CreateBatch(tx *gorm.DB, batch *entity.MemberImportBatch) error
	CreateRows(tx *gorm.DB, rows []entity.MemberImportRow) error
	GetBatch(tx *gorm.DB, cooperativeID uuid.UUID, batchID uuid.UUID) (*entity.MemberImportBatch, error)
	UpdateBatch(tx *gorm.DB, batch *entity.MemberImportBatch) error
	GetRow(tx *gorm.DB, batchID uuid.UUID, rowID uuid.UUID) (*entity.MemberImportRow, error)
	UpdateRow(tx *gorm.DB, row *entity.MemberImportRow) error
	ListRows(tx *gorm.DB, req model.GetMemberImportRequest) ([]entity.MemberImportRow, int64, error)
	ListRowsForSubmit(tx *gorm.DB, batchID uuid.UUID) ([]entity.MemberImportRow, error)
	RecalculateBatchSummary(tx *gorm.DB, batchID uuid.UUID) (totalRows, successRows, errorRows int, err error)
}

type MemberImportRepository struct {
	db *gorm.DB
}

func NewMemberImportRepository(db *gorm.DB) IMemberImportRepository {
	return &MemberImportRepository{db: db}
}

func (r *MemberImportRepository) CreateBatch(tx *gorm.DB, batch *entity.MemberImportBatch) error {
	err := tx.Debug().Create(batch).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberImportRepository) CreateRows(tx *gorm.DB, rows []entity.MemberImportRow) error {
	if len(rows) == 0 {
		return nil
	}

	err := tx.Debug().Create(&rows).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberImportRepository) GetBatch(tx *gorm.DB, cooperativeID uuid.UUID, batchID uuid.UUID) (*entity.MemberImportBatch, error) {
	var batch entity.MemberImportBatch
	err := tx.Debug().
		Where("cooperative_id = ?", cooperativeID).
		Where("import_batch_id = ?", batchID).
		First(&batch).Error
	if err != nil {
		return nil, err
	}

	return &batch, nil
}

func (r *MemberImportRepository) UpdateBatch(tx *gorm.DB, batch *entity.MemberImportBatch) error {
	err := tx.Debug().Save(batch).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberImportRepository) GetRow(tx *gorm.DB, batchID uuid.UUID, rowID uuid.UUID) (*entity.MemberImportRow, error) {
	var row entity.MemberImportRow
	err := tx.Debug().
		Where("import_batch_id = ?", batchID).
		Where("import_row_id = ?", rowID).
		First(&row).Error
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *MemberImportRepository) UpdateRow(tx *gorm.DB, row *entity.MemberImportRow) error {
	err := tx.Debug().Save(row).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MemberImportRepository) ListRows(tx *gorm.DB, req model.GetMemberImportRequest) ([]entity.MemberImportRow, int64, error) {
	var rows []entity.MemberImportRow
	var total int64

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	query := tx.Debug().
		Model(&entity.MemberImportRow{}).
		Where("import_batch_id = ?", req.ImportBatchID).
		Where("status <> ?", "DELETED")

	if req.Status != "" && req.Status != "SEMUA" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("`row_number` ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *MemberImportRepository) ListRowsForSubmit(tx *gorm.DB, batchID uuid.UUID) ([]entity.MemberImportRow, error) {
	var rows []entity.MemberImportRow
	err := tx.Debug().
		Where("import_batch_id = ?", batchID).
		Where("status = ?", "VALID").
		Order("`row_number` ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *MemberImportRepository) RecalculateBatchSummary(tx *gorm.DB, batchID uuid.UUID) (int, int, int, error) {
	var totalRows int64
	var successRows int64
	var errorRows int64

	base := tx.Debug().Model(&entity.MemberImportRow{}).
		Where("import_batch_id = ?", batchID).
		Where("status <> ?", "DELETED")

	if err := base.Count(&totalRows).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := base.Where("status = ?", "VALID").Count(&successRows).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := base.Where("status = ?", "ERROR").Count(&errorRows).Error; err != nil {
		return 0, 0, 0, err
	}

	return int(totalRows), int(successRows), int(errorRows), nil
}

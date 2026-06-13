package repository

import (
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICreditAccessRepository interface {
	GetActivePartnerByName(tx *gorm.DB, name string) (*entity.Partner, error)
	ListCreditAccessRequests(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) ([]CreditAccessRequestRow, error)
	GetCreditAccessRequestDetail(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, requestID uuid.UUID) (*CreditAccessRequestRow, error)
	GetCreditAccessRequestForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, requestID uuid.UUID) (*entity.CreditAccessRequest, error)
	GetActiveConsentByRequest(tx *gorm.DB, requestID uuid.UUID) (*entity.MemberDataConsent, error)
	CreateCreditAccessRequest(tx *gorm.DB, request *entity.CreditAccessRequest) error
	UpdateCreditAccessRequest(tx *gorm.DB, request *entity.CreditAccessRequest) error
	CreateMemberDataConsent(tx *gorm.DB, consent *entity.MemberDataConsent) error
	UpdateMemberDataConsent(tx *gorm.DB, consent *entity.MemberDataConsent) error
}

type CreditAccessRepository struct {
	db *gorm.DB
}

type CreditAccessRequestRow struct {
	RequestID             uuid.UUID
	PartnerID             uuid.UUID
	PartnerName           string
	PartnerType           string
	OJKRegistrationNumber string
	MCSGrade              string
	Status                string
	RequestedAmount       int64
	Purpose               string
	DataScopeJSON         string
	RequestedAt           time.Time
	GrantedAt             *time.Time
	AccessExpiresAt       *time.Time
	DurationDays          int
}

func NewCreditAccessRepository(db *gorm.DB) ICreditAccessRepository {
	return &CreditAccessRepository{db: db}
}

func (r *CreditAccessRepository) GetActivePartnerByName(tx *gorm.DB, name string) (*entity.Partner, error) {
	var partner entity.Partner
	err := tx.Debug().
		Where("name = ?", name).
		Where("status = ?", "ACTIVE").
		First(&partner).Error
	if err != nil {
		return nil, err
	}

	return &partner, nil
}

func (r *CreditAccessRepository) ListCreditAccessRequests(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID) ([]CreditAccessRequestRow, error) {
	var rows []CreditAccessRequestRow
	err := baseCreditAccessRequestDetailQuery(tx).
		Where("credit_access_requests.cooperative_id = ?", cooperativeID).
		Where("credit_access_requests.member_id = ?", memberID).
		Order("credit_access_requests.requested_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *CreditAccessRepository) GetCreditAccessRequestDetail(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, requestID uuid.UUID) (*CreditAccessRequestRow, error) {
	var row CreditAccessRequestRow
	err := baseCreditAccessRequestDetailQuery(tx).
		Where("credit_access_requests.request_id = ?", requestID).
		Where("credit_access_requests.cooperative_id = ?", cooperativeID).
		Where("credit_access_requests.member_id = ?", memberID).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.RequestID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &row, nil
}

func (r *CreditAccessRepository) GetCreditAccessRequestForUpdate(tx *gorm.DB, cooperativeID uuid.UUID, memberID uuid.UUID, requestID uuid.UUID) (*entity.CreditAccessRequest, error) {
	var request entity.CreditAccessRequest
	err := tx.Debug().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("request_id = ?", requestID).
		Where("cooperative_id = ?", cooperativeID).
		Where("member_id = ?", memberID).
		First(&request).Error
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *CreditAccessRepository) GetActiveConsentByRequest(tx *gorm.DB, requestID uuid.UUID) (*entity.MemberDataConsent, error) {
	var consent entity.MemberDataConsent
	err := tx.Debug().
		Where("request_id = ?", requestID).
		Where("is_active = ?", true).
		First(&consent).Error
	if err != nil {
		return nil, err
	}

	return &consent, nil
}

func (r *CreditAccessRepository) CreateCreditAccessRequest(tx *gorm.DB, request *entity.CreditAccessRequest) error {
	err := tx.Debug().Create(request).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccessRepository) UpdateCreditAccessRequest(tx *gorm.DB, request *entity.CreditAccessRequest) error {
	err := tx.Debug().Save(request).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccessRepository) CreateMemberDataConsent(tx *gorm.DB, consent *entity.MemberDataConsent) error {
	err := tx.Debug().Create(consent).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccessRepository) UpdateMemberDataConsent(tx *gorm.DB, consent *entity.MemberDataConsent) error {
	err := tx.Debug().Save(consent).Error
	if err != nil {
		return err
	}

	return nil
}

func baseCreditAccessRequestDetailQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("credit_access_requests").
		Select(`
			credit_access_requests.request_id,
			partners.partner_id,
			partners.name AS partner_name,
			partners.partner_type,
			partners.ojk_registration_number,
			members.mcs_grade,
			credit_access_requests.status,
			credit_access_requests.requested_amount,
			credit_access_requests.purpose,
			credit_access_requests.data_scope_json,
			credit_access_requests.requested_at,
			credit_access_requests.granted_at,
			credit_access_requests.access_expires_at,
			COALESCE(member_data_consents.duration_days, 0) AS duration_days
		`).
		Joins("JOIN partners ON partners.partner_id = credit_access_requests.partner_id").
		Joins("JOIN members ON members.member_id = credit_access_requests.member_id").
		Joins("LEFT JOIN member_data_consents ON member_data_consents.request_id = credit_access_requests.request_id")
}

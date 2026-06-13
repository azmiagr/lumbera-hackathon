package service

import (
	"errors"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const mcsXGBoostModelVersion = "MCS_XGBOOST_V1"

type IMCSService interface {
	TriggerMemberMCS(req model.TriggerMemberMCSRequest) (*model.TriggerMemberMCSResponse, error)
	ApplyMCSCallback(req model.MCSCallbackRequest) (*model.MCSCallbackResponse, error)
}

type MCSService struct {
	deps serviceDependency
}

func NewMCSService(deps serviceDependency) IMCSService {
	return &MCSService{deps: deps}
}

func (s *MCSService) TriggerMemberMCS(req model.TriggerMemberMCSRequest) (*model.TriggerMemberMCSResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.MemberID == uuid.Nil {
		return nil, appErrors.BadRequest("member_id wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	_, err := s.deps.repository.MemberRepository.GetActiveMember(tx, req.CooperativeID, req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil anggota")
	}

	payload, err := s.deps.repository.MCSRepository.GetLatestTrainingFeatures(tx, req.CooperativeID, req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("fitur MCS anggota belum tersedia")
		}
		return nil, appErrors.InternalServer("gagal mengambil fitur MCS")
	}

	requestID := uuid.New()
	payload["request_id"] = requestID
	payload["member_id"] = req.MemberID

	err = s.deps.mcsAPI.SendMCSRequest(payload)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengirim request MCS ke API scoring")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menjalankan MCS")
	}

	return &model.TriggerMemberMCSResponse{
		RequestID:    requestID,
		MemberID:     req.MemberID,
		Status:       "PROCESSING",
		WebhookState: "SENT",
	}, nil
}

func (s *MCSService) ApplyMCSCallback(req model.MCSCallbackRequest) (*model.MCSCallbackResponse, error) {
	if req.RequestID == uuid.Nil || req.MemberID == uuid.Nil {
		return nil, appErrors.BadRequest("request_id dan member_id wajib diisi")
	}

	if req.MCSScore < 300 || req.MCSScore > 850 {
		return nil, appErrors.BadRequest("mcs_score harus di antara 300 sampai 850")
	}

	grade := strings.ToUpper(strings.TrimSpace(req.MCSGrade))
	if !isValidMCSGrade(grade) {
		return nil, appErrors.BadRequest("mcs_grade tidak valid")
	}

	status := strings.ToUpper(strings.TrimSpace(req.CalculationStatus))
	if status == "" {
		status = "COMPLETE"
	}
	if status != "COMPLETE" && status != "FAILED" {
		return nil, appErrors.BadRequest("calculation_status tidak valid")
	}

	modelVersion := strings.TrimSpace(req.ModelVersion)
	if modelVersion == "" {
		modelVersion = mcsXGBoostModelVersion
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	member, err := s.deps.repository.MemberRepository.GetActiveMember(tx, uuid.Nil, req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("anggota aktif tidak ditemukan")
		}
		return nil, appErrors.InternalServer("gagal mengambil anggota")
	}

	now := time.Now()
	snapshot := &entity.MCSScoreSnapshot{
		SnapshotID:             uuid.New(),
		RequestID:              req.RequestID,
		CooperativeID:          member.CooperativeID,
		MemberID:               req.MemberID,
		MCSScore:               req.MCSScore,
		MCSGrade:               grade,
		Eligible:               req.Eligible,
		EligibilityProbability: req.EligibilityProbability,
		ModelVersion:           modelVersion,
		CalculationStatus:      status,
		Explanation:            strings.TrimSpace(req.Explanation),
		CalculatedAt:           now,
	}

	err = s.deps.repository.MCSRepository.CreateScoreSnapshot(tx, snapshot)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan snapshot MCS")
	}

	if status == "COMPLETE" {
		err := s.deps.repository.MCSRepository.UpdateMemberCurrentScore(tx, req.MemberID, req.MCSScore, grade, now)
		if err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui skor MCS anggota")
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal memproses callback MCS")
	}

	return &model.MCSCallbackResponse{
		MemberID:           req.MemberID,
		MCSScore:           req.MCSScore,
		MCSGrade:           grade,
		LastScoreUpdatedAt: now,
	}, nil
}

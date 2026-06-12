package service

import (
	"crypto/rand"
	"encoding/base64"
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

const (
	registrationOTPExpiry     = 5 * time.Minute
	registrationDraftTTL      = 24 * time.Hour
	maxRegistrationOTPAttempt = 5
)

type IOfficerRegistrationService interface {
	StartRegistration(req model.StartOfficerRegistrationRequest) (*model.StartOfficerRegistrationResponse, error)
	VerifyOTP(req model.VerifyOfficerRegistrationOTPRequest) error
	SetPIN(req model.SetOfficerRegistrationPINRequest) (*model.SetOfficerRegistrationPINResponse, error)
	UpdatePersonalData(req model.UpdatePersonalDataRequest) (*model.OnboardingStepResponse, error)
	UpdateCooperativeType(req model.UpdateCooperativeTypeRequest) (*model.OnboardingStepResponse, error)
	UpdateCooperativeProfile(req model.UpdateCooperativeProfileRequest) (*model.OnboardingStepResponse, error)
	UpdateFinancialConfiguration(req model.UpdateFinancialConfigurationRequest) (*model.OnboardingStepResponse, error)
	UpdateCooperativeBankAccount(req model.UpdateCooperativeBankAccountRequest) (*model.OnboardingStepResponse, error)
	ActivateOnboardingDraft(req model.ActivateOnboardingDraftRequest) (*model.ActivateOnboardingDraftResponse, error)
	GetOnboardingState(req model.GetOnboardingStateRequest) (*model.GetOnboardingStateResponse, error)
}

type OfficerRegistrationService struct {
	deps serviceDependency
}

func NewOfficerRegistrationService(deps serviceDependency) IOfficerRegistrationService {
	return &OfficerRegistrationService{deps: deps}
}

func (s *OfficerRegistrationService) StartRegistration(req model.StartOfficerRegistrationRequest) (*model.StartOfficerRegistrationResponse, error) {
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, appErrors.BadRequest("nomor handphone wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	_, err := s.deps.repository.UserRepository.GetUser(tx, model.GetUserParam{
		PhoneNumber: phoneNumber,
	})
	if err == nil {
		return nil, appErrors.Conflict("nomor handphone sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memverifikasi nomor handphone")
	}

	draft, err := s.deps.repository.OnboardingDraftRepository.GetActiveOnboardingDraftByPhone(tx, phoneNumber)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("gagal memverifikasi proses onboarding")
		}

		draft = &entity.OnboardingDraft{
			OnboardingDraftID: uuid.New(),
			PhoneNumber:       phoneNumber,
			CurrentStep:       0,
			Status:            "OTP_PENDING",
			ExpiresAt:         time.Now().Add(registrationDraftTTL),
		}

		err = s.deps.repository.OnboardingDraftRepository.CreateOnboardingDraft(tx, draft)
		if err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan data onboarding")
		}
	} else {
		draft.Status = "OTP_PENDING"
		draft.ExpiresAt = time.Now().Add(registrationDraftTTL)
		draft.PhoneVerifiedAt = nil
		draft.PINHash = ""
		draft.SessionTokenHash = ""
		draft.PINSetAt = nil

		err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
		if err != nil {
			return nil, appErrors.InternalServer("gagal mereset proses onboarding")
		}
	}

	err = s.deps.repository.PhoneVerificationRepository.DeletePhoneVerificationChallenges(tx, phoneNumber, "REGISTRATION")
	if err != nil {
		return nil, appErrors.InternalServer("gagal menghapus OTP lama")
	}

	otp, err := generateNumericOTP(6)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk membuat kode otp")
	}

	otpHash, err := s.deps.bcrypt.GenerateFromPassword(otp)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk membuat hash otp")
	}

	challenge := &entity.PhoneVerificationChallenge{
		ChallengeID:  uuid.New(),
		PhoneNumber:  phoneNumber,
		OTPHash:      otpHash,
		Purpose:      "REGISTRATION",
		AttemptCount: 0,
		ExpiresAt:    time.Now().Add(registrationOTPExpiry),
	}

	err = s.deps.repository.PhoneVerificationRepository.CreatePhoneVerificationChallenge(tx, challenge)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan data challenge")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan data user dan challenge")
	}

	message := fmt.Sprintf("Kode OTP LUMBERA Anda adalah *%s*. Berlaku selama 5 menit. Jangan bagikan kode ini kepada siapa pun.", otp)
	err = s.deps.whatsapp.SendMessage(phoneNumber, message)
	if err != nil {
		return nil, err
	}

	return &model.StartOfficerRegistrationResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		PhoneNumber:       phoneNumber,
		ExpiresInSeconds:  int(registrationOTPExpiry.Seconds()),
	}, nil
}

func (s *OfficerRegistrationService) VerifyOTP(req model.VerifyOfficerRegistrationOTPRequest) error {
	if req.OnboardingDraftID == uuid.Nil {
		return appErrors.BadRequest("id draft tidak valid")
	}

	if strings.TrimSpace(req.OTP) == "" {
		return appErrors.BadRequest("otp wajib diisi")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.deps.repository.OnboardingDraftRepository.GetOnboardingDraft(tx, model.GetOnboardingDraftParam{
		OnboardingDraftID: req.OnboardingDraftID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("draft registrasi tidak ditemukan")
		}
		return err
	}

	if draft.Status != "OTP_PENDING" {
		return appErrors.BadRequest("draft registrasi tidak berada pada tahap OTP")
	}

	if time.Now().After(draft.ExpiresAt) {
		draft.Status = "EXPIRED"
		_ = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
		_ = tx.Commit().Error
		return appErrors.BadRequest("draft registrasi sudah kedaluwarsa")
	}

	challenge, err := s.deps.repository.PhoneVerificationRepository.GetPhoneVerificationChallenge(tx, model.GetPhoneVerificationChallengeParam{
		PhoneNumber: draft.PhoneNumber,
		Purpose:     "REGISTRATION",
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("OTP tidak ditemukan")
		}
		return err
	}

	if challenge.VerifiedAt != nil {
		return appErrors.BadRequest("OTP sudah digunakan")
	}

	if time.Now().After(challenge.ExpiresAt) {
		return appErrors.BadRequest("OTP sudah kedaluwarsa")
	}

	if challenge.AttemptCount >= maxRegistrationOTPAttempt {
		return appErrors.Forbidden("percobaan OTP sudah melewati batas")
	}

	err = s.deps.bcrypt.CompareAndHashPassword(challenge.OTPHash, req.OTP)
	if err != nil {
		challenge.AttemptCount++
		updateErr := s.deps.repository.PhoneVerificationRepository.UpdatePhoneVerificationChallenge(tx, challenge)
		if updateErr != nil {
			return updateErr
		}

		commitErr := tx.Commit().Error
		if commitErr != nil {
			return commitErr
		}
		return appErrors.Unauthorized("OTP tidak valid")
	}

	now := time.Now()
	challenge.VerifiedAt = &now
	draft.PhoneVerifiedAt = &now
	draft.Status = "OTP_VERIFIED"

	err = s.deps.repository.PhoneVerificationRepository.UpdatePhoneVerificationChallenge(tx, challenge)
	if err != nil {
		return err
	}

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

func (s *OfficerRegistrationService) SetPIN(req model.SetOfficerRegistrationPINRequest) (*model.SetOfficerRegistrationPINResponse, error) {
	if req.OnboardingDraftID == uuid.Nil {
		return nil, appErrors.BadRequest("id draft tidak valid")
	}

	if len(req.PIN) != 6 || len(req.ConfirmPIN) != 6 {
		return nil, appErrors.BadRequest("PIN harus 6 digit")
	}

	if req.PIN != req.ConfirmPIN {
		return nil, appErrors.BadRequest("konfirmasi PIN tidak sama")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.deps.repository.OnboardingDraftRepository.GetOnboardingDraft(tx, model.GetOnboardingDraftParam{
		OnboardingDraftID: req.OnboardingDraftID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("draft registrasi tidak ditemukan")
		}
		return nil, err
	}

	if draft.Status != "OTP_VERIFIED" {
		return nil, appErrors.BadRequest("nomor handphone belum diverifikasi")
	}

	if time.Now().After(draft.ExpiresAt) {
		draft.Status = "EXPIRED"
		_ = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
		_ = tx.Commit().Error
		return nil, appErrors.BadRequest("draft registrasi sudah kedaluwarsa")
	}

	pinHash, err := s.deps.bcrypt.GenerateFromPassword(req.PIN)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk membuat hash pin")
	}

	onboardingToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk membuat onboarding token")
	}

	onboardingTokenHash, err := s.deps.bcrypt.GenerateFromPassword(onboardingToken)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk membuat hash onboarding token")
	}

	nextStep := draft.CurrentStep
	if nextStep < 1 {
		nextStep = 1
	}

	now := time.Now()
	draft.PINHash = pinHash
	draft.SessionTokenHash = onboardingTokenHash
	draft.PINSetAt = &now
	draft.CurrentStep = nextStep
	if nextStep > 1 {
		draft.Status = "IN_PROGRESS"
	} else {
		draft.Status = "PIN_SET"
	}

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk menyimpan pin")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk menyimpan pin")
	}

	return &model.SetOfficerRegistrationPINResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		OnboardingToken:   onboardingToken,
		NextStep:          nextStep,
	}, nil
}

func (s *OfficerRegistrationService) UpdatePersonalData(req model.UpdatePersonalDataRequest) (*model.OnboardingStepResponse, error) {
	if req.KTPFile == nil || req.FullName == "" || req.NIKEncrypted == "" || req.NIKHash == "" || req.PositionCode == "" {
		return nil, appErrors.BadRequest("data diri belum lengkap")
	}

	if !isValidCooperativePosition(req.PositionCode) {
		return nil, appErrors.BadRequest("jabatan koperasi tidak valid")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	role, err := s.deps.repository.RoleRepository.GetRole(tx, model.GetRoleParam{
		Code:      constants.RoleCodePengurusKoperasi,
		ScopeType: constants.RoleScopeCooperative,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("role pengurus koperasi belum tersedia")
		}
		return nil, err
	}

	if req.ExistingCooperativeCode != "" {
		_, err := s.deps.repository.CooperativeRepository.GetCooperative(tx, model.GetCooperativeParam{
			CooperativeCode: req.ExistingCooperativeCode,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.NotFound("kode koperasi tidak ditemukan")
			}
			return nil, err
		}
	}

	ktpImageURL, err := s.deps.supabase.UploadImage(req.KTPFile, "ktp")
	if err != nil {
		return nil, err
	}

	draft.KTPImageURL = ktpImageURL
	draft.FullName = req.FullName
	draft.NIKEncrypted = req.NIKEncrypted
	draft.NIKHash = req.NIKHash
	draft.NIKMasked = req.NIKMasked
	draft.RoleCode = role.Code
	draft.PositionCode = req.PositionCode
	draft.ExistingCooperativeCode = req.ExistingCooperativeCode
	draft.CurrentStep = 1
	draft.Status = "IN_PROGRESS"

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk menyimpan data diri")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal untuk menyimpan data diri")
	}

	return &model.OnboardingStepResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		CurrentStep:       1,
		NextStep:          "COOPERATIVE_TYPE",
	}, nil
}

func (s *OfficerRegistrationService) UpdateCooperativeType(req model.UpdateCooperativeTypeRequest) (*model.OnboardingStepResponse, error) {
	if !isValidCooperativeType(req.CooperativeType) {
		return nil, appErrors.BadRequest("jenis koperasi tidak valid")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	if draft.CurrentStep < 1 {
		return nil, appErrors.BadRequest("lengkapi data diri terlebih dahulu")
	}

	draft.CooperativeType = req.CooperativeType
	draft.CurrentStep = 2
	draft.Status = "IN_PROGRESS"

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.OnboardingStepResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		CurrentStep:       2,
		NextStep:          "COOPERATIVE_PROFILE",
	}, nil
}

func (s *OfficerRegistrationService) UpdateCooperativeProfile(req model.UpdateCooperativeProfileRequest) (*model.OnboardingStepResponse, error) {
	if req.CooperativeName == "" || req.RegistrationNumber == "" || req.Address == "" || req.EstablishedYear == 0 {
		return nil, appErrors.BadRequest("profil koperasi belum lengkap")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	if draft.CurrentStep < 2 {
		return nil, appErrors.BadRequest("pilih jenis koperasi terlebih dahulu")
	}

	draft.CooperativeName = req.CooperativeName
	draft.RegistrationNumber = req.RegistrationNumber
	draft.Address = req.Address
	draft.EstablishedYear = req.EstablishedYear
	draft.CurrentStep = 3
	draft.Status = "IN_PROGRESS"

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.OnboardingStepResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		CurrentStep:       3,
		NextStep:          "FINANCIAL_CONFIGURATION",
	}, nil
}

func (s *OfficerRegistrationService) UpdateFinancialConfiguration(req model.UpdateFinancialConfigurationRequest) (*model.OnboardingStepResponse, error) {
	if req.MaxLoanAmountPerMember <= 0 {
		return nil, appErrors.BadRequest("batas pinjaman wajib lebih dari 0")
	}
	if req.LoanInterestRateBpsPerMonth <= 0 {
		return nil, appErrors.BadRequest("bunga pinjaman wajib lebih dari 0")
	}
	if req.LateFeeRateBpsPerDay < 0 {
		return nil, appErrors.BadRequest("denda tidak boleh negatif")
	}
	if req.MaxLoanTermMonths <= 0 {
		return nil, appErrors.BadRequest("masa pinjaman wajib lebih dari 0")
	}
	if req.MandatorySavingsPerMonth <= 0 {
		return nil, appErrors.BadRequest("simpanan wajib wajib lebih dari 0")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	if draft.CurrentStep < 3 {
		return nil, appErrors.BadRequest("lengkapi profil koperasi terlebih dahulu")
	}

	draft.MaxLoanAmountPerMember = req.MaxLoanAmountPerMember
	draft.LoanInterestRateBpsPerMonth = req.LoanInterestRateBpsPerMonth
	draft.LateFeeRateBpsPerDay = req.LateFeeRateBpsPerDay
	draft.MaxLoanTermMonths = req.MaxLoanTermMonths
	draft.MandatorySavingsPerMonth = req.MandatorySavingsPerMonth
	draft.CurrentStep = 4
	draft.Status = "IN_PROGRESS"

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan konfigurasi keuangan")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan konfigurasi keuangan")
	}

	return &model.OnboardingStepResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		CurrentStep:       4,
		NextStep:          "COOPERATIVE_BANK_ACCOUNT",
	}, nil
}

func (s *OfficerRegistrationService) UpdateCooperativeBankAccount(req model.UpdateCooperativeBankAccountRequest) (*model.OnboardingStepResponse, error) {
	if req.BankName == "" || req.BankAccountNumber == "" || req.BankAccountHolderName == "" {
		return nil, appErrors.BadRequest("rekening koperasi belum lengkap")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	if draft.CurrentStep < 4 {
		return nil, appErrors.BadRequest("lengkapi konfigurasi keuangan terlebih dahulu")
	}

	draft.BankName = req.BankName
	draft.BankAccountNumber = req.BankAccountNumber
	draft.BankAccountHolderName = req.BankAccountHolderName
	draft.CurrentStep = 5
	draft.Status = "IN_PROGRESS"

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan rekening koperasi")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan rekening koperasi")
	}

	return &model.OnboardingStepResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		CurrentStep:       5,
		NextStep:          "CONFIRMATION",
	}, nil
}

func (s *OfficerRegistrationService) ActivateOnboardingDraft(req model.ActivateOnboardingDraftRequest) (*model.ActivateOnboardingDraftResponse, error) {
	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	if draft.CurrentStep < 5 {
		return nil, appErrors.BadRequest("lengkapi semua data onboarding terlebih dahulu")
	}

	if draft.Status == "ACTIVATED" {
		return nil, appErrors.Conflict("draft onboarding sudah diaktivasi")
	}

	err = s.validateDraftReadyForActivation(tx, draft)
	if err != nil {
		return nil, err
	}

	role, err := s.deps.repository.RoleRepository.GetRole(tx, model.GetRoleParam{
		Code:      draft.RoleCode,
		ScopeType: constants.RoleScopeCooperative,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("jabatan koperasi tidak valid")
		}
		return nil, err
	}

	userID := uuid.New()
	user := &entity.User{
		UserID:      userID,
		FullName:    draft.FullName,
		PhoneNumber: draft.PhoneNumber,
		Status:      "ACTIVE",
		UserType:    "COOPERATIVE",
	}

	err = s.deps.repository.UserRepository.CreateUser(tx, user)
	if err != nil {
		return nil, err
	}

	userIdentity := &entity.UserIdentity{
		IdentityID:   uuid.New(),
		UserID:       userID,
		NIKEncrypted: draft.NIKEncrypted,
		NIKHash:      draft.NIKHash,
		KTPImageURL:  draft.KTPImageURL,
	}

	err = s.deps.repository.UserIdentityRepository.CreateUserIdentity(tx, userIdentity)
	if err != nil {
		return nil, err
	}

	cooperativeID := uuid.New()
	cooperativeCode, err := s.generateUniqueCooperativeCode(tx, draft.CooperativeName)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat kode koperasi")
	}

	cooperative := &entity.Cooperative{
		CooperativeID:         cooperativeID,
		Name:                  draft.CooperativeName,
		CooperativeType:       draft.CooperativeType,
		RegistrationNumber:    draft.RegistrationNumber,
		CooperativeCode:       cooperativeCode,
		EstablishedYear:       draft.EstablishedYear,
		Status:                "ACTIVE",
		Address:               draft.Address,
		BankName:              draft.BankName,
		BankAccountNumber:     draft.BankAccountNumber,
		BankAccountHolderName: draft.BankAccountHolderName,
	}

	err = s.deps.repository.CooperativeRepository.CreateCooperative(tx, cooperative)
	if err != nil {
		return nil, err
	}

	financialConfiguration := &entity.FinancialConfiguration{
		FinancialConfigurationID:    uuid.New(),
		CooperativeID:               cooperativeID,
		MaxLoanAmountPerMember:      draft.MaxLoanAmountPerMember,
		LoanInterestRateBpsPerMonth: draft.LoanInterestRateBpsPerMonth,
		LateFeeRateBpsPerDay:        draft.LateFeeRateBpsPerDay,
		MaxLoanTermMonths:           draft.MaxLoanTermMonths,
		MandatorySavingsPerMonth:    draft.MandatorySavingsPerMonth,
	}

	err = s.deps.repository.FinancialConfigurationRepository.CreateFinancialConfiguration(tx, financialConfiguration)
	if err != nil {
		return nil, err
	}

	membershipID := uuid.New()
	membership := &entity.UserCooperativeMembership{
		CooperativeMembershipID: membershipID,
		UserID:                  userID,
		CooperativeID:           cooperativeID,
		RoleID:                  role.RoleID,
		PositionCode:            draft.PositionCode,
		Status:                  "ACTIVE",
		JoinedAt:                time.Now(),
	}

	err = s.deps.repository.UserCooperativeMembershipRepository.CreateUserCooperativeMembership(tx, membership)
	if err != nil {
		return nil, err
	}

	userPin := &entity.UserPINCredential{
		PinCredentialID: uuid.New(),
		UserID:          userID,
		PINHash:         draft.PINHash,
		FailedAttempts:  0,
		LastChangedAt:   time.Now(),
	}

	err = s.deps.repository.UserPinRepository.CreateUserPin(tx, userPin)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	draft.Status = "ACTIVATED"
	draft.CurrentStep = 6
	draft.ActivatedAt = &now
	draft.ActivatedUserID = &userID
	draft.ActivatedCooperativeID = &cooperativeID

	err = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.ActivateOnboardingDraftResponse{
		UserID:        userID,
		CooperativeID: cooperativeID,
		MembershipID:  membershipID,
		NextStep:      "COOPERATIVE_READY",
	}, nil
}

func (s *OfficerRegistrationService) GetOnboardingState(req model.GetOnboardingStateRequest) (*model.GetOnboardingStateResponse, error) {
	tx := s.deps.db.Begin()
	defer tx.Rollback()

	draft, err := s.getVerifiedDraft(tx, req.OnboardingDraftID, req.OnboardingToken)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.GetOnboardingStateResponse{
		OnboardingDraftID: draft.OnboardingDraftID,
		PhoneNumber:       draft.PhoneNumber,
		Status:            draft.Status,
		CurrentStep:       draft.CurrentStep,
		NextStep:          resolveOnboardingNextStep(draft),
		CompletedSteps:    resolveOnboardingCompletedSteps(draft),
		DraftData:         buildOnboardingDraftData(draft),
		Summary:           buildOnboardingSummary(draft),
	}, nil
}

func buildOnboardingSummary(draft *entity.OnboardingDraft) model.OnboardingSummary {
	return model.OnboardingSummary{
		PersonalData: model.OnboardingPersonalSummary{
			FullName:      draft.FullName,
			NIKMasked:     draft.NIKMasked,
			PhoneNumber:   formatIndonesianPhoneNumber(draft.PhoneNumber),
			PositionLabel: resolvePositionLabel(draft.PositionCode),
		},
		Cooperative: model.OnboardingCooperativeSummary{
			CooperativeType:   draft.CooperativeType,
			CooperativeName:   draft.CooperativeName,
			LoanConfiguration: formatLoanConfiguration(draft.MaxLoanAmountPerMember, draft.LoanInterestRateBpsPerMonth),
			BankAccount:       formatBankAccountSummary(draft.BankName, draft.BankAccountNumber),
		},
	}
}

func buildOnboardingDraftData(draft *entity.OnboardingDraft) model.OnboardingDraftData {
	return model.OnboardingDraftData{
		PersonalData: model.OnboardingPersonalDataState{
			FullName:                draft.FullName,
			NIKHash:                 draft.NIKHash,
			NIKMasked:               draft.NIKMasked,
			KTPImageURL:             draft.KTPImageURL,
			PositionCode:            draft.PositionCode,
			ExistingCooperativeCode: draft.ExistingCooperativeCode,
		},
		CooperativeType: model.OnboardingCooperativeTypeState{
			CooperativeType: draft.CooperativeType,
		},
		CooperativeProfile: model.OnboardingCooperativeProfileState{
			CooperativeName:    draft.CooperativeName,
			RegistrationNumber: draft.RegistrationNumber,
			Address:            draft.Address,
			EstablishedYear:    draft.EstablishedYear,
		},
		FinancialConfiguration: model.OnboardingFinancialConfigurationState{
			MaxLoanAmountPerMember:      draft.MaxLoanAmountPerMember,
			LoanInterestRateBpsPerMonth: draft.LoanInterestRateBpsPerMonth,
			LateFeeRateBpsPerDay:        draft.LateFeeRateBpsPerDay,
			MaxLoanTermMonths:           draft.MaxLoanTermMonths,
			MandatorySavingsPerMonth:    draft.MandatorySavingsPerMonth,
		},
		BankAccount: model.OnboardingBankAccountState{
			BankName:              draft.BankName,
			BankAccountNumber:     draft.BankAccountNumber,
			BankAccountHolderName: draft.BankAccountHolderName,
		},
	}
}

func resolveOnboardingNextStep(draft *entity.OnboardingDraft) string {
	if draft.Status == "ACTIVATED" {
		return "COOPERATIVE_READY"
	}

	if draft.FullName == "" || draft.NIKEncrypted == "" || draft.NIKHash == "" || draft.KTPImageURL == "" || draft.PositionCode == "" {
		return "PERSONAL_DATA"
	}

	if draft.CooperativeType == "" {
		return "COOPERATIVE_TYPE"
	}

	if draft.CooperativeName == "" || draft.RegistrationNumber == "" || draft.Address == "" || draft.EstablishedYear == 0 {
		return "COOPERATIVE_PROFILE"
	}

	if draft.MaxLoanAmountPerMember <= 0 ||
		draft.LoanInterestRateBpsPerMonth <= 0 ||
		draft.MaxLoanTermMonths <= 0 ||
		draft.MandatorySavingsPerMonth <= 0 {
		return "FINANCIAL_CONFIGURATION"
	}

	if draft.BankName == "" || draft.BankAccountNumber == "" || draft.BankAccountHolderName == "" {
		return "COOPERATIVE_BANK_ACCOUNT"
	}

	return "CONFIRMATION"
}

func resolveOnboardingCompletedSteps(draft *entity.OnboardingDraft) []string {
	completedSteps := []string{}

	if draft.PhoneVerifiedAt != nil {
		completedSteps = append(completedSteps, "PHONE_VERIFICATION")
	}

	if draft.PINSetAt != nil {
		completedSteps = append(completedSteps, "PIN")
	}

	if draft.FullName != "" && draft.NIKEncrypted != "" && draft.NIKHash != "" && draft.KTPImageURL != "" && draft.PositionCode != "" {
		completedSteps = append(completedSteps, "PERSONAL_DATA")
	}

	if draft.CooperativeType != "" {
		completedSteps = append(completedSteps, "COOPERATIVE_TYPE")
	}

	if draft.CooperativeName != "" && draft.RegistrationNumber != "" && draft.Address != "" && draft.EstablishedYear != 0 {
		completedSteps = append(completedSteps, "COOPERATIVE_PROFILE")
	}

	if draft.MaxLoanAmountPerMember > 0 &&
		draft.LoanInterestRateBpsPerMonth > 0 &&
		draft.MaxLoanTermMonths > 0 &&
		draft.MandatorySavingsPerMonth > 0 {
		completedSteps = append(completedSteps, "FINANCIAL_CONFIGURATION")
	}

	if draft.BankName != "" && draft.BankAccountNumber != "" && draft.BankAccountHolderName != "" {
		completedSteps = append(completedSteps, "COOPERATIVE_BANK_ACCOUNT")
	}

	if draft.Status == "ACTIVATED" {
		completedSteps = append(completedSteps, "ACTIVATION")
	}

	return completedSteps
}

func resolvePositionLabel(positionCode string) string {
	switch positionCode {
	case constants.PositionCodeChairman:
		return "Ketua"
	case constants.PositionCodeTreasurer:
		return "Bendahara"
	case constants.PositionCodeSecretary:
		return "Sekretaris"
	case constants.PositionCodeStaff:
		return "Staf"
	default:
		return positionCode
	}
}

func formatIndonesianPhoneNumber(phoneNumber string) string {
	if strings.HasPrefix(phoneNumber, "62") {
		return "0" + strings.TrimPrefix(phoneNumber, "62")
	}

	return phoneNumber
}

func formatLoanConfiguration(maxLoanAmount int64, interestRateBps int) string {
	if maxLoanAmount <= 0 && interestRateBps <= 0 {
		return ""
	}

	return fmt.Sprintf("%s · %s/bln", formatRupiah(maxLoanAmount), formatBpsPercent(interestRateBps))
}

func formatRupiah(amount int64) string {
	if amount <= 0 {
		return "Rp 0"
	}

	raw := fmt.Sprintf("%d", amount)
	result := ""

	for i, digit := range reverseString(raw) {
		if i > 0 && i%3 == 0 {
			result = "." + result
		}
		result = string(digit) + result
	}

	return "Rp " + result
}

func formatBpsPercent(bps int) string {
	if bps <= 0 {
		return "0%"
	}

	integerPart := bps / 100
	decimalPart := bps % 100

	if decimalPart == 0 {
		return fmt.Sprintf("%d%%", integerPart)
	}

	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%d,%02d", integerPart, decimalPart), "0"), ",") + "%"
}

func formatBankAccountSummary(bankName string, accountNumber string) string {
	if bankName == "" && accountNumber == "" {
		return ""
	}

	if accountNumber == "" {
		return bankName
	}

	lastDigits := accountNumber
	if len(accountNumber) > 4 {
		lastDigits = accountNumber[len(accountNumber)-4:]
	}

	if bankName == "" {
		return "xxxx-" + lastDigits
	}

	return bankName + " · xxxx-" + lastDigits
}

func reverseString(value string) string {
	runes := []rune(value)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}

	return string(runes)
}

func (s *OfficerRegistrationService) validateDraftReadyForActivation(tx *gorm.DB, draft *entity.OnboardingDraft) error {
	if draft.PhoneNumber == "" || draft.PINHash == "" {
		return appErrors.BadRequest("akun belum lengkap")
	}

	if draft.KTPImageURL == "" || draft.FullName == "" || draft.NIKEncrypted == "" || draft.NIKHash == "" || draft.RoleCode == "" || draft.PositionCode == "" {
		return appErrors.BadRequest("data diri belum lengkap")
	}

	if draft.CooperativeType == "" || draft.CooperativeName == "" || draft.RegistrationNumber == "" || draft.Address == "" || draft.EstablishedYear == 0 {
		return appErrors.BadRequest("data koperasi belum lengkap")
	}

	if draft.MaxLoanAmountPerMember <= 0 || draft.LoanInterestRateBpsPerMonth <= 0 || draft.MaxLoanTermMonths <= 0 || draft.MandatorySavingsPerMonth <= 0 {
		return appErrors.BadRequest("konfigurasi keuangan belum lengkap")
	}

	if draft.BankName == "" || draft.BankAccountNumber == "" || draft.BankAccountHolderName == "" {
		return appErrors.BadRequest("rekening koperasi belum lengkap")
	}

	_, err := s.deps.repository.UserRepository.GetUser(tx, model.GetUserParam{
		PhoneNumber: draft.PhoneNumber,
	})
	if err == nil {
		return appErrors.Conflict("nomor handphone sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	_, err = s.deps.repository.UserIdentityRepository.GetUserIdentity(tx, model.GetUserIdentityParam{
		NIKHash: draft.NIKHash,
	})
	if err == nil {
		return appErrors.Conflict("NIK sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	_, err = s.deps.repository.CooperativeRepository.GetCooperative(tx, model.GetCooperativeParam{
		RegistrationNumber: draft.RegistrationNumber,
	})
	if err == nil {
		return appErrors.Conflict("nomor badan hukum sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *OfficerRegistrationService) generateUniqueCooperativeCode(tx *gorm.DB, cooperativeName string) (string, error) {
	baseCode := buildCooperativeCodeBase(cooperativeName)
	for attempt := 1; attempt <= 99; attempt++ {
		candidate := baseCode
		if attempt > 1 {
			suffix := fmt.Sprintf("-%02d", attempt)
			candidate = truncateCooperativeCode(baseCode, len(suffix)) + suffix
		}

		_, err := s.deps.repository.CooperativeRepository.GetCooperative(tx, model.GetCooperativeParam{
			CooperativeCode: candidate,
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}

	return truncateCooperativeCode(baseCode, 9) + "-" + uuid.NewString()[:8], nil
}

func buildCooperativeCodeBase(cooperativeName string) string {
	normalizedName := strings.ToUpper(strings.TrimSpace(cooperativeName))
	var builder strings.Builder
	previousDash := false

	for _, char := range normalizedName {
		isAlphaNumeric := (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
		if isAlphaNumeric {
			builder.WriteRune(char)
			previousDash = false
			continue
		}

		if builder.Len() > 0 && !previousDash {
			builder.WriteRune('-')
			previousDash = true
		}
	}

	code := strings.Trim(builder.String(), "-")
	if code == "" {
		code = "KOPERASI"
	}

	return truncateCooperativeCode(code, 0)
}

func truncateCooperativeCode(code string, reservedLength int) string {
	maxLength := 50 - reservedLength
	if maxLength < 1 {
		maxLength = 1
	}

	if len(code) <= maxLength {
		return strings.Trim(code, "-")
	}

	return strings.Trim(code[:maxLength], "-")
}

func generateNumericOTP(length int) (string, error) {
	const digits = "0123456789"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = digits[int(bytes[i])%len(digits)]
	}

	return string(bytes), nil
}

func generateSecureToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func normalizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if strings.HasPrefix(phone, "0") {
		return "62" + strings.TrimPrefix(phone, "0")
	}

	if strings.HasPrefix(phone, "+") {
		return strings.TrimPrefix(phone, "+")
	}

	return phone
}

func (s *OfficerRegistrationService) getVerifiedDraft(tx *gorm.DB, draftID uuid.UUID, token string) (*entity.OnboardingDraft, error) {
	if draftID == uuid.Nil {
		return nil, appErrors.BadRequest("onboarding_draft_id wajib diisi")
	}

	if token == "" {
		return nil, appErrors.Unauthorized("onboarding token wajib diisi")
	}

	draft, err := s.deps.repository.OnboardingDraftRepository.GetOnboardingDraft(tx, model.GetOnboardingDraftParam{
		OnboardingDraftID: draftID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("draft onboarding tidak ditemukan")
		}
		return nil, err
	}

	if time.Now().After(draft.ExpiresAt) {
		draft.Status = "EXPIRED"
		_ = s.deps.repository.OnboardingDraftRepository.UpdateOnboardingDraft(tx, draft)
		return nil, appErrors.BadRequest("draft onboarding sudah kedaluwarsa")
	}

	if draft.Status != "PIN_SET" && draft.Status != "IN_PROGRESS" {
		return nil, appErrors.Forbidden("draft onboarding belum siap dilanjutkan")
	}

	if draft.SessionTokenHash == "" {
		return nil, appErrors.Forbidden("draft onboarding belum memiliki session token")
	}

	if err := s.deps.bcrypt.CompareAndHashPassword(draft.SessionTokenHash, token); err != nil {
		return nil, appErrors.Unauthorized("onboarding token tidak valid")
	}

	return draft, nil
}

func isValidCooperativeType(value string) bool {
	switch value {
	case "KSP", "PANGAN_BULKY", "COLD_CHAIN", "TOKO_GERAI", "UTILITY", "PETERNAKAN":
		return true
	default:
		return false
	}
}

func isValidCooperativePosition(value string) bool {
	switch value {
	case constants.PositionCodeChairman,
		constants.PositionCodeTreasurer,
		constants.PositionCodeSecretary,
		constants.PositionCodeStaff:
		return true
	default:
		return false
	}
}

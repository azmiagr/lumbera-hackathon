package service

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/azmiagr/lumbera-hackathon/entity"
	"github.com/azmiagr/lumbera-hackathon/model"
	constants "github.com/azmiagr/lumbera-hackathon/pkg/constant"
	appErrors "github.com/azmiagr/lumbera-hackathon/pkg/errors"
	"github.com/azmiagr/lumbera-hackathon/pkg/identity"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type IMemberService interface {
	ListMembers(req model.ListMembersRequest) (*model.ListMembersResponse, error)
	CreateMember(req model.CreateMemberRequest) (*model.CreateMemberResponse, error)
	UploadMemberImport(req model.UploadMemberImportRequest) (*model.UploadMemberImportResponse, error)
	GetMemberImport(req model.GetMemberImportRequest) (*model.GetMemberImportResponse, error)
	UpdateMemberImportRow(req model.UpdateMemberImportRowRequest) (*model.MemberImportRowResponse, error)
	DeleteMemberImportRow(req model.DeleteMemberImportRowRequest) error
	SubmitMemberImport(req model.SubmitMemberImportRequest) (*model.SubmitMemberImportResponse, error)
	DownloadMemberImportTemplate() ([]byte, string, error)
}

type MemberService struct {
	deps serviceDependency
}

var scientificNumberPattern = regexp.MustCompile(`(?i)^[0-9]+(\.[0-9]+)?e\+[0-9]+$`)

func NewMemberService(deps serviceDependency) IMemberService {
	return &MemberService{deps: deps}
}

func (s *MemberService) ListMembers(req model.ListMembersRequest) (*model.ListMembersResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat melihat daftar anggota")
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	grade := strings.ToUpper(strings.TrimSpace(req.Grade))
	if grade != "" && grade != "SEMUA" && !isValidMCSGrade(grade) {
		return nil, appErrors.BadRequest("grade anggota tidak valid")
	}
	req.Grade = grade

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "" && status != "ACTIVE" && status != "INACTIVE" && status != "SUSPENDED" && status != "RESIGNED" {
		return nil, appErrors.BadRequest("status anggota tidak valid")
	}
	req.Status = status

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	items, total, err := s.deps.repository.MemberRepository.ListMembers(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil daftar anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil daftar anggota")
	}

	for i := range items {
		items[i].Initials = buildMemberInitials(items[i].FullName)
		items[i].MembershipYears = calculateMembershipYears(items[i].JoinedDate)
	}

	return &model.ListMembersResponse{
		Items: items,
		Page:  req.Page,
		Limit: req.Limit,
		Total: total,
	}, nil
}

func (s *MemberService) CreateMember(req model.CreateMemberRequest) (*model.CreateMemberResponse, error) {
	if req.UserID == uuid.Nil || req.CooperativeID == uuid.Nil {
		return nil, appErrors.Unauthorized("akses tidak valid")
	}

	if req.RoleCode != constants.RoleCodePengurusKoperasi {
		return nil, appErrors.Forbidden("hanya pengurus koperasi yang dapat mendaftarkan anggota")
	}

	fullName := strings.TrimSpace(req.FullName)
	phoneNumber := normalizePhoneNumber(req.PhoneNumber)
	nik := identity.NormalizeNIK(req.NIK)
	address := strings.TrimSpace(req.Address)

	if fullName == "" || phoneNumber == "" || nik == "" || address == "" || req.JoinedDate == nil {
		return nil, appErrors.BadRequest("data anggota belum lengkap")
	}

	if !isSixteenDigitNIK(nik) {
		return nil, appErrors.BadRequest("NIK harus 16 digit")
	}

	nikHash, err := identity.HashNIK(nik)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat hash NIK")
	}

	nikEncrypted, err := identity.EncryptNIK(nik)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengenkripsi NIK")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	_, err = s.deps.repository.UserRepository.GetUser(tx, model.GetUserParam{
		PhoneNumber: phoneNumber,
	})
	if err == nil {
		return nil, appErrors.Conflict("nomor handphone sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi nomor handphone")
	}

	_, err = s.deps.repository.UserIdentityRepository.GetUserIdentity(tx, model.GetUserIdentityParam{
		NIKHash: nikHash,
	})
	if err == nil {
		return nil, appErrors.Conflict("NIK sudah terdaftar")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("gagal memvalidasi NIK")
	}

	role, err := s.deps.repository.RoleRepository.GetRole(tx, model.GetRoleParam{
		Code:      constants.RoleCodeAnggota,
		ScopeType: constants.RoleScopeCooperative,
	})
	if err != nil {
		return nil, appErrors.InternalServer("role anggota belum tersedia")
	}

	memberNumber, err := s.generateMemberNumber(tx, req.CooperativeID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal membuat nomor anggota")
	}

	userID := uuid.New()
	user := &entity.User{
		UserID:      userID,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
		Status:      "PIN_REQUIRED",
		UserType:    "COOPERATIVE",
	}

	err = s.deps.repository.UserRepository.CreateUser(tx, user)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan user anggota")
	}

	userIdentity := &entity.UserIdentity{
		IdentityID:   uuid.New(),
		UserID:       userID,
		NIKEncrypted: nikEncrypted,
		NIKHash:      nikHash,
		Address:      address,
	}

	err = s.deps.repository.UserIdentityRepository.CreateUserIdentity(tx, userIdentity)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan identitas anggota")
	}

	memberID := uuid.New()
	member := &entity.Member{
		MemberID:      memberID,
		CooperativeID: req.CooperativeID,
		UserID:        userID,
		MemberNumber:  memberNumber,
		JoinedDate:    req.JoinedDate,
		MemberStatus:  "ACTIVE",
		MCSGrade:      "C",
	}

	err = s.deps.repository.MemberRepository.CreateMember(tx, member)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan anggota")
	}

	membership := &entity.UserCooperativeMembership{
		CooperativeMembershipID: uuid.New(),
		UserID:                  userID,
		CooperativeID:           req.CooperativeID,
		MemberID:                &memberID,
		RoleID:                  role.RoleID,
		PositionCode:            constants.PositionCodeStaff,
		Status:                  "ACTIVE",
		JoinedAt:                time.Now(),
	}

	err = s.deps.repository.UserCooperativeMembershipRepository.CreateUserCooperativeMembership(tx, membership)
	if err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan membership anggota")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("gagal mendaftarkan anggota")
	}

	return &model.CreateMemberResponse{
		UserID:        userID,
		MemberID:      memberID,
		CooperativeID: req.CooperativeID,
		FullName:      fullName,
		PhoneNumber:   phoneNumber,
		MemberNumber:  memberNumber,
		JoinedDate:    req.JoinedDate,
		MemberStatus:  "ACTIVE",
		AccountStatus: "PIN_REQUIRED",
	}, nil
}

func (s *MemberService) UploadMemberImport(req model.UploadMemberImportRequest) (*model.UploadMemberImportResponse, error) {
	if err := validateMemberImportAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.File == nil {
		return nil, appErrors.BadRequest("file excel wajib diupload")
	}
	if !isXLSXFile(req.File) {
		return nil, appErrors.BadRequest("file harus berformat .xlsx")
	}

	rows, err := parseMemberImportXLSX(req.File)
	if err != nil {
		return nil, appErrors.BadRequest("gagal membaca file excel")
	}
	if len(rows) == 0 {
		return nil, appErrors.BadRequest("file tidak memiliki data anggota")
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	batch := &entity.MemberImportBatch{
		ImportBatchID: uuid.New(),
		CooperativeID: req.CooperativeID,
		UploadedBy:    req.UserID,
		FileName:      req.File.Filename,
		Status:        constants.MemberImportBatchStatusDraft,
	}

	if err := s.deps.repository.MemberImportRepository.CreateBatch(tx, batch); err != nil {
		return nil, appErrors.InternalServer("gagal membuat batch import")
	}

	importRows := make([]entity.MemberImportRow, 0, len(rows))
	for index, row := range rows {
		importRow := s.buildValidatedImportRow(tx, batch.ImportBatchID, index+2, row)
		importRows = append(importRows, importRow)
	}

	if err := s.deps.repository.MemberImportRepository.CreateRows(tx, importRows); err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan data import")
	}

	if err := s.refreshImportBatchSummary(tx, batch); err != nil {
		return nil, appErrors.InternalServer("gagal menghitung ringkasan import")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal menyimpan import anggota")
	}

	return &model.UploadMemberImportResponse{
		Summary: mapMemberImportSummary(batch),
		Rows:    mapMemberImportRows(importRows),
	}, nil
}

func (s *MemberService) GetMemberImport(req model.GetMemberImportRequest) (*model.GetMemberImportResponse, error) {
	if err := validateMemberImportAccess(req.AuthContext); err != nil {
		return nil, err
	}
	if req.ImportBatchID == uuid.Nil {
		return nil, appErrors.BadRequest("import_batch_id tidak valid")
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	batch, err := s.deps.repository.MemberImportRepository.GetBatch(tx, req.CooperativeID, req.ImportBatchID)
	if err != nil {
		return nil, appErrors.NotFound("batch import tidak ditemukan")
	}

	rows, total, err := s.deps.repository.MemberImportRepository.ListRows(tx, req)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil data import")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal mengambil data import")
	}

	return &model.GetMemberImportResponse{
		Summary: mapMemberImportSummary(batch),
		Rows:    mapMemberImportRows(rows),
		Page:    req.Page,
		Limit:   req.Limit,
		Total:   total,
	}, nil
}

func (s *MemberService) UpdateMemberImportRow(req model.UpdateMemberImportRowRequest) (*model.MemberImportRowResponse, error) {
	if err := validateMemberImportAccess(req.AuthContext); err != nil {
		return nil, err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	batch, err := s.deps.repository.MemberImportRepository.GetBatch(tx, req.CooperativeID, req.ImportBatchID)
	if err != nil {
		return nil, appErrors.NotFound("batch import tidak ditemukan")
	}
	if batch.Status != constants.MemberImportBatchStatusDraft {
		return nil, appErrors.BadRequest("batch import sudah tidak dapat diedit")
	}

	row, err := s.deps.repository.MemberImportRepository.GetRow(tx, req.ImportBatchID, req.ImportRowID)
	if err != nil {
		return nil, appErrors.NotFound("row import tidak ditemukan")
	}
	if row.Status == constants.MemberImportRowStatusDeleted || row.Status == constants.MemberImportRowStatusImported {
		return nil, appErrors.BadRequest("row import tidak dapat diedit")
	}

	parsed := parsedMemberImportRow{
		FullName:    req.FullName,
		NIK:         req.NIK,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		JoinedDate:  req.JoinedDate,
	}
	validated := s.buildValidatedImportRow(tx, req.ImportBatchID, row.RowNumber, parsed)

	row.FullName = validated.FullName
	row.NIKEncrypted = validated.NIKEncrypted
	row.NIKHash = validated.NIKHash
	row.NIKMasked = validated.NIKMasked
	row.PhoneNumber = validated.PhoneNumber
	row.Address = validated.Address
	row.JoinedDate = validated.JoinedDate
	row.Status = validated.Status
	row.ErrorMessage = validated.ErrorMessage

	if err := s.deps.repository.MemberImportRepository.UpdateRow(tx, row); err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui row import")
	}

	if err := s.refreshImportBatchSummary(tx, batch); err != nil {
		return nil, appErrors.InternalServer("gagal menghitung ringkasan import")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal memperbarui row import")
	}

	response := mapMemberImportRow(*row)
	return &response, nil
}

func (s *MemberService) DeleteMemberImportRow(req model.DeleteMemberImportRowRequest) error {
	if err := validateMemberImportAccess(req.AuthContext); err != nil {
		return err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	batch, err := s.deps.repository.MemberImportRepository.GetBatch(tx, req.CooperativeID, req.ImportBatchID)
	if err != nil {
		return appErrors.NotFound("batch import tidak ditemukan")
	}
	if batch.Status != constants.MemberImportBatchStatusDraft {
		return appErrors.BadRequest("batch import sudah tidak dapat diedit")
	}

	row, err := s.deps.repository.MemberImportRepository.GetRow(tx, req.ImportBatchID, req.ImportRowID)
	if err != nil {
		return appErrors.NotFound("row import tidak ditemukan")
	}

	row.Status = constants.MemberImportRowStatusDeleted
	if err := s.deps.repository.MemberImportRepository.UpdateRow(tx, row); err != nil {
		return appErrors.InternalServer("gagal menghapus row import")
	}

	if err := s.refreshImportBatchSummary(tx, batch); err != nil {
		return appErrors.InternalServer("gagal menghitung ringkasan import")
	}

	if err := tx.Commit().Error; err != nil {
		return appErrors.InternalServer("gagal menghapus row import")
	}

	return nil
}

func (s *MemberService) SubmitMemberImport(req model.SubmitMemberImportRequest) (*model.SubmitMemberImportResponse, error) {
	if err := validateMemberImportAccess(req.AuthContext); err != nil {
		return nil, err
	}

	tx := s.deps.db.Begin()
	defer tx.Rollback()

	batch, err := s.deps.repository.MemberImportRepository.GetBatch(tx, req.CooperativeID, req.ImportBatchID)
	if err != nil {
		return nil, appErrors.NotFound("batch import tidak ditemukan")
	}
	if batch.Status != constants.MemberImportBatchStatusDraft {
		return nil, appErrors.BadRequest("batch import sudah disubmit")
	}
	if batch.ErrorRows > 0 {
		return nil, appErrors.BadRequest("masih ada data error yang perlu diperbaiki atau dihapus")
	}

	rows, err := s.deps.repository.MemberImportRepository.ListRowsForSubmit(tx, req.ImportBatchID)
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil data valid")
	}
	if len(rows) == 0 {
		return nil, appErrors.BadRequest("tidak ada data valid untuk diimport")
	}

	role, err := s.deps.repository.RoleRepository.GetRole(tx, model.GetRoleParam{
		Code:      constants.RoleCodeAnggota,
		ScopeType: constants.RoleScopeCooperative,
	})
	if err != nil {
		return nil, appErrors.InternalServer("role anggota belum tersedia")
	}

	importedRows := 0
	for i := range rows {
		memberNumber, err := s.generateMemberNumber(tx, req.CooperativeID)
		if err != nil {
			return nil, appErrors.InternalServer("gagal membuat nomor anggota")
		}

		userID := uuid.New()
		user := &entity.User{
			UserID:      userID,
			FullName:    rows[i].FullName,
			PhoneNumber: rows[i].PhoneNumber,
			Status:      "PIN_REQUIRED",
			UserType:    "COOPERATIVE",
		}
		if err := s.deps.repository.UserRepository.CreateUser(tx, user); err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan user anggota")
		}

		userIdentity := &entity.UserIdentity{
			IdentityID:   uuid.New(),
			UserID:       userID,
			NIKEncrypted: rows[i].NIKEncrypted,
			NIKHash:      rows[i].NIKHash,
			Address:      rows[i].Address,
		}
		if err := s.deps.repository.UserIdentityRepository.CreateUserIdentity(tx, userIdentity); err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan identitas anggota")
		}

		memberID := uuid.New()
		member := &entity.Member{
			MemberID:      memberID,
			CooperativeID: req.CooperativeID,
			UserID:        userID,
			MemberNumber:  memberNumber,
			JoinedDate:    rows[i].JoinedDate,
			MemberStatus:  "ACTIVE",
			MCSGrade:      "C",
		}
		if err := s.deps.repository.MemberRepository.CreateMember(tx, member); err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan anggota")
		}

		membership := &entity.UserCooperativeMembership{
			CooperativeMembershipID: uuid.New(),
			UserID:                  userID,
			CooperativeID:           req.CooperativeID,
			MemberID:                &memberID,
			RoleID:                  role.RoleID,
			PositionCode:            constants.PositionCodeStaff,
			Status:                  "ACTIVE",
			JoinedAt:                time.Now(),
		}
		if err := s.deps.repository.UserCooperativeMembershipRepository.CreateUserCooperativeMembership(tx, membership); err != nil {
			return nil, appErrors.InternalServer("gagal menyimpan membership anggota")
		}

		rows[i].Status = constants.MemberImportRowStatusImported
		if err := s.deps.repository.MemberImportRepository.UpdateRow(tx, &rows[i]); err != nil {
			return nil, appErrors.InternalServer("gagal memperbarui status import")
		}
		importedRows++
	}

	now := time.Now()
	batch.Status = constants.MemberImportBatchStatusSubmitted
	batch.SubmittedAt = &now
	if err := s.deps.repository.MemberImportRepository.UpdateBatch(tx, batch); err != nil {
		return nil, appErrors.InternalServer("gagal menyelesaikan import")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("gagal submit import anggota")
	}

	return &model.SubmitMemberImportResponse{
		ImportBatchID: req.ImportBatchID,
		ImportedRows:  importedRows,
		SkippedRows:   0,
	}, nil
}

func (s *MemberService) DownloadMemberImportTemplate() ([]byte, string, error) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "Template"
	_ = file.SetSheetName("Sheet1", sheet)

	headers := []string{"NAMA LENGKAP", "NIK", "NO. HANDPHONE", "ALAMAT", "TANGGAL BERGABUNG"}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	textStyle, err := file.NewStyle(&excelize.Style{
		NumFmt: 49,
	})
	if err != nil {
		return nil, "", err
	}
	_ = file.SetColStyle(sheet, "B", textStyle)
	_ = file.SetColStyle(sheet, "C", textStyle)

	examples := [][]string{
		{"Bara Hermawan", "1234567890123456", "081234567890", "Jl. Bunga Kenari No.1, Kota Malang, Jawa Timur", "11 Juni 2024"},
	}
	for rowIndex, row := range examples {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			_ = file.SetCellValue(sheet, cell, value)
		}
	}

	_ = file.SetColWidth(sheet, "A", "A", 28)
	_ = file.SetColWidth(sheet, "B", "C", 20)
	_ = file.SetColWidth(sheet, "D", "D", 55)
	_ = file.SetColWidth(sheet, "E", "E", 22)

	buffer := bytes.NewBuffer(nil)
	if err := file.Write(buffer); err != nil {
		return nil, "", appErrors.InternalServer("gagal membuat template import")
	}

	return buffer.Bytes(), "template-import-anggota.xlsx", nil
}

func validateMemberImportAccess(auth model.AuthContext) error {
	if auth.UserID == uuid.Nil || auth.CooperativeID == uuid.Nil {
		return appErrors.Unauthorized("akses tidak valid")
	}
	if auth.RoleCode != constants.RoleCodePengurusKoperasi {
		return appErrors.Forbidden("hanya pengurus koperasi yang dapat import anggota")
	}
	return nil
}

func isXLSXFile(file *multipart.FileHeader) bool {
	return strings.EqualFold(filepath.Ext(file.Filename), ".xlsx")
}

func getExcelCell(row []string, index int) string {
	if len(row) <= index {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func isEmptyImportExcelRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseImportDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"2/1/2006",
		"02 Jan 2006",
		"2 Jan 2006",
		"02 January 2006",
		"2 January 2006",
		"02 Januari 2006",
		"2 Januari 2006",
	}

	normalized := normalizeIndonesianMonthName(value)
	for _, format := range formats {
		parsed, err := time.Parse(format, normalized)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

func normalizeIndonesianMonthName(value string) string {
	replacer := strings.NewReplacer(
		"Januari", "January",
		"Februari", "February",
		"Maret", "March",
		"Mei", "May",
		"Juni", "June",
		"Juli", "July",
		"Agustus", "August",
		"Oktober", "October",
		"Desember", "December",
	)
	return replacer.Replace(value)
}

func (s *MemberService) refreshImportBatchSummary(tx *gorm.DB, batch *entity.MemberImportBatch) error {
	totalRows, successRows, errorRows, err := s.deps.repository.MemberImportRepository.RecalculateBatchSummary(tx, batch.ImportBatchID)
	if err != nil {
		return err
	}

	batch.TotalRows = totalRows
	batch.SuccessRows = successRows
	batch.ErrorRows = errorRows
	return s.deps.repository.MemberImportRepository.UpdateBatch(tx, batch)
}

func mapMemberImportSummary(batch *entity.MemberImportBatch) model.MemberImportSummary {
	return model.MemberImportSummary{
		ImportBatchID: batch.ImportBatchID,
		FileName:      batch.FileName,
		Status:        batch.Status,
		TotalRows:     batch.TotalRows,
		SuccessRows:   batch.SuccessRows,
		ErrorRows:     batch.ErrorRows,
	}
}

func mapMemberImportRows(rows []entity.MemberImportRow) []model.MemberImportRowResponse {
	result := make([]model.MemberImportRowResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapMemberImportRow(row))
	}
	return result
}

func mapMemberImportRow(row entity.MemberImportRow) model.MemberImportRowResponse {
	return model.MemberImportRowResponse{
		ImportRowID:  row.ImportRowID,
		RowNumber:    row.RowNumber,
		FullName:     row.FullName,
		NIKMasked:    row.NIKMasked,
		PhoneNumber:  row.PhoneNumber,
		Address:      row.Address,
		JoinedDate:   row.JoinedDate,
		Status:       row.Status,
		ErrorMessage: row.ErrorMessage,
	}
}

type parsedMemberImportRow struct {
	FullName    string
	NIK         string
	PhoneNumber string
	Address     string
	JoinedDate  *time.Time
}

func parseMemberImportXLSX(file *multipart.FileHeader) ([]parsedMemberImportRow, error) {
	openedFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer openedFile.Close()

	workbook, err := excelize.OpenReader(openedFile)
	if err != nil {
		return nil, err
	}
	defer workbook.Close()

	sheetName := workbook.GetSheetName(0)
	rawRows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	rows := make([]parsedMemberImportRow, 0)
	for index, rawRow := range rawRows {
		if index == 0 {
			continue
		}
		if isEmptyImportExcelRow(rawRow) {
			continue
		}

		rows = append(rows, parsedMemberImportRow{
			FullName:    getExcelCell(rawRow, 0),
			NIK:         getExcelCell(rawRow, 1),
			PhoneNumber: getExcelCell(rawRow, 2),
			Address:     getExcelCell(rawRow, 3),
			JoinedDate:  parseImportDate(getExcelCell(rawRow, 4)),
		})
	}

	return rows, nil
}

func (s *MemberService) buildValidatedImportRow(tx *gorm.DB, batchID uuid.UUID, rowNumber int, row parsedMemberImportRow) entity.MemberImportRow {
	fullName := strings.TrimSpace(row.FullName)
	nik := identity.NormalizeNIK(row.NIK)
	phoneNumber := normalizeImportedPhoneNumber(row.PhoneNumber)
	address := strings.TrimSpace(row.Address)

	importRow := entity.MemberImportRow{
		ImportRowID:   uuid.New(),
		ImportBatchID: batchID,
		RowNumber:     rowNumber,
		FullName:      fullName,
		PhoneNumber:   phoneNumber,
		Address:       address,
		JoinedDate:    row.JoinedDate,
		Status:        constants.MemberImportRowStatusValid,
	}

	errors := make([]string, 0)

	if fullName == "" {
		errors = append(errors, "nama lengkap wajib diisi")
	}
	if phoneNumber == "" {
		errors = append(errors, "nomor handphone wajib diisi")
	}
	if nik == "" {
		errors = append(errors, "NIK wajib diisi")
	} else if isScientificNumber(nik) {
		errors = append(errors, "NIK terbaca sebagai angka Excel/scientific notation, format kolom NIK sebagai Text")
	} else if !isSixteenDigitNIK(nik) {
		errors = append(errors, "NIK harus 16 digit")
	}
	if address == "" {
		errors = append(errors, "alamat wajib diisi")
	}
	if row.JoinedDate == nil {
		errors = append(errors, "tanggal bergabung wajib diisi")
	}

	if nik != "" && isSixteenDigitNIK(nik) {
		nikHash, err := identity.HashNIK(nik)
		if err != nil {
			errors = append(errors, "gagal membuat hash NIK")
		} else {
			importRow.NIKHash = nikHash
			importRow.NIKMasked = maskNIK(nik)

			if _, err := s.deps.repository.UserIdentityRepository.GetUserIdentity(tx, model.GetUserIdentityParam{NIKHash: nikHash}); err == nil {
				errors = append(errors, "NIK sudah terdaftar")
			} else if !errorsIsRecordNotFound(err) {
				errors = append(errors, "gagal memvalidasi NIK")
			}
		}

		nikEncrypted, err := identity.EncryptNIK(nik)
		if err != nil {
			errors = append(errors, "gagal mengenkripsi NIK")
		} else {
			importRow.NIKEncrypted = nikEncrypted
		}
	}

	if phoneNumber != "" {
		if _, err := s.deps.repository.UserRepository.GetUser(tx, model.GetUserParam{PhoneNumber: phoneNumber}); err == nil {
			errors = append(errors, "nomor handphone sudah terdaftar")
		} else if !errorsIsRecordNotFound(err) {
			errors = append(errors, "gagal memvalidasi nomor handphone")
		}
	}

	if len(errors) > 0 {
		importRow.Status = constants.MemberImportRowStatusError
		importRow.ErrorMessage = strings.Join(errors, "; ")
	}

	return importRow
}

func errorsIsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func normalizeImportedPhoneNumber(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "-", "")

	if scientificNumberPattern.MatchString(value) {
		parsed, ok := new(big.Float).SetString(value)
		if ok {
			integerValue, _ := parsed.Int(nil)
			value = integerValue.String()
		}
	}

	return normalizePhoneNumber(value)
}

func isScientificNumber(value string) bool {
	return scientificNumberPattern.MatchString(strings.TrimSpace(value))
}

func maskNIK(nik string) string {
	if len(nik) <= 4 {
		return nik
	}
	return strings.Repeat("*", len(nik)-4) + nik[len(nik)-4:]
}

func (s *MemberService) generateMemberNumber(tx *gorm.DB, cooperativeID uuid.UUID) (string, error) {
	total, err := s.deps.repository.MemberRepository.CountMembersByCooperative(tx, cooperativeID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%04d", total+1), nil
}

func isSixteenDigitNIK(nik string) bool {
	if len(nik) != 16 {
		return false
	}

	for _, char := range nik {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func isValidMCSGrade(grade string) bool {
	switch grade {
	case "AA", "A", "B", "C", "D":
		return true
	default:
		return false
	}
}

func buildMemberInitials(fullName string) string {
	words := strings.Fields(fullName)
	if len(words) == 0 {
		return ""
	}

	if len(words) == 1 {
		return strings.ToUpper(firstLetter(words[0]))
	}

	return strings.ToUpper(firstLetter(words[0]) + firstLetter(words[len(words)-1]))
}

func firstLetter(value string) string {
	for _, char := range value {
		return string(char)
	}

	return ""
}

func calculateMembershipYears(joinedDate *time.Time) int {
	if joinedDate == nil {
		return 0
	}

	now := time.Now()
	years := now.Year() - joinedDate.Year()
	if now.YearDay() < joinedDate.YearDay() {
		years--
	}

	if years < 0 {
		return 0
	}

	return years
}

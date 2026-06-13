package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type UploadMemberImportRequest struct {
	AuthContext
	File *multipart.FileHeader `form:"file"`
}

type MemberImportSummary struct {
	ImportBatchID uuid.UUID `json:"import_batch_id"`
	FileName      string    `json:"file_name"`
	Status        string    `json:"status"`
	TotalRows     int       `json:"total_rows"`
	SuccessRows   int       `json:"success_rows"`
	ErrorRows     int       `json:"error_rows"`
}

type MemberImportRowResponse struct {
	ImportRowID  uuid.UUID  `json:"import_row_id"`
	RowNumber    int        `json:"row_number"`
	FullName     string     `json:"full_name"`
	NIKMasked    string     `json:"nik_masked"`
	PhoneNumber  string     `json:"phone_number"`
	Address      string     `json:"address"`
	JoinedDate   *time.Time `json:"joined_date"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message"`
}

type UploadMemberImportResponse struct {
	Summary MemberImportSummary       `json:"summary"`
	Rows    []MemberImportRowResponse `json:"rows"`
}

type GetMemberImportRequest struct {
	AuthContext
	ImportBatchID uuid.UUID `json:"-"`
	Status        string    `form:"status"`
	Page          int       `form:"page"`
	Limit         int       `form:"limit"`
}

type GetMemberImportResponse struct {
	Summary MemberImportSummary       `json:"summary"`
	Rows    []MemberImportRowResponse `json:"rows"`
	Page    int                       `json:"page"`
	Limit   int                       `json:"limit"`
	Total   int64                     `json:"total"`
}

type UpdateMemberImportRowRequest struct {
	AuthContext
	ImportBatchID uuid.UUID  `json:"-"`
	ImportRowID   uuid.UUID  `json:"-"`
	FullName      string     `json:"full_name"`
	NIK           string     `json:"nik"`
	PhoneNumber   string     `json:"phone_number"`
	Address       string     `json:"address"`
	JoinedDate    *time.Time `json:"joined_date"`
}

type DeleteMemberImportRowRequest struct {
	AuthContext
	ImportBatchID uuid.UUID `json:"-"`
	ImportRowID   uuid.UUID `json:"-"`
}

type SubmitMemberImportRequest struct {
	AuthContext
	ImportBatchID uuid.UUID `json:"-"`
}

type SubmitMemberImportResponse struct {
	ImportBatchID uuid.UUID `json:"import_batch_id"`
	ImportedRows  int       `json:"imported_rows"`
	SkippedRows   int       `json:"skipped_rows"`
}

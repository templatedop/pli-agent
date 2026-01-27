package request

import (
	"time"
)

// AddLicenseRequest represents request to add a new license
// AGT-030: Add License
type AddLicenseRequest struct {
	LicenseLine                 string     `json:"license_line" validate:"required,oneof=LIFE NON_LIFE HEALTH"`
	LicenseType                 string     `json:"license_type" validate:"required,oneof=PROVISIONAL PERMANENT"`
	LicenseNumber               string     `json:"license_number" validate:"required,min=5,max=50"`
	ResidentStatus              string     `json:"resident_status" validate:"required,oneof=RESIDENT NON_RESIDENT"`
	LicenseDate                 time.Time  `json:"license_date" validate:"required"`
	AuthorityDate               *time.Time `json:"authority_date,omitempty"`
	IsPrimary                   bool       `json:"is_primary"`
	LicentiateExamPassed        bool       `json:"licentiate_exam_passed"`
	LicentiateExamDate          *time.Time `json:"licentiate_exam_date,omitempty"`
	LicentiateCertificateNumber *string    `json:"licentiate_certificate_number,omitempty"`
	Metadata                    *string    `json:"metadata,omitempty"`
	CreatedBy                   string     `json:"created_by" validate:"required"`
}

// UpdateLicenseRequest represents request to update license
// AGT-032: Update License
type UpdateLicenseRequest struct {
	LicenseLine                 *string    `json:"license_line,omitempty" validate:"omitempty,oneof=LIFE NON_LIFE HEALTH"`
	LicenseType                 *string    `json:"license_type,omitempty" validate:"omitempty,oneof=PROVISIONAL PERMANENT"`
	LicenseNumber               *string    `json:"license_number,omitempty" validate:"omitempty,min=5,max=50"`
	ResidentStatus              *string    `json:"resident_status,omitempty" validate:"omitempty,oneof=RESIDENT NON_RESIDENT"`
	LicenseDate                 *time.Time `json:"license_date,omitempty"`
	AuthorityDate               *time.Time `json:"authority_date,omitempty"`
	LicenseStatus               *string    `json:"license_status,omitempty" validate:"omitempty,oneof=ACTIVE EXPIRED SUSPENDED TERMINATED RENEWED"`
	LicentiateExamPassed        *bool      `json:"licentiate_exam_passed,omitempty"`
	LicentiateExamDate          *time.Time `json:"licentiate_exam_date,omitempty"`
	LicentiateCertificateNumber *string    `json:"licentiate_certificate_number,omitempty"`
	IsPrimary                   *bool      `json:"is_primary,omitempty"`
	Metadata                    *string    `json:"metadata,omitempty"`
	UpdatedBy                   string     `json:"updated_by" validate:"required"`
}

// RenewLicenseRequest represents request to renew license
// AGT-033: Renew License
// BR-AGT-PRF-012: Complex renewal rules
type RenewLicenseRequest struct {
	RenewalType           string     `json:"renewal_type" validate:"required,oneof=PROVISIONAL_RENEWAL CONVERT_TO_PERMANENT PERMANENT_RENEWAL"`
	ExamPassed            bool       `json:"exam_passed"`
	ExamDate              *time.Time `json:"exam_date,omitempty"`
	ExamCertificateNumber *string    `json:"exam_certificate_number,omitempty"`
	Reason                string     `json:"reason" validate:"required"`
	UpdatedBy             string     `json:"updated_by" validate:"required"`
}

// GetExpiringLicensesQuery represents query parameters for expiring licenses
// AGT-036: Get Expiring Licenses
type GetExpiringLicensesQuery struct {
	Days       int    `query:"days" validate:"min=1,max=365"`
	OfficeCode string `query:"office_code"`
	Page       int    `query:"page" validate:"min=1"`
	Limit      int    `query:"limit" validate:"min=1,max=1000"`
}

// BatchDeactivateExpiredRequest represents request to batch deactivate expired licenses
// AGT-038: Batch Deactivate Expired Licenses
type BatchDeactivateExpiredRequest struct {
	ProcessedBy string `json:"processed_by" validate:"required"`
	DryRun      bool   `json:"dry_run"` // If true, only returns licenses to be deactivated without updating
}

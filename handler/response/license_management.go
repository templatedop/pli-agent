package response

import (
	"time"

	"pli-agent-api/core/port"
)

// LicenseDTO represents a license with calculated expiry info
type LicenseDTO struct {
	LicenseID                   string     `json:"license_id"`
	AgentID                     string     `json:"agent_id"`
	LicenseLine                 string     `json:"license_line"`
	LicenseType                 string     `json:"license_type"`
	LicenseNumber               string     `json:"license_number"`
	ResidentStatus              string     `json:"resident_status"`
	LicenseDate                 time.Time  `json:"license_date"`
	RenewalDate                 time.Time  `json:"renewal_date"`
	AuthorityDate               *time.Time `json:"authority_date,omitempty"`
	RenewalCount                int        `json:"renewal_count"`
	LicenseStatus               string     `json:"license_status"`
	IsPrimary                   bool       `json:"is_primary"`
	LicentiateExamPassed        bool       `json:"licentiate_exam_passed"`
	LicentiateExamDate          *time.Time `json:"licentiate_exam_date,omitempty"`
	LicentiateCertificateNumber *string    `json:"licentiate_certificate_number,omitempty"`
	Metadata                    *string    `json:"metadata,omitempty"`

	// Computed fields
	DaysUntilExpiry int    `json:"days_until_expiry"` // Negative if expired
	ExpiryStatus    string `json:"expiry_status"`     // VALID, EXPIRING_SOON, EXPIRED
	CanRenew        bool   `json:"can_renew"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgentLicensesResponse returns list of licenses for an agent
// AGT-029: Get Agent Licenses
type AgentLicensesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentID                   string       `json:"agent_id"`
	Licenses                  []LicenseDTO `json:"licenses"`
	TotalCount                int          `json:"total_count"`
	HasPrimaryLicense         bool         `json:"has_primary_license"`
}

// AddLicenseResponse returns newly created license
// AGT-030: Add License
type AddLicenseResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	License                   LicenseDTO `json:"license"`
}

// LicenseDetailsResponse returns detailed license information
// AGT-031: Get License Details
type LicenseDetailsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	License                   LicenseDTO         `json:"license"`
	RenewalHistory            []RenewalRecordDTO `json:"renewal_history"`
	CanRenew                  bool               `json:"can_renew"`
	RenewalEligibilityReason  string             `json:"renewal_eligibility_reason"`
}

// RenewalRecordDTO represents a renewal event in history
type RenewalRecordDTO struct {
	RenewedAt   time.Time  `json:"renewed_at"`
	RenewedBy   string     `json:"renewed_by"`
	RenewalType string     `json:"renewal_type"` // PROVISIONAL_RENEWAL, CONVERT_TO_PERMANENT, PERMANENT_RENEWAL
	OldType     string     `json:"old_type"`
	NewType     string     `json:"new_type"`
	OldExpiry   time.Time  `json:"old_expiry"`
	NewExpiry   time.Time  `json:"new_expiry"`
	ExamPassed  bool       `json:"exam_passed"`
	ExamDate    *time.Time `json:"exam_date,omitempty"`
}

// UpdateLicenseResponse returns updated license
// AGT-032: Update License
type UpdateLicenseResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	License                   LicenseDTO `json:"license"`
}

// RenewLicenseResponse returns renewed license
// AGT-033: Renew License
type RenewLicenseResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	License                   LicenseDTO `json:"license"`
	RenewalType               string     `json:"renewal_type"`
	PreviousExpiry            time.Time  `json:"previous_expiry"`
	NewExpiry                 time.Time  `json:"new_expiry"`
	RenewalMessage            string     `json:"renewal_message"`
}

// LicenseTypeDTO represents a license type with metadata
type LicenseTypeDTO struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	ValidityYears       int    `json:"validity_years"`
	RenewalIntervalDays int    `json:"renewal_interval_days"`
	MaxRenewals         int    `json:"max_renewals"` // -1 for unlimited
}

// LicenseTypesResponse returns available license types
// AGT-035: Get License Types
type LicenseTypesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	LicenseTypes              []LicenseTypeDTO `json:"license_types"`
}

// ExpiringLicenseDTO represents a license expiring soon
type ExpiringLicenseDTO struct {
	LicenseID     string    `json:"license_id"`
	AgentID       string    `json:"agent_id"`
	AgentCode     string    `json:"agent_code"`
	AgentName     string    `json:"agent_name"`
	LicenseLine   string    `json:"license_line"`
	LicenseType   string    `json:"license_type"`
	LicenseNumber string    `json:"license_number"`
	RenewalDate   time.Time `json:"renewal_date"`
	DaysRemaining int       `json:"days_remaining"`
	RenewalCount  int       `json:"renewal_count"`
	OfficeCode    string    `json:"office_code"`
	OfficeName    string    `json:"office_name"`
	ContactMobile string    `json:"contact_mobile"`
	ContactEmail  string    `json:"contact_email"`
}

// ExpiringLicensesResponse returns licenses expiring soon
// AGT-036: Get Expiring Licenses
type ExpiringLicensesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Licenses                  []ExpiringLicenseDTO `json:"licenses"`
	Pagination                PaginationMetadata   `json:"pagination"`
	Summary                   struct {
		TotalExpiring    int `json:"total_expiring"`
		ExpiringIn7Days  int `json:"expiring_in_7_days"`
		ExpiringIn15Days int `json:"expiring_in_15_days"`
		ExpiringIn30Days int `json:"expiring_in_30_days"`
	} `json:"summary"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// ReminderScheduleDTO represents reminder schedule for a license
type ReminderScheduleDTO struct {
	ReminderDate time.Time  `json:"reminder_date"`
	DaysBefore   int        `json:"days_before"`
	Status       string     `json:"status"` // PENDING, SENT, FAILED
	SentAt       *time.Time `json:"sent_at,omitempty"`
}

// LicenseRemindersResponse returns reminder schedule
// AGT-037: Get License Reminders
type LicenseRemindersResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	LicenseID                 string                `json:"license_id"`
	RenewalDate               time.Time             `json:"renewal_date"`
	DaysUntilExpiry           int                   `json:"days_until_expiry"`
	Reminders                 []ReminderScheduleDTO `json:"reminders"`
}

// BatchDeactivationSummary represents summary of batch deactivation
type BatchDeactivationSummary struct {
	TotalProcessed      int       `json:"total_processed"`
	SuccessfullyUpdated int       `json:"successfully_updated"`
	Failed              int       `json:"failed"`
	AffectedAgentIDs    []string  `json:"affected_agent_ids"`
	ProcessedAt         time.Time `json:"processed_at"`
}

// BatchDeactivateExpiredResponse returns batch deactivation result
// AGT-038: Batch Deactivate Expired Licenses
type BatchDeactivateExpiredResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Summary                   BatchDeactivationSummary `json:"summary"`
	DryRun                    bool                     `json:"dry_run"`
	Message                   string                   `json:"message"`
}

package domain

import (
	"database/sql"
	"time"
)

// AgentLicense represents agent license entity
// E-06: Agent License Entity
// BR-AGT-PRF-012: License Renewal Period Rules
type AgentLicense struct {
	// Primary Key
	LicenseID string `json:"license_id" db:"license_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// License Details (BR-AGT-PRF-012, VR-AGT-PRF-013, VR-AGT-PRF-014)
	LicenseLine     string `json:"license_line" db:"license_line"`           // LIFE
	LicenseType     string `json:"license_type" db:"license_type"`           // PROVISIONAL, PERMANENT
	LicenseNumber   string `json:"license_number" db:"license_number"`       // Unique
	ResidentStatus  string `json:"resident_status" db:"resident_status"`     // RESIDENT, NON_RESIDENT

	// License Dates (BR-AGT-PRF-030)
	LicenseDate   time.Time `json:"license_date" db:"license_date"`
	RenewalDate   time.Time `json:"renewal_date" db:"renewal_date"`
	AuthorityDate time.Time `json:"authority_date" db:"authority_date"`

	// Renewal Tracking (BR-AGT-PRF-012)
	RenewalCount  int    `json:"renewal_count" db:"renewal_count"`
	LicenseStatus string `json:"license_status" db:"license_status"` // ACTIVE, EXPIRED, RENEWED

	// License Exam Status (BR-AGT-PRF-012)
	LicentiatExamPassed          bool           `json:"licentiate_exam_passed" db:"licentiate_exam_passed"`
	LicentiateExamDate           sql.NullTime   `json:"licentiate_exam_date" db:"licentiate_exam_date"`
	LicenticateCertificateNumber sql.NullString `json:"licenticate_certificate_number" db:"licentiate_certificate_number"`

	// Primary License Flag
	IsPrimary bool `json:"is_primary" db:"is_primary"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"` // JSONB

	// Audit Fields
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at" db:"updated_at"`
	CreatedBy string         `json:"created_by" db:"created_by"`
	UpdatedBy sql.NullString `json:"updated_by" db:"updated_by"`
	DeletedAt sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	Version   int            `json:"version" db:"version"`
}

// License Line constants
const (
	LicenseLineLife    = "LIFE"
	LicenseLineGeneral = "GENERAL"
)

// License Type constants (BR-AGT-PRF-012)
const (
	LicenseTypeProvisional = "PROVISIONAL"
	LicenseTypePermanent   = "PERMANENT"
)

// Resident Status constants
const (
	ResidentStatusResident    = "RESIDENT"
	ResidentStatusNonResident = "NON_RESIDENT"
)

// License Status constants (BR-AGT-PRF-012, BR-AGT-PRF-013)
const (
	LicenseStatusActive  = "ACTIVE"
	LicenseStatusExpired = "EXPIRED"
	LicenseStatusRenewed = "RENEWED"
)

// Renewal Period Rules (BR-AGT-PRF-012)
const (
	RenewalPeriodProvisional = 365  // 1 year in days
	RenewalPeriodPermanent   = 1825 // 5 years in days
	MaxProvisionalRenewals   = 2    // Maximum provisional renewals before exam required
)

// Expiry Status for UI/frontend display
const (
	ExpiryStatusGreen  = "GREEN"  // More than 30 days
	ExpiryStatusYellow = "YELLOW" // 15-30 days
	ExpiryStatusRed    = "RED"    // Less than 15 days
)

// IsProvisional checks if license is provisional
func (l *AgentLicense) IsProvisional() bool {
	return l.LicenseType == LicenseTypeProvisional
}

// IsPermanent checks if license is permanent
func (l *AgentLicense) IsPermanent() bool {
	return l.LicenseType == LicenseTypePermanent
}

// IsActive checks if license is active
func (l *AgentLicense) IsActive() bool {
	return l.LicenseStatus == LicenseStatusActive
}

// IsExpired checks if license has expired
func (l *AgentLicense) IsExpired() bool {
	return l.LicenseStatus == LicenseStatusExpired || time.Now().After(l.RenewalDate)
}

// DaysUntilExpiry calculates days until renewal date
func (l *AgentLicense) DaysUntilExpiry() int {
	duration := time.Until(l.RenewalDate)
	return int(duration.Hours() / 24)
}

// GetExpiryStatus returns expiry status for UI display
// BR-AGT-PRF-014: License Renewal Reminders
func (l *AgentLicense) GetExpiryStatus() string {
	days := l.DaysUntilExpiry()
	if days > 30 {
		return ExpiryStatusGreen
	} else if days >= 15 {
		return ExpiryStatusYellow
	}
	return ExpiryStatusRed
}

// CanRenewProvisional checks if provisional license can be renewed
// BR-AGT-PRF-012: Maximum 2 provisional renewals before exam required
func (l *AgentLicense) CanRenewProvisional() bool {
	return l.IsProvisional() && l.RenewalCount < MaxProvisionalRenewals
}

// NeedsExamForPermanent checks if exam is required for permanent license
func (l *AgentLicense) NeedsExamForPermanent() bool {
	return l.IsProvisional() && !l.LicentiatExamPassed
}

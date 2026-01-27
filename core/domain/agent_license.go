package domain

import (
	"database/sql"
	"time"
)

// AgentLicense represents agent license entity
// E-06: Agent License Entity
// BR-AGT-PRF-012, BR-AGT-PRF-013, BR-AGT-PRF-014, BR-AGT-PRF-030
type AgentLicense struct {
	// Primary Key
	LicenseID string `json:"license_id" db:"license_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// License Details (BR-AGT-PRF-012, VR-AGT-PRF-013, VR-AGT-PRF-014)
	LicenseLine    string `json:"license_line" db:"license_line"`       // LIFE
	LicenseType    string `json:"license_type" db:"license_type"`       // PROVISIONAL, PERMANENT
	LicenseNumber  string `json:"license_number" db:"license_number"`   // VR-AGT-PRF-013: Unique
	ResidentStatus string `json:"resident_status" db:"resident_status"` // RESIDENT, NON_RESIDENT

	// License Dates (BR-AGT-PRF-030)
	LicenseDate   time.Time `json:"license_date" db:"license_date"`
	RenewalDate   time.Time `json:"renewal_date" db:"renewal_date"` // BR-AGT-PRF-012, BR-AGT-PRF-014
	AuthorityDate time.Time `json:"authority_date" db:"authority_date"`

	// Renewal Tracking (BR-AGT-PRF-012)
	RenewalCount  int    `json:"renewal_count" db:"renewal_count"`   // Max 2 for provisional
	LicenseStatus string `json:"license_status" db:"license_status"` // ACTIVE, EXPIRED, RENEWED

	// License Exam Status (BR-AGT-PRF-012)
	LicentiateExamPassed        bool           `json:"licentiate_exam_passed" db:"licentiate_exam_passed"`
	LicentiateExamDate          sql.NullTime   `json:"licentiate_exam_date" db:"licentiate_exam_date"`
	LicentiateCertificateNumber sql.NullString `json:"licentiate_certificate_number" db:"licentiate_certificate_number"`

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

// LicenseLine constants
const (
	LicenseLineLife = "LIFE"
)

// LicenseType constants (BR-AGT-PRF-012)
const (
	LicenseTypeProvisional = "PROVISIONAL"
	LicenseTypePermanent   = "PERMANENT"
)

// ResidentStatus constants
const (
	ResidentStatusResident    = "RESIDENT"
	ResidentStatusNonResident = "NON_RESIDENT"
)

// LicenseStatus constants (BR-AGT-PRF-012, BR-AGT-PRF-013)
const (
	LicenseStatusActive  = "ACTIVE"
	LicenseStatusExpired = "EXPIRED"
	LicenseStatusRenewed = "RENEWED"
)

// IsProvisional checks if license is provisional
// BR-AGT-PRF-012: License Renewal Period Rules
func (l *AgentLicense) IsProvisional() bool {
	return l.LicenseType == LicenseTypeProvisional
}

// CanBeRenewed checks if provisional license can be renewed
// BR-AGT-PRF-012: Max 2 renewals for provisional
func (l *AgentLicense) CanBeRenewed() bool {
	return l.IsProvisional() && l.RenewalCount < 2
}

// IsExpired checks if license has expired
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func (l *AgentLicense) IsExpired() bool {
	return time.Now().After(l.RenewalDate) && l.LicenseStatus != LicenseStatusRenewed
}

// GetDaysUntilRenewal returns days until renewal date
// BR-AGT-PRF-014: License Renewal Reminder Schedule
func (l *AgentLicense) GetDaysUntilRenewal() int {
	return int(time.Until(l.RenewalDate).Hours() / 24)
}

// AgentLicenseWithProfile combines license and basic agent profile info
// Used for optimized queries that need both license and profile data in single SELECT
// Eliminates N+1 query problem when fetching expiring licenses with agent details
type AgentLicenseWithProfile struct {
	// License fields
	LicenseID     string    `json:"license_id" db:"license_id"`
	AgentID       string    `json:"agent_id" db:"agent_id"`
	LicenseLine   string    `json:"license_line" db:"license_line"`
	LicenseType   string    `json:"license_type" db:"license_type"`
	LicenseNumber string    `json:"license_number" db:"license_number"`
	LicenseDate   time.Time `json:"license_date" db:"license_date"`
	RenewalDate   time.Time `json:"renewal_date" db:"renewal_date"`
	RenewalCount  int       `json:"renewal_count" db:"renewal_count"`
	LicenseStatus string    `json:"license_status" db:"license_status"`

	// Agent profile fields (from JOIN)
	AgentCode  string `json:"agent_code" db:"agent_code"`
	FirstName  string `json:"first_name" db:"first_name"`
	MiddleName string `json:"middle_name" db:"middle_name"`
	LastName   string `json:"last_name" db:"last_name"`
	OfficeCode string `json:"office_code" db:"office_code"`
}

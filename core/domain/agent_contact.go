package domain

import (
	"database/sql"
	"time"
)

// AgentContact represents agent contact entity
// E-03: Agent Contact Entity
// BR-AGT-PRF-010
type AgentContact struct {
	// Primary Key
	ContactID string `json:"contact_id" db:"contact_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Contact Type (BR-AGT-PRF-010)
	ContactType string `json:"contact_type" db:"contact_type"` // MOBILE, OFFICIAL_LANDLINE, RESIDENT_LANDLINE

	// Contact Number (VR-AGT-PRF-011)
	ContactNumber string `json:"contact_number" db:"contact_number"` // 10 digits for mobile

	// Primary Contact Flag
	IsPrimary bool `json:"is_primary" db:"is_primary"`

	// Effective Date
	EffectiveFrom time.Time `json:"effective_from" db:"effective_from"`

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

// ContactType constants (VR-AGT-PRF-011)
const (
	ContactTypeMobile           = "MOBILE"
	ContactTypeOfficialLandline = "OFFICIAL_LANDLINE"
	ContactTypeResidentLandline = "RESIDENT_LANDLINE"
)

package domain

import (
	"database/sql"
	"time"
)

// AgentAddress represents agent address entity
// E-02: Agent Address Entity
// BR-AGT-PRF-008, BR-AGT-PRF-009
type AgentAddress struct {
	// Primary Key
	AddressID string `json:"address_id" db:"address_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Address Type (BR-AGT-PRF-008)
	AddressType string `json:"address_type" db:"address_type"` // OFFICIAL, PERMANENT, COMMUNICATION

	// Address Fields (VR-AGT-PRF-007, VR-AGT-PRF-008)
	AddressLine1 string         `json:"address_line1" db:"address_line1"`
	AddressLine2 sql.NullString `json:"address_line2" db:"address_line2"`
	Village      sql.NullString `json:"village" db:"village"`
	Taluka       sql.NullString `json:"taluka" db:"taluka"`
	City         string         `json:"city" db:"city"`
	District     sql.NullString `json:"district" db:"district"`
	State        string         `json:"state" db:"state"`
	Country      string         `json:"country" db:"country"`
	Pincode      string         `json:"pincode" db:"pincode"` // VR-AGT-PRF-008: 6 digits

	// Communication Address Flag (BR-AGT-PRF-009)
	IsSameAsPermanent bool `json:"is_same_as_permanent" db:"is_same_as_permanent"`

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

// AddressType constants (BR-AGT-PRF-008)
const (
	AddressTypeOfficial      = "OFFICIAL"
	AddressTypePermanent     = "PERMANENT"
	AddressTypeCommunication = "COMMUNICATION"
)

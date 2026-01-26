package domain

import (
	"database/sql"
	"time"
)

// AgentEmail represents agent email entity
// E-04: Agent Email Entity
// BR-AGT-PRF-011
type AgentEmail struct {
	// Primary Key
	EmailID string `json:"email_id" db:"email_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Email Type (BR-AGT-PRF-011)
	EmailType string `json:"email_type" db:"email_type"` // OFFICIAL, PERMANENT, COMMUNICATION

	// Email Address (VR-AGT-PRF-012)
	EmailAddress string `json:"email_address" db:"email_address"`

	// Primary Email Flag
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

// EmailType constants (BR-AGT-PRF-011)
const (
	EmailTypeOfficial      = "OFFICIAL"
	EmailTypePermanent     = "PERMANENT"
	EmailTypeCommunication = "COMMUNICATION"
)

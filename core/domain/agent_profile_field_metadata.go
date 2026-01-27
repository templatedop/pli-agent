package domain

import (
	"database/sql"
	"time"
)

// AgentProfileFieldMetadata represents metadata about a profile field
// Phase 6.2: Dynamic field metadata for form generation
type AgentProfileFieldMetadata struct {
	// Primary Key
	FieldID int `json:"field_id" db:"field_id"`

	// Field identification
	FieldName string `json:"field_name" db:"field_name"`
	Section   string `json:"section" db:"section"` // personal_info, address, contact, email, bank, license

	// Display information
	DisplayName  string         `json:"display_name" db:"display_name"`
	DisplayOrder int            `json:"display_order" db:"display_order"`
	FieldType    string         `json:"field_type" db:"field_type"` // text, number, date, email, phone, select, textarea, checkbox
	Placeholder  sql.NullString `json:"placeholder" db:"placeholder"`
	HelpText     sql.NullString `json:"help_text" db:"help_text"`

	// Editability and validation
	IsEditable       bool `json:"is_editable" db:"is_editable"`
	IsRequired       bool `json:"is_required" db:"is_required"`
	RequiresApproval bool `json:"requires_approval" db:"requires_approval"`

	// Validation rules and options (JSONB)
	ValidationRules sql.NullString `json:"validation_rules" db:"validation_rules"` // JSONB stored as string
	SelectOptions   sql.NullString `json:"select_options" db:"select_options"`     // JSONB stored as string

	// Metadata
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Table name
const AgentProfileFieldMetadataTable = "agent_profile_field_metadata"

// Section constants
const (
	FieldSectionPersonalInfo = "personal_info"
	FieldSectionAddress      = "address"
	FieldSectionContact      = "contact"
	FieldSectionEmail        = "email"
	FieldSectionBank         = "bank"
	FieldSectionLicense      = "license"
)

// Field type constants
const (
	FieldTypeText     = "text"
	FieldTypeNumber   = "number"
	FieldTypeDate     = "date"
	FieldTypeEmail    = "email"
	FieldTypePhone    = "phone"
	FieldTypeSelect   = "select"
	FieldTypeTextarea = "textarea"
	FieldTypeCheckbox = "checkbox"
)

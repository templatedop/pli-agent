package domain

import (
	"database/sql"
	"time"
)

// AgentAuditLog represents agent audit log entity
// E-08: Agent Audit Log Entity
// BR-AGT-PRF-005, BR-AGT-PRF-006
type AgentAuditLog struct {
	// Primary Key
	AuditID string `json:"audit_id" db:"audit_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Action Details (BR-AGT-PRF-005, BR-AGT-PRF-006)
	ActionType   string         `json:"action_type" db:"action_type"`
	FieldName    sql.NullString `json:"field_name" db:"field_name"`
	OldValue     sql.NullString `json:"old_value" db:"old_value"`
	NewValue     sql.NullString `json:"new_value" db:"new_value"`
	ActionReason sql.NullString `json:"action_reason" db:"action_reason"`

	// Performed By (BR-AGT-PRF-005)
	PerformedBy string    `json:"performed_by" db:"performed_by"`
	PerformedAt time.Time `json:"performed_at" db:"performed_at"`

	// IP Address for Security
	IPAddress sql.NullString `json:"ip_address" db:"ip_address"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"` // JSONB

	// Audit Fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuditAction constants
const (
	AuditActionCreate        = "CREATE"
	AuditActionUpdate        = "UPDATE"
	AuditActionDelete        = "DELETE"
	AuditActionStatusChange  = "STATUS_CHANGE"
	AuditActionLicenseAdd    = "LICENSE_ADD"
	AuditActionLicenseUpdate = "LICENSE_UPDATE"
	AuditActionBankUpdate    = "BANK_UPDATE"
	AuditActionAddressUpdate = "ADDRESS_UPDATE"
	AuditActionContactUpdate = "CONTACT_UPDATE"
	AuditActionEmailUpdate   = "EMAIL_UPDATE"
	AuditActionLogin         = "LOGIN"
	AuditActionLogout        = "LOGOUT"
	AuditActionTerminate     = "TERMINATE"
	AuditActionActivate      = "ACTIVATE"
	AuditActionSuspend       = "SUSPEND"
)

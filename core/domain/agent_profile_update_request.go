package domain

import (
	"database/sql"
	"time"
)

// AgentProfileUpdateRequest represents a profile update request awaiting approval
// Used for critical field updates (name, PAN, Aadhar) that require approval
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Validation
type AgentProfileUpdateRequest struct {
	// Primary Key
	RequestID string `json:"request_id" db:"request_id"` // UUID

	// Agent Information
	AgentID string `json:"agent_id" db:"agent_id"` // FK to agent_profiles

	// Request Details
	Section      string         `json:"section" db:"section"`             // personal_info, address, contact, email
	FieldUpdates sql.NullString `json:"field_updates" db:"field_updates"` // JSONB: {"field_name": "new_value"}
	Reason       sql.NullString `json:"reason" db:"reason"`               // Reason for update
	RequestedBy  string         `json:"requested_by" db:"requested_by"`   // User ID who requested
	RequestedAt  time.Time      `json:"requested_at" db:"requested_at"`

	// Approval Status
	Status     string         `json:"status" db:"status"`           // PENDING, APPROVED, REJECTED
	ApprovedBy sql.NullString `json:"approved_by" db:"approved_by"` // User ID who approved
	ApprovedAt sql.NullTime   `json:"approved_at" db:"approved_at"`
	RejectedBy sql.NullString `json:"rejected_by" db:"rejected_by"` // User ID who rejected
	RejectedAt sql.NullTime   `json:"rejected_at" db:"rejected_at"`
	Comments   sql.NullString `json:"comments" db:"comments"` // Approval/rejection comments

	// Metadata
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

// Update request status constants
const (
	UpdateRequestStatusPending  = "PENDING"
	UpdateRequestStatusApproved = "APPROVED"
	UpdateRequestStatusRejected = "REJECTED"
)

// Table name
const AgentProfileUpdateRequestTable = "agent_profile_update_requests"

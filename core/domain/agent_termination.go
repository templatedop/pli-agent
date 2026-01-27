package domain

import (
	"database/sql"
	"time"
)

// AgentTerminationRecord represents a complete termination record
// AGT-039: Terminate Agent
// BR-AGT-PRF-016: Status Update with Reason
// BR-AGT-PRF-017: Agent Termination Workflow
type AgentTerminationRecord struct {
	// Primary Key
	TerminationID string `json:"termination_id" db:"termination_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Termination Details
	TerminationDate      time.Time `json:"termination_date" db:"termination_date"`
	EffectiveDate        time.Time `json:"effective_date" db:"effective_date"`
	TerminationReason    string    `json:"termination_reason" db:"termination_reason"` // Min 20 chars
	TerminationReasonCode string    `json:"termination_reason_code" db:"termination_reason_code"`
	TerminatedBy         string    `json:"terminated_by" db:"terminated_by"`

	// Workflow Tracking (WF-AGT-PRF-004)
	WorkflowID     sql.NullString `json:"workflow_id" db:"workflow_id"`
	WorkflowStatus string          `json:"workflow_status" db:"workflow_status"`

	// Actions Performed
	StatusUpdated      bool `json:"status_updated" db:"status_updated"`
	PortalDisabled     bool `json:"portal_disabled" db:"portal_disabled"`
	CommissionStopped  bool `json:"commission_stopped" db:"commission_stopped"`
	LetterGenerated    bool `json:"letter_generated" db:"letter_generated"`
	DataArchived       bool `json:"data_archived" db:"data_archived"`
	NotificationsSent  bool `json:"notifications_sent" db:"notifications_sent"`

	// Generated Documents
	TerminationLetterURL         sql.NullString `json:"termination_letter_url" db:"termination_letter_url"`
	TerminationLetterGeneratedAt sql.NullTime   `json:"termination_letter_generated_at" db:"termination_letter_generated_at"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"`

	// Audit Fields
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
	Version   int          `json:"version" db:"version"`
}

// TerminationReasonCode constants
const (
	TerminationReasonResignation     = "RESIGNATION"
	TerminationReasonMisconduct      = "MISCONDUCT"
	TerminationReasonNonPerformance  = "NON_PERFORMANCE"
	TerminationReasonFraud           = "FRAUD"
	TerminationReasonLicenseExpired  = "LICENSE_EXPIRED"
	TerminationReasonOther           = "OTHER"
)

// WorkflowStatus constants
const (
	WorkflowStatusPending    = "PENDING"
	WorkflowStatusInProgress = "IN_PROGRESS"
	WorkflowStatusCompleted  = "COMPLETED"
	WorkflowStatusFailed     = "FAILED"
)

// AgentReinstatementRequest represents a reinstatement request
// AGT-041: Reinstate Agent
// WF-AGT-PRF-011: Reinstatement Workflow
type AgentReinstatementRequest struct {
	// Primary Key
	ReinstatementID string `json:"reinstatement_id" db:"reinstatement_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Request Details
	RequestDate         time.Time `json:"request_date" db:"request_date"`
	ReinstatementReason string    `json:"reinstatement_reason" db:"reinstatement_reason"` // Min 10 chars
	RequestedBy         string    `json:"requested_by" db:"requested_by"`

	// Approval Workflow
	Status          string         `json:"status" db:"status"`
	ApprovedBy      sql.NullString `json:"approved_by" db:"approved_by"`
	ApprovedAt      sql.NullTime   `json:"approved_at" db:"approved_at"`
	RejectedBy      sql.NullString `json:"rejected_by" db:"rejected_by"`
	RejectedAt      sql.NullTime   `json:"rejected_at" db:"rejected_at"`
	RejectionReason sql.NullString `json:"rejection_reason" db:"rejection_reason"`

	// Workflow Tracking (WF-AGT-PRF-011)
	WorkflowID sql.NullString `json:"workflow_id" db:"workflow_id"`

	// Conditions and Terms
	ReinstatementConditions sql.NullString `json:"reinstatement_conditions" db:"reinstatement_conditions"`
	ProbationPeriodDays     sql.NullInt32  `json:"probation_period_days" db:"probation_period_days"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"`

	// Audit Fields
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
	Version   int          `json:"version" db:"version"`
}

// ReinstatementStatus constants
const (
	ReinstatementStatusPending   = "PENDING"
	ReinstatementStatusApproved  = "APPROVED"
	ReinstatementStatusRejected  = "REJECTED"
	ReinstatementStatusCompleted = "COMPLETED"
)

// AgentDataArchive represents archived agent data
// BR-AGT-PRF-017: 7-year data retention
type AgentDataArchive struct {
	// Primary Key
	ArchiveID string `json:"archive_id" db:"archive_id"`

	// Agent Reference
	AgentID string `json:"agent_id" db:"agent_id"`

	// Archive Details
	ArchiveDate    time.Time `json:"archive_date" db:"archive_date"`
	ArchiveType    string    `json:"archive_type" db:"archive_type"`
	RetentionUntil time.Time `json:"retention_until" db:"retention_until"` // 7 years

	// Archived Data
	DataSnapshot  string         `json:"data_snapshot" db:"data_snapshot"` // JSONB
	DataChecksum  sql.NullString `json:"data_checksum" db:"data_checksum"`

	// Storage Reference
	StorageLocation sql.NullString `json:"storage_location" db:"storage_location"`
	StorageSizeBytes sql.NullInt64  `json:"storage_size_bytes" db:"storage_size_bytes"`

	// Metadata
	ArchivedBy string         `json:"archived_by" db:"archived_by"`
	Metadata   sql.NullString `json:"metadata" db:"metadata"`

	// Audit Fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ArchiveType constants
const (
	ArchiveTypeTermination    = "TERMINATION"
	ArchiveTypeReinstatement  = "REINSTATEMENT"
	ArchiveTypePeriodic       = "PERIODIC"
	ArchiveTypeManual         = "MANUAL"
)

// IsPending checks if termination workflow is pending
func (t *AgentTerminationRecord) IsPending() bool {
	return t.WorkflowStatus == WorkflowStatusPending || t.WorkflowStatus == WorkflowStatusInProgress
}

// IsCompleted checks if all termination actions are completed
func (t *AgentTerminationRecord) IsCompleted() bool {
	return t.StatusUpdated && t.PortalDisabled && t.CommissionStopped &&
		t.LetterGenerated && t.DataArchived && t.NotificationsSent
}

// IsPending checks if reinstatement request is pending approval
func (r *AgentReinstatementRequest) IsPending() bool {
	return r.Status == ReinstatementStatusPending
}

// IsApproved checks if reinstatement request is approved
func (r *AgentReinstatementRequest) IsApproved() bool {
	return r.Status == ReinstatementStatusApproved
}

// IsRejected checks if reinstatement request is rejected
func (r *AgentReinstatementRequest) IsRejected() bool {
	return r.Status == ReinstatementStatusRejected
}

// IsExpired checks if archive retention period has expired
func (a *AgentDataArchive) IsExpired() bool {
	return time.Now().After(a.RetentionUntil)
}

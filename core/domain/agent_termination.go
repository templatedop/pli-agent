package domain

import (
	"time"
)

// AgentTermination represents an agent termination record
// E-07: Agent Termination Entity
// BR-AGT-PRF-017: Agent Termination Workflow
type AgentTermination struct {
	TerminationID         string     `db:"termination_id"`
	AgentID               string     `db:"agent_id"`
	TerminationReasonCode string     `db:"termination_reason_code"`
	TerminationReasonText string     `db:"termination_reason_text"`
	EffectiveDate         time.Time  `db:"effective_date"`
	TerminationLetterURL  string     `db:"termination_letter_url"`
	PortalDisabledAt      *time.Time `db:"portal_disabled_at"`
	CommissionStoppedAt   *time.Time `db:"commission_stopped_at"`
	ArchiveID             string     `db:"archive_id"`
	WorkflowID            string     `db:"workflow_id"`
	WorkflowStatus        string     `db:"workflow_status"`
	CreatedBy             string     `db:"created_by"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedBy             string     `db:"updated_by"`
	UpdatedAt             time.Time  `db:"updated_at"`
	DeletedAt             *time.Time `db:"deleted_at"`
	Version               int        `db:"version"`
}

// Termination Reason Codes
// VR-AGT-PRF-020: Termination Reason Mandatory
const (
	TerminationReasonResignation    = "RESIGNATION"
	TerminationReasonMisconduct     = "MISCONDUCT"
	TerminationReasonNonPerformance = "NON_PERFORMANCE"
	TerminationReasonFraud          = "FRAUD"
	TerminationReasonOther          = "OTHER"
)

// Workflow Status Constants
const (
	WorkflowStatusPending   = "PENDING"
	WorkflowStatusRunning   = "RUNNING"
	WorkflowStatusCompleted = "COMPLETED"
	WorkflowStatusFailed    = "FAILED"
)

// IsTerminated checks if agent has been terminated
func (t *AgentTermination) IsTerminated() bool {
	return t.EffectiveDate.Before(time.Now()) || t.EffectiveDate.Equal(time.Now())
}

// IsPortalDisabled checks if portal access is disabled
func (t *AgentTermination) IsPortalDisabled() bool {
	return t.PortalDisabledAt != nil
}

// IsCommissionStopped checks if commission processing is stopped
func (t *AgentTermination) IsCommissionStopped() bool {
	return t.CommissionStoppedAt != nil
}

// IsWorkflowCompleted checks if termination workflow is completed
func (t *AgentTermination) IsWorkflowCompleted() bool {
	return t.WorkflowStatus == WorkflowStatusCompleted
}

// GetTerminationReasonName returns human-readable termination reason
func (t *AgentTermination) GetTerminationReasonName() string {
	switch t.TerminationReasonCode {
	case TerminationReasonResignation:
		return "Resignation"
	case TerminationReasonMisconduct:
		return "Misconduct"
	case TerminationReasonNonPerformance:
		return "Non-Performance"
	case TerminationReasonFraud:
		return "Fraud"
	case TerminationReasonOther:
		return "Other"
	default:
		return "Unknown"
	}
}

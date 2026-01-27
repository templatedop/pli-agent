package domain

import (
	"database/sql"
	"time"
)

// AgentProfileSession represents a profile creation workflow session
// WF-AGT-PRF-001: Profile Creation Workflow
type AgentProfileSession struct {
	// Primary Key
	SessionID string `json:"session_id" db:"session_id"`

	// Agent Type
	AgentType string `json:"agent_type" db:"agent_type"`

	// Workflow State
	WorkflowState      string `json:"workflow_state" db:"workflow_state"`
	CurrentStep        string `json:"current_step" db:"current_step"`
	NextStep           string `json:"next_step" db:"next_step"`
	ProgressPercentage int    `json:"progress_percentage" db:"progress_percentage"`

	// Session Data
	FormData         sql.NullString `json:"form_data" db:"form_data"`                 // JSONB
	ValidationErrors sql.NullString `json:"validation_errors" db:"validation_errors"` // JSONB

	// Temporal Integration
	TemporalWorkflowID sql.NullString `json:"temporal_workflow_id" db:"temporal_workflow_id"`
	TemporalRunID      sql.NullString `json:"temporal_run_id" db:"temporal_run_id"`

	// Status
	Status string `json:"status" db:"status"` // ACTIVE, EXPIRED, COMPLETED, CANCELLED

	// Timestamps
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at" db:"updated_at"`
	ExpiresAt   time.Time    `json:"expires_at" db:"expires_at"`
	CompletedAt sql.NullTime `json:"completed_at" db:"completed_at"`

	// User Tracking
	InitiatedBy   string         `json:"initiated_by" db:"initiated_by"`
	LastUpdatedBy sql.NullString `json:"last_updated_by" db:"last_updated_by"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"` // JSONB

	// Created Agent
	CreatedAgentID sql.NullString `json:"created_agent_id" db:"created_agent_id"`
}

// Session Status Constants
const (
	SessionStatusActive    = "ACTIVE"
	SessionStatusExpired   = "EXPIRED"
	SessionStatusCompleted = "COMPLETED"
	SessionStatusCancelled = "CANCELLED"
)

// Workflow State Constants
const (
	WorkflowStateInitiated          = "INITIATED"
	WorkflowStateHRMSFetching       = "HRMS_FETCHING"
	WorkflowStateHRMSFetched        = "HRMS_FETCHED"
	WorkflowStateCoordinatorLinking = "COORDINATOR_LINKING"
	WorkflowStateProfileValidation  = "PROFILE_VALIDATION"
	WorkflowStateProfileSubmitting  = "PROFILE_SUBMITTING"
	WorkflowStateCompleted          = "COMPLETED"
	WorkflowStateCancelled          = "CANCELLED"
	WorkflowStateExpired            = "EXPIRED"
)

// IsActive checks if session is active
func (s *AgentProfileSession) IsActive() bool {
	return s.Status == SessionStatusActive && time.Now().Before(s.ExpiresAt)
}

// IsExpired checks if session has expired
func (s *AgentProfileSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

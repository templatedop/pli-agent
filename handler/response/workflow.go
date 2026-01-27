package response

import (
	"time"

	"pli-agent-api/core/port"
)

// ========================================================================
// WORKFLOW API RESPONSES (AGT-016 to AGT-021)
// ========================================================================

// WorkflowState represents the current workflow state
type WorkflowState struct {
	CurrentStep        string   `json:"current_step"`
	NextStep           string   `json:"next_step"`
	AllowedActions     []string `json:"allowed_actions"`
	ProgressPercentage int      `json:"progress_percentage"` // 0-100
}

// SessionStatusResponse returns session status
// AGT-016: Get Session Status
// WF-AGT-PRF-001: Profile Creation Workflow
type SessionStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	SessionID                 string         `json:"session_id"`
	Status                    string         `json:"status"` // ACTIVE, EXPIRED, COMPLETED, CANCELLED
	WorkflowState             *WorkflowState `json:"workflow_state"`
	LastSavedAt               *time.Time     `json:"last_saved_at,omitempty"`
}

// SaveSessionResponse returns session save result
// AGT-017: Save Session Checkpoint
// WF-AGT-PRF-001: Profile Creation Workflow
type SaveSessionResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Saved                     bool       `json:"saved"`
	SessionExpiresAt          *time.Time `json:"session_expires_at,omitempty"`
}

// ResumeSessionResponse returns saved session data
// AGT-018: Resume Session
// WF-AGT-PRF-001: Profile Creation Workflow
type ResumeSessionResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	SessionID                 string                 `json:"session_id"`
	AgentType                 string                 `json:"agent_type"`
	FormData                  map[string]interface{} `json:"form_data"`
	WorkflowState             *WorkflowState         `json:"workflow_state"`
}

// CancelSessionResponse returns session cancellation result
// AGT-019: Cancel Session
// WF-AGT-PRF-001: Profile Creation Workflow
type CancelSessionResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Cancelled                 bool   `json:"cancelled"`
	Message                   string `json:"message"`
}

// VerificationStatus represents verification status for various fields
type VerificationStatus struct {
	PANVerified    bool `json:"pan_verified"`
	HRMSVerified   bool `json:"hrms_verified"`
	OfficeVerified bool `json:"office_verified"`
}

// SLATracking represents SLA tracking information
type SLATracking struct {
	SLAStatus          string          `json:"sla_status"` // GREEN, YELLOW, RED
	TimeElapsedMinutes int             `json:"time_elapsed_minutes"`
	NextActionsDue     []NextActionDue `json:"next_actions_due"`
}

// NextActionDue represents a due action
type NextActionDue struct {
	Action  string     `json:"action"`
	DueDate *time.Time `json:"due_date,omitempty"`
}

// CreationStatusResponse returns agent creation status
// AGT-020: Get Creation Status
// FR-AGT-PRF-009: Agent Profile Status Tracking
type CreationStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentID                   string              `json:"agent_id"`
	Status                    string              `json:"status"`
	CurrentStage              string              `json:"current_stage"`
	VerificationStatus        *VerificationStatus `json:"verification_status"`
	SLATracking               *SLATracking        `json:"sla_tracking,omitempty"`
}

// WelcomeNotificationResponse returns welcome notification result
// AGT-021: Resend Welcome Notification
// INT-AGT-005: Notification Service Integration
type WelcomeNotificationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	NotificationID            string     `json:"notification_id"`
	ChannelsSent              []string   `json:"channels_sent"`
	SentAt                    *time.Time `json:"sent_at"`
}

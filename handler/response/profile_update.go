package response

import (
	"time"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
)

// ==================== PHASE 6: PROFILE UPDATE APIs (AGT-022 to AGT-028) ====================

// AgentSummary represents agent information in search results
// AGT-022: Multi-criteria Agent Search
type AgentSummary struct {
	AgentID      string `json:"agent_id"`
	AgentCode    string `json:"agent_code,omitempty"`
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name,omitempty"`
	LastName     string `json:"last_name"`
	AgentType    string `json:"agent_type"`
	Status       string `json:"status"`
	OfficeCode   string `json:"office_code"`
	PANNumber    string `json:"pan_number"`
	MobileNumber string `json:"mobile_number,omitempty"`
}

// AgentSearchResponse response for agent search
// AGT-022: Multi-criteria Agent Search
type AgentSearchResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Data []AgentSummary `json:"data"`
}

// AgentProfileDetailsResponse response for agent profile details
// AGT-023: Get Agent Profile Details
type AgentProfileDetailsResponse struct {
	port.StatusCodeAndMessage
	Profile   domain.AgentProfile   `json:"profile"`
	Addresses []domain.AgentAddress `json:"addresses,omitempty"`
	Contacts  []domain.AgentContact `json:"contacts,omitempty"`
	Emails    []domain.AgentEmail   `json:"emails,omitempty"`
}

// UpdateFormResponse response for update form data
// AGT-024: Get Profile Update Form
type UpdateFormResponse struct {
	port.StatusCodeAndMessage
	Section        string                 `json:"section"`
	FormData       map[string]interface{} `json:"form_data"`
	EditableFields []string               `json:"editable_fields"`
	CriticalFields []string               `json:"critical_fields"` // Fields requiring approval
}

// UpdateProfileResponse response for profile section update
// AGT-025: Update Profile Section
type UpdateProfileResponse struct {
	port.StatusCodeAndMessage
	ApprovalRequired  bool    `json:"approval_required"`
	ApprovalRequestID *string `json:"approval_request_id,omitempty"`
	Status            string  `json:"status"` // UPDATED, PENDING_APPROVAL
	UpdatedFields     []string `json:"updated_fields,omitempty"`
	Version           int      `json:"version,omitempty"`
}

// ApprovalActionResponse response for approve/reject actions
// AGT-026: Approve Profile Update
// AGT-027: Reject Profile Update
type ApprovalActionResponse struct {
	port.StatusCodeAndMessage
	ApprovalRequestID string     `json:"approval_request_id"`
	Status            string     `json:"status"` // APPROVED, REJECTED
	ReviewedBy        string     `json:"reviewed_by"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewComments    string     `json:"review_comments,omitempty"`
}

// AuditLogEntry represents a single audit log entry
// AGT-028: Get Audit History
type AuditLogEntry struct {
	AuditID      string    `json:"audit_id"`
	ActionType   string    `json:"action_type"`
	ActionReason string    `json:"action_reason,omitempty"`
	FieldName    string    `json:"field_name,omitempty"`
	OldValue     string    `json:"old_value,omitempty"`
	NewValue     string    `json:"new_value,omitempty"`
	PerformedBy  string    `json:"performed_by"`
	PerformedAt  time.Time `json:"performed_at"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
}

// AuditHistoryResponse response for audit history
// AGT-028: Get Audit History
type AuditHistoryResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Data []AuditLogEntry `json:"data"`
}

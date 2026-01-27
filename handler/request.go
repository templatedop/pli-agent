package handler

import (
	"database/sql"
	"time"
)

// ========================================================================
// VALIDATION API REQUESTS (AGT-012 to AGT-015)
// ========================================================================

// CheckPANUniquenessRequest validates PAN uniqueness
// AGT-012: Check PAN Uniqueness
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// VR-AGT-PRF-002: PAN Uniqueness
type CheckPANUniquenessRequest struct {
	PANNumber      string `json:"pan_number" validate:"required,len=10,uppercase"`
	ExcludeAgentID string `json:"exclude_agent_id" validate:"omitempty,uuid4"`
}

// ValidateEmployeeIDRequest validates employee ID against HRMS
// AGT-013: Validate Employee ID (HRMS)
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
// VR-AGT-PRF-023: HRMS Employee ID Validation
type ValidateEmployeeIDRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
}

// ValidateIFSCRequest validates IFSC code
// AGT-014: Validate IFSC Code
// BR-AGT-PRF-018: Bank Account Details for Commission Disbursement
// VR-AGT-PRF-017: IFSC Code Format Validation
type ValidateIFSCRequest struct {
	IFSCCode string `json:"ifsc_code" validate:"required,len=11,uppercase"`
}

// ========================================================================
// WORKFLOW API REQUESTS (AGT-016 to AGT-019)
// ========================================================================

// SessionIDUri represents session ID in URI
type SessionIDUri struct {
	SessionID string `uri:"session_id" validate:"required,uuid4"`
}

// AgentIDUri represents agent ID in URI
type AgentIDUri struct {
	AgentID string `uri:"agent_id" validate:"required,uuid4"`
}

// OfficeCodeUri represents office code in URI
type OfficeCodeUri struct {
	OfficeCode string `uri:"office_code" validate:"required"`
}

// SaveSessionRequest saves profile creation session checkpoint
// AGT-017: Save Session Checkpoint
// WF-AGT-PRF-001: Profile Creation Workflow
type SaveSessionRequest struct {
	SessionID     string                 `uri:"session_id" validate:"required,uuid4"`
	CurrentScreen string                 `json:"current_screen" validate:"required"`
	FormData      map[string]interface{} `json:"form_data" validate:"required"`
}

// ========================================================================
// NOTIFICATION API REQUESTS (AGT-021)
// ========================================================================

// ResendWelcomeNotificationRequest resends welcome notification
// AGT-021: Resend Welcome Notification
// INT-AGT-005: Notification Service Integration
type ResendWelcomeNotificationRequest struct {
	AgentID  string   `uri:"agent_id" validate:"required,uuid4"`
	Channels []string `json:"channels" validate:"omitempty,dive,oneof=EMAIL SMS"`
}

// ========================================================================
// PROFILE CREATION WORKFLOW REQUESTS (for Phase 5)
// ========================================================================

// InitiateProfileRequest initiates agent profile creation
// AGT-001: Initiate Agent Profile Creation
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-031: Workflow Orchestration
type InitiateProfileRequest struct {
	AgentType   string `json:"agent_type" validate:"required,oneof=ADVISOR ADVISOR_COORDINATOR DEPARTMENTAL_EMPLOYEE FIELD_OFFICER DIRECT_AGENT GDS"`
	InitiatedBy string `json:"initiated_by" validate:"required"`
}

// FetchHRMSEmployeeRequest fetches employee data from HRMS
// AGT-002: Fetch HRMS Employee Data
// FR-AGT-PRF-002: HRMS Data Auto-Population
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
type FetchHRMSEmployeeRequest struct {
	SessionID  string `uri:"session_id" validate:"required,uuid4"`
	EmployeeID string `json:"employee_id" validate:"required"`
	FetchMode  string `json:"fetch_mode" validate:"required,oneof=AUTO_HRMS MANUAL_ENTRY"`
}

// LinkCoordinatorRequest links advisor to coordinator
// AGT-004: Link Advisor to Coordinator
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
type LinkCoordinatorRequest struct {
	SessionID            string       `uri:"session_id" validate:"required,uuid4"`
	CoordinatorID        string       `json:"coordinator_id" validate:"required,uuid4"`
	LinkageEffectiveDate sql.NullTime `json:"linkage_effective_date" validate:"omitempty"`
}

// GetAdvisorCoordinatorsQuery query params for coordinator list
// AGT-003: Get Advisor Coordinators List
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
type GetAdvisorCoordinatorsQuery struct {
	Status     string `query:"status" validate:"omitempty,oneof=ACTIVE INACTIVE"`
	CircleID   string `query:"circle_id" validate:"omitempty"`
	DivisionID string `query:"division_id" validate:"omitempty"`
	Page       uint64 `query:"page" validate:"omitempty,min=1"`
	Limit      uint64 `query:"limit" validate:"omitempty,min=1,max=100"`
}

// SubmitProfileRequest submits profile for creation
// AGT-006: Submit Profile for Creation
// FR-AGT-PRF-001: New Profile Creation
// WF-002: Agent Onboarding Workflow
type SubmitProfileRequest struct {
	SessionID   string `uri:"session_id" validate:"required,uuid4"`
	SubmittedBy string `json:"submitted_by" validate:"required"`
}

// ========================================================================
// PROFILE UPDATE API REQUESTS (AGT-022 to AGT-028)
// ========================================================================

// SearchAgentsRequest performs multi-criteria agent search
// AGT-022: Search Agents
// FR-AGT-PRF-004: Multi-criteria agent search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
type SearchAgentsRequest struct {
	AgentID      *string `query:"agent_id"`
	Name         *string `query:"name"`
	PANNumber    *string `query:"pan_number"`
	MobileNumber *string `query:"mobile_number"`
	Email        *string `query:"email"`
	Status       *string `query:"status"`
	OfficeCode   *string `query:"office_code"`
	Page         int     `query:"page" validate:"omitempty,min=1" default:"1"`
	Limit        int     `query:"limit" validate:"omitempty,min=1,max=100" default:"20"`
}

// UpdateSectionRequest updates a profile section
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
type UpdateSectionRequest struct {
	AgentID   string                 `uri:"agent_id" validate:"required"`
	Section   string                 `uri:"section" validate:"required,oneof=personal_info address contact email"`
	Updates   map[string]interface{} `json:"updates" validate:"required"`
	UpdatedBy string                 `json:"updated_by" validate:"required"`
	Reason    string                 `json:"reason,omitempty"`
}

// ApprovalRequest approves or rejects a profile update
// AGT-026: Approve Profile Update
// AGT-027: Reject Profile Update
type ApprovalRequest struct {
	ApprovalRequestID string `uri:"approval_request_id" validate:"required,uuid4"`
	Action            string `json:"action" validate:"required,oneof=APPROVE REJECT"`
	Comments          string `json:"comments" validate:"required,min=10"`
	ApprovedBy        string `json:"approved_by,omitempty"`
	RejectedBy        string `json:"rejected_by,omitempty"`
}

// AuditHistoryRequest retrieves audit history with filters
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Profile Change History and Audit Trail
type AuditHistoryRequest struct {
	AgentID  string  `uri:"agent_id" validate:"required"`
	FromDate *string `query:"from_date"`
	ToDate   *string `query:"to_date"`
	Page     int     `query:"page" validate:"omitempty,min=1" default:"1"`
	Limit    int     `query:"limit" validate:"omitempty,min=1,max=100" default:"50"`
}

// GetAuditHistoryRequest retrieves audit history with filters (Phase 9)
// AGT-028: Get Audit History
type GetAuditHistoryRequest struct {
	FromDate *time.Time `query:"from_date"`
	ToDate   *time.Time `query:"to_date"`
	Page     *int       `query:"page"`
	Limit    *int       `query:"limit"`
}

// GetTimelineRequest retrieves agent activity timeline
// AGT-076: Agent Activity Timeline
// Phase 9: Search & Dashboard APIs
type GetTimelineRequest struct {
	ActivityType *string    `query:"activity_type"` // PROFILE_CHANGE, LICENSE_UPDATE, STATUS_CHANGE
	FromDate     *time.Time `query:"from_date"`
	ToDate       *time.Time `query:"to_date"`
	Page         *int       `query:"page"`
	Limit        *int       `query:"limit"`
}

// GetNotificationsRequest retrieves agent notification history
// AGT-077: Agent Notification History
// Phase 9: Search & Dashboard APIs
type GetNotificationsRequest struct {
	NotificationType *string    `query:"notification_type"` // EMAIL, SMS, INTERNAL
	FromDate         *time.Time `query:"from_date"`
	ToDate           *time.Time `query:"to_date"`
	Page             *int       `query:"page"`
	Limit            *int       `query:"limit"`
}

// ========================================================================
// PHASE 10: BATCH & WEBHOOK API REQUESTS
// ========================================================================

// ConfigureExportRequest configures export parameters
// AGT-064: Configure Export Parameters
// FR-AGT-PRF-025: Profile Export
type ConfigureExportRequest struct {
	ExportName   string      `json:"export_name" validate:"required,min=3,max=255"`
	Filters      interface{} `json:"filters" validate:"required"`      // ExportFilters struct
	Fields       []string    `json:"fields" validate:"required,min=1"` // Fields to export
	OutputFormat string      `json:"output_format" validate:"required,oneof=EXCEL PDF CSV"`
	CreatedBy    string      `json:"created_by" validate:"required"`
}

// ExecuteExportRequest executes export asynchronously
// AGT-065: Execute Export Asynchronously
type ExecuteExportRequest struct {
	ExportConfigID string `json:"export_config_id" validate:"required,uuid4"`
	RequestedBy    string `json:"requested_by" validate:"required"`
}

// HRMSWebhookRequest receives HRMS employee updates
// AGT-078: HRMS Webhook Receiver
// INT-AGT-001: HRMS System Integration
type HRMSWebhookRequest struct {
	EventID      string               `json:"event_id" validate:"required,uuid4"`
	EventType    string               `json:"event_type" validate:"required,oneof=EMPLOYEE_CREATED EMPLOYEE_UPDATED EMPLOYEE_TRANSFERRED EMPLOYEE_TERMINATED"`
	Timestamp    time.Time            `json:"timestamp" validate:"required"`
	Signature    string               `json:"signature" validate:"required"`
	EmployeeData HRMSEmployeeDataReq  `json:"employee_data" validate:"required"`
}

// HRMSEmployeeDataReq represents employee data in webhook
type HRMSEmployeeDataReq struct {
	EmployeeID   string `json:"employee_id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Designation  string `json:"designation"`
	OfficeCode   string `json:"office_code"`
	Status       string `json:"status"`
	EmailAddress string `json:"email_address,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`
}

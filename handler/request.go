package handler

import "database/sql"

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

// ==================== PHASE 6: PROFILE UPDATE APIs (AGT-022 to AGT-028) ====================

// SearchAgentsQuery query params for agent search
// AGT-022: Multi-criteria Agent Search
// FR-AGT-PRF-004: Agent Search Functionality
// BR-AGT-PRF-022: Multi-Criteria Search Support
type SearchAgentsQuery struct {
	AgentID      string `query:"agent_id" validate:"omitempty,uuid4"`
	Name         string `query:"name" validate:"omitempty"`
	PANNumber    string `query:"pan_number" validate:"omitempty,len=10"`
	MobileNumber string `query:"mobile_number" validate:"omitempty,len=10"`
	Status       string `query:"status" validate:"omitempty,oneof=ACTIVE SUSPENDED TERMINATED DEACTIVATED"`
	OfficeCode   string `query:"office_code" validate:"omitempty"`
	Page         int    `query:"page" validate:"omitempty,min=1"`
	Limit        int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// AgentIDURI uri param for agent ID
// AGT-023: Get Agent Profile
type AgentIDURI struct {
	AgentID string `uri:"agent_id" validate:"required,uuid4"`
}

// GetUpdateFormRequest request for getting update form
// AGT-024: Get Profile Update Form
// FR-AGT-PRF-006: Personal Information Update
type GetUpdateFormRequest struct {
	AgentID string `uri:"agent_id" validate:"required,uuid4"`
	Section string `query:"section" validate:"required,oneof=personal_info address contact license bank_details status"`
}

// UpdateProfileSectionRequest request for updating profile section
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
type UpdateProfileSectionRequest struct {
	AgentID      string                 `uri:"agent_id" validate:"required,uuid4"`
	Section      string                 `uri:"section" validate:"required,oneof=personal_info address contact license bank_details status"`
	Changes      map[string]interface{} `json:"changes" validate:"required"`
	UpdateReason string                 `json:"update_reason" validate:"omitempty,min=10"`
	UpdatedBy    string                 `json:"updated_by" validate:"required"`
}

// ApprovalActionRequest request for approve/reject actions
// AGT-026: Approve Profile Update
// AGT-027: Reject Profile Update
type ApprovalActionRequest struct {
	ApprovalRequestID string  `uri:"approval_request_id" validate:"required,uuid4"`
	ReviewedBy        string  `json:"reviewed_by" validate:"required"`
	ReviewComments    *string `json:"review_comments" validate:"omitempty"`
}

// GetAuditHistoryRequest request for fetching audit history
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Audit History Tracking
type GetAuditHistoryRequest struct {
	AgentID string `uri:"agent_id" validate:"required,uuid4"`
	Page    int    `query:"page" validate:"omitempty,min=1"`
	Limit   int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// ========================================================================
// LICENSE MANAGEMENT API REQUESTS (AGT-029 to AGT-038)
// ========================================================================

// LicenseIDUri represents license ID in URI
type LicenseIDUri struct {
	LicenseID string `uri:"license_id" validate:"required,uuid4"`
}

// GetAgentLicensesQuery filters for agent licenses list
// AGT-029: Get Agent Licenses
type GetAgentLicensesQuery struct {
	Status string `query:"status" validate:"omitempty,oneof=ACTIVE EXPIRED RENEWED"`
}

// AddLicenseRequest represents request to add new license
// AGT-030: Add License
// BR-AGT-PRF-012: License Renewal Period Rules
// VR-AGT-PRF-031 to VR-AGT-PRF-036
type AddLicenseRequest struct {
	AgentID                      string  `uri:"agent_id" validate:"required,uuid4"`
	LicenseLine                  string  `json:"license_line" validate:"required,oneof=Life General"`
	LicenseType                  string  `json:"license_type" validate:"required,oneof=Provisional Permanent"`
	LicenseNumber                string  `json:"license_number" validate:"required"`
	ResidentStatus               string  `json:"resident_status" validate:"required,oneof=Resident Non_Resident"`
	LicenseDate                  string  `json:"license_date" validate:"required"` // format: date
	AuthorityDate                string  `json:"authority_date" validate:"required"` // format: date
	LicentiateExamPassed         bool    `json:"licentiate_exam_passed"`
	IsPrimary                    bool    `json:"is_primary"`
}

// UpdateLicenseRequest represents request to update license
// AGT-032: Update License
type UpdateLicenseRequest struct {
	AgentID                      string  `uri:"agent_id" validate:"required,uuid4"`
	LicenseID                    string  `uri:"license_id" validate:"required,uuid4"`
	LicenseNumber                *string `json:"license_number" validate:"omitempty"`
	ResidentStatus               *string `json:"resident_status" validate:"omitempty,oneof=Resident Non_Resident"`
	AuthorityDate                *string `json:"authority_date" validate:"omitempty"` // format: date
	IsPrimary                    *bool   `json:"is_primary" validate:"omitempty"`
	UpdatedBy                    string  `json:"updated_by" validate:"required"`
}

// RenewLicenseRequest represents request to renew license
// AGT-033: Renew License
// BR-AGT-PRF-012: Complex renewal rules
type RenewLicenseRequest struct {
	AgentID               string  `uri:"agent_id" validate:"required,uuid4"`
	LicenseID             string  `uri:"license_id" validate:"required,uuid4"`
	RenewalType           string  `json:"renewal_type" validate:"required,oneof=PROVISIONAL POST_EXAM ANNUAL"`
	LicentiateExamPassed  bool    `json:"licentiate_exam_passed"`
	ExamDate              *string `json:"exam_date" validate:"omitempty"` // format: date
	ExamCertificateNumber *string `json:"exam_certificate_number" validate:"omitempty"`
	RenewedBy             string  `json:"renewed_by" validate:"required"`
}

// GetExpiringLicensesQuery filters for expiring licenses
// AGT-036: Get Expiring Licenses
// BR-AGT-PRF-014: License Renewal Reminders
type GetExpiringLicensesQuery struct {
	Days       int    `query:"days" validate:"omitempty,min=1,max=365"`
	OfficeCode string `query:"office_code" validate:"omitempty"`
	Page       int    `query:"page" validate:"omitempty,min=1"`
	Limit      int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// TriggerExpiryDeactivationRequest triggers batch job to deactivate expired licenses
// AGT-038: Trigger License Expiry Deactivation
// BR-AGT-PRF-013: Auto-Deactivation on Expiry
type TriggerExpiryDeactivationRequest struct {
	BatchDate string `json:"batch_date" validate:"required"` // format: date
	DryRun    bool   `json:"dry_run"`
}

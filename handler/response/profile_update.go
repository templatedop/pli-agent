package response

import (
	"time"

	"pli-agent-api/core/port"
)

// ========================================================================
// PROFILE UPDATE API RESPONSES (AGT-022 to AGT-028)
// ========================================================================

// SearchAgentsResponse returns paginated search results
// AGT-022: Search Agents
// FR-AGT-PRF-004: Multi-criteria agent search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
type SearchAgentsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Results                   []AgentSearchResult `json:"results"`
	Pagination                PaginationMetadata  `json:"pagination"`
}

// AgentSearchResult represents a single search result
type AgentSearchResult struct {
	AgentID     string `json:"agent_id"`
	Name        string `json:"name"`
	AgentType   string `json:"agent_type"`
	PAN         string `json:"pan"`
	Mobile      string `json:"mobile,omitempty"`
	Email       string `json:"email,omitempty"`
	Status      string `json:"status"`
	Coordinator string `json:"advisor_coordinator,omitempty"`
	Office      string `json:"office,omitempty"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	CurrentPage    int `json:"current_page"`
	TotalPages     int `json:"total_pages"`
	TotalResults   int `json:"total_results"`
	ResultsPerPage int `json:"results_per_page"`
}

// AgentProfileResponse returns complete agent profile
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
type AgentProfileResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentProfile              AgentProfileDTO   `json:"agent_profile"`
	WorkflowState             *WorkflowStateDTO `json:"workflow_state,omitempty"`
}

// AgentProfileDTO represents complete agent profile data
type AgentProfileDTO struct {
	AgentID      string          `json:"agent_id"`
	ProfileType  string          `json:"profile_type"`
	FullName     string          `json:"full_name"`
	PANNumber    string          `json:"pan_number"`
	Status       string          `json:"status"`
	PersonalInfo PersonalInfoDTO `json:"personal_info"`
	Addresses    []AddressDTO    `json:"addresses"`
	Contacts     []ContactDTO    `json:"contacts"`
	Emails       []EmailDTO      `json:"emails"`
	Office       *OfficeDTO      `json:"office,omitempty"`
}

// PersonalInfoDTO represents personal information
type PersonalInfoDTO struct {
	FirstName     string     `json:"first_name"`
	MiddleName    string     `json:"middle_name,omitempty"`
	LastName      string     `json:"last_name"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	Gender        string     `json:"gender,omitempty"`
	AadharNumber  string     `json:"aadhar_number,omitempty"`
	MaritalStatus string     `json:"marital_status,omitempty"`
	Category      string     `json:"category,omitempty"`
	Title         string     `json:"title,omitempty"`
}

// AddressDTO represents an address
type AddressDTO struct {
	AddressID   string `json:"address_id"`
	AddressType string `json:"address_type"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2,omitempty"`
	Line3       string `json:"line3,omitempty"`
	City        string `json:"city"`
	District    string `json:"district,omitempty"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Pincode     string `json:"pincode"`
	IsPrimary   bool   `json:"is_primary"`
}

// ContactDTO represents a contact
type ContactDTO struct {
	ContactID       string `json:"contact_id"`
	ContactType     string `json:"contact_type"`
	MobileNumber    string `json:"mobile_number"`
	AlternateNumber string `json:"alternate_number,omitempty"`
	IsPrimary       bool   `json:"is_primary"`
	IsVerified      bool   `json:"is_verified"`
}

// EmailDTO represents an email
type EmailDTO struct {
	EmailID      string `json:"email_id"`
	EmailType    string `json:"email_type"`
	EmailAddress string `json:"email_address"`
	IsPrimary    bool   `json:"is_primary"`
	IsVerified   bool   `json:"is_verified"`
}

// OfficeDTO represents office information
type OfficeDTO struct {
	OfficeCode string `json:"office_code"`
	OfficeName string `json:"office_name"`
	OfficeType string `json:"office_type,omitempty"`
}

// WorkflowStateDTO represents workflow state
type WorkflowStateDTO struct {
	CurrentStep        string `json:"current_step"`
	NextStep           string `json:"next_step,omitempty"`
	ProgressPercentage int    `json:"progress_percentage"`
}

// UpdateFormResponse returns pre-populated update form
// AGT-024: Get Update Form
type UpdateFormResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentID                   string                   `json:"agent_id"`
	Sections                  []SectionDTO             `json:"sections"`
	CurrentData               map[string]interface{}   `json:"current_data"`
	EditableFields            map[string]FieldMetadata `json:"editable_fields"`
}

// SectionDTO represents an editable section
type SectionDTO struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Fields      []string `json:"fields"`
}

// FieldMetadata contains field editing metadata
type FieldMetadata struct {
	Name             string                   `json:"name"`
	DisplayName      string                   `json:"display_name"`
	Type             string                   `json:"type"`
	Required         bool                     `json:"required"`
	Editable         bool                     `json:"editable"`
	RequiresApproval bool                     `json:"requires_approval"`
	ValidationRules  map[string]interface{}   `json:"validation_rules,omitempty"`
	SelectOptions    []map[string]interface{} `json:"select_options,omitempty"`
	Placeholder      string                   `json:"placeholder,omitempty"`
	HelpText         string                   `json:"help_text,omitempty"`
}

// UpdateSectionResponse returns update result
// AGT-025: Update Profile Section
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
type UpdateSectionResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentID                   string                `json:"agent_id"`
	Section                   string                `json:"section"`
	Status                    string                `json:"status"` // "UPDATED", "PENDING_APPROVAL"
	UpdatedFields             []string              `json:"updated_fields"`
	ApprovalRequired          bool                  `json:"approval_required"`
	ApprovalRequestID         *string               `json:"approval_request_id,omitempty"`
	UpdatedProfile            *AgentProfileDTO      `json:"updated_profile,omitempty"`
	ChangedFields             map[string]ChangeInfo `json:"changed_fields"`
}

// ChangeInfo represents a field change
type ChangeInfo struct {
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// ApprovalResponse returns approval/rejection result
// AGT-026: Approve Profile Update
// AGT-027: Reject Profile Update
type ApprovalResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	ApprovalRequestID         string    `json:"approval_request_id"`
	Status                    string    `json:"status"` // "APPROVED", "REJECTED"
	AgentID                   string    `json:"agent_id"`
	ApprovedBy                string    `json:"approved_by,omitempty"`
	RejectedBy                string    `json:"rejected_by,omitempty"`
	ProcessedAt               time.Time `json:"processed_at"`
	Message                   string    `json:"message"`
}

// AuditHistoryResponse returns paginated audit history
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Profile Change History and Audit Trail
type AuditHistoryResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	AgentID                   string             `json:"agent_id"`
	AuditLogs                 []AuditLogDTO      `json:"audit_logs"`
	Pagination                PaginationMetadata `json:"pagination"`
}

// AuditLogDTO represents a single audit log entry
type AuditLogDTO struct {
	AuditID     string    `json:"audit_id"`
	Action      string    `json:"action"`
	FieldName   string    `json:"field_name"`
	OldValue    string    `json:"old_value,omitempty"`
	NewValue    string    `json:"new_value,omitempty"`
	PerformedBy string    `json:"performed_by"`
	PerformedAt time.Time `json:"performed_at"`
	IPAddress   string    `json:"ip_address,omitempty"`
}

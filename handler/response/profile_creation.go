package response

import (
	"time"

	"pli-agent-api/core/port"
)

// ========================================================================
// PROFILE CREATION API RESPONSES (AGT-001 to AGT-006)
// ========================================================================

// InitiateProfileResponse returns session initiation result
// AGT-001: Initiate Agent Profile Creation
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-031: Workflow Orchestration
// WF-002: Agent Onboarding Workflow
type InitiateProfileResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	SessionID                 string         `json:"session_id"`
	AgentType                 string         `json:"agent_type"`
	WorkflowState             *WorkflowState `json:"workflow_state"`
	ExpiresAt                 *time.Time     `json:"expires_at"`
}

// FetchHRMSResponse returns HRMS data fetch result
// AGT-002: Fetch HRMS Employee Data
// FR-AGT-PRF-002: HRMS Data Auto-Population
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
// INT-AGT-001: HRMS Integration
type FetchHRMSResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	EmployeeData              *EmployeeData `json:"employee_data"`
	AutoPopulated             bool          `json:"auto_populated"`
	Message                   string        `json:"message"`
}

// AdvisorCoordinator represents an advisor coordinator in the list
type AdvisorCoordinator struct {
	AgentID    string `json:"agent_id"`
	AgentCode  string `json:"agent_code"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	CircleID   string `json:"circle_id,omitempty"`
	DivisionID string `json:"division_id,omitempty"`
}

// AdvisorCoordinatorsResponse returns list of advisor coordinators
// AGT-003: Get Advisor Coordinators List
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
type AdvisorCoordinatorsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []AdvisorCoordinator `json:"data"`
}

// LinkCoordinatorResponse returns coordinator linkage result
// AGT-004: Link Advisor to Coordinator
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
type LinkCoordinatorResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	CoordinatorLinked         bool   `json:"coordinator_linked"`
	CoordinatorName           string `json:"coordinator_name"`
	Message                   string `json:"message"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidateProfileResponse returns profile validation result
// AGT-005: Validate Profile Details
// FR-AGT-PRF-002: Profile Data Validation
// VR-AGT-PRF-002 to VR-AGT-PRF-030: All validation rules
type ValidateProfileResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	IsValid                   bool              `json:"is_valid"`
	ValidationErrors          []ValidationError `json:"validation_errors,omitempty"`
	Message                   string            `json:"message"`
}

// SubmitProfileResponse returns profile submission result
// AGT-006: Submit Profile for Creation
// FR-AGT-PRF-001: New Profile Creation
// WF-002: Agent Onboarding Workflow
type SubmitProfileResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	SubmissionID              string `json:"submission_id"`
	Status                    string `json:"status"` // PROCESSING, PENDING_APPROVAL, COMPLETED
	Message                   string `json:"message"`
	WorkflowID                string `json:"workflow_id"`
}

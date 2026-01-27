package domain

import (
	"database/sql"
	"time"
)

// AgentProfile represents the main agent profile entity
// E-01: Agent Profile Entity
// BR-AGT-PRF-001 to BR-AGT-PRF-030
type AgentProfile struct {
	// Primary Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Profile Identification
	AgentCode  sql.NullString `json:"agent_code" db:"agent_code"`
	AgentType  string         `json:"agent_type" db:"agent_type"` // BR-AGT-PRF-001 to BR-AGT-PRF-004
	EmployeeID sql.NullString `json:"employee_id" db:"employee_id"`

	// Office and Hierarchy (BR-AGT-PRF-002)
	OfficeCode           string         `json:"office_code" db:"office_code"`
	CircleID             sql.NullString `json:"circle_id" db:"circle_id"`
	DivisionID           sql.NullString `json:"division_id" db:"division_id"`
	AdvisorCoordinatorID sql.NullString `json:"advisor_coordinator_id" db:"advisor_coordinator_id"`

	// Personal Information (VR-AGT-PRF-001 to VR-AGT-PRF-007)
	Title         sql.NullString `json:"title" db:"title"`
	FirstName     string         `json:"first_name" db:"first_name"`
	MiddleName    sql.NullString `json:"middle_name" db:"middle_name"`
	LastName      string         `json:"last_name" db:"last_name"`
	Gender        string         `json:"gender" db:"gender"`
	DateOfBirth   time.Time      `json:"date_of_birth" db:"date_of_birth"`
	Category      sql.NullString `json:"category" db:"category"`
	MaritalStatus sql.NullString `json:"marital_status" db:"marital_status"`

	// Identification Numbers (VR-AGT-PRF-003, VR-AGT-PRF-004)
	AadharNumber sql.NullString `json:"aadhar_number" db:"aadhar_number"` // VR-AGT-PRF-004
	PANNumber    string         `json:"pan_number" db:"pan_number"`       // BR-AGT-PRF-006, VR-AGT-PRF-003

	// Professional Information
	DesignationRank   sql.NullString `json:"designation_rank" db:"designation_rank"`
	ServiceNumber     sql.NullString `json:"service_number" db:"service_number"`
	ProfessionalTitle sql.NullString `json:"professional_title" db:"professional_title"`

	// Status Management (BR-AGT-PRF-016, BR-AGT-PRF-017)
	Status       string         `json:"status" db:"status"` // BR-AGT-PRF-016
	StatusDate   time.Time      `json:"status_date" db:"status_date"`
	StatusReason sql.NullString `json:"status_reason" db:"status_reason"` // BR-AGT-PRF-016: Mandatory for SUSPENDED/TERMINATED/DEACTIVATED

	// Distribution Channel and Product Authorization (BR-AGT-PRF-026)
	DistributionChannel          sql.NullString `json:"distribution_channel" db:"distribution_channel"`
	ProductClass                 sql.NullString `json:"product_class" db:"product_class"`                                   // BR-AGT-PRF-026
	ExternalIdentificationNumber sql.NullString `json:"external_identification_number" db:"external_identification_number"` // BR-AGT-PRF-027

	// Goals and Performance (BR-AGT-PRF-024)
	Goals sql.NullString `json:"goals" db:"goals"` // JSONB stored as string

	// Workflow State Management (WF-AGT-PRF-001 to WF-AGT-PRF-012)
	WorkflowState        sql.NullString `json:"workflow_state" db:"workflow_state"`
	WorkflowStateHistory sql.NullString `json:"workflow_state_history" db:"workflow_state_history"` // JSONB stored as string

	// Metadata and Search
	Metadata     sql.NullString `json:"metadata" db:"metadata"` // JSONB stored as string
	SearchVector sql.NullString `json:"search_vector" db:"search_vector"`

	// Audit Fields (BR-AGT-PRF-005, BR-AGT-PRF-006)
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at" db:"updated_at"`
	CreatedBy string         `json:"created_by" db:"created_by"`
	UpdatedBy sql.NullString `json:"updated_by" db:"updated_by"`
	DeletedAt sql.NullTime   `json:"deleted_at" db:"deleted_at"` // Soft delete
	Version   int            `json:"version" db:"version"`       // Optimistic locking
}

// AgentType constants (BR-AGT-PRF-001 to BR-AGT-PRF-004)
const (
	AgentTypeAdvisor              = "ADVISOR"
	AgentTypeAdvisorCoordinator   = "ADVISOR_COORDINATOR"
	AgentTypeDepartmentalEmployee = "DEPARTMENTAL_EMPLOYEE"
	AgentTypeFieldOfficer         = "FIELD_OFFICER"
)

// AgentStatus constants (BR-AGT-PRF-016, BR-AGT-PRF-017)
const (
	AgentStatusActive      = "ACTIVE"
	AgentStatusSuspended   = "SUSPENDED"
	AgentStatusTerminated  = "TERMINATED"
	AgentStatusDeactivated = "DEACTIVATED"
	AgentStatusExpired     = "EXPIRED"
)

// Gender constants (VR-AGT-PRF-005)
const (
	GenderMale   = "Male"
	GenderFemale = "Female"
	GenderOther  = "Other"
)

// MaritalStatus constants (VR-AGT-PRF-006)
const (
	MaritalStatusSingle   = "Single"
	MaritalStatusMarried  = "Married"
	MaritalStatusWidowed  = "Widowed"
	MaritalStatusDivorced = "Divorced"
)

// GetFullName returns the full name of the agent
func (a *AgentProfile) GetFullName() string {
	fullName := a.FirstName
	if a.MiddleName.Valid {
		fullName += " " + a.MiddleName.String
	}
	fullName += " " + a.LastName
	return fullName
}

// IsActive checks if the agent is active
// BR-AGT-PRF-016: Status Management
func (a *AgentProfile) IsActive() bool {
	return a.Status == AgentStatusActive
}

// RequiresAdvisorCoordinator checks if this agent type requires a coordinator link
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
func (a *AgentProfile) RequiresAdvisorCoordinator() bool {
	return a.AgentType == AgentTypeAdvisor
}

// IsGeographicAssignmentRequired checks if geographic assignment is required
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
func (a *AgentProfile) IsGeographicAssignmentRequired() bool {
	return a.AgentType == AgentTypeAdvisorCoordinator
}

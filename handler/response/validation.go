package response

import "pli-agent-api/core/port"

// ========================================================================
// VALIDATION API RESPONSES (AGT-012 to AGT-015)
// ========================================================================

// ExistingAgentInfo represents existing agent information when PAN is not unique
type ExistingAgentInfo struct {
	AgentID   string `json:"agent_id"`
	AgentCode string `json:"agent_code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}

// PANUniquenessResponse returns PAN uniqueness check result
// AGT-012: Check PAN Uniqueness
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// VR-AGT-PRF-002: PAN Uniqueness
type PANUniquenessResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	IsUnique                  bool               `json:"is_unique"`
	Message                   string             `json:"message"`
	ExistingAgent             *ExistingAgentInfo `json:"existing_agent,omitempty"`
}

// EmployeeData represents employee data from HRMS
type EmployeeData struct {
	EmployeeID   string `json:"employee_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	Designation  string `json:"designation"`
	OfficeCode   string `json:"office_code"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Status       string `json:"status"`
}

// EmployeeIDValidationResponse returns employee ID validation result
// AGT-013: Validate Employee ID (HRMS)
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
// VR-AGT-PRF-023: HRMS Employee ID Validation
// INT-AGT-001: HRMS Integration
type EmployeeIDValidationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	IsValid                   bool          `json:"is_valid"`
	EmployeeStatus            string        `json:"employee_status,omitempty"`
	EmployeeData              *EmployeeData `json:"employee_data,omitempty"`
}

// BankDetails represents bank details for IFSC validation
type BankDetails struct {
	IFSCCode   string `json:"ifsc_code"`
	BankName   string `json:"bank_name"`
	BranchName string `json:"branch_name"`
	City       string `json:"city"`
	State      string `json:"state"`
}

// IFSCValidationResponse returns IFSC validation result
// AGT-014: Validate IFSC Code
// BR-AGT-PRF-018: Bank Account Details for Commission Disbursement
// VR-AGT-PRF-017: IFSC Code Format Validation
type IFSCValidationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	IsValid                   bool         `json:"is_valid"`
	BankDetails               *BankDetails `json:"bank_details,omitempty"`
}

// OfficeDetails represents office details
type OfficeDetails struct {
	OfficeCode string `json:"office_code"`
	OfficeName string `json:"office_name"`
	OfficeType string `json:"office_type"`
	CircleID   string `json:"circle_id"`
	DivisionID string `json:"division_id"`
	State      string `json:"state"`
	City       string `json:"city"`
}

// OfficeValidationResponse returns office code validation result
// AGT-015: Validate Office Code
// BR-AGT-PRF-034: Office Association
// VR-AGT-PRF-027: Valid Office Code
type OfficeValidationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	IsValid                   bool           `json:"is_valid"`
	OfficeDetails             *OfficeDetails `json:"office_details,omitempty"`
}

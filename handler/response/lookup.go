package response

import "pli-agent-api/core/port"

// ========================================================================
// LOOKUP API RESPONSES (AGT-007 to AGT-011)
// ========================================================================

// AgentType represents an agent type lookup item
type AgentType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

// AgentTypesResponse returns list of agent types
// AGT-007: Fetch Agent Types
// VR-AGT-PRF-025: Profile Type Valid
type AgentTypesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []AgentType `json:"data"`
}

// Category represents a category lookup item
type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CategoriesResponse returns list of categories
// AGT-009: Get Categories
// VR-AGT-PRF-033: Category Validation
type CategoriesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []Category `json:"data"`
}

// Designation represents a designation lookup item
type Designation struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// DesignationsResponse returns list of designations
// AGT-010: Get Designations
// VR-AGT-PRF-034: Designation Validation
type DesignationsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []Designation `json:"data"`
}

// OfficeType represents an office type lookup item
type OfficeType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// OfficeTypesResponse returns list of office types
// AGT-011: Get Office Types
// BR-AGT-PRF-034: Office Association
type OfficeTypesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []OfficeType `json:"data"`
}

// State represents a state lookup item
type State struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// StatesResponse returns list of states
// AGT-011b: Get States
// VR-AGT-PRF-030: State Validation
type StatesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []State `json:"data"`
}

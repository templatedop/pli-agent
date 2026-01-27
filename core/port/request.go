package port

// MetadataRequest provides common pagination and sorting parameters
// Embed this in list/search request structs
//
// FR-AGT-PRF-021: Multi-Criteria Agent Search
// VR-AGT-PRF-034: Pagination Support
type MetadataRequest struct {
	Skip     uint64 `form:"skip,default=0" validate:"omitempty"`
	Limit    uint64 `form:"limit,default=10" validate:"omitempty,max=100"`
	OrderBy  string `form:"orderBy" validate:"omitempty"`
	SortType string `form:"sortType" validate:"omitempty,oneof=ASC DESC asc desc"`
}

// FilterRequest provides common filtering parameters for agent searches
// FR-AGT-PRF-021: Multi-Criteria Agent Search
type FilterRequest struct {
	Status     string `form:"status" validate:"omitempty"`
	AgentType  string `form:"agent_type" validate:"omitempty"`
	CircleID   string `form:"circle_id" validate:"omitempty"`
	DivisionID string `form:"division_id" validate:"omitempty"`
}

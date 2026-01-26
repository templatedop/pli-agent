package handler

import (
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentLookupHandler handles all lookup APIs
// AGT-007 to AGT-011: Lookup APIs
type AgentLookupHandler struct {
	*serverHandler.Base
}

// NewAgentLookupHandler creates a new lookup handler
func NewAgentLookupHandler() *AgentLookupHandler {
	base := serverHandler.New("Agent Lookup APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentLookupHandler{Base: base}
}

// Routes defines all lookup API routes
func (h *AgentLookupHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/agents/lookup/agent-types", h.FetchAgentTypes).Name("Fetch Agent Types"),
		serverRoute.GET("/agent-types", h.GetAgentTypes).Name("Get Agent Types Dropdown"),
		serverRoute.GET("/categories", h.GetCategories).Name("Get Categories"),
		serverRoute.GET("/designations", h.GetDesignations).Name("Get Designations"),
		serverRoute.GET("/office-types", h.GetOfficeTypes).Name("Get Office Types"),
		serverRoute.GET("/states", h.GetStates).Name("Get States"),
	}
}

// FetchAgentTypes returns list of agent types
// AGT-007: Fetch Agent Types
// VR-AGT-PRF-025: Profile Type Valid
func (h *AgentLookupHandler) FetchAgentTypes(sctx *serverRoute.Context, _ struct{}) (*resp.AgentTypesResponse, error) {
	log.Info(sctx.Ctx, "Fetching agent types")

	// Static lookup data for agent types
	// BR-AGT-PRF-001 to BR-AGT-PRF-004: Agent Type Business Rules
	agentTypes := []resp.AgentType{
		{
			Code:        "ADVISOR",
			Name:        "Advisor",
			Description: "Insurance advisors who work under coordinator supervision",
			Active:      true,
		},
		{
			Code:        "ADVISOR_COORDINATOR",
			Name:        "Advisor Coordinator",
			Description: "Coordinators who manage multiple advisors in assigned geographic areas",
			Active:      true,
		},
		{
			Code:        "DEPARTMENTAL_EMPLOYEE",
			Name:        "Departmental Employee",
			Description: "Postal department employees selling insurance products",
			Active:      true,
		},
		{
			Code:        "FIELD_OFFICER",
			Name:        "Field Officer",
			Description: "Field officers managing operations in assigned regions",
			Active:      true,
		},
		{
			Code:        "DIRECT_AGENT",
			Name:        "Direct Agent",
			Description: "Direct agents working independently",
			Active:      true,
		},
		{
			Code:        "GDS",
			Name:        "Gramin Dak Sevak",
			Description: "Rural postal service agents",
			Active:      true,
		},
	}

	return &resp.AgentTypesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 agentTypes,
	}, nil
}

// GetAgentTypes returns agent types dropdown
// AGT-008: Get Agent Types
func (h *AgentLookupHandler) GetAgentTypes(sctx *serverRoute.Context, _ struct{}) (*resp.AgentTypesResponse, error) {
	// Reuse FetchAgentTypes logic
	return h.FetchAgentTypes(sctx, struct{}{})
}

// GetCategories returns categories dropdown
// AGT-009: Get Categories
// VR-AGT-PRF-033: Category Validation
func (h *AgentLookupHandler) GetCategories(sctx *serverRoute.Context, _ struct{}) (*resp.CategoriesResponse, error) {
	log.Info(sctx.Ctx, "Fetching categories")

	// Static lookup data for categories
	categories := []resp.Category{
		{Code: "GEN", Name: "General"},
		{Code: "SC", Name: "Scheduled Caste"},
		{Code: "ST", Name: "Scheduled Tribe"},
		{Code: "OBC", Name: "Other Backward Class"},
		{Code: "EWS", Name: "Economically Weaker Section"},
	}

	return &resp.CategoriesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 categories,
	}, nil
}

// GetDesignations returns designations dropdown
// AGT-010: Get Designations
// VR-AGT-PRF-034: Designation Validation
func (h *AgentLookupHandler) GetDesignations(sctx *serverRoute.Context, _ struct{}) (*resp.DesignationsResponse, error) {
	log.Info(sctx.Ctx, "Fetching designations")

	// Static lookup data for designations
	designations := []resp.Designation{
		{Code: "POSTMASTER", Name: "Postmaster"},
		{Code: "ASST_POSTMASTER", Name: "Assistant Postmaster"},
		{Code: "POST_OFFICE_INCHARGE", Name: "Post Office In-Charge"},
		{Code: "INSPECTOR", Name: "Inspector"},
		{Code: "SUPERINTENDENT", Name: "Superintendent"},
		{Code: "BRANCH_POSTMASTER", Name: "Branch Postmaster"},
		{Code: "SUB_POSTMASTER", Name: "Sub Postmaster"},
		{Code: "SORTING_ASSISTANT", Name: "Sorting Assistant"},
		{Code: "POSTAL_ASSISTANT", Name: "Postal Assistant"},
		{Code: "MULTI_TASKING_STAFF", Name: "Multi Tasking Staff"},
	}

	return &resp.DesignationsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 designations,
	}, nil
}

// GetOfficeTypes returns office types dropdown
// AGT-011: Get Office Types
// BR-AGT-PRF-034: Office Association
func (h *AgentLookupHandler) GetOfficeTypes(sctx *serverRoute.Context, _ struct{}) (*resp.OfficeTypesResponse, error) {
	log.Info(sctx.Ctx, "Fetching office types")

	// Static lookup data for office types
	officeTypes := []resp.OfficeType{
		{Code: "HEAD_OFFICE", Name: "Head Office"},
		{Code: "SUB_OFFICE", Name: "Sub Office"},
		{Code: "BRANCH_OFFICE", Name: "Branch Office"},
		{Code: "POST_OFFICE", Name: "Post Office"},
		{Code: "SUB_POST_OFFICE", Name: "Sub Post Office"},
		{Code: "EXTRA_DEPARTMENTAL", Name: "Extra Departmental Office"},
	}

	return &resp.OfficeTypesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 officeTypes,
	}, nil
}

// GetStates returns states dropdown
// AGT-011b: Get States
// VR-AGT-PRF-030: State Validation
func (h *AgentLookupHandler) GetStates(sctx *serverRoute.Context, _ struct{}) (*resp.StatesResponse, error) {
	log.Info(sctx.Ctx, "Fetching states")

	// Static lookup data for Indian states and UTs
	states := []resp.State{
		{Code: "AN", Name: "Andaman and Nicobar Islands"},
		{Code: "AP", Name: "Andhra Pradesh"},
		{Code: "AR", Name: "Arunachal Pradesh"},
		{Code: "AS", Name: "Assam"},
		{Code: "BR", Name: "Bihar"},
		{Code: "CH", Name: "Chandigarh"},
		{Code: "CT", Name: "Chhattisgarh"},
		{Code: "DN", Name: "Dadra and Nagar Haveli and Daman and Diu"},
		{Code: "DL", Name: "Delhi"},
		{Code: "GA", Name: "Goa"},
		{Code: "GJ", Name: "Gujarat"},
		{Code: "HR", Name: "Haryana"},
		{Code: "HP", Name: "Himachal Pradesh"},
		{Code: "JK", Name: "Jammu and Kashmir"},
		{Code: "JH", Name: "Jharkhand"},
		{Code: "KA", Name: "Karnataka"},
		{Code: "KL", Name: "Kerala"},
		{Code: "LA", Name: "Ladakh"},
		{Code: "LD", Name: "Lakshadweep"},
		{Code: "MP", Name: "Madhya Pradesh"},
		{Code: "MH", Name: "Maharashtra"},
		{Code: "MN", Name: "Manipur"},
		{Code: "ML", Name: "Meghalaya"},
		{Code: "MZ", Name: "Mizoram"},
		{Code: "NL", Name: "Nagaland"},
		{Code: "OR", Name: "Odisha"},
		{Code: "PY", Name: "Puducherry"},
		{Code: "PB", Name: "Punjab"},
		{Code: "RJ", Name: "Rajasthan"},
		{Code: "SK", Name: "Sikkim"},
		{Code: "TN", Name: "Tamil Nadu"},
		{Code: "TG", Name: "Telangana"},
		{Code: "TR", Name: "Tripura"},
		{Code: "UP", Name: "Uttar Pradesh"},
		{Code: "UT", Name: "Uttarakhand"},
		{Code: "WB", Name: "West Bengal"},
	}

	return &resp.StatesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 states,
	}, nil
}

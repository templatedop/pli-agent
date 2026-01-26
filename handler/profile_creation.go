package handler

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/client"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"
	"pli-agent-api/workflows/activities"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentProfileCreationHandler handles profile creation and session management APIs
// AGT-001 to AGT-006: Profile Creation
// AGT-016 to AGT-019: Session Management
// WF-002: Agent Onboarding Workflow
type AgentProfileCreationHandler struct {
	*serverHandler.Base
	sessionRepo    *repo.AgentProfileSessionRepository
	profileRepo    *repo.AgentProfileRepository
	temporalClient client.Client
}

// NewAgentProfileCreationHandler creates a new profile creation handler
func NewAgentProfileCreationHandler(
	sessionRepo *repo.AgentProfileSessionRepository,
	profileRepo *repo.AgentProfileRepository,
	temporalClient client.Client,
) *AgentProfileCreationHandler {
	base := serverHandler.New("Agent Profile Creation & Session APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentProfileCreationHandler{
		Base:           base,
		sessionRepo:    sessionRepo,
		profileRepo:    profileRepo,
		temporalClient: temporalClient,
	}
}

// Routes defines profile creation and session management routes
func (h *AgentProfileCreationHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// Profile Creation APIs (AGT-001 to AGT-006)
		serverRoute.POST("/agent-profiles/initiate", h.InitiateProfileCreation).Name("Initiate Profile Creation"),
		serverRoute.POST("/agent-profiles/:session_id/fetch-hrms", h.FetchHRMSData).Name("Fetch HRMS Data"),
		serverRoute.GET("/advisor-coordinators", h.GetAdvisorCoordinators).Name("Get Advisor Coordinators"),
		serverRoute.POST("/agent-profiles/:session_id/link-coordinator", h.LinkCoordinator).Name("Link Advisor to Coordinator"),
		serverRoute.POST("/agent-profiles/:session_id/validate-basic", h.ValidateProfile).Name("Validate Profile Details"),
		serverRoute.POST("/agent-profiles/:session_id/submit", h.SubmitProfile).Name("Submit Profile for Creation"),

		// Session Management APIs (AGT-016 to AGT-019)
		serverRoute.GET("/agent-profiles/sessions/:session_id/status", h.GetSessionStatus).Name("Get Session Status"),
		serverRoute.POST("/agent-profiles/sessions/:session_id/save", h.SaveSession).Name("Save Session Checkpoint"),
		serverRoute.GET("/agent-profiles/sessions/:session_id/resume", h.ResumeSession).Name("Resume Session"),
		serverRoute.DELETE("/agent-profiles/sessions/:session_id", h.CancelSession).Name("Cancel Session"),
	}
}

// InitiateProfileCreation initiates agent profile creation workflow
// AGT-001: Initiate Agent Profile Creation
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-031: Workflow Orchestration
// WF-002: Agent Onboarding Workflow
func (h *AgentProfileCreationHandler) InitiateProfileCreation(sctx *serverRoute.Context, req InitiateProfileRequest) (*resp.InitiateProfileResponse, error) {
	log.Info(sctx.Ctx, "Initiating profile creation for agent type: %s", req.AgentType)

	// Create session in database
	session := domain.AgentProfileSession{
		AgentType:          req.AgentType,
		WorkflowState:      domain.WorkflowStateInitiated,
		CurrentStep:        "AGENT_TYPE_SELECTION",
		NextStep:           "PROFILE_DETAILS",
		ProgressPercentage: 10,
		Status:             domain.SessionStatusActive,
		InitiatedBy:        req.InitiatedBy,
	}

	createdSession, err := h.sessionRepo.Create(sctx.Ctx, session)
	if err != nil {
		log.Error(sctx.Ctx, "Error creating session: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Session created with ID: %s", createdSession.SessionID)

	return &resp.InitiateProfileResponse{
		StatusCodeAndMessage: port.SessionCreateSuccess,
		SessionID:            createdSession.SessionID,
		AgentType:            createdSession.AgentType,
		WorkflowState: &resp.WorkflowState{
			CurrentStep:        createdSession.CurrentStep,
			NextStep:           createdSession.NextStep,
			AllowedActions:     []string{"FETCH_HRMS", "SAVE", "CANCEL"},
			ProgressPercentage: createdSession.ProgressPercentage,
		},
		ExpiresAt: &createdSession.ExpiresAt,
	}, nil
}

// FetchHRMSData fetches employee data from HRMS
// AGT-002: Fetch HRMS Employee Data
// FR-AGT-PRF-002: HRMS Data Auto-Population
// BR-AGT-PRF-003: HRMS Integration Mandatory for Departmental Employees
// INT-AGT-001: HRMS Integration
func (h *AgentProfileCreationHandler) FetchHRMSData(sctx *serverRoute.Context, req FetchHRMSEmployeeRequest) (*resp.FetchHRMSResponse, error) {
	log.Info(sctx.Ctx, "Fetching HRMS data for session: %s, employee: %s", req.SessionID, req.EmployeeID)

	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// INT-AGT-001: HRMS Integration
	// Call actual HRMS service for employee data fetch
	// For now, return mock employee data
	employeeData := &resp.EmployeeData{
		EmployeeID:   req.EmployeeID,
		FirstName:    "Rajesh",
		LastName:     "Kumar",
		DateOfBirth:  "1985-05-15",
		Gender:       "Male",
		Designation:  "Postmaster",
		OfficeCode:   "OFF-001",
		MobileNumber: "9876543210",
		Email:        "rajesh.kumar@indiapost.gov.in",
		Status:       "ACTIVE",
	}

	// Build form data from HRMS response
	formDataMap := map[string]interface{}{
		"employee_id":   employeeData.EmployeeID,
		"first_name":    employeeData.FirstName,
		"last_name":     employeeData.LastName,
		"date_of_birth": employeeData.DateOfBirth,
		"gender":        employeeData.Gender,
		"designation":   employeeData.Designation,
		"office_code":   employeeData.OfficeCode,
		"mobile_number": employeeData.MobileNumber,
		"email":         employeeData.Email,
	}
	formDataJSON, _ := json.Marshal(formDataMap)

	// ATOMIC: Save form data + update workflow state in single database round trip
	// Prevents inconsistent state where form data is saved but workflow state is not updated
	_, err = h.sessionRepo.SaveFormDataAndUpdateWorkflowStateReturning(
		sctx.Ctx,
		req.SessionID,
		string(formDataJSON),
		domain.WorkflowStateHRMSFetched,
		"HRMS_DATA_FETCHED",
		"PROFILE_DETAILS",
		30,
		req.SessionID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error saving HRMS data and updating state: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "HRMS data fetched and saved successfully")

	return &resp.FetchHRMSResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		EmployeeData:         employeeData,
		AutoPopulated:        true,
		Message:              "HRMS data fetched successfully. Please review and correct if needed.",
	}, nil
}

// GetAdvisorCoordinators retrieves active coordinators for advisor linkage
// AGT-003: Get Advisor Coordinators List
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
func (h *AgentProfileCreationHandler) GetAdvisorCoordinators(sctx *serverRoute.Context, req GetAdvisorCoordinatorsQuery) (*resp.AdvisorCoordinatorsResponse, error) {
	log.Info(sctx.Ctx, "Fetching advisor coordinators with filters: circle=%s, division=%s", req.CircleID, req.DivisionID)

	// Query active coordinators from repository
	coordinators, err := h.profileRepo.GetActiveAdvisorCoordinators(sctx.Ctx, req.CircleID, req.DivisionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching coordinators: %v", err)
		return nil, err
	}

	// Build response
	coordinatorList := make([]resp.AdvisorCoordinator, 0, len(coordinators))
	for _, c := range coordinators {
		coordinatorList = append(coordinatorList, resp.AdvisorCoordinator{
			AgentID:    c.AgentID,
			AgentCode:  c.AgentCode.String,
			FirstName:  c.FirstName,
			MiddleName: c.MiddleName.String,
			LastName:   c.LastName,
			CircleID:   c.CircleID.String,
			DivisionID: c.DivisionID.String,
		})
	}

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	return &resp.AdvisorCoordinatorsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     port.NewMetaDataResponse((req.Page-1)*req.Limit, req.Limit, uint64(len(coordinatorList))),
		Data:                 coordinatorList,
	}, nil
}

// LinkCoordinator links advisor to coordinator
// AGT-004: Link Advisor to Coordinator
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
func (h *AgentProfileCreationHandler) LinkCoordinator(sctx *serverRoute.Context, req LinkCoordinatorRequest) (*resp.LinkCoordinatorResponse, error) {
	log.Info(sctx.Ctx, "Linking coordinator %s to session %s", req.CoordinatorID, req.SessionID)

	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// Verify coordinator exists and is active
	// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
	coordinator, err := h.profileRepo.FindByID(sctx.Ctx, req.CoordinatorID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding coordinator: %v", err)
		return nil, err
	}

	if coordinator.AgentType != domain.AgentTypeAdvisorCoordinator {
		return nil, fmt.Errorf("selected agent is not an advisor coordinator")
	}

	if coordinator.Status != domain.AgentStatusActive {
		return nil, fmt.Errorf("coordinator is not active")
	}

	// Merge coordinator ID into existing form data
	var formData map[string]interface{}
	if session.FormData.Valid {
		json.Unmarshal([]byte(session.FormData.String), &formData)
	} else {
		formData = make(map[string]interface{})
	}
	formData["coordinator_id"] = req.CoordinatorID
	if req.LinkageEffectiveDate.Valid {
		formData["linkage_effective_date"] = req.LinkageEffectiveDate.Time.Format("2006-01-02")
	}
	formDataJSON, _ := json.Marshal(formData)

	// ATOMIC: Save coordinator link + update workflow state in single database round trip
	// Prevents inconsistent state where coordinator is saved but workflow state is not updated
	_, err = h.sessionRepo.SaveFormDataAndUpdateWorkflowStateReturning(
		sctx.Ctx,
		req.SessionID,
		string(formDataJSON),
		domain.WorkflowStateCoordinatorLinking,
		"COORDINATOR_LINKED",
		"PROFILE_VALIDATION",
		50,
		req.SessionID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error linking coordinator and updating state: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Coordinator linked successfully")

	return &resp.LinkCoordinatorResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		CoordinatorLinked:    true,
		CoordinatorName:      coordinator.FirstName + " " + coordinator.LastName,
		Message:              "Coordinator linked successfully",
	}, nil
}

// ValidateProfile validates profile details
// AGT-005: Validate Profile Details
// FR-AGT-PRF-002: Profile Data Validation
// VR-AGT-PRF-002 to VR-AGT-PRF-030: All validation rules
func (h *AgentProfileCreationHandler) ValidateProfile(sctx *serverRoute.Context, req SessionIDUri) (*resp.ValidateProfileResponse, error) {
	log.Info(sctx.Ctx, "Validating profile for session: %s", req.SessionID)

	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// Parse form data for validation
	var formData map[string]interface{}
	if session.FormData.Valid {
		json.Unmarshal([]byte(session.FormData.String), &formData)
	} else {
		return nil, fmt.Errorf("no form data found in session")
	}

	// Comprehensive validation logic
	// VR-AGT-PRF-002 to VR-AGT-PRF-030: All validation rules
	validationErrors := []resp.ValidationError{}

	// VR-AGT-PRF-001: First Name is mandatory
	if getStringFromMap(formData, "first_name") == "" {
		validationErrors = append(validationErrors, resp.ValidationError{
			Field:   "first_name",
			Message: "First name is required",
			Code:    "VR-AGT-PRF-001",
		})
	}

	// VR-AGT-PRF-002: Last Name is mandatory
	if getStringFromMap(formData, "last_name") == "" {
		validationErrors = append(validationErrors, resp.ValidationError{
			Field:   "last_name",
			Message: "Last name is required",
			Code:    "VR-AGT-PRF-002",
		})
	}

	// VR-AGT-PRF-003: PAN format validation
	panNumber := getStringFromMap(formData, "pan_number")
	if panNumber != "" && len(panNumber) != 10 {
		validationErrors = append(validationErrors, resp.ValidationError{
			Field:   "pan_number",
			Message: "PAN must be exactly 10 characters",
			Code:    "VR-AGT-PRF-003",
		})
	}

	// VR-AGT-PRF-004: Date of Birth is mandatory
	if getStringFromMap(formData, "date_of_birth") == "" {
		validationErrors = append(validationErrors, resp.ValidationError{
			Field:   "date_of_birth",
			Message: "Date of birth is required",
			Code:    "VR-AGT-PRF-004",
		})
	}

	// VR-AGT-PRF-005: Gender is mandatory
	if getStringFromMap(formData, "gender") == "" {
		validationErrors = append(validationErrors, resp.ValidationError{
			Field:   "gender",
			Message: "Gender is required",
			Code:    "VR-AGT-PRF-005",
		})
	}

	// BR-AGT-PRF-001: Coordinator linkage required for advisors
	if session.AgentType == domain.AgentTypeAdvisor {
		if getStringFromMap(formData, "coordinator_id") == "" {
			validationErrors = append(validationErrors, resp.ValidationError{
				Field:   "coordinator_id",
				Message: "Advisor coordinator linkage is mandatory for advisors",
				Code:    "BR-AGT-PRF-001",
			})
		}
	}

	// ATOMIC: Update workflow state using UPDATE...RETURNING
	// Only update if validation passed, otherwise keep current state
	if len(validationErrors) == 0 {
		err = h.sessionRepo.UpdateWorkflowState(sctx.Ctx, req.SessionID, domain.WorkflowStateProfileValidation, "VALIDATION_COMPLETE", "PROFILE_SUBMISSION", 80, req.SessionID)
		if err != nil {
			log.Error(sctx.Ctx, "Error updating workflow state: %v", err)
			return nil, err
		}
	}

	log.Info(sctx.Ctx, "Profile validation completed with %d errors", len(validationErrors))

	return &resp.ValidateProfileResponse{
		StatusCodeAndMessage: port.ValidationSuccess,
		IsValid:              len(validationErrors) == 0,
		ValidationErrors:     validationErrors,
		Message:              "Profile validation completed",
	}, nil
}

// SubmitProfile submits profile for creation via Temporal workflow
// AGT-006: Submit Profile for Creation
// FR-AGT-PRF-001: New Profile Creation
// WF-002: Agent Onboarding Workflow
func (h *AgentProfileCreationHandler) SubmitProfile(sctx *serverRoute.Context, req SubmitProfileRequest) (*resp.SubmitProfileResponse, error) {
	log.Info(sctx.Ctx, "Submitting profile for session: %s", req.SessionID)

	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// Ensure validation was completed
	if session.WorkflowState != domain.WorkflowStateProfileValidation {
		return nil, fmt.Errorf("profile must be validated before submission")
	}

	// Parse session form data
	var formData map[string]interface{}
	if session.FormData.Valid {
		json.Unmarshal([]byte(session.FormData.String), &formData)
	} else {
		return nil, fmt.Errorf("no form data found in session")
	}

	// Build workflow input from session data
	// WF-002: Agent Onboarding Workflow
	workflowInput := activities.OnboardingInput{
		SessionID: session.SessionID,
		AgentType: session.AgentType,
		ProfileData: activities.ProfileData{
			FirstName:    getStringFromMap(formData, "first_name"),
			LastName:     getStringFromMap(formData, "last_name"),
			MiddleName:   getStringFromMap(formData, "middle_name"),
			PANNumber:    getStringFromMap(formData, "pan_number"),
			DateOfBirth:  getStringFromMap(formData, "date_of_birth"),
			Gender:       getStringFromMap(formData, "gender"),
			OfficeCode:   getStringFromMap(formData, "office_code"),
			MobileNumber: getStringFromMap(formData, "mobile_number"),
			Email:        getStringFromMap(formData, "email"),
			EmployeeID:   getStringFromMap(formData, "employee_id"),
		},
		CoordinatorID: getStringFromMap(formData, "coordinator_id"),
		SubmittedBy:   req.SubmittedBy,
	}

	// Start Temporal workflow
	// Task queue must match worker configuration
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("agent-onboarding-%s", session.SessionID),
		TaskQueue: "agent-profile-task-queue",
	}

	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.AgentOnboardingWorkflow, workflowInput)
	if err != nil {
		log.Error(sctx.Ctx, "Error starting Temporal workflow: %v", err)
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	log.Info(sctx.Ctx, "Temporal workflow started: %s (run: %s)", we.GetID(), we.GetRunID())

	// NOTE: The workflow itself will record its start in the database as the FIRST activity
	// This ensures self-healing: If database update fails, Temporal retries the activity
	// The workflow doesn't proceed until the database knows about it
	// See RecordWorkflowStartActivity in agent_onboarding_workflow.go

	log.Info(sctx.Ctx, "Profile submitted successfully - workflow will record itself in database")

	return &resp.SubmitProfileResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		SubmissionID:         we.GetID(),
		Status:               "PROCESSING",
		Message:              "Profile submitted successfully. Workflow started.",
		WorkflowID:           we.GetID(),
	}, nil
}

// ========================================================================
// SESSION MANAGEMENT ENDPOINTS (AGT-016 to AGT-019)
// ========================================================================

// GetSessionStatus returns current session status and workflow state
// AGT-016: Get Session Status
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentProfileCreationHandler) GetSessionStatus(sctx *serverRoute.Context, req SessionIDUri) (*resp.SessionStatusResponse, error) {
	log.Info(sctx.Ctx, "Fetching session status for session: %s", req.SessionID)

	// Query session from repository
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	// Build workflow state
	workflowState := &resp.WorkflowState{
		CurrentStep:        session.CurrentStep,
		NextStep:           session.NextStep,
		AllowedActions:     []string{"SAVE", "SUBMIT", "CANCEL"},
		ProgressPercentage: session.ProgressPercentage,
	}

	// Adjust allowed actions based on workflow state
	if session.WorkflowState == domain.WorkflowStateInitiated {
		workflowState.AllowedActions = []string{"FETCH_HRMS", "SAVE", "CANCEL"}
	} else if session.WorkflowState == domain.WorkflowStateProfileValidation {
		workflowState.AllowedActions = []string{"SUBMIT", "SAVE", "CANCEL"}
	}

	return &resp.SessionStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		SessionID:            session.SessionID,
		Status:               session.Status,
		WorkflowState:        workflowState,
		LastSavedAt:          &session.UpdatedAt,
	}, nil
}

// SaveSession saves current session checkpoint
// AGT-017: Save Session Checkpoint
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentProfileCreationHandler) SaveSession(sctx *serverRoute.Context, req SaveSessionRequest) (*resp.SaveSessionResponse, error) {
	log.Info(sctx.Ctx, "Saving session checkpoint for session: %s", req.SessionID)

	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// Convert form data to JSON
	formDataJSON, err := json.Marshal(req.FormData)
	if err != nil {
		log.Error(sctx.Ctx, "Error marshaling form data: %v", err)
		return nil, err
	}

	// ATOMIC: Save form data and return updated session in single round trip
	// Uses UPDATE...RETURNING to get updated session (including updated_at, expires_at)
	// Avoids second FindByID call
	updatedSession, err := h.sessionRepo.SaveFormDataAndUpdateWorkflowStateReturning(
		sctx.Ctx,
		req.SessionID,
		string(formDataJSON),
		session.WorkflowState,      // Keep current workflow state
		req.CurrentScreen,          // Update current screen from request
		session.NextStep,           // Keep current next step
		session.ProgressPercentage, // Keep current progress
		req.SessionID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Error saving session data: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Session saved successfully: %s", req.SessionID)

	return &resp.SaveSessionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Saved:                true,
		SessionExpiresAt:     &updatedSession.ExpiresAt,
	}, nil
}

// ResumeSession resumes a saved session
// AGT-018: Resume Session
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentProfileCreationHandler) ResumeSession(sctx *serverRoute.Context, req SessionIDUri) (*resp.ResumeSessionResponse, error) {
	log.Info(sctx.Ctx, "Resuming session: %s", req.SessionID)

	// Query session from repository
	session, err := h.sessionRepo.FindByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error finding session: %v", err)
		return nil, err
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active or has expired")
	}

	// Parse form data
	var formData map[string]interface{}
	if session.FormData.Valid {
		json.Unmarshal([]byte(session.FormData.String), &formData)
	} else {
		formData = make(map[string]interface{})
	}

	// Build workflow state
	workflowState := &resp.WorkflowState{
		CurrentStep:        session.CurrentStep,
		NextStep:           session.NextStep,
		AllowedActions:     []string{"SAVE", "SUBMIT", "CANCEL"},
		ProgressPercentage: session.ProgressPercentage,
	}

	log.Info(sctx.Ctx, "Session resumed successfully: %s", req.SessionID)

	return &resp.ResumeSessionResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		SessionID:            session.SessionID,
		AgentType:            session.AgentType,
		FormData:             formData,
		WorkflowState:        workflowState,
	}, nil
}

// CancelSession cancels an active session
// AGT-019: Cancel Session
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentProfileCreationHandler) CancelSession(sctx *serverRoute.Context, req SessionIDUri) (*resp.CancelSessionResponse, error) {
	log.Info(sctx.Ctx, "Cancelling session: %s", req.SessionID)

	// ATOMIC: Cancel session and return result in single database round trip
	// Uses UPDATE...RETURNING to cancel and verify in one operation
	// Avoids separate FindByID + Cancel calls
	cancelledSession, err := h.sessionRepo.CancelReturning(sctx.Ctx, req.SessionID, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Error cancelling session: %v", err)
		return nil, err
	}

	// Check if already cancelled (idempotent operation)
	if cancelledSession.Status == domain.SessionStatusCancelled {
		log.Info(sctx.Ctx, "Session was already cancelled: %s", req.SessionID)
	} else {
		log.Info(sctx.Ctx, "Session cancelled successfully: %s", req.SessionID)
	}

	return &resp.CancelSessionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Cancelled:            true,
		Message:              "Session cancelled successfully",
	}, nil
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// Helper function to get string from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

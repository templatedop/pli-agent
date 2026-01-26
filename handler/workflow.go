package handler

import (
	"time"

	"github.com/google/uuid"

	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentWorkflowHandler handles all workflow and session management APIs
// AGT-016 to AGT-021: Workflow, Status, and Notification APIs
type AgentWorkflowHandler struct {
	*serverHandler.Base
	// TODO: Add Temporal client when implementing workflows
	// temporalClient client.Client
}

// NewAgentWorkflowHandler creates a new workflow handler
func NewAgentWorkflowHandler() *AgentWorkflowHandler {
	base := serverHandler.New("Agent Workflow APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentWorkflowHandler{
		Base: base,
	}
}

// Routes defines all workflow API routes
func (h *AgentWorkflowHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/agent-profiles/sessions/:session_id/status", h.GetSessionStatus).Name("Get Session Status"),
		serverRoute.POST("/agent-profiles/sessions/:session_id/save", h.SaveSession).Name("Save Session Checkpoint"),
		serverRoute.GET("/agent-profiles/sessions/:session_id/resume", h.ResumeSession).Name("Resume Session"),
		serverRoute.DELETE("/agent-profiles/sessions/:session_id", h.CancelSession).Name("Cancel Session"),
		serverRoute.GET("/agent-profiles/creation-status/:agent_id", h.GetCreationStatus).Name("Get Creation Status"),
		serverRoute.POST("/agents/:agent_id/notifications/resend-welcome", h.ResendWelcomeNotification).Name("Resend Welcome Notification"),
	}
}

// GetSessionStatus returns profile creation session status
// AGT-016: Get Session Status
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentWorkflowHandler) GetSessionStatus(sctx *serverRoute.Context, req SessionIDUri) (*resp.SessionStatusResponse, error) {
	log.Info(sctx.Ctx, "Fetching session status for session ID: %s", req.SessionID)

	// TODO: WF-AGT-PRF-001 - Query session from Temporal workflow or database
	// For now, return mock response

	now := time.Now()
	workflowState := &resp.WorkflowState{
		CurrentStep:        "PROFILE_DETAILS",
		NextStep:           "ADDRESS_DETAILS",
		AllowedActions:     []string{"SAVE", "SUBMIT", "CANCEL"},
		ProgressPercentage: 40,
	}

	return &resp.SessionStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		SessionID:            req.SessionID,
		Status:               "ACTIVE",
		WorkflowState:        workflowState,
		LastSavedAt:          &now,
	}, nil
}

// SaveSession saves profile creation session checkpoint
// AGT-017: Save Session Checkpoint
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentWorkflowHandler) SaveSession(sctx *serverRoute.Context, req SaveSessionRequest) (*resp.SaveSessionResponse, error) {
	log.Info(sctx.Ctx, "Saving session checkpoint for session ID: %s, screen: %s", req.SessionID, req.CurrentScreen)

	// TODO: WF-AGT-PRF-001 - Save session data to Temporal workflow or database
	// For now, return mock response

	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour expiry

	return &resp.SaveSessionResponse{
		StatusCodeAndMessage: port.CustomEnv.WithMessage("Session saved successfully"),
		Saved:                true,
		SessionExpiresAt:     &expiresAt,
	}, nil
}

// ResumeSession retrieves saved session data
// AGT-018: Resume Session
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentWorkflowHandler) ResumeSession(sctx *serverRoute.Context, req SessionIDUri) (*resp.ResumeSessionResponse, error) {
	log.Info(sctx.Ctx, "Resuming session for session ID: %s", req.SessionID)

	// TODO: WF-AGT-PRF-001 - Retrieve session data from Temporal workflow or database
	// For now, return mock response

	formData := map[string]interface{}{
		"first_name":   "Rajesh",
		"last_name":    "Kumar",
		"date_of_birth": "1985-05-15",
		"email":        "rajesh.kumar@example.com",
		"mobile":       "9876543210",
	}

	workflowState := &resp.WorkflowState{
		CurrentStep:        "PROFILE_DETAILS",
		NextStep:           "ADDRESS_DETAILS",
		AllowedActions:     []string{"SAVE", "SUBMIT", "CANCEL"},
		ProgressPercentage: 40,
	}

	return &resp.ResumeSessionResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		SessionID:            req.SessionID,
		AgentType:            "ADVISOR",
		FormData:             formData,
		WorkflowState:        workflowState,
	}, nil
}

// CancelSession cancels profile creation session
// AGT-019: Cancel Session
// WF-AGT-PRF-001: Profile Creation Workflow
func (h *AgentWorkflowHandler) CancelSession(sctx *serverRoute.Context, req SessionIDUri) (*resp.CancelSessionResponse, error) {
	log.Info(sctx.Ctx, "Cancelling session for session ID: %s", req.SessionID)

	// TODO: WF-AGT-PRF-001 - Cancel Temporal workflow and cleanup session data
	// For now, return mock response

	return &resp.CancelSessionResponse{
		StatusCodeAndMessage: port.CustomEnv.WithMessage("Session cancelled successfully"),
		Cancelled:            true,
		Message:              "Profile creation session has been cancelled",
	}, nil
}

// GetCreationStatus returns agent creation status
// AGT-020: Get Creation Status
// FR-AGT-PRF-009: Agent Profile Status Tracking
func (h *AgentWorkflowHandler) GetCreationStatus(sctx *serverRoute.Context, req AgentIDUri) (*resp.CreationStatusResponse, error) {
	log.Info(sctx.Ctx, "Fetching creation status for agent ID: %s", req.AgentID)

	// TODO: Query agent profile and workflow status
	// For now, return mock response

	verificationStatus := &resp.VerificationStatus{
		PANVerified:    true,
		HRMSVerified:   true,
		OfficeVerified: false,
	}

	nextAction := resp.NextActionDue{
		Action:  "Complete Office Verification",
		DueDate: nil,
	}

	slaTracking := &resp.SLATracking{
		SLAStatus:          "YELLOW",
		TimeElapsedMinutes: 45,
		NextActionsDue:     []resp.NextActionDue{nextAction},
	}

	return &resp.CreationStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		AgentID:              req.AgentID,
		Status:               "IN_PROGRESS",
		CurrentStage:         "VERIFICATION",
		VerificationStatus:   verificationStatus,
		SLATracking:          slaTracking,
	}, nil
}

// ResendWelcomeNotification resends welcome notification to agent
// AGT-021: Resend Welcome Notification
// INT-AGT-005: Notification Service Integration
func (h *AgentWorkflowHandler) ResendWelcomeNotification(sctx *serverRoute.Context, req ResendWelcomeNotificationRequest) (*resp.WelcomeNotificationResponse, error) {
	log.Info(sctx.Ctx, "Resending welcome notification for agent ID: %s", req.AgentID)

	// TODO: INT-AGT-005 - Call notification service to send email and SMS
	// For now, return mock response

	// Default channels if not specified
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{"EMAIL", "SMS"}
	}

	notificationID := uuid.New().String()
	sentAt := time.Now()

	return &resp.WelcomeNotificationResponse{
		StatusCodeAndMessage: port.CustomEnv.WithMessage("Welcome notification sent successfully"),
		NotificationID:       notificationID,
		ChannelsSent:         channels,
		SentAt:               &sentAt,
	}, nil
}

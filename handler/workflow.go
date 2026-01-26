package handler

import (
	"time"

	"github.com/google/uuid"

	"pli-agent-api/core/port"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentWorkflowHandler handles agent status and notification APIs
// AGT-020 to AGT-021: Status and Notification APIs
// NOTE: Session management APIs (AGT-016 to AGT-019) moved to Phase 5 with Temporal WF-002
type AgentWorkflowHandler struct {
	*serverHandler.Base
	profileRepo *repo.AgentProfileRepository
	// TODO: Add Temporal client when implementing WF-002 in Phase 5
	// temporalClient client.Client
	// TODO: Add notification service client for INT-AGT-005
	// notificationClient NotificationService
}

// NewAgentWorkflowHandler creates a new workflow handler
func NewAgentWorkflowHandler(profileRepo *repo.AgentProfileRepository) *AgentWorkflowHandler {
	base := serverHandler.New("Agent Status & Notification APIs").SetPrefix("/v1").AddPrefix("")
	return &AgentWorkflowHandler{
		Base:        base,
		profileRepo: profileRepo,
	}
}

// Routes defines status and notification API routes
// NOTE: Session management routes (AGT-016 to AGT-019) will be implemented in Phase 5
// with Temporal WF-002: Agent Onboarding Workflow
func (h *AgentWorkflowHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// AGT-020: Get Agent Creation Status (Connected to Repository)
		serverRoute.GET("/agent-profiles/creation-status/:agent_id", h.GetCreationStatus).Name("Get Agent Creation Status"),

		// AGT-021: Resend Welcome Notification (Mock - needs INT-AGT-005)
		serverRoute.POST("/agents/:agent_id/notifications/resend-welcome", h.ResendWelcomeNotification).Name("Resend Welcome Notification"),

		// TODO: Phase 5 - Session Management APIs (with Temporal WF-002)
		// AGT-016: GET /agent-profiles/sessions/:session_id/status
		// AGT-017: POST /agent-profiles/sessions/:session_id/save
		// AGT-018: GET /agent-profiles/sessions/:session_id/resume
		// AGT-019: DELETE /agent-profiles/sessions/:session_id
	}
}

// GetCreationStatus returns agent creation status
// AGT-020: Get Creation Status
// FR-AGT-PRF-009: Agent Profile Status Tracking
// FIXED: Now connects to AgentProfileRepository for real data
func (h *AgentWorkflowHandler) GetCreationStatus(sctx *serverRoute.Context, req AgentIDUri) (*resp.CreationStatusResponse, error) {
	log.Info(sctx.Ctx, "Fetching creation status for agent ID: %s", req.AgentID)

	// Query actual agent profile from repository
	profile, err := h.profileRepo.FindByID(sctx.Ctx, req.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching agent profile: %v", err)
		return nil, err
	}

	// Build verification status from actual profile data
	verificationStatus := &resp.VerificationStatus{
		PANVerified:    profile.PANNumber != "",
		HRMSVerified:   profile.EmployeeID.Valid && profile.EmployeeID.String != "",
		OfficeVerified: profile.OfficeCode != "",
	}

	// Calculate SLA tracking based on creation time
	timeElapsed := time.Since(profile.CreatedAt)
	slaStatus := "GREEN"
	if timeElapsed > 24*time.Hour {
		slaStatus = "RED"
	} else if timeElapsed > 4*time.Hour {
		slaStatus = "YELLOW"
	}

	// Determine next actions based on verification status
	nextActions := []resp.NextActionDue{}
	if !verificationStatus.PANVerified {
		nextActions = append(nextActions, resp.NextActionDue{
			Action:  "Complete PAN Verification",
			DueDate: nil,
		})
	}
	if !verificationStatus.HRMSVerified {
		nextActions = append(nextActions, resp.NextActionDue{
			Action:  "Complete HRMS Verification",
			DueDate: nil,
		})
	}
	if !verificationStatus.OfficeVerified {
		nextActions = append(nextActions, resp.NextActionDue{
			Action:  "Complete Office Verification",
			DueDate: nil,
		})
	}

	slaTracking := &resp.SLATracking{
		SLAStatus:          slaStatus,
		TimeElapsedMinutes: int(timeElapsed.Minutes()),
		NextActionsDue:     nextActions,
	}

	// Determine current stage based on status and workflow state
	currentStage := "CREATED"
	if profile.WorkflowState.Valid && profile.WorkflowState.String != "" {
		currentStage = profile.WorkflowState.String
	}

	return &resp.CreationStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		AgentID:              profile.AgentID,
		Status:               profile.Status,
		CurrentStage:         currentStage,
		VerificationStatus:   verificationStatus,
		SLATracking:          slaTracking,
	}, nil
}

// ResendWelcomeNotification resends welcome notification to agent
// AGT-021: Resend Welcome Notification
// INT-AGT-005: Notification Service Integration
//
// MOCK IMPLEMENTATION: This is a mock response for development/testing.
//
// Production Requirements:
// 1. Inject notification service client (email + SMS gateway)
// 2. Query agent profile for contact details (email, mobile)
// 3. Call notification service with welcome template
// 4. Store notification log in database
// 5. Handle failures and retries
//
// Integration Points:
// - Email Service: SMTP or third-party (SendGrid, AWS SES)
// - SMS Service: Third-party gateway (Twilio, AWS SNS)
// - Template Service: For welcome message templates
func (h *AgentWorkflowHandler) ResendWelcomeNotification(sctx *serverRoute.Context, req ResendWelcomeNotificationRequest) (*resp.WelcomeNotificationResponse, error) {
	log.Info(sctx.Ctx, "Resending welcome notification for agent ID: %s", req.AgentID)

	// Verify agent exists
	profile, err := h.profileRepo.FindByID(sctx.Ctx, req.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching agent profile: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Agent found: %s %s, Status: %s", profile.FirstName, profile.LastName, profile.Status)

	// TODO: INT-AGT-005 - Call notification service
	// Example production code:
	// err = h.notificationClient.SendWelcomeEmail(profile.Email, profile.FirstName)
	// err = h.notificationClient.SendWelcomeSMS(profile.Mobile, profile.FirstName)

	// Default channels if not specified
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{"EMAIL", "SMS"}
	}

	// Mock response - in production, this should be the actual notification ID from service
	notificationID := uuid.New().String()
	sentAt := time.Now()

	log.Info(sctx.Ctx, "Welcome notification mock sent to agent %s via channels: %v", req.AgentID, channels)

	return &resp.WelcomeNotificationResponse{
		StatusCodeAndMessage: port.CustomEnv.WithMessage("Welcome notification sent successfully (mock)"),
		NotificationID:       notificationID,
		ChannelsSent:         channels,
		SentAt:               &sentAt,
	}, nil
}

package handler

import (
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
	req "pli-agent-api/handler/request"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"
	"pli-agent-api/workflows"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"
)

// AgentStatusManagementHandler handles agent status management APIs
// AGT-039 to AGT-041: Status Management
// BR-AGT-PRF-016: Status Update with Reason
// BR-AGT-PRF-017: Agent Termination Workflow
// WF-AGT-PRF-004: Termination Workflow (Temporal)
// WF-AGT-PRF-011: Reinstatement Workflow (Temporal)
type AgentStatusManagementHandler struct {
	*serverHandler.Base
	terminationRepo *repo.AgentTerminationRepository
	profileRepo     *repo.AgentProfileRepository
	temporalClient  client.Client
}

// NewAgentStatusManagementHandler creates a new status management handler
func NewAgentStatusManagementHandler(
	terminationRepo *repo.AgentTerminationRepository,
	profileRepo *repo.AgentProfileRepository,
	temporalClient client.Client,
) *AgentStatusManagementHandler {
	return &AgentStatusManagementHandler{
		Base:            &serverHandler.Base{},
		terminationRepo: terminationRepo,
		profileRepo:     profileRepo,
		temporalClient:  temporalClient,
	}
}

// RegisterRoutes registers all status management routes
func (h *AgentStatusManagementHandler) RegisterRoutes() []serverRoute.Route {
	return []serverRoute.Route{
		// AGT-039: Terminate Agent
		serverRoute.NewRoute("POST", "/agents/:agent_id/terminate", h.TerminateAgent),
		// AGT-040: Get Termination Letter
		serverRoute.NewRoute("GET", "/agents/:agent_id/termination-letter", h.GetTerminationLetter),
		// AGT-041: Reinstate Agent
		serverRoute.NewRoute("POST", "/agents/:agent_id/reinstate", h.ReinstateAgent),
	}
}

// TerminateAgent terminates an agent and initiates termination workflow
// AGT-039: Terminate Agent
// BR-AGT-PRF-016: Status Update with Reason (min 20 chars)
// BR-AGT-PRF-017: Agent Termination Workflow
// WF-AGT-PRF-004: Termination Workflow (Temporal orchestration)
//
// OPTIMIZED: Single database hit using CTE pattern
// Updates agent_profiles + creates termination_record + audit_log atomically
func (h *AgentStatusManagementHandler) TerminateAgent(
	sctx *serverRoute.Context,
	uri AgentIDUri,
	request req.TerminateAgentRequest,
) (*resp.TerminateAgentResponse, error) {
	log.Info(sctx.Ctx, "Terminating agent: %s, reason code: %s", uri.AgentID, request.TerminationReasonCode)

	// Verify agent exists and is not already terminated
	profile, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Agent not found: %v", err)
		return nil, err
	}

	if profile.Status == domain.AgentStatusTerminated {
		log.Error(sctx.Ctx, "Agent already terminated: %s", uri.AgentID)
		return nil, fmt.Errorf("agent is already terminated")
	}

	// Generate workflow ID if not provided
	workflowID := request.WorkflowID
	if workflowID == "" {
		workflowID = fmt.Sprintf("termination-workflow-%s-%d", uri.AgentID, request.EffectiveDate.Unix())
	}

	// Terminate agent - SINGLE database hit
	// Updates: agent_profiles.status + termination_record + audit_log
	terminationRecord, err := h.terminationRepo.TerminateAgent(
		sctx.Ctx,
		uri.AgentID,
		request.TerminationReason,
		request.TerminationReasonCode,
		request.EffectiveDate,
		request.TerminatedBy,
		workflowID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to terminate agent: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Agent terminated successfully: %s, termination_id: %s", uri.AgentID, terminationRecord.TerminationID)

	// Initiate Temporal workflow for termination orchestration
	// WF-AGT-PRF-004: Agent Termination Workflow
	// This workflow handles: portal disable, commission stop, letter generation, data archival, notifications
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "agent-profile-task-queue",
	}

	workflowInput := workflows.TerminationWorkflowInput{
		AgentID:               uri.AgentID,
		TerminationReason:     request.TerminationReason,
		TerminationReasonCode: request.TerminationReasonCode,
		EffectiveDate:         request.EffectiveDate,
		TerminatedBy:          request.TerminatedBy,
		TerminationRecordID:   terminationRecord.TerminationID,
	}

	_, err = h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.AgentTerminationWorkflow, workflowInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start termination workflow: %v", err)
		// Don't fail the request - the database is already updated
		// Workflow can be retried manually if needed
	} else {
		log.Info(sctx.Ctx, "Termination workflow started: %s", workflowID)
	}

	return &resp.TerminateAgentResponse{
		StatusCodeAndMessage: port.AgentTerminationSuccess,
		Message:              "Agent termination initiated successfully",
		TerminationRecord:    resp.ToTerminationRecordDTO(*terminationRecord),
		NextSteps: []string{
			"Portal access will be disabled",
			"Commission processing will stop",
			"Termination letter will be generated and sent to agent",
			"Agent data will be archived for 7-year retention",
			"Notifications will be sent to relevant stakeholders",
		},
	}, nil
}

// GetTerminationLetter retrieves termination letter for an agent
// AGT-040: Get Termination Letter
// BR-AGT-PRF-017: Agent Termination Workflow
func (h *AgentStatusManagementHandler) GetTerminationLetter(
	sctx *serverRoute.Context,
	uri AgentIDUri,
) (*resp.GetTerminationLetterResponse, error) {
	log.Info(sctx.Ctx, "Fetching termination letter for agent: %s", uri.AgentID)

	// Verify agent exists
	profile, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Agent not found: %v", err)
		return nil, err
	}

	// Verify agent is terminated
	if profile.Status != domain.AgentStatusTerminated {
		log.Error(sctx.Ctx, "Agent is not terminated: %s", uri.AgentID)
		return nil, fmt.Errorf("agent is not terminated, cannot retrieve termination letter")
	}

	// Fetch termination record
	terminationRecord, err := h.terminationRepo.FindTerminationRecord(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Termination record not found: %v", err)
		return nil, err
	}

	// Check if letter is generated
	if !terminationRecord.LetterGenerated || !terminationRecord.TerminationLetterURL.Valid {
		log.Warn(sctx.Ctx, "Termination letter not yet generated for agent: %s", uri.AgentID)
		return nil, fmt.Errorf("termination letter is still being generated, please try again later")
	}

	log.Info(sctx.Ctx, "Termination letter retrieved for agent: %s", uri.AgentID)

	return &resp.GetTerminationLetterResponse{
		StatusCodeAndMessage: port.TerminationLetterRetrieved,
		Message:              "Termination letter retrieved successfully",
		TerminationRecord:    resp.ToTerminationRecordDTO(*terminationRecord),
		LetterURL:            terminationRecord.TerminationLetterURL.String,
		LetterGeneratedAt:    terminationRecord.TerminationLetterGeneratedAt.Time,
	}, nil
}

// ReinstateAgent creates a reinstatement request for a terminated agent
// AGT-041: Reinstate Agent
// WF-AGT-PRF-011: Reinstatement Workflow (Temporal orchestration)
//
// OPTIMIZED: Single database hit using CTE pattern
// Creates reinstatement_request + audit_log atomically
func (h *AgentStatusManagementHandler) ReinstateAgent(
	sctx *serverRoute.Context,
	uri AgentIDUri,
	request req.ReinstateAgentRequest,
) (*resp.ReinstateAgentResponse, error) {
	log.Info(sctx.Ctx, "Creating reinstatement request for agent: %s", uri.AgentID)

	// Verify agent exists and is terminated
	profile, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Agent not found: %v", err)
		return nil, err
	}

	if profile.Status != domain.AgentStatusTerminated {
		log.Error(sctx.Ctx, "Agent is not terminated: %s", uri.AgentID)
		return nil, fmt.Errorf("agent must be terminated to request reinstatement")
	}

	// Check if there's already a pending reinstatement request
	pendingRequests, err := h.terminationRepo.FindPendingReinstatementRequests(sctx.Ctx)
	if err != nil {
		log.Error(sctx.Ctx, "Error checking pending requests: %v", err)
		return nil, err
	}

	for _, pendingReq := range pendingRequests {
		if pendingReq.AgentID == uri.AgentID {
			log.Warn(sctx.Ctx, "Pending reinstatement request already exists for agent: %s", uri.AgentID)
			return nil, fmt.Errorf("a pending reinstatement request already exists for this agent")
		}
	}

	// Generate workflow ID if not provided
	workflowID := request.WorkflowID
	if workflowID == "" {
		workflowID = fmt.Sprintf("reinstatement-workflow-%s", uri.AgentID)
	}

	// Create reinstatement request - SINGLE database hit
	// Creates: reinstatement_request + audit_log
	reinstatementRequest, err := h.terminationRepo.CreateReinstatementRequest(
		sctx.Ctx,
		uri.AgentID,
		request.ReinstatementReason,
		request.RequestedBy,
		workflowID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create reinstatement request: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Reinstatement request created successfully: %s, request_id: %s",
		uri.AgentID, reinstatementRequest.ReinstatementID)

	// Initiate Temporal workflow for reinstatement approval
	// WF-AGT-PRF-011: Agent Reinstatement Workflow
	// This workflow handles: approval routing, notifications, status restoration, portal access, notifications
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "agent-profile-task-queue",
		// Long timeout for human approval (30 days)
		WorkflowExecutionTimeout: 30 * 24 * time.Hour,
	}

	workflowInput := workflows.ReinstatementWorkflowInput{
		ReinstatementID:     reinstatementRequest.ReinstatementID,
		AgentID:             uri.AgentID,
		ReinstatementReason: request.ReinstatementReason,
		RequestedBy:         request.RequestedBy,
	}

	_, err = h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.AgentReinstatementWorkflow, workflowInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start reinstatement workflow: %v", err)
		// Don't fail the request - the database is already updated
		// Workflow can be retried manually if needed
	} else {
		log.Info(sctx.Ctx, "Reinstatement workflow started: %s (waiting for approval)", workflowID)
	}

	return &resp.ReinstateAgentResponse{
		StatusCodeAndMessage: port.ReinstatementRequestCreated,
		Message:              "Reinstatement request created successfully",
		ReinstatementRequest: resp.ToReinstatementRequestDTO(*reinstatementRequest),
		NextSteps: []string{
			"Request is pending approval from management",
			"Manager will review the reinstatement request",
			"Agent will be notified of the approval decision",
			"If approved, agent status will be restored to ACTIVE",
			"Commission processing will be re-enabled upon approval",
		},
	}, nil
}

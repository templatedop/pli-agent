package handler

import (
	"fmt"
	"time"

	"pli-agent-api/core/domain"
	req "pli-agent-api/handler/request"
	resp "pli-agent-api/handler/response"
	repo "pli-agent-api/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// AgentSearchDashboardHandler handles agent search and dashboard APIs
// Phase 9: Search & Dashboard APIs
// AGT-022 to AGT-028, AGT-068, AGT-073, AGT-076, AGT-077
// FR-AGT-PRF-004: Agent Search
// FR-AGT-PRF-005: Profile Dashboard View
// FR-AGT-PRF-018: Agent Dashboard
// FR-AGT-PRF-022: Audit History
type AgentSearchDashboardHandler struct {
	*serverHandler.Base
	profileRepo      *repo.AgentProfileRepository
	auditLogRepo     *repo.AgentAuditLogRepository
	notificationRepo *repo.AgentNotificationRepository
	licenseRepo      *repo.AgentLicenseRepository
}

// NewAgentSearchDashboardHandler creates a new search and dashboard handler
func NewAgentSearchDashboardHandler(
	profileRepo *repo.AgentProfileRepository,
	auditLogRepo *repo.AgentAuditLogRepository,
	notificationRepo *repo.AgentNotificationRepository,
	licenseRepo *repo.AgentLicenseRepository,
) *AgentSearchDashboardHandler {
	return &AgentSearchDashboardHandler{
		Base:             &serverHandler.Base{},
		profileRepo:      profileRepo,
		auditLogRepo:     auditLogRepo,
		notificationRepo: notificationRepo,
		licenseRepo:      licenseRepo,
	}
}

// RegisterRoutes registers all search and dashboard routes
func (h *AgentSearchDashboardHandler) RegisterRoutes() []serverRoute.Route {
	return []serverRoute.Route{
		// AGT-022: Multi-criteria Agent Search
		serverRoute.NewRoute("GET", "/agents/search", h.SearchAgents),
		// AGT-023: Get Complete Agent Profile
		serverRoute.NewRoute("GET", "/agents/:agent_id", h.GetAgentProfile),
		// AGT-028: Get Audit History
		serverRoute.NewRoute("GET", "/agents/:agent_id/audit-history", h.GetAuditHistory),
		// AGT-068: Agent Dashboard
		serverRoute.NewRoute("GET", "/dashboard/agent/:agent_id", h.GetAgentDashboard),
		// AGT-073: Get Agent Hierarchy
		serverRoute.NewRoute("GET", "/agents/:agent_id/hierarchy", h.GetAgentHierarchy),
		// AGT-076: Agent Activity Timeline
		serverRoute.NewRoute("GET", "/agents/:agent_id/timeline", h.GetAgentTimeline),
		// AGT-077: Agent Notification History
		serverRoute.NewRoute("GET", "/agents/:agent_id/notifications", h.GetAgentNotifications),
	}
}

// SearchAgents performs multi-criteria agent search with pagination
// AGT-022: Multi-criteria Agent Search
// FR-AGT-PRF-004: Agent Search
// BR-AGT-PRF-022: Multi-Criteria Search
// OPTIMIZED: Single query with JOINs (no N+1 problem)
func (h *AgentSearchDashboardHandler) SearchAgents(
	sctx *serverRoute.Context,
	request req.SearchAgentsRequest,
) (*resp.SearchAgentsResponse, error) {
	log.Info(sctx.Ctx, "Searching agents with filters: %+v", request)

	// Default pagination
	page := 1
	if request.Page != nil {
		page = *request.Page
	}
	limit := 20
	if request.Limit != nil {
		limit = *request.Limit
	}

	// Build search filters
	filters := repo.AgentSearchFilters{
		AgentID:      request.AgentID,
		Name:         request.Name,
		PANNumber:    request.PANNumber,
		MobileNumber: request.MobileNumber,
		Status:       request.Status,
		OfficeCode:   request.OfficeCode,
		AgentType:    request.AgentType,
	}

	// Perform search - SINGLE database hit
	results, totalCount, err := h.profileRepo.Search(sctx.Ctx, filters, page, limit)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to search agents: %v", err)
		return nil, fmt.Errorf("failed to search agents: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (int(totalCount) + limit - 1) / limit

	return &resp.SearchAgentsResponse{
		Results: results,
		Pagination: domain.PaginationMetadata{
			CurrentPage:    page,
			TotalPages:     totalPages,
			TotalResults:   int(totalCount),
			ResultsPerPage: limit,
		},
	}, nil
}

// GetAgentProfile retrieves complete agent profile with all related entities
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
// OPTIMIZED: Single query using batch (4 sub-queries in one round trip)
func (h *AgentSearchDashboardHandler) GetAgentProfile(
	sctx *serverRoute.Context,
	uri AgentIDUri,
) (*resp.AgentProfileCompleteResponse, error) {
	log.Info(sctx.Ctx, "Getting complete profile for agent: %s", uri.AgentID)

	// Fetch profile with all related entities - SINGLE database round trip
	profile, addresses, contacts, emails, err := h.profileRepo.GetProfileWithRelatedEntities(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get agent profile: %v", err)
		return nil, fmt.Errorf("failed to get agent profile: %w", err)
	}

	// Fetch licenses (additional query - can be batched in future optimization)
	licenses, err := h.licenseRepo.FindByAgentID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Warn(sctx.Ctx, "Failed to get licenses for agent %s: %v", uri.AgentID, err)
		licenses = []domain.AgentLicense{} // Continue without licenses
	}

	// Calculate workflow state
	workflowInfo := calculateWorkflowInfo(profile)

	return &resp.AgentProfileCompleteResponse{
		AgentProfile: *profile,
		Addresses:    addresses,
		Contacts:     contacts,
		Emails:       emails,
		Licenses:     licenses,
		WorkflowInfo: workflowInfo,
	}, nil
}

// GetAuditHistory retrieves audit history with pagination and date filters
// AGT-028: Get Audit History
// FR-AGT-PRF-022: Profile Change History and Audit Trail
// BR-AGT-PRF-005: Audit Logging
// OPTIMIZED: Single query with pagination
func (h *AgentSearchDashboardHandler) GetAuditHistory(
	sctx *serverRoute.Context,
	uri AgentIDUri,
	request req.GetAuditHistoryRequest,
) (*resp.GetAuditHistoryResponse, error) {
	log.Info(sctx.Ctx, "Getting audit history for agent: %s", uri.AgentID)

	// Default pagination
	page := 1
	if request.Page != nil {
		page = *request.Page
	}
	limit := 50
	if request.Limit != nil {
		limit = *request.Limit
	}

	// Fetch audit logs - SINGLE database hit with pagination
	auditLogs, totalCount, err := h.auditLogRepo.GetHistory(
		sctx.Ctx,
		uri.AgentID,
		request.FromDate,
		request.ToDate,
		page,
		limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get audit history: %v", err)
		return nil, fmt.Errorf("failed to get audit history: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit

	return &resp.GetAuditHistoryResponse{
		AgentID:   uri.AgentID,
		AuditLogs: auditLogs,
		Pagination: domain.PaginationMetadata{
			CurrentPage:    page,
			TotalPages:     totalPages,
			TotalResults:   totalCount,
			ResultsPerPage: limit,
		},
	}, nil
}

// GetAgentDashboard retrieves agent dashboard with metrics and tasks
// AGT-068: Agent Dashboard
// FR-AGT-PRF-018: Agent Dashboard
// FR-AGT-PRF-021: Self-Service Update
func (h *AgentSearchDashboardHandler) GetAgentDashboard(
	sctx *serverRoute.Context,
	uri AgentIDUri,
) (*resp.AgentDashboardResponse, error) {
	log.Info(sctx.Ctx, "Getting dashboard for agent: %s", uri.AgentID)

	// Fetch agent profile
	profile, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get agent profile: %v", err)
		return nil, fmt.Errorf("failed to get agent profile: %w", err)
	}

	// Build profile summary
	profileSummary := domain.ProfileSummary{
		AgentID:   profile.AgentID,
		Name:      fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
		AgentType: profile.AgentType,
		Status:    profile.Status,
		PANNumber: profile.PANNumber,
	}

	// Fetch licenses for pending tasks
	licenses, err := h.licenseRepo.FindByAgentID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Warn(sctx.Ctx, "Failed to get licenses: %v", err)
		licenses = []domain.AgentLicense{}
	}

	// Calculate pending tasks
	pendingTasks := calculatePendingTasks(profile, licenses)

	// Fetch recent notifications
	recentNotifications, err := h.notificationRepo.GetRecentByAgentID(sctx.Ctx, uri.AgentID, 10)
	if err != nil {
		log.Warn(sctx.Ctx, "Failed to get recent notifications: %v", err)
		recentNotifications = []domain.AgentNotification{}
	}

	// TODO: Fetch performance metrics from external system
	// For now, return placeholder metrics
	performanceMetrics := domain.PerformanceMetrics{
		PoliciesSold:     0,
		PremiumCollected: 0.0,
	}
	performanceMetrics.TargetsAchieved.MonthlyTarget = 100
	performanceMetrics.TargetsAchieved.Achieved = 0
	performanceMetrics.TargetsAchieved.Percentage = 0.0

	return &resp.AgentDashboardResponse{
		AgentID:             uri.AgentID,
		ProfileSummary:      profileSummary,
		PerformanceMetrics:  performanceMetrics,
		PendingTasks:        pendingTasks,
		RecentNotifications: recentNotifications,
	}, nil
}

// GetAgentHierarchy retrieves agent's hierarchy chain
// AGT-073: Get Agent Hierarchy
// Phase 9: Search & Dashboard APIs
// OPTIMIZED: Single recursive CTE query
func (h *AgentSearchDashboardHandler) GetAgentHierarchy(
	sctx *serverRoute.Context,
	uri AgentIDUri,
) (*resp.AgentHierarchyResponse, error) {
	log.Info(sctx.Ctx, "Getting hierarchy for agent: %s", uri.AgentID)

	// Fetch agent profile to get agent type
	profile, err := h.profileRepo.FindByID(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get agent profile: %v", err)
		return nil, fmt.Errorf("failed to get agent profile: %w", err)
	}

	// Get hierarchy chain - SINGLE database hit with recursive CTE
	hierarchyChain, err := h.profileRepo.GetHierarchy(sctx.Ctx, uri.AgentID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get hierarchy: %v", err)
		return nil, fmt.Errorf("failed to get hierarchy: %w", err)
	}

	return &resp.AgentHierarchyResponse{
		AgentID:        uri.AgentID,
		AgentType:      profile.AgentType,
		HierarchyChain: hierarchyChain,
	}, nil
}

// GetAgentTimeline retrieves agent activity timeline with filters
// AGT-076: Agent Activity Timeline
// Phase 9: Search & Dashboard APIs
// OPTIMIZED: Single query with UNION combining multiple event sources
func (h *AgentSearchDashboardHandler) GetAgentTimeline(
	sctx *serverRoute.Context,
	uri AgentIDUri,
	request req.GetTimelineRequest,
) (*resp.AgentTimelineResponse, error) {
	log.Info(sctx.Ctx, "Getting timeline for agent: %s", uri.AgentID)

	// Default pagination
	page := 1
	if request.Page != nil {
		page = *request.Page
	}
	limit := 50
	if request.Limit != nil {
		limit = *request.Limit
	}

	// Get timeline events - SINGLE database hit with UNION
	timelineEvents, totalCount, err := h.auditLogRepo.GetTimeline(
		sctx.Ctx,
		uri.AgentID,
		request.ActivityType,
		request.FromDate,
		request.ToDate,
		page,
		limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get timeline: %v", err)
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit

	return &resp.AgentTimelineResponse{
		AgentID:  uri.AgentID,
		Timeline: timelineEvents,
		Pagination: domain.PaginationMetadata{
			CurrentPage:    page,
			TotalPages:     totalPages,
			TotalResults:   totalCount,
			ResultsPerPage: limit,
		},
	}, nil
}

// GetAgentNotifications retrieves agent notification history
// AGT-077: Agent Notification History
// Phase 9: Search & Dashboard APIs
// FR-AGT-PRF-021: Self-Service Update notifications
// OPTIMIZED: Single query with pagination
func (h *AgentSearchDashboardHandler) GetAgentNotifications(
	sctx *serverRoute.Context,
	uri AgentIDUri,
	request req.GetNotificationsRequest,
) (*resp.AgentNotificationsResponse, error) {
	log.Info(sctx.Ctx, "Getting notifications for agent: %s", uri.AgentID)

	// Default pagination
	page := 1
	if request.Page != nil {
		page = *request.Page
	}
	limit := 50
	if request.Limit != nil {
		limit = *request.Limit
	}

	// Fetch notifications - SINGLE database hit with pagination
	notifications, pagination, err := h.notificationRepo.GetByAgentID(
		sctx.Ctx,
		uri.AgentID,
		request.NotificationType,
		request.FromDate,
		request.ToDate,
		page,
		limit,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get notifications: %v", err)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return &resp.AgentNotificationsResponse{
		AgentID:       uri.AgentID,
		Notifications: notifications,
		Pagination:    *pagination,
	}, nil
}

// Helper: Calculate workflow info for profile
func calculateWorkflowInfo(profile *domain.AgentProfile) *domain.WorkflowInfo {
	// Calculate progress based on workflow state
	progressMap := map[string]float64{
		"DRAFT":            10.0,
		"PENDING_REVIEW":   30.0,
		"UNDER_REVIEW":     50.0,
		"PENDING_APPROVAL": 70.0,
		"APPROVED":         90.0,
		"ACTIVE":           100.0,
	}

	progress, exists := progressMap[profile.WorkflowState]
	if !exists {
		progress = 0.0
	}

	return &domain.WorkflowInfo{
		CurrentStep:        profile.WorkflowState,
		ProgressPercentage: progress,
	}
}

// Helper: Calculate pending tasks for dashboard
func calculatePendingTasks(profile *domain.AgentProfile, licenses []domain.AgentLicense) []domain.PendingTask {
	tasks := []domain.PendingTask{}
	now := time.Now()

	// Check for expiring licenses (within 30 days)
	for _, license := range licenses {
		if license.Status == domain.LicenseStatusActive && license.ExpiryDate.Time.After(now) {
			daysUntilExpiry := int(license.ExpiryDate.Time.Sub(now).Hours() / 24)
			if daysUntilExpiry <= 30 {
				priority := "LOW"
				if daysUntilExpiry <= 7 {
					priority = "HIGH"
				} else if daysUntilExpiry <= 15 {
					priority = "MEDIUM"
				}

				tasks = append(tasks, domain.PendingTask{
					Task:     fmt.Sprintf("License renewal pending for %s", license.LicenseType),
					Priority: priority,
					DueDate:  license.ExpiryDate.Time,
					Overdue:  false,
				})
			}
		}
	}

	// Check for incomplete profile fields
	if profile.WorkflowState != "ACTIVE" {
		tasks = append(tasks, domain.PendingTask{
			Task:     "Complete profile onboarding",
			Priority: "MEDIUM",
			DueDate:  time.Now().AddDate(0, 0, 7), // 7 days from now
			Overdue:  false,
		})
	}

	return tasks
}

package response

import (
	"pli-agent-api/core/domain"
)

// ========================================================================
// PHASE 9: SEARCH & DASHBOARD API RESPONSES
// ========================================================================

// SearchAgentsResponse for AGT-022
// FR-AGT-PRF-004: Multi-criteria agent search
// BR-AGT-PRF-022: Multi-Criteria Search
type SearchAgentsResponse struct {
	Results    []domain.AgentSearchResult `json:"results"`
	Pagination domain.PaginationMetadata  `json:"pagination"`
}

// AgentProfileCompleteResponse for AGT-023
// FR-AGT-PRF-005: Profile Dashboard View
// Returns complete agent profile with all related entities
type AgentProfileCompleteResponse struct {
	AgentProfile domain.AgentProfile     `json:"agent_profile"`
	Addresses    []domain.AgentAddress   `json:"addresses"`
	Contacts     []domain.AgentContact   `json:"contacts"`
	Emails       []domain.AgentEmail     `json:"emails"`
	Licenses     []domain.AgentLicense   `json:"licenses"`
	WorkflowInfo *domain.WorkflowInfo    `json:"workflow_info,omitempty"`
}

// GetAuditHistoryResponse for AGT-028
// FR-AGT-PRF-022: Profile Change History and Audit Trail
// BR-AGT-PRF-005: Audit Logging
type GetAuditHistoryResponse struct {
	AgentID    string                    `json:"agent_id"`
	AuditLogs  []domain.AgentAuditLog    `json:"audit_logs"`
	Pagination domain.PaginationMetadata `json:"pagination"`
}

// AgentDashboardResponse for AGT-068
// FR-AGT-PRF-018: Agent Dashboard
// FR-AGT-PRF-021: Self-Service Update
type AgentDashboardResponse struct {
	AgentID             string                      `json:"agent_id"`
	ProfileSummary      domain.ProfileSummary       `json:"profile_summary"`
	PerformanceMetrics  domain.PerformanceMetrics   `json:"performance_metrics"`
	PendingTasks        []domain.PendingTask        `json:"pending_tasks"`
	RecentNotifications []domain.AgentNotification  `json:"recent_notifications"`
}

// AgentHierarchyResponse for AGT-073
// Phase 9: Get agent hierarchy chain
type AgentHierarchyResponse struct {
	AgentID        string                  `json:"agent_id"`
	AgentType      string                  `json:"agent_type"`
	HierarchyChain []domain.HierarchyNode  `json:"hierarchy_chain"`
}

// AgentTimelineResponse for AGT-076
// Phase 9: Agent activity timeline
type AgentTimelineResponse struct {
	AgentID    string                    `json:"agent_id"`
	Timeline   []domain.TimelineEvent    `json:"timeline"`
	Pagination domain.PaginationMetadata `json:"pagination"`
}

// AgentNotificationsResponse for AGT-077
// Phase 9: Agent notification history
// FR-AGT-PRF-021: Self-Service Update notifications
type AgentNotificationsResponse struct {
	AgentID       string                     `json:"agent_id"`
	Notifications []domain.AgentNotification `json:"notifications"`
	Pagination    domain.PaginationMetadata  `json:"pagination"`
}

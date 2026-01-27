package handler

import (
	"errors"
	"testing"
	"time"

	"pli-agent-api/core/domain"
	req "pli-agent-api/handler/request"
	"pli-agent-api/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// TestSearchAgents_Success tests successful agent search
func TestSearchAgents_Success(t *testing.T) {
	// Setup
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	// Test data
	searchResults := []domain.AgentSearchResult{
		*testutil.CreateTestSearchResult(),
		*testutil.CreateTestSearchResult(),
	}

	// Mock expectations
	mockProfileRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
		Return(searchResults, int64(2), nil)

	// Test request
	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.SearchAgentsRequest{
		Status: stringPtr("ACTIVE"),
		Page:   intPtr(1),
		Limit:  intPtr(20),
	}

	// Execute
	response, err := handler.SearchAgents(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Results, 2)
	assert.Equal(t, 1, response.Pagination.CurrentPage)
	assert.Equal(t, 1, response.Pagination.TotalPages)
	assert.Equal(t, 2, response.Pagination.TotalResults)

	mockProfileRepo.AssertExpectations(t)
}

// TestSearchAgents_WithFilters tests search with multiple filters
func TestSearchAgents_WithFilters(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	searchResults := []domain.AgentSearchResult{*testutil.CreateTestSearchResult()}

	mockProfileRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
		Return(searchResults, int64(1), nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.SearchAgentsRequest{
		Status:     stringPtr("ACTIVE"),
		AgentType:  stringPtr("ADVISOR"),
		OfficeCode: stringPtr("OFF-001"),
		Name:       stringPtr("John"),
		Page:       intPtr(1),
		Limit:      intPtr(20),
	}

	response, err := handler.SearchAgents(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Results, 1)
	mockProfileRepo.AssertExpectations(t)
}

// TestSearchAgents_EmptyResults tests search with no results
func TestSearchAgents_EmptyResults(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	mockProfileRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
		Return([]domain.AgentSearchResult{}, int64(0), nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.SearchAgentsRequest{
		Page:  intPtr(1),
		Limit: intPtr(20),
	}

	response, err := handler.SearchAgents(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Empty(t, response.Results)
	assert.Equal(t, 0, response.Pagination.TotalResults)
	mockProfileRepo.AssertExpectations(t)
}

// TestSearchAgents_Error tests search with database error
func TestSearchAgents_Error(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	mockProfileRepo.On("Search", mock.Anything, mock.Anything, 1, 20).
		Return([]domain.AgentSearchResult{}, int64(0), errors.New("database error"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.SearchAgentsRequest{
		Page:  intPtr(1),
		Limit: intPtr(20),
	}

	response, err := handler.SearchAgents(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to search agents")
	mockProfileRepo.AssertExpectations(t)
}

// TestGetAgentProfile_Success tests successful profile retrieval
func TestGetAgentProfile_Success(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "test-agent-id"
	profile := testutil.CreateTestAgentProfile(agentID)
	addresses := []domain.AgentAddress{*testutil.CreateTestAgentAddress(agentID)}
	contacts := []domain.AgentContact{*testutil.CreateTestAgentContact(agentID)}
	emails := []domain.AgentEmail{*testutil.CreateTestAgentEmail(agentID)}
	licenses := []domain.AgentLicense{*testutil.CreateTestAgentLicense(agentID)}

	mockProfileRepo.On("GetProfileWithRelatedEntities", mock.Anything, agentID).
		Return(profile, addresses, contacts, emails, nil)
	mockLicenseRepo.On("FindByAgentID", mock.Anything, agentID).
		Return(licenses, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}

	response, err := handler.GetAgentProfile(ctx, uri)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, agentID, response.AgentProfile.AgentID)
	assert.Len(t, response.Addresses, 1)
	assert.Len(t, response.Contacts, 1)
	assert.Len(t, response.Emails, 1)
	assert.Len(t, response.Licenses, 1)
	assert.NotNil(t, response.WorkflowInfo)

	mockProfileRepo.AssertExpectations(t)
	mockLicenseRepo.AssertExpectations(t)
}

// TestGetAgentProfile_NotFound tests profile not found error
func TestGetAgentProfile_NotFound(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "non-existent-id"
	mockProfileRepo.On("GetProfileWithRelatedEntities", mock.Anything, agentID).
		Return(nil, nil, nil, nil, errors.New("agent not found"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}

	response, err := handler.GetAgentProfile(ctx, uri)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get agent profile")
	mockProfileRepo.AssertExpectations(t)
}

// TestGetAuditHistory_Success tests successful audit history retrieval
func TestGetAuditHistory_Success(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "test-agent-id"
	auditLogs := []domain.AgentAuditLog{
		*testutil.CreateTestAuditLog(agentID),
		*testutil.CreateTestAuditLog(agentID),
	}

	mockAuditLogRepo.On("GetHistory", mock.Anything, agentID, (*time.Time)(nil), (*time.Time)(nil), 1, 50).
		Return(auditLogs, 2, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}
	request := req.GetAuditHistoryRequest{}

	response, err := handler.GetAuditHistory(ctx, uri, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, agentID, response.AgentID)
	assert.Len(t, response.AuditLogs, 2)
	assert.Equal(t, 1, response.Pagination.CurrentPage)
	assert.Equal(t, 2, response.Pagination.TotalResults)

	mockAuditLogRepo.AssertExpectations(t)
}

// TestGetAgentHierarchy_Success tests successful hierarchy retrieval
func TestGetAgentHierarchy_Success(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "test-agent-id"
	profile := testutil.CreateTestAgentProfile(agentID)
	hierarchyChain := []domain.HierarchyNode{
		*testutil.CreateTestHierarchyNode(1),
		*testutil.CreateTestHierarchyNode(2),
	}

	mockProfileRepo.On("FindByID", mock.Anything, agentID).Return(profile, nil)
	mockProfileRepo.On("GetHierarchy", mock.Anything, agentID).Return(hierarchyChain, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}

	response, err := handler.GetAgentHierarchy(ctx, uri)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, agentID, response.AgentID)
	assert.Len(t, response.HierarchyChain, 2)
	assert.Equal(t, profile.AgentType, response.AgentType)

	mockProfileRepo.AssertExpectations(t)
}

// TestGetAgentTimeline_Success tests successful timeline retrieval
func TestGetAgentTimeline_Success(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "test-agent-id"
	timelineEvents := []domain.TimelineEvent{
		*testutil.CreateTestTimelineEvent(),
		*testutil.CreateTestTimelineEvent(),
	}

	mockAuditLogRepo.On("GetTimeline", mock.Anything, agentID, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 1, 50).
		Return(timelineEvents, 2, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}
	request := req.GetTimelineRequest{}

	response, err := handler.GetAgentTimeline(ctx, uri, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, agentID, response.AgentID)
	assert.Len(t, response.Timeline, 2)
	assert.Equal(t, 2, response.Pagination.TotalResults)

	mockAuditLogRepo.AssertExpectations(t)
}

// TestGetAgentNotifications_Success tests successful notifications retrieval
func TestGetAgentNotifications_Success(t *testing.T) {
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockAuditLogRepo := new(testutil.MockAuditLogRepository)
	mockNotificationRepo := new(testutil.MockNotificationRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)

	handler := NewAgentSearchDashboardHandler(
		mockProfileRepo,
		mockAuditLogRepo,
		mockNotificationRepo,
		mockLicenseRepo,
	)

	agentID := "test-agent-id"
	notifications := []domain.AgentNotification{
		*testutil.CreateTestNotification(agentID),
		*testutil.CreateTestNotification(agentID),
	}
	pagination := &domain.PaginationMetadata{
		CurrentPage:    1,
		TotalPages:     1,
		TotalResults:   2,
		ResultsPerPage: 50,
	}

	mockNotificationRepo.On("GetByAgentID", mock.Anything, agentID, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 1, 50).
		Return(notifications, pagination, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := AgentIDUri{AgentID: agentID}
	request := req.GetNotificationsRequest{}

	response, err := handler.GetAgentNotifications(ctx, uri, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, agentID, response.AgentID)
	assert.Len(t, response.Notifications, 2)
	assert.Equal(t, 2, response.Pagination.TotalResults)

	mockNotificationRepo.AssertExpectations(t)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

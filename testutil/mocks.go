package testutil

import (
	"context"
	"time"

	"pli-agent-api/core/domain"

	"github.com/stretchr/testify/mock"
)

// MockProfileRepository is a mock implementation of AgentProfileRepository
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) FindByID(ctx context.Context, agentID string) (*domain.AgentProfile, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AgentProfile), args.Error(1)
}

func (m *MockProfileRepository) Search(ctx context.Context, filters interface{}, page, limit int) ([]domain.AgentSearchResult, int64, error) {
	args := m.Called(ctx, filters, page, limit)
	return args.Get(0).([]domain.AgentSearchResult), args.Get(1).(int64), args.Error(2)
}

func (m *MockProfileRepository) GetProfileWithRelatedEntities(ctx context.Context, agentID string) (*domain.AgentProfile, []domain.AgentAddress, []domain.AgentContact, []domain.AgentEmail, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, nil, nil, nil, args.Error(4)
	}
	return args.Get(0).(*domain.AgentProfile),
		args.Get(1).([]domain.AgentAddress),
		args.Get(2).([]domain.AgentContact),
		args.Get(3).([]domain.AgentEmail),
		args.Error(4)
}

func (m *MockProfileRepository) GetHierarchy(ctx context.Context, agentID string) ([]domain.HierarchyNode, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).([]domain.HierarchyNode), args.Error(1)
}

// MockAuditLogRepository is a mock implementation of AgentAuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) GetHistory(ctx context.Context, agentID string, fromDate, toDate *time.Time, page, limit int) ([]domain.AgentAuditLog, int, error) {
	args := m.Called(ctx, agentID, fromDate, toDate, page, limit)
	return args.Get(0).([]domain.AgentAuditLog), args.Get(1).(int), args.Error(2)
}

func (m *MockAuditLogRepository) GetTimeline(ctx context.Context, agentID string, activityType *string, fromDate, toDate *time.Time, page, limit int) ([]domain.TimelineEvent, int, error) {
	args := m.Called(ctx, agentID, activityType, fromDate, toDate, page, limit)
	return args.Get(0).([]domain.TimelineEvent), args.Get(1).(int), args.Error(2)
}

// MockNotificationRepository is a mock implementation of AgentNotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) GetByAgentID(ctx context.Context, agentID string, notificationType *string, fromDate, toDate *time.Time, page, limit int) ([]domain.AgentNotification, *domain.PaginationMetadata, error) {
	args := m.Called(ctx, agentID, notificationType, fromDate, toDate, page, limit)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]domain.AgentNotification), args.Get(1).(*domain.PaginationMetadata), args.Error(2)
}

func (m *MockNotificationRepository) GetRecentByAgentID(ctx context.Context, agentID string, limit int) ([]domain.AgentNotification, error) {
	args := m.Called(ctx, agentID, limit)
	return args.Get(0).([]domain.AgentNotification), args.Error(1)
}

// MockLicenseRepository is a mock implementation of AgentLicenseRepository
type MockLicenseRepository struct {
	mock.Mock
}

func (m *MockLicenseRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentLicense, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).([]domain.AgentLicense), args.Error(1)
}

// MockExportRepository is a mock implementation of AgentExportRepository
type MockExportRepository struct {
	mock.Mock
}

func (m *MockExportRepository) CreateConfig(ctx context.Context, config *domain.AgentExportConfig) (*domain.AgentExportConfig, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AgentExportConfig), args.Error(1)
}

func (m *MockExportRepository) CreateJob(ctx context.Context, job *domain.AgentExportJob) (*domain.AgentExportJob, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AgentExportJob), args.Error(1)
}

func (m *MockExportRepository) GetJobByID(ctx context.Context, exportID string) (*domain.AgentExportJob, error) {
	args := m.Called(ctx, exportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AgentExportJob), args.Error(1)
}

func (m *MockExportRepository) GetConfigByID(ctx context.Context, configID string) (*domain.AgentExportConfig, error) {
	args := m.Called(ctx, configID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AgentExportConfig), args.Error(1)
}

func (m *MockExportRepository) EstimateRecordCount(ctx context.Context, filtersJSON string) (int, error) {
	args := m.Called(ctx, filtersJSON)
	return args.Get(0).(int), args.Error(1)
}

// MockWebhookRepository is a mock implementation of HRMSWebhookRepository
type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) CreateEvent(ctx context.Context, event *domain.HRMSWebhookEvent) (*domain.HRMSWebhookEvent, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HRMSWebhookEvent), args.Error(1)
}

func (m *MockWebhookRepository) UpdateEventStatus(ctx context.Context, eventID string, status string, processingResult, errorMessage *string) error {
	args := m.Called(ctx, eventID, status, processingResult, errorMessage)
	return args.Error(0)
}

// MockTemporalClient is a mock implementation of Temporal client
type MockTemporalClient struct {
	mock.Mock
}

func (m *MockTemporalClient) ExecuteWorkflow(ctx context.Context, options interface{}, workflow interface{}, args ...interface{}) (interface{}, error) {
	mockArgs := m.Called(ctx, options, workflow, args)
	return mockArgs.Get(0), mockArgs.Error(1)
}

func (m *MockTemporalClient) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	args := m.Called(ctx, workflowID, runID, signalName, arg)
	return args.Error(0)
}

func (m *MockTemporalClient) Close() {
	m.Called()
}

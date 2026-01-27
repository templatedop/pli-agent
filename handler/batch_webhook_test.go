package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// ========================================================================
// EXPORT CONFIGURATION TESTS (AGT-064)
// ========================================================================

func TestConfigureExport_Success(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	// Expected config result
	exportConfig := testutil.CreateTestExportConfig()
	exportConfig.EstimatedRecords = 150

	// Mock expectations
	mockExportRepo.On("EstimateRecordCount", mock.Anything, mock.Anything).
		Return(150, nil)
	mockExportRepo.On("CreateConfig", mock.Anything, mock.Anything).
		Return(exportConfig, nil)

	// Test request
	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ConfigureExportRequest{
		ExportName:   "Test Export",
		OutputFormat: "XLSX",
		Filters: map[string]interface{}{
			"status": "ACTIVE",
		},
		Fields:    []string{"agent_id", "name", "status"},
		CreatedBy: "admin-user",
	}

	// Execute
	response, err := handler.ConfigureExport(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, exportConfig.ExportConfigID, response.ExportConfigID)
	assert.Equal(t, "Test Export", response.ExportName)
	assert.Equal(t, 150, response.EstimatedRecords)
	assert.Greater(t, response.EstimatedTimeSeconds, 0)

	mockExportRepo.AssertExpectations(t)
}

func TestConfigureExport_EstimationError(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportConfig := testutil.CreateTestExportConfig()
	exportConfig.EstimatedRecords = 0

	// Mock estimation failure but config creation succeeds
	mockExportRepo.On("EstimateRecordCount", mock.Anything, mock.Anything).
		Return(0, errors.New("estimation error"))
	mockExportRepo.On("CreateConfig", mock.Anything, mock.Anything).
		Return(exportConfig, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ConfigureExportRequest{
		ExportName:   "Test Export",
		OutputFormat: "CSV",
		Filters:      map[string]interface{}{},
		Fields:       []string{"agent_id"},
		CreatedBy:    "admin-user",
	}

	// Execute - should still succeed with 0 estimated records
	response, err := handler.ConfigureExport(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, response.EstimatedRecords)

	mockExportRepo.AssertExpectations(t)
}

func TestConfigureExport_CreateConfigError(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	mockExportRepo.On("EstimateRecordCount", mock.Anything, mock.Anything).
		Return(100, nil)
	mockExportRepo.On("CreateConfig", mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ConfigureExportRequest{
		ExportName:   "Test Export",
		OutputFormat: "XLSX",
		Filters:      map[string]interface{}{},
		Fields:       []string{"agent_id"},
		CreatedBy:    "admin-user",
	}

	// Execute
	response, err := handler.ConfigureExport(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create export config")

	mockExportRepo.AssertExpectations(t)
}

// ========================================================================
// EXPORT EXECUTION TESTS (AGT-065)
// ========================================================================

func TestExecuteExport_Success(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportConfig := testutil.CreateTestExportConfig()
	exportJob := testutil.CreateTestExportJob()
	exportJob.Status = domain.ExportStatusInProgress

	mockExportRepo.On("GetConfigByID", mock.Anything, exportConfig.ExportConfigID).
		Return(exportConfig, nil)
	mockExportRepo.On("CreateJob", mock.Anything, mock.Anything).
		Return(exportJob, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ExecuteExportRequest{
		ExportConfigID: exportConfig.ExportConfigID,
		RequestedBy:    "admin-user",
	}

	// Execute
	response, err := handler.ExecuteExport(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, exportJob.ExportID, response.ExportID)
	assert.Equal(t, domain.ExportStatusInProgress, response.Status)
	assert.Contains(t, response.Message, "Export started")

	mockExportRepo.AssertExpectations(t)
}

func TestExecuteExport_ConfigNotFound(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	mockExportRepo.On("GetConfigByID", mock.Anything, "non-existent-id").
		Return(nil, errors.New("config not found"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ExecuteExportRequest{
		ExportConfigID: "non-existent-id",
		RequestedBy:    "admin-user",
	}

	// Execute
	response, err := handler.ExecuteExport(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "export config not found")

	mockExportRepo.AssertExpectations(t)
}

func TestExecuteExport_CreateJobError(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportConfig := testutil.CreateTestExportConfig()

	mockExportRepo.On("GetConfigByID", mock.Anything, exportConfig.ExportConfigID).
		Return(exportConfig, nil)
	mockExportRepo.On("CreateJob", mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.ExecuteExportRequest{
		ExportConfigID: exportConfig.ExportConfigID,
		RequestedBy:    "admin-user",
	}

	// Execute
	response, err := handler.ExecuteExport(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create export job")

	mockExportRepo.AssertExpectations(t)
}

// ========================================================================
// EXPORT STATUS TESTS (AGT-066)
// ========================================================================

func TestGetExportStatus_Success(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportJob := testutil.CreateTestExportJob()
	exportJob.Status = domain.ExportStatusInProgress
	exportJob.ProgressPercentage = 45
	exportJob.RecordsProcessed = 450
	exportJob.TotalRecords = 1000

	mockExportRepo.On("GetJobByID", mock.Anything, exportJob.ExportID).
		Return(exportJob, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: exportJob.ExportID}

	// Execute
	response, err := handler.GetExportStatus(ctx, uri)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, exportJob.ExportID, response.ExportID)
	assert.Equal(t, domain.ExportStatusInProgress, response.Status)
	assert.Equal(t, 45, response.ProgressPercentage)
	assert.Equal(t, 450, response.RecordsProcessed)
	assert.Equal(t, 1000, response.TotalRecords)

	mockExportRepo.AssertExpectations(t)
}

func TestGetExportStatus_CompletedWithFile(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportJob := testutil.CreateTestExportJob()
	exportJob.Status = domain.ExportStatusCompleted
	exportJob.ProgressPercentage = 100
	exportJob.RecordsProcessed = 1000
	exportJob.TotalRecords = 1000

	mockExportRepo.On("GetJobByID", mock.Anything, exportJob.ExportID).
		Return(exportJob, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: exportJob.ExportID}

	// Execute
	response, err := handler.GetExportStatus(ctx, uri)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, domain.ExportStatusCompleted, response.Status)
	assert.Equal(t, 100, response.ProgressPercentage)
	assert.NotNil(t, response.FileURL)
	assert.NotNil(t, response.CompletedAt)

	mockExportRepo.AssertExpectations(t)
}

func TestGetExportStatus_NotFound(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	mockExportRepo.On("GetJobByID", mock.Anything, "non-existent-id").
		Return(nil, errors.New("export job not found"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: "non-existent-id"}

	// Execute
	response, err := handler.GetExportStatus(ctx, uri)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "export job not found")

	mockExportRepo.AssertExpectations(t)
}

// ========================================================================
// EXPORT DOWNLOAD TESTS (AGT-067)
// ========================================================================

func TestDownloadExport_Success(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportJob := testutil.CreateTestExportJob()
	exportJob.Status = domain.ExportStatusCompleted

	mockExportRepo.On("GetJobByID", mock.Anything, exportJob.ExportID).
		Return(exportJob, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: exportJob.ExportID}

	// Execute
	response, err := handler.DownloadExport(ctx, uri)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.FileURL)
	assert.Contains(t, response.FileName, exportJob.ExportID)
	assert.Greater(t, response.FileSizeBytes, int64(0))

	mockExportRepo.AssertExpectations(t)
}

func TestDownloadExport_NotCompleted(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	exportJob := testutil.CreateTestExportJob()
	exportJob.Status = domain.ExportStatusInProgress

	mockExportRepo.On("GetJobByID", mock.Anything, exportJob.ExportID).
		Return(exportJob, nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: exportJob.ExportID}

	// Execute
	response, err := handler.DownloadExport(ctx, uri)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "export not completed yet")

	mockExportRepo.AssertExpectations(t)
}

func TestDownloadExport_NotFound(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	mockExportRepo.On("GetJobByID", mock.Anything, "non-existent-id").
		Return(nil, errors.New("export job not found"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	uri := ExportIDUri{ExportID: "non-existent-id"}

	// Execute
	response, err := handler.DownloadExport(ctx, uri)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "export job not found")

	mockExportRepo.AssertExpectations(t)
}

// ========================================================================
// HRMS WEBHOOK TESTS (AGT-078)
// ========================================================================

func TestHandleHRMSWebhook_Success(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	webhookEvent := testutil.CreateTestWebhookEvent()
	webhookEvent.Status = domain.WebhookStatusReceived

	mockWebhookRepo.On("CreateEvent", mock.Anything, mock.Anything).
		Return(webhookEvent, nil)
	mockWebhookRepo.On("UpdateEventStatus", mock.Anything, webhookEvent.EventID, domain.WebhookStatusProcessed, mock.Anything, mock.Anything).
		Return(nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.HRMSWebhookRequest{
		EventID:   webhookEvent.EventID,
		EventType: domain.WebhookEventEmployeeUpdated,
		EmployeeData: domain.HRMSEmployeeData{
			EmployeeID:   "EMP-001",
			EmployeeName: "John Doe",
			Department:   "Sales",
			Designation:  "Senior Advisor",
		},
		Signature: handler.generateTestSignature(t, domain.HRMSEmployeeData{
			EmployeeID:   "EMP-001",
			EmployeeName: "John Doe",
			Department:   "Sales",
			Designation:  "Senior Advisor",
		}),
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "PROCESSED", response.Status)
	assert.Equal(t, webhookEvent.EventID, response.EventID)

	mockWebhookRepo.AssertExpectations(t)
}

func TestHandleHRMSWebhook_InvalidSignature(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	request := req.HRMSWebhookRequest{
		EventID:   "test-event-id",
		EventType: domain.WebhookEventEmployeeUpdated,
		EmployeeData: domain.HRMSEmployeeData{
			EmployeeID:   "EMP-001",
			EmployeeName: "John Doe",
		},
		Signature: "invalid-signature-12345",
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid webhook signature")
}

func TestHandleHRMSWebhook_CreateEventError(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	mockWebhookRepo.On("CreateEvent", mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	employeeData := domain.HRMSEmployeeData{
		EmployeeID:   "EMP-001",
		EmployeeName: "John Doe",
	}
	request := req.HRMSWebhookRequest{
		EventID:      "test-event-id",
		EventType:    domain.WebhookEventEmployeeUpdated,
		EmployeeData: employeeData,
		Signature:    handler.generateTestSignature(t, employeeData),
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to store webhook event")

	mockWebhookRepo.AssertExpectations(t)
}

func TestHandleHRMSWebhook_EmployeeTerminated(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	webhookEvent := testutil.CreateTestWebhookEvent()
	webhookEvent.EventType = domain.WebhookEventEmployeeTerminated

	mockWebhookRepo.On("CreateEvent", mock.Anything, mock.Anything).
		Return(webhookEvent, nil)
	mockWebhookRepo.On("UpdateEventStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	terminationTime := time.Now()
	employeeData := domain.HRMSEmployeeData{
		EmployeeID:   "EMP-001",
		EmployeeName: "John Doe",
		TerminationDate: &terminationTime,
	}
	request := req.HRMSWebhookRequest{
		EventID:      webhookEvent.EventID,
		EventType:    domain.WebhookEventEmployeeTerminated,
		EmployeeData: employeeData,
		Signature:    handler.generateTestSignature(t, employeeData),
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "PROCESSED", response.Status)

	mockWebhookRepo.AssertExpectations(t)
}

func TestHandleHRMSWebhook_EmployeeTransferred(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	webhookEvent := testutil.CreateTestWebhookEvent()
	webhookEvent.EventType = domain.WebhookEventEmployeeTransferred

	mockWebhookRepo.On("CreateEvent", mock.Anything, mock.Anything).
		Return(webhookEvent, nil)
	mockWebhookRepo.On("UpdateEventStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	employeeData := domain.HRMSEmployeeData{
		EmployeeID:      "EMP-001",
		EmployeeName:    "John Doe",
		OfficeCode:      "OFF-NEW-001",
		TransferDate:    timePtr(time.Now()),
	}
	request := req.HRMSWebhookRequest{
		EventID:      webhookEvent.EventID,
		EventType:    domain.WebhookEventEmployeeTransferred,
		EmployeeData: employeeData,
		Signature:    handler.generateTestSignature(t, employeeData),
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "PROCESSED", response.Status)

	mockWebhookRepo.AssertExpectations(t)
}

func TestHandleHRMSWebhook_EmployeeCreated(t *testing.T) {
	mockExportRepo := new(testutil.MockExportRepository)
	mockWebhookRepo := new(testutil.MockWebhookRepository)
	mockProfileRepo := new(testutil.MockProfileRepository)
	mockLicenseRepo := new(testutil.MockLicenseRepository)
	mockTemporalClient := new(testutil.MockTemporalClient)

	handler := NewAgentBatchWebhookHandler(
		mockExportRepo,
		mockWebhookRepo,
		mockProfileRepo,
		mockLicenseRepo,
		mockTemporalClient,
	)

	webhookEvent := testutil.CreateTestWebhookEvent()
	webhookEvent.EventType = domain.WebhookEventEmployeeCreated

	mockWebhookRepo.On("CreateEvent", mock.Anything, mock.Anything).
		Return(webhookEvent, nil)
	mockWebhookRepo.On("UpdateEventStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	ctx := &serverRoute.Context{Ctx: testutil.TestContext(t)}
	employeeData := domain.HRMSEmployeeData{
		EmployeeID:   "EMP-NEW-001",
		EmployeeName: "Jane Smith",
		Department:   "Marketing",
	}
	request := req.HRMSWebhookRequest{
		EventID:      webhookEvent.EventID,
		EventType:    domain.WebhookEventEmployeeCreated,
		EmployeeData: employeeData,
		Signature:    handler.generateTestSignature(t, employeeData),
	}

	// Execute
	response, err := handler.HandleHRMSWebhook(ctx, request)

	// Assert - Employee created events are acknowledged but require manual profile creation
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "PROCESSED", response.Status)

	mockWebhookRepo.AssertExpectations(t)
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// Helper to generate valid HMAC signature for testing
func (h *AgentBatchWebhookHandler) generateTestSignature(t *testing.T, data domain.HRMSEmployeeData) string {
	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal employee data: %v", err)
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func timePtr(t time.Time) *time.Time {
	return &t
}

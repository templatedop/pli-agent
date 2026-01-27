package testutil

import (
	"database/sql"
	"time"

	"pli-agent-api/core/domain"

	"github.com/google/uuid"
)

// Test fixtures for creating consistent test data across all tests

// CreateTestAgentProfile creates a test agent profile with default values
func CreateTestAgentProfile(agentID string) *domain.AgentProfile {
	if agentID == "" {
		agentID = uuid.New().String()
	}

	return &domain.AgentProfile{
		AgentID:       agentID,
		AgentCode:     sql.NullString{String: "AGT-TEST-001", Valid: true},
		AgentType:     "ADVISOR",
		EmployeeID:    sql.NullString{String: "EMP12345", Valid: true},
		OfficeCode:    "OFF-001",
		FirstName:     "John",
		MiddleName:    sql.NullString{String: "M", Valid: true},
		LastName:      "Doe",
		Gender:        "MALE",
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		PANNumber:     "ABCDE1234F",
		Status:        "ACTIVE",
		StatusDate:    time.Now(),
		WorkflowState: "ACTIVE",
		CreatedAt:     time.Now(),
		CreatedBy:     "test-user",
	}
}

// CreateTestAgentAddress creates a test agent address
func CreateTestAgentAddress(agentID string) *domain.AgentAddress {
	return &domain.AgentAddress{
		AddressID:   uuid.New().String(),
		AgentID:     agentID,
		AddressType: "PERMANENT",
		Line1:       "123 Main St",
		Line2:       sql.NullString{String: "Apt 4B", Valid: true},
		City:        "Mumbai",
		State:       "Maharashtra",
		Pincode:     "400001",
		Country:     "India",
		IsPrimary:   true,
		CreatedAt:   time.Now(),
	}
}

// CreateTestAgentContact creates a test agent contact
func CreateTestAgentContact(agentID string) *domain.AgentContact {
	return &domain.AgentContact{
		ContactID:     uuid.New().String(),
		AgentID:       agentID,
		ContactType:   "PERSONAL",
		MobileNumber:  "9876543210",
		IsPrimary:     true,
		IsVerified:    true,
		CreatedAt:     time.Now(),
	}
}

// CreateTestAgentEmail creates a test agent email
func CreateTestAgentEmail(agentID string) *domain.AgentEmail {
	return &domain.AgentEmail{
		EmailRecordID: uuid.New().String(),
		AgentID:       agentID,
		EmailType:     "PERSONAL",
		EmailID:       "john.doe@example.com",
		IsPrimary:     true,
		IsVerified:    true,
		CreatedAt:     time.Now(),
	}
}

// CreateTestAgentLicense creates a test agent license
func CreateTestAgentLicense(agentID string) *domain.AgentLicense {
	return &domain.AgentLicense{
		LicenseID:   uuid.New().String(),
		AgentID:     agentID,
		LicenseType: "LIFE_INSURANCE",
		LicenseNumber: sql.NullString{String: "LIC123456", Valid: true},
		IssueDate:   sql.NullTime{Time: time.Now().AddDate(-1, 0, 0), Valid: true},
		ExpiryDate:  sql.NullTime{Time: time.Now().AddDate(1, 0, 0), Valid: true},
		Status:      "ACTIVE",
		CreatedAt:   time.Now(),
	}
}

// CreateTestAuditLog creates a test audit log entry
func CreateTestAuditLog(agentID string) *domain.AgentAuditLog {
	return &domain.AgentAuditLog{
		AuditID:     uuid.New().String(),
		AgentID:     agentID,
		ActionType:  "PROFILE_UPDATE",
		FieldName:   sql.NullString{String: "mobile_number", Valid: true},
		OldValue:    sql.NullString{String: "9876543210", Valid: true},
		NewValue:    sql.NullString{String: "9876543211", Valid: true},
		PerformedBy: "test-admin",
		PerformedAt: time.Now(),
		CreatedAt:   time.Now(),
	}
}

// CreateTestNotification creates a test notification
func CreateTestNotification(agentID string) *domain.AgentNotification {
	return &domain.AgentNotification{
		NotificationID:   uuid.New().String(),
		AgentID:          agentID,
		NotificationType: "EMAIL",
		Template:         "WELCOME_EMAIL",
		Recipient:        "john.doe@example.com",
		Subject:          sql.NullString{String: "Welcome to Agent Portal", Valid: true},
		SentAt:           time.Now(),
		Status:           "SENT",
		CreatedAt:        time.Now(),
	}
}

// CreateTestExportConfig creates a test export configuration
func CreateTestExportConfig() *domain.AgentExportConfig {
	return &domain.AgentExportConfig{
		ExportConfigID:   uuid.New().String(),
		ExportName:       "Test Export",
		Filters:          sql.NullString{String: `{"status":"ACTIVE"}`, Valid: true},
		Fields:           sql.NullString{String: `["agent_id","name","status"]`, Valid: true},
		OutputFormat:     "EXCEL",
		EstimatedRecords: 100,
		CreatedBy:        "test-user",
		CreatedAt:        time.Now(),
	}
}

// CreateTestExportJob creates a test export job
func CreateTestExportJob() *domain.AgentExportJob {
	config := CreateTestExportConfig()
	return &domain.AgentExportJob{
		ExportID:           uuid.New().String(),
		ExportConfigID:     config.ExportConfigID,
		RequestedBy:        "test-user",
		Status:             "IN_PROGRESS",
		ProgressPercentage: 0,
		RecordsProcessed:   0,
		TotalRecords:       100,
		WorkflowID:         sql.NullString{String: "workflow-123", Valid: true},
		FileURL:            sql.NullString{String: "https://storage.example.com/exports/test.xlsx", Valid: true},
		FileSizeBytes:      sql.NullInt64{Int64: 1024000, Valid: true},
		StartedAt:          time.Now(),
	}
}

// CreateTestWebhookEvent creates a test webhook event
func CreateTestWebhookEvent() *domain.HRMSWebhookEvent {
	return &domain.HRMSWebhookEvent{
		EventID:        uuid.New().String(),
		EventType:      "EMPLOYEE_UPDATED",
		EmployeeID:     "EMP12345",
		EmployeeData:   sql.NullString{String: `{"employee_id":"EMP12345","name":"John Doe"}`, Valid: true},
		Signature:      "test-signature",
		SignatureValid: true,
		ReceivedAt:     time.Now(),
		Status:         "RECEIVED",
		RetryCount:     0,
	}
}

// CreateTestSearchResult creates a test search result
func CreateTestSearchResult() *domain.AgentSearchResult {
	return &domain.AgentSearchResult{
		AgentID:      uuid.New().String(),
		AgentCode:    sql.NullString{String: "AGT-TEST-001", Valid: true},
		Name:         "John Doe",
		AgentType:    "ADVISOR",
		PANNumber:    "ABCDE1234F",
		MobileNumber: sql.NullString{String: "9876543210", Valid: true},
		EmailAddress: sql.NullString{String: "john@example.com", Valid: true},
		Status:       "ACTIVE",
		OfficeCode:   "OFF-001",
		CreatedAt:    time.Now(),
	}
}

// CreateTestHierarchyNode creates a test hierarchy node
func CreateTestHierarchyNode(level int) *domain.HierarchyNode {
	return &domain.HierarchyNode{
		AgentID:   uuid.New().String(),
		AgentCode: sql.NullString{String: "AGT-TEST-001", Valid: true},
		Name:      "John Doe",
		AgentType: "ADVISOR",
		Level:     level,
	}
}

// CreateTestTimelineEvent creates a test timeline event
func CreateTestTimelineEvent() *domain.TimelineEvent {
	return &domain.TimelineEvent{
		Timestamp:   time.Now(),
		EventType:   "PROFILE_CHANGE",
		Description: "Updated mobile number",
		PerformedBy: sql.NullString{String: "test-admin", Valid: true},
	}
}

// CreateMultipleTestProfiles creates multiple test profiles
func CreateMultipleTestProfiles(count int) []*domain.AgentProfile {
	profiles := make([]*domain.AgentProfile, count)
	for i := 0; i < count; i++ {
		profile := CreateTestAgentProfile("")
		profile.FirstName = "Agent"
		profile.LastName = string(rune('A' + i))
		profiles[i] = profile
	}
	return profiles
}

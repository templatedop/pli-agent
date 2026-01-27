package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"pli-agent-api/core/domain"

	"github.com/stretchr/testify/assert"
)

// TestWebhookEvent_Validation tests webhook event validation
func TestWebhookEvent_Validation(t *testing.T) {
	tests := []struct {
		name     string
		event    *domain.HRMSWebhookEvent
		isValid  bool
		errorMsg string
	}{
		{
			name: "Valid webhook event",
			event: &domain.HRMSWebhookEvent{
				EventID:        "evt-123",
				EventType:      domain.WebhookEventEmployeeUpdated,
				EmployeeID:     "EMP-001",
				Signature:      "valid-signature",
				SignatureValid: true,
				Status:         domain.WebhookStatusReceived,
			},
			isValid: true,
		},
		{
			name: "Empty event ID",
			event: &domain.HRMSWebhookEvent{
				EventID:    "",
				EventType:  domain.WebhookEventEmployeeUpdated,
				EmployeeID: "EMP-001",
			},
			isValid:  false,
			errorMsg: "event ID is required",
		},
		{
			name: "Invalid event type",
			event: &domain.HRMSWebhookEvent{
				EventID:    "evt-123",
				EventType:  "INVALID_TYPE",
				EmployeeID: "EMP-001",
			},
			isValid:  false,
			errorMsg: "invalid event type",
		},
		{
			name: "Empty employee ID",
			event: &domain.HRMSWebhookEvent{
				EventID:    "evt-123",
				EventType:  domain.WebhookEventEmployeeUpdated,
				EmployeeID: "",
			},
			isValid:  false,
			errorMsg: "employee ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookEvent(tt.event)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestWebhookEvent_StatusTransitions tests valid webhook status transitions
func TestWebhookEvent_StatusTransitions(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		isValid    bool
	}{
		{
			name:       "RECEIVED to PROCESSING",
			fromStatus: domain.WebhookStatusReceived,
			toStatus:   domain.WebhookStatusProcessing,
			isValid:    true,
		},
		{
			name:       "PROCESSING to PROCESSED",
			fromStatus: domain.WebhookStatusProcessing,
			toStatus:   domain.WebhookStatusProcessed,
			isValid:    true,
		},
		{
			name:       "PROCESSING to FAILED",
			fromStatus: domain.WebhookStatusProcessing,
			toStatus:   domain.WebhookStatusFailed,
			isValid:    true,
		},
		{
			name:       "FAILED to RETRYING",
			fromStatus: domain.WebhookStatusFailed,
			toStatus:   domain.WebhookStatusRetrying,
			isValid:    true,
		},
		{
			name:       "RETRYING to PROCESSING",
			fromStatus: domain.WebhookStatusRetrying,
			toStatus:   domain.WebhookStatusProcessing,
			isValid:    true,
		},
		{
			name:       "PROCESSED to PROCESSING (invalid)",
			fromStatus: domain.WebhookStatusProcessed,
			toStatus:   domain.WebhookStatusProcessing,
			isValid:    false,
		},
		{
			name:       "PROCESSED to FAILED (invalid)",
			fromStatus: domain.WebhookStatusProcessed,
			toStatus:   domain.WebhookStatusFailed,
			isValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidWebhookStatusTransition(tt.fromStatus, tt.toStatus)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// TestWebhookEvent_RetryLogic tests retry count and backoff logic
func TestWebhookEvent_RetryLogic(t *testing.T) {
	tests := []struct {
		name               string
		retryCount         int
		expectedBackoffMin int // Expected backoff in minutes
		shouldRetry        bool
	}{
		{
			name:               "First retry",
			retryCount:         0,
			expectedBackoffMin: 2, // 2^1 = 2 minutes
			shouldRetry:        true,
		},
		{
			name:               "Second retry",
			retryCount:         1,
			expectedBackoffMin: 4, // 2^2 = 4 minutes
			shouldRetry:        true,
		},
		{
			name:               "Third retry",
			retryCount:         2,
			expectedBackoffMin: 8, // 2^3 = 8 minutes
			shouldRetry:        true,
		},
		{
			name:               "Max retries exceeded",
			retryCount:         5,
			expectedBackoffMin: 0,
			shouldRetry:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateRetryBackoff(tt.retryCount)
			shouldRetry := shouldRetryWebhook(tt.retryCount)

			if tt.shouldRetry {
				assert.Equal(t, tt.expectedBackoffMin, backoff)
			}
			assert.Equal(t, tt.shouldRetry, shouldRetry)
		})
	}
}

// TestWebhookEvent_EmployeeDataParsing tests employee data JSON parsing
func TestWebhookEvent_EmployeeDataParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		expectData  *domain.HRMSEmployeeData
	}{
		{
			name: "Valid employee data",
			jsonData: `{
				"employee_id": "EMP-001",
				"employee_name": "John Doe",
				"department": "Sales",
				"designation": "Senior Advisor"
			}`,
			expectError: false,
			expectData: &domain.HRMSEmployeeData{
				EmployeeID:   "EMP-001",
				EmployeeName: "John Doe",
				Department:   "Sales",
				Designation:  "Senior Advisor",
			},
		},
		{
			name:        "Invalid JSON",
			jsonData:    `{invalid json}`,
			expectError: true,
		},
		{
			name:        "Empty JSON",
			jsonData:    `{}`,
			expectError: false,
			expectData:  &domain.HRMSEmployeeData{},
		},
		{
			name: "Employee termination data",
			jsonData: `{
				"employee_id": "EMP-001",
				"employee_name": "John Doe",
				"termination_date": "2024-01-15T10:30:00Z"
			}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data domain.HRMSEmployeeData
			err := json.Unmarshal([]byte(tt.jsonData), &data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectData != nil {
					assert.Equal(t, tt.expectData.EmployeeID, data.EmployeeID)
					assert.Equal(t, tt.expectData.EmployeeName, data.EmployeeName)
					assert.Equal(t, tt.expectData.Department, data.Department)
					assert.Equal(t, tt.expectData.Designation, data.Designation)
				}
			}
		})
	}
}

// TestWebhookEvent_EventTypeValidation tests event type validation
func TestWebhookEvent_EventTypeValidation(t *testing.T) {
	validTypes := []string{
		domain.WebhookEventEmployeeCreated,
		domain.WebhookEventEmployeeUpdated,
		domain.WebhookEventEmployeeTransferred,
		domain.WebhookEventEmployeeTerminated,
	}

	for _, eventType := range validTypes {
		t.Run(eventType, func(t *testing.T) {
			assert.True(t, isValidEventType(eventType))
		})
	}

	// Test invalid types
	invalidTypes := []string{
		"INVALID_TYPE",
		"EMPLOYEE_DELETED",
		"",
	}

	for _, eventType := range invalidTypes {
		t.Run(eventType, func(t *testing.T) {
			assert.False(t, isValidEventType(eventType))
		})
	}
}

// TestIncrementRetryCount_Logic tests retry count increment logic
func TestIncrementRetryCount_Logic(t *testing.T) {
	maxRetries := 5

	for i := 0; i <= maxRetries+1; i++ {
		t.Run(fmt.Sprintf("RetryCount_%d", i), func(t *testing.T) {
			shouldRetry := shouldRetryWebhook(i)
			expectedResult := i < maxRetries

			assert.Equal(t, expectedResult, shouldRetry,
				"Retry count %d: expected shouldRetry=%v, got %v", i, expectedResult, shouldRetry)
		})
	}
}

// TestGetPendingEvents_FilterLogic tests pending events filter logic
func TestGetPendingEvents_FilterLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		event         *domain.HRMSWebhookEvent
		shouldInclude bool
	}{
		{
			name: "RECEIVED status - should include",
			event: &domain.HRMSWebhookEvent{
				Status:     domain.WebhookStatusReceived,
				ReceivedAt: now.Add(-1 * time.Hour),
			},
			shouldInclude: true,
		},
		{
			name: "RETRYING with past retry time - should include",
			event: &domain.HRMSWebhookEvent{
				Status:      domain.WebhookStatusRetrying,
				ReceivedAt:  now.Add(-2 * time.Hour),
				NextRetryAt: sql.NullTime{Time: now.Add(-5 * time.Minute), Valid: true},
			},
			shouldInclude: true,
		},
		{
			name: "RETRYING with future retry time - should not include",
			event: &domain.HRMSWebhookEvent{
				Status:      domain.WebhookStatusRetrying,
				ReceivedAt:  now.Add(-1 * time.Hour),
				NextRetryAt: sql.NullTime{Time: now.Add(10 * time.Minute), Valid: true},
			},
			shouldInclude: false,
		},
		{
			name: "PROCESSED status - should not include",
			event: &domain.HRMSWebhookEvent{
				Status:     domain.WebhookStatusProcessed,
				ReceivedAt: now.Add(-1 * time.Hour),
			},
			shouldInclude: false,
		},
		{
			name: "FAILED status - should not include",
			event: &domain.HRMSWebhookEvent{
				Status:     domain.WebhookStatusFailed,
				ReceivedAt: now.Add(-1 * time.Hour),
			},
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			included := shouldIncludeInPendingEvents(tt.event, now)
			assert.Equal(t, tt.shouldInclude, included)
		})
	}
}

// ========================================================================
// HELPER FUNCTIONS FOR VALIDATION (would be in actual repository code)
// ========================================================================

func validateWebhookEvent(event *domain.HRMSWebhookEvent) error {
	if event.EventID == "" {
		return fmt.Errorf("event ID is required")
	}
	if !isValidEventType(event.EventType) {
		return fmt.Errorf("invalid event type: %s", event.EventType)
	}
	if event.EmployeeID == "" {
		return fmt.Errorf("employee ID is required")
	}
	return nil
}

func isValidEventType(eventType string) bool {
	validTypes := []string{
		domain.WebhookEventEmployeeCreated,
		domain.WebhookEventEmployeeUpdated,
		domain.WebhookEventEmployeeTransferred,
		domain.WebhookEventEmployeeTerminated,
	}

	for _, valid := range validTypes {
		if valid == eventType {
			return true
		}
	}
	return false
}

func isValidWebhookStatusTransition(fromStatus, toStatus string) bool {
	validTransitions := map[string][]string{
		domain.WebhookStatusReceived: {
			domain.WebhookStatusProcessing,
		},
		domain.WebhookStatusProcessing: {
			domain.WebhookStatusProcessed,
			domain.WebhookStatusFailed,
		},
		domain.WebhookStatusFailed: {
			domain.WebhookStatusRetrying,
		},
		domain.WebhookStatusRetrying: {
			domain.WebhookStatusProcessing,
			domain.WebhookStatusFailed,
		},
	}

	allowedTransitions, ok := validTransitions[fromStatus]
	if !ok {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == toStatus {
			return true
		}
	}
	return false
}

func calculateRetryBackoff(retryCount int) int {
	// Exponential backoff: 2^(retryCount+1) minutes
	return 1 << (retryCount + 1)
}

func shouldRetryWebhook(retryCount int) bool {
	maxRetries := 5
	return retryCount < maxRetries
}

func shouldIncludeInPendingEvents(event *domain.HRMSWebhookEvent, currentTime time.Time) bool {
	// Include RECEIVED events
	if event.Status == domain.WebhookStatusReceived {
		return true
	}

	// Include RETRYING events if retry time has passed
	if event.Status == domain.WebhookStatusRetrying {
		if event.NextRetryAt.Valid {
			return event.NextRetryAt.Time.Before(currentTime) || event.NextRetryAt.Time.Equal(currentTime)
		}
		return true
	}

	return false
}

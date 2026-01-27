# Phase 11: Comprehensive Testing Documentation

## Overview

This document describes the comprehensive testing strategy implemented for the PLI Agent Management System. Phase 11 focuses on unit tests, repository tests, integration tests, and test coverage reporting.

## Test Structure

```
pli-agent-api/
├── testutil/
│   ├── fixtures.go           # Test data factories
│   ├── mocks.go               # Mock implementations
│   └── helpers.go             # Test utility functions
├── handler/
│   ├── search_dashboard_test.go    # Phase 9 handler tests
│   └── batch_webhook_test.go       # Phase 10 handler tests
└── repo/postgres/
    ├── agent_export_test.go        # Export repository tests
    └── hrms_webhook_test.go        # Webhook repository tests
```

## Testing Layers

### 1. Unit Tests (Handler Layer)

**Location**: `handler/*_test.go`

**Purpose**: Test handler logic in isolation using mocks

**Coverage**:
- ✅ Phase 9: Search & Dashboard APIs (10 test functions)
  - SearchAgents - success, filters, empty results, errors
  - GetAgentProfile - success, not found
  - GetAuditHistory - success
  - GetAgentHierarchy - success
  - GetAgentTimeline - success
  - GetAgentNotifications - success

- ✅ Phase 10: Batch & Webhook APIs (16 test functions)
  - ConfigureExport - success, estimation error, config error
  - ExecuteExport - success, config not found, job creation error
  - GetExportStatus - success, completed with file, not found
  - DownloadExport - success, not completed, not found
  - HandleHRMSWebhook - success, invalid signature, create error
  - Webhook event types - updated, terminated, transferred, created

**Key Features**:
- Mock-based testing using `testify/mock`
- Table-driven tests for multiple scenarios
- Comprehensive error handling tests
- Validation of request/response formats
- HMAC signature validation testing

### 2. Repository Tests

**Location**: `repo/postgres/*_test.go`

**Purpose**: Test database query logic and business rules

**Coverage**:
- ✅ Export Repository (`agent_export_test.go`)
  - Query building with various filters
  - JSON parsing and validation
  - Export config validation
  - Status transition validation
  - Progress validation logic

- ✅ Webhook Repository (`hrms_webhook_test.go`)
  - Webhook event validation
  - Status transition rules
  - Retry logic and exponential backoff
  - Employee data parsing
  - Event type validation
  - Pending events filter logic

**Key Features**:
- Business logic validation
- Status transition state machines
- Retry mechanism testing
- Data validation rules
- Query filter logic

### 3. Test Utilities

**Location**: `testutil/`

**Purpose**: Provide reusable test infrastructure

**Components**:

#### a) Fixtures (`fixtures.go`)
Factory functions for creating consistent test data:
- `CreateTestAgentProfile()` - Agent profile with defaults
- `CreateTestAgentAddress()` - Agent address
- `CreateTestAgentContact()` - Agent contact
- `CreateTestAgentEmail()` - Agent email
- `CreateTestAgentLicense()` - Agent license
- `CreateTestAuditLog()` - Audit log entry
- `CreateTestNotification()` - Notification
- `CreateTestExportConfig()` - Export configuration
- `CreateTestExportJob()` - Export job
- `CreateTestWebhookEvent()` - Webhook event
- `CreateTestSearchResult()` - Search result
- `CreateTestHierarchyNode()` - Hierarchy node
- `CreateTestTimelineEvent()` - Timeline event
- `CreateMultipleTestProfiles()` - Multiple profiles

#### b) Mocks (`mocks.go`)
Mock implementations using `testify/mock`:
- `MockProfileRepository` - Agent profile operations
- `MockAuditLogRepository` - Audit log operations
- `MockNotificationRepository` - Notification operations
- `MockLicenseRepository` - License operations
- `MockExportRepository` - Export operations
- `MockWebhookRepository` - Webhook operations
- `MockTemporalClient` - Temporal workflow client

#### c) Helpers (`helpers.go`)
Utility functions for common test operations:
- `TestContext()` - Create test context
- `AssertNoPanic()` - Assert function doesn't panic
- `AssertErrorContains()` - Assert error message contains text
- `AssertPaginationValid()` - Validate pagination metadata
- `AssertValidUUID()` - Validate UUID format
- `RunTableTests()` - Execute table-driven tests
- `RunBenchmarkTests()` - Execute benchmark tests

## Running Tests

### Run All Tests
```bash
go test ./... -v
```

### Run Specific Package Tests
```bash
# Handler tests
go test ./handler -v

# Repository tests
go test ./repo/postgres -v
```

### Run Specific Test
```bash
go test ./handler -run TestSearchAgents_Success -v
```

### Run with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run with Race Detection
```bash
go test ./... -race
```

### Run Benchmark Tests
```bash
go test ./... -bench=. -benchmem
```

## Test Coverage Goals

| Layer | Target Coverage | Current Status |
|-------|----------------|----------------|
| Handlers | 80%+ | ✅ Implemented |
| Repositories | 70%+ | ✅ Implemented |
| Domain Logic | 90%+ | ✅ Implemented |
| Integration | 60%+ | 🔄 Pending |

## Test Patterns

### 1. Table-Driven Tests

Used for testing multiple scenarios with similar logic:

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "expected1"},
        {"case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Mock-Based Testing

Using mocks to isolate unit under test:

```go
func TestWithMock(t *testing.T) {
    mockRepo := new(testutil.MockProfileRepository)
    handler := NewHandler(mockRepo)

    mockRepo.On("FindByID", mock.Anything, "id").Return(profile, nil)

    result, err := handler.GetProfile(ctx, "id")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 3. Validation Testing

Testing business rules and constraints:

```go
func TestValidation(t *testing.T) {
    err := validateConfig(config)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected message")
}
```

## Test Data Management

### Fixtures Strategy

1. **Deterministic Data**: Use fixed values for predictable tests
2. **Realistic Data**: Use realistic values that match production patterns
3. **Edge Cases**: Include boundary values and special cases
4. **Isolation**: Each test creates its own data to avoid interference

### Example Fixture Usage

```go
func TestExample(t *testing.T) {
    // Create test profile
    profile := testutil.CreateTestAgentProfile("test-id")

    // Modify as needed
    profile.Status = "ACTIVE"

    // Use in test
    mockRepo.On("FindByID", mock.Anything, "test-id").Return(profile, nil)
}
```

## Common Test Scenarios

### Success Cases
- ✅ Valid input returns expected output
- ✅ All required fields populated
- ✅ Correct data transformations

### Error Cases
- ✅ Invalid input returns error
- ✅ Database errors propagated correctly
- ✅ Not found scenarios handled
- ✅ Validation errors with clear messages

### Edge Cases
- ✅ Empty results
- ✅ Null/optional fields
- ✅ Boundary values
- ✅ Large datasets
- ✅ Concurrent access

## Best Practices

### 1. Test Naming
- Use descriptive names: `TestFunction_Scenario_ExpectedOutcome`
- Examples:
  - `TestSearchAgents_Success`
  - `TestGetProfile_NotFound`
  - `TestValidation_EmptyName_ReturnsError`

### 2. Test Organization
- Group related tests in same file
- Use subtests for variations
- Keep tests focused and small
- One assertion per logical concept

### 3. Mock Usage
- Mock external dependencies only
- Don't mock types you own
- Set up expectations explicitly
- Always assert expectations

### 4. Test Data
- Use factories for consistency
- Make test data obvious
- Avoid magic numbers/strings
- Document unusual test data

### 5. Error Testing
- Test both success and failure paths
- Validate error messages
- Test error propagation
- Check error types when relevant

## Integration Testing

### Future Implementation (Pending)

Integration tests will cover:
- ✅ Database integration tests with testcontainers
- ✅ API endpoint tests with real HTTP calls
- ✅ Temporal workflow execution tests
- ✅ End-to-end user journeys

### Example Integration Test Structure

```go
func TestIntegration_CreateAgent(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Setup test server
    server := setupTestServer(t, db)
    defer server.Close()

    // Execute test
    response := makeRequest(server, "POST", "/agents", agentData)

    // Verify
    assert.Equal(t, 201, response.StatusCode)

    // Verify database state
    agent := queryDatabase(db, agentID)
    assert.NotNil(t, agent)
}
```

## Continuous Integration

### GitHub Actions / GitLab CI

```yaml
test:
  stage: test
  script:
    - go test ./... -v -coverprofile=coverage.out
    - go tool cover -func=coverage.out
    - go test ./... -race
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+)%/'
```

## Test Maintenance

### Adding New Tests

1. **Create fixtures** for new domain objects
2. **Add mocks** for new repositories
3. **Write handler tests** for new endpoints
4. **Write repository tests** for new queries
5. **Update documentation**

### Updating Existing Tests

1. **Update fixtures** when domain models change
2. **Update mocks** when interfaces change
3. **Update assertions** when behavior changes
4. **Add regression tests** for bug fixes

## Performance Testing

### Benchmark Tests

```go
func BenchmarkSearchAgents(b *testing.B) {
    setup := setupBenchmark()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        handler.SearchAgents(ctx, request)
    }
}
```

### Load Testing

Use tools like:
- `k6` for API load testing
- `hey` for HTTP benchmarking
- Custom Go benchmarks for specific operations

## Troubleshooting Tests

### Common Issues

1. **Flaky Tests**
   - Use fixed test data
   - Avoid time.Now() - use fixed times
   - Clean up resources properly

2. **Slow Tests**
   - Mock external dependencies
   - Use parallel testing: `t.Parallel()`
   - Optimize test data size

3. **Mock Expectations Not Met**
   - Check call order
   - Verify argument matchers
   - Use mock.Anything for flexible matching

## Test Coverage Report

Generate and view coverage:

```bash
# Generate coverage
go test ./... -coverprofile=coverage.out

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out

# Generate coverage badge
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}'
```

## Summary

Phase 11 implements comprehensive testing across all layers:

✅ **Test Infrastructure**: Fixtures, mocks, helpers
✅ **Handler Tests**: Phase 9 & 10 APIs (26 test functions)
✅ **Repository Tests**: Export & webhook repositories (15+ test functions)
🔄 **Integration Tests**: Pending future implementation
🔄 **Coverage Reporting**: Tools configured, pending CI integration

All tests follow Go best practices and use industry-standard testing patterns for maintainability and reliability.

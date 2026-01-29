# Request/Response Model Structure Guide

## 📁 Directory Structure

```
lib/data/models/
├── requests/              # All API request models
│   ├── agent/
│   │   ├── create_agent_request.dart
│   │   ├── update_agent_request.dart
│   │   └── search_agent_request.dart
│   ├── license/
│   │   ├── add_license_request.dart
│   │   └── renew_license_request.dart
│   ├── bank/
│   │   └── add_bank_details_request.dart
│   ├── goal/
│   │   └── set_goals_request.dart
│   └── common/
│       ├── otp_request.dart
│       └── pagination_request.dart
│
├── responses/             # All API response models
│   ├── agent/
│   │   ├── agent_profile_response.dart
│   │   ├── search_results_response.dart
│   │   └── workflow_state_response.dart
│   ├── license/
│   │   ├── license_response.dart
│   │   └── renewal_calculation_response.dart
│   ├── bank/
│   │   └── bank_details_response.dart
│   ├── goal/
│   │   └── goals_response.dart
│   └── common/
│       ├── api_response.dart
│       ├── error_response.dart
│       ├── pagination_response.dart
│       └── validation_response.dart
│
└── MODEL_STRUCTURE_GUIDE.md  # This file
```

---

## 📝 Naming Conventions

### Request Models
- **Pattern**: `{action}_{entity}_request.dart`
- **Class Name**: `{Action}{Entity}Request`

Examples:
```dart
create_agent_request.dart     → CreateAgentRequest
update_agent_request.dart     → UpdateAgentRequest
add_license_request.dart      → AddLicenseRequest
validate_pan_request.dart     → ValidatePanRequest
```

### Response Models
- **Pattern**: `{entity}_{data_type}_response.dart`
- **Class Name**: `{Entity}{DataType}Response`

Examples:
```dart
agent_profile_response.dart   → AgentProfileResponse
search_results_response.dart  → SearchResultsResponse
license_response.dart         → LicenseResponse
validation_response.dart      → ValidationResponse
```

---

## 🎯 Model Template Structure

### Request Model Template

```dart
// File: lib/data/models/requests/agent/create_agent_request.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'create_agent_request.freezed.dart';
part 'create_agent_request.g.dart';

/// Request model for creating agent profile
///
/// API: POST /agent-profiles/{session_id}/submit
/// Journey: UJ-001 (Agent Profile Creation)
///
/// HOW TO UPDATE:
/// 1. Modify fields in the factory constructor
/// 2. Run: flutter pub run build_runner build --delete-conflicting-outputs
/// 3. All usages will automatically update
@freezed
class CreateAgentRequest with _$CreateAgentRequest {
  const factory CreateAgentRequest({
    // ============================================================================
    // REQUIRED FIELDS - Update these when API changes
    // ============================================================================

    @JsonKey(name: 'submit_action') required String submitAction,
    required bool confirmation,
    @JsonKey(name: 'submitted_by') required String submittedBy,

    // ============================================================================
    // OPTIONAL FIELDS - Add/remove as needed
    // ============================================================================

    @JsonKey(name: 'additional_notes') String? additionalNotes,

    // ============================================================================
    // NESTED OBJECTS - Update when nested structure changes
    // ============================================================================

    @JsonKey(name: 'personal_info') PersonalInfoModel? personalInfo,
    List<AddressModel>? addresses,
    List<ContactModel>? contacts,
  }) = _CreateAgentRequest;

  /// From JSON factory
  factory CreateAgentRequest.fromJson(Map<String, dynamic> json) =>
      _$CreateAgentRequestFromJson(json);
}

// ============================================================================
// CHANGE LOG
// ============================================================================
//
// 2026-01-29: Initial creation
// - Added submit_action, confirmation, submitted_by fields
// - Added optional personal_info, addresses, contacts
//
// HOW TO ADD NEW FIELD:
// 1. Add field in factory constructor
// 2. Add @JsonKey annotation if API name differs
// 3. Run build_runner
// 4. Document change in this log
// ============================================================================
```

### Response Model Template

```dart
// File: lib/data/models/responses/agent/agent_profile_response.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'agent_profile_response.freezed.dart';
part 'agent_profile_response.g.dart';

/// Response model for agent profile
///
/// API: GET /agents/{agent_id}
/// Journey: UJ-002 (Agent Profile View/Update)
///
/// HOW TO UPDATE:
/// 1. Modify fields in the factory constructor
/// 2. Run: flutter pub run build_runner build --delete-conflicting-outputs
/// 3. All usages will automatically update
@freezed
class AgentProfileResponse with _$AgentProfileResponse {
  const factory AgentProfileResponse({
    // ============================================================================
    // PROFILE IDENTIFICATION - Update when API changes
    // ============================================================================

    @JsonKey(name: 'agent_id') required String agentId,
    @JsonKey(name: 'profile_id') required String profileId,
    @JsonKey(name: 'agent_type') required String agentType,
    @JsonKey(name: 'full_name') required String fullName,
    required String status,

    // ============================================================================
    // PROFILE DETAILS - Add/remove fields as API evolves
    // ============================================================================

    @JsonKey(name: 'pan_number') String? panNumber,
    @JsonKey(name: 'status_date') String? statusDate,
    @JsonKey(name: 'effective_date') String? effectiveDate,
    @JsonKey(name: 'created_at') String? createdAt,
    @JsonKey(name: 'created_by') String? createdBy,

    // ============================================================================
    // NESTED OBJECTS - Update structure as needed
    // ============================================================================

    @JsonKey(name: 'personal_info') PersonalInfoModel? personalInfo,
    List<AddressModel>? addresses,
    List<ContactModel>? contacts,
    @JsonKey(name: 'office_association') OfficeModel? officeAssociation,
    @JsonKey(name: 'advisor_coordinator') CoordinatorModel? advisorCoordinator,

    // ============================================================================
    // WORKFLOW & STATE - Usually consistent across APIs
    // ============================================================================

    @JsonKey(name: 'workflow_state') WorkflowStateModel? workflowState,
  }) = _AgentProfileResponse;

  /// From JSON factory
  factory AgentProfileResponse.fromJson(Map<String, dynamic> json) =>
      _$AgentProfileResponseFromJson(json);
}

// ============================================================================
// CHANGE LOG
// ============================================================================
//
// 2026-01-29: Initial creation
// - Added core profile fields
// - Added nested objects for personal_info, addresses, contacts
// - Added workflow_state for state tracking
//
// HOW TO MODIFY:
// 1. Add/remove/update field in factory constructor
// 2. Update @JsonKey if API field name changes
// 3. Run build_runner
// 4. Document change here
// ============================================================================
```

---

## 🔧 How to Make Changes

### 1. Update Request Model

When API request structure changes:

```dart
// BEFORE
@freezed
class CreateAgentRequest with _$CreateAgentRequest {
  const factory CreateAgentRequest({
    required String submitAction,
  }) = _CreateAgentRequest;
}

// AFTER (Added new field)
@freezed
class CreateAgentRequest with _$CreateAgentRequest {
  const factory CreateAgentRequest({
    required String submitAction,
    String? newField,  // ← Added this
  }) = _CreateAgentRequest;
}
```

### 2. Update Response Model

When API response structure changes:

```dart
// BEFORE
@freezed
class AgentProfileResponse with _$AgentProfileResponse {
  const factory AgentProfileResponse({
    required String agentId,
  }) = _AgentProfileResponse;
}

// AFTER (Added new field)
@freezed
class AgentProfileResponse with _$AgentProfileResponse {
  const factory AgentProfileResponse({
    required String agentId,
    @JsonKey(name: 'email_address') String? emailAddress,  // ← Added this
  }) = _AgentProfileResponse;
}
```

### 3. Update JSON Key Mapping

When API field name changes:

```dart
// API changed "agent_id" to "id"

// BEFORE
@JsonKey(name: 'agent_id') required String agentId,

// AFTER
@JsonKey(name: 'id') required String agentId,
```

### 4. Regenerate Code

After any model change:

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

---

## 📋 Common Model Patterns

### Pagination Request

```dart
@freezed
class PaginationRequest with _$PaginationRequest {
  const factory PaginationRequest({
    @Default(1) int page,
    @Default(20) int limit,
    @JsonKey(name: 'sort_by') String? sortBy,
    @JsonKey(name: 'sort_order') String? sortOrder,
  }) = _PaginationRequest;
}
```

### Pagination Response

```dart
@freezed
class PaginationResponse with _$PaginationResponse {
  const factory PaginationResponse({
    required int page,
    required int limit,
    @JsonKey(name: 'total_count') required int totalCount,
    @JsonKey(name: 'total_pages') required int totalPages,
    @JsonKey(name: 'has_next') required bool hasNext,
    @JsonKey(name: 'has_previous') required bool hasPrevious,
  }) = _PaginationResponse;
}
```

### Generic API Response Wrapper

```dart
@freezed
class ApiResponse<T> with _$ApiResponse<T> {
  const factory ApiResponse({
    required bool success,
    String? message,
    T? data,
    @JsonKey(name: 'error_code') String? errorCode,
  }) = _ApiResponse<T>;
}
```

### Error Response

```dart
@freezed
class ErrorResponse with _$ErrorResponse {
  const factory ErrorResponse({
    required String code,
    required String message,
    String? details,
    @JsonKey(name: 'suggested_actions') List<String>? suggestedActions,
  }) = _ErrorResponse;
}
```

### Workflow State

```dart
@freezed
class WorkflowStateModel with _$WorkflowStateModel {
  const factory WorkflowStateModel({
    @JsonKey(name: 'current_step') required String currentStep,
    @JsonKey(name: 'next_step') String? nextStep,
    @JsonKey(name: 'allowed_actions') List<String>? allowedActions,
    @JsonKey(name: 'progress_percentage') int? progressPercentage,
  }) = _WorkflowStateModel;
}
```

---

## 📦 Model Organization by Journey

### UJ-001: Agent Profile Creation

**Requests:**
- `InitiateProfileRequest`
- `FetchHrmsRequest`
- `LinkCoordinatorRequest`
- `ValidateProfileRequest`
- `SubmitProfileRequest`

**Responses:**
- `ProfileSessionResponse`
- `HrmsDataResponse`
- `CoordinatorLinkageResponse`
- `ValidationResultsResponse`
- `AgentProfileResponse`

### UJ-002: Agent Profile Update

**Requests:**
- `SearchAgentRequest`
- `UpdateSectionRequest`
- `ApproveUpdateRequest`

**Responses:**
- `SearchResultsResponse`
- `UpdateStatusResponse`
- `ApprovalResponse`

### UJ-003: License Management

**Requests:**
- `AddLicenseRequest`
- `RenewLicenseRequest`
- `UpdateLicenseRequest`

**Responses:**
- `LicenseResponse`
- `RenewalCalculationResponse`
- `LicenseListResponse`

---

## 🎨 JSON Key Annotations

### When to Use @JsonKey

```dart
// Use when API field name differs from Dart field name
@JsonKey(name: 'agent_id') String agentId,        // API: agent_id, Dart: agentId
@JsonKey(name: 'pan_number') String panNumber,    // API: pan_number, Dart: panNumber

// Use when field has default value
@JsonKey(defaultValue: 'ACTIVE') String status,

// Use when field can be null in JSON
@JsonKey(includeIfNull: false) String? optionalField,

// Use for custom JSON conversion
@JsonKey(fromJson: _dateFromJson, toJson: _dateToJson) DateTime? date,
```

---

## 🧪 Testing Models

### Unit Test Template

```dart
void main() {
  group('CreateAgentRequest', () {
    test('should serialize to JSON correctly', () {
      // Arrange
      final request = CreateAgentRequest(
        submitAction: 'CREATE',
        confirmation: true,
        submittedBy: 'admin_123',
      );

      // Act
      final json = request.toJson();

      // Assert
      expect(json['submit_action'], 'CREATE');
      expect(json['confirmation'], true);
      expect(json['submitted_by'], 'admin_123');
    });

    test('should deserialize from JSON correctly', () {
      // Arrange
      final json = {
        'submit_action': 'CREATE',
        'confirmation': true,
        'submitted_by': 'admin_123',
      };

      // Act
      final request = CreateAgentRequest.fromJson(json);

      // Assert
      expect(request.submitAction, 'CREATE');
      expect(request.confirmation, true);
      expect(request.submittedBy, 'admin_123');
    });
  });
}
```

---

## 📚 Best Practices

### 1. Always Use Freezed
✅ Provides immutability
✅ Auto-generates copyWith, equality, toString
✅ Type-safe JSON serialization

### 2. Document Each Model
✅ Add comment with API endpoint
✅ Add comment with journey reference
✅ Add change log

### 3. Use Descriptive Names
✅ `CreateAgentRequest` not `AgentReq`
✅ `AgentProfileResponse` not `AgentResp`
✅ `SearchResultsResponse` not `Results`

### 4. Group Related Fields
✅ Required fields first
✅ Optional fields next
✅ Nested objects last
✅ Workflow state last

### 5. Maintain Change Log
✅ Document when field added
✅ Document when field removed
✅ Document when field renamed

---

## 🔄 Migration Guide

### When API Changes

**Scenario 1: Field Added**
```dart
// Just add the field
String? newField,

// Rebuild
flutter pub run build_runner build --delete-conflicting-outputs
```

**Scenario 2: Field Removed**
```dart
// Comment out or delete the field
// String? oldField,  // Deprecated

// Rebuild
flutter pub run build_runner build --delete-conflicting-outputs
```

**Scenario 3: Field Renamed**
```dart
// Update @JsonKey
@JsonKey(name: 'new_field_name') String fieldName,

// Rebuild
flutter pub run build_runner build --delete-conflicting-outputs
```

**Scenario 4: Type Changed**
```dart
// Update type
// BEFORE: String? age,
// AFTER:
int? age,

// Rebuild
flutter pub run build_runner build --delete-conflicting-outputs
```

---

## 📖 Quick Reference

### Common Commands
```bash
# Generate code
flutter pub run build_runner build --delete-conflicting-outputs

# Watch for changes (auto-generate)
flutter pub run build_runner watch

# Clean generated files
flutter pub run build_runner clean
```

### File Naming
- Requests: `{action}_{entity}_request.dart`
- Responses: `{entity}_{type}_response.dart`
- Models: `{entity}_model.dart`

### Class Naming
- Requests: `{Action}{Entity}Request`
- Responses: `{Entity}{Type}Response`
- Models: `{Entity}Model`

---

**Last Updated**: 2026-01-29
**Maintained by**: PLI Development Team

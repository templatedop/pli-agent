// File: lib/data/models/responses/agent/agent_profile_response.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'agent_profile_response.freezed.dart';
part 'agent_profile_response.g.dart';

/// Response model for agent profile
///
/// **API Endpoint**: GET /agents/{agent_id}
/// **Journey**: UJ-002 (Agent Profile View/Update)
/// **API ID**: AGT-023
///
/// **Usage Example**:
/// ```dart
/// // Parse from API response
/// final response = AgentProfileResponse.fromJson(jsonData);
///
/// // Access fields
/// print(response.agentId);        // 'AGT-2026-000567'
/// print(response.fullName);       // 'Rajesh Kumar Sharma'
/// print(response.status);         // 'ACTIVE'
/// ```
///
/// **HOW TO UPDATE THIS MODEL**:
/// 1. Modify fields in the factory constructor below
/// 2. Run: `flutter pub run build_runner build --delete-conflicting-outputs`
/// 3. All usages will automatically update with new structure
/// 4. Update the CHANGE LOG section at bottom
///
/// **JSON Mapping**:
/// - API uses snake_case (e.g., `agent_id`)
/// - Dart uses camelCase (e.g., `agentId`)
/// - @JsonKey handles conversion automatically
@freezed
class AgentProfileResponse with _$AgentProfileResponse {
  const factory AgentProfileResponse({
    // ========================================================================
    // PROFILE IDENTIFICATION
    // ========================================================================
    // Core identifiers that uniquely identify the agent

    /// Unique agent identifier
    /// **API Field**: agent_id
    /// **Format**: AGT-YYYY-NNNNNN
    /// **Example**: 'AGT-2026-000567'
    @JsonKey(name: 'agent_id') required String agentId,

    /// Internal profile ID
    /// **API Field**: profile_id
    /// **Format**: UUID
    @JsonKey(name: 'profile_id') required String profileId,

    /// Type of agent
    /// **API Field**: agent_type
    /// **Values**: ADVISOR, ADVISOR_COORDINATOR, DEPARTMENTAL_EMPLOYEE, etc.
    @JsonKey(name: 'agent_type') required String agentType,

    /// Full name of the agent
    /// **API Field**: full_name
    /// **Example**: 'Rajesh Kumar Sharma'
    @JsonKey(name: 'full_name') required String fullName,

    /// Current status of the agent
    /// **API Field**: status
    /// **Values**: ACTIVE, INACTIVE, SUSPENDED, TERMINATED, DEACTIVATED
    required String status,

    // ========================================================================
    // PROFILE DETAILS
    // ========================================================================
    // Additional profile information

    /// PAN number (optional)
    /// **API Field**: pan_number
    /// **Format**: AAAAA9999A
    /// **Example**: 'ABCDE1234F'
    @JsonKey(name: 'pan_number') String? panNumber,

    /// Date when status was last changed
    /// **API Field**: status_date
    /// **Format**: DD-MM-YYYY
    @JsonKey(name: 'status_date') String? statusDate,

    /// Profile effective date
    /// **API Field**: effective_date
    /// **Format**: DD-MM-YYYY
    @JsonKey(name: 'effective_date') String? effectiveDate,

    /// Profile creation timestamp
    /// **API Field**: created_at
    /// **Format**: ISO 8601
    @JsonKey(name: 'created_at') String? createdAt,

    /// User who created the profile
    /// **API Field**: created_by
    @JsonKey(name: 'created_by') String? createdBy,

    // ========================================================================
    // NESTED OBJECTS
    // ========================================================================
    // Complex objects embedded in the response

    /// Personal information details
    /// **API Field**: personal_info
    /// **Type**: PersonalInfoModel (nested object)
    @JsonKey(name: 'personal_info') Map<String, dynamic>? personalInfo,

    /// List of addresses
    /// **API Field**: addresses
    /// **Type**: Array of AddressModel
    List<Map<String, dynamic>>? addresses,

    /// List of contact numbers
    /// **API Field**: contacts
    /// **Type**: Array of ContactModel
    List<Map<String, dynamic>>? contacts,

    /// List of email addresses
    /// **API Field**: emails
    /// **Type**: Array of EmailModel
    List<Map<String, dynamic>>? emails,

    /// Office association details
    /// **API Field**: office
    /// **Type**: OfficeModel (nested object)
    Map<String, dynamic>? office,

    /// Advisor coordinator details (for Advisors)
    /// **API Field**: advisor_coordinator
    /// **Type**: CoordinatorModel (nested object)
    @JsonKey(name: 'advisor_coordinator') Map<String, dynamic>? advisorCoordinator,

    /// License details
    /// **API Field**: licenses
    /// **Type**: Array of LicenseModel
    List<Map<String, dynamic>>? licenses,

    /// Bank details (masked)
    /// **API Field**: bank_details
    /// **Type**: Array of BankDetailsModel
    @JsonKey(name: 'bank_details') List<Map<String, dynamic>>? bankDetails,

    // ========================================================================
    // WORKFLOW & STATE
    // ========================================================================
    // Workflow tracking and state management

    /// Current workflow state
    /// **API Field**: workflow_state
    /// **Type**: WorkflowStateModel (nested object)
    @JsonKey(name: 'workflow_state') Map<String, dynamic>? workflowState,
  }) = _AgentProfileResponse;

  /// Create from JSON response
  /// Used when deserializing API response
  factory AgentProfileResponse.fromJson(Map<String, dynamic> json) =>
      _$AgentProfileResponseFromJson(json);
}

// ============================================================================
// CHANGE LOG - Document all changes here
// ============================================================================
//
// 2026-01-29 v1.0.0 - Initial creation
//   - Added: Core profile identification fields
//   - Added: Profile details (pan_number, dates, etc.)
//   - Added: Nested objects (personal_info, addresses, contacts, etc.)
//   - Added: Workflow state tracking
//   - API Version: v1
//   - Based on: agent_profile_management_api_part1.yaml, part2.yaml
//
// FIELD MAPPING REFERENCE:
// ┌─────────────────────────┬──────────────────────────┬─────────────┐
// │ Dart Field Name         │ API Field Name           │ Type        │
// ├─────────────────────────┼──────────────────────────┼─────────────┤
// │ agentId                 │ agent_id                 │ String      │
// │ profileId               │ profile_id               │ String      │
// │ agentType               │ agent_type               │ String      │
// │ fullName                │ full_name                │ String      │
// │ status                  │ status                   │ String      │
// │ panNumber               │ pan_number               │ String?     │
// │ statusDate              │ status_date              │ String?     │
// │ effectiveDate           │ effective_date           │ String?     │
// │ createdAt               │ created_at               │ String?     │
// │ createdBy               │ created_by               │ String?     │
// │ personalInfo            │ personal_info            │ Object?     │
// │ addresses               │ addresses                │ Array?      │
// │ contacts                │ contacts                 │ Array?      │
// │ emails                  │ emails                   │ Array?      │
// │ office                  │ office                   │ Object?     │
// │ advisorCoordinator      │ advisor_coordinator      │ Object?     │
// │ licenses                │ licenses                 │ Array?      │
// │ bankDetails             │ bank_details             │ Array?      │
// │ workflowState           │ workflow_state           │ Object?     │
// └─────────────────────────┴──────────────────────────┴─────────────┘
//
// HOW TO UPDATE WHEN API CHANGES:
//
// 1. API ADDS A NEW FIELD:
//    - Add field in appropriate section
//    - Add @JsonKey annotation
//    - Add JSDoc comment
//    - Run build_runner
//    - Document in change log
//
// 2. API REMOVES A FIELD:
//    - Comment out or delete the field
//    - Run build_runner
//    - Document in change log
//
// 3. API RENAMES A FIELD:
//    - Keep Dart field name same (for stability)
//    - Update @JsonKey(name: 'new_api_name')
//    - Run build_runner
//    - Document in change log
//
// 4. API CHANGES FIELD TYPE:
//    - Update Dart type
//    - May need custom JSON converter
//    - Run build_runner
//    - Document in change log
//    - Test thoroughly
//
// EXAMPLE FUTURE CHANGES:
// 2026-02-15 v1.1.0 - Added mobile_verified field
//   - Added: mobile_verified (bool?)
//   - Reason: API now returns mobile verification status
//   - Location: Profile Details section
//   - API Version: v1.1
//
// 2026-03-01 v1.2.0 - Renamed pan_number to pan_card
//   - Changed: @JsonKey(name: 'pan_card') for panNumber field
//   - Reason: API standardization
//   - Dart field name unchanged for backward compatibility
//   - API Version: v1.2
//
// ============================================================================

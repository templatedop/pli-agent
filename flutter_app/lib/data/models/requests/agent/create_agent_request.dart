// File: lib/data/models/requests/agent/create_agent_request.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'create_agent_request.freezed.dart';
part 'create_agent_request.g.dart';

/// Request model for creating agent profile
///
/// **API Endpoint**: POST /agent-profiles/{session_id}/submit
/// **Journey**: UJ-001 (Agent Profile Creation - Step 5)
/// **API ID**: AGT-006
///
/// **Usage Example**:
/// ```dart
/// final request = CreateAgentRequest(
///   submitAction: 'CREATE',
///   confirmation: true,
///   submittedBy: 'admin_user_123',
/// );
///
/// final json = request.toJson();
/// // Send to API
/// ```
///
/// **HOW TO UPDATE THIS MODEL**:
/// 1. Modify fields in the factory constructor below
/// 2. Run: `flutter pub run build_runner build --delete-conflicting-outputs`
/// 3. All usages will automatically update with new structure
/// 4. Update the CHANGE LOG section at bottom
///
/// **JSON Mapping**:
/// - Dart field names use camelCase (e.g., `submittedBy`)
/// - API field names use snake_case (e.g., `submitted_by`)
/// - @JsonKey annotation handles the conversion automatically
@freezed
class CreateAgentRequest with _$CreateAgentRequest {
  const factory CreateAgentRequest({
    // ========================================================================
    // REQUIRED FIELDS
    // ========================================================================
    // These fields MUST be provided when creating the request

    /// Submit action type
    /// **API Field**: submit_action
    /// **Values**: 'CREATE', 'SAVE_DRAFT'
    @JsonKey(name: 'submit_action') required String submitAction,

    /// User confirmation flag
    /// **API Field**: confirmation
    /// **Values**: true, false
    required bool confirmation,

    /// User ID who is submitting the profile
    /// **API Field**: submitted_by
    /// **Example**: 'admin_user_123'
    @JsonKey(name: 'submitted_by') required String submittedBy,

    // ========================================================================
    // OPTIONAL FIELDS
    // ========================================================================
    // These fields can be null/omitted

    /// Additional notes or comments
    /// **API Field**: additional_notes
    /// **Max Length**: 500 characters
    @JsonKey(name: 'additional_notes') String? additionalNotes,

    /// Whether to skip notifications
    /// **API Field**: skip_notifications
    /// **Default**: false
    @JsonKey(name: 'skip_notifications') bool? skipNotifications,
  }) = _CreateAgentRequest;

  /// Create from JSON response
  /// Used when deserializing API response
  factory CreateAgentRequest.fromJson(Map<String, dynamic> json) =>
      _$CreateAgentRequestFromJson(json);
}

// ============================================================================
// CHANGE LOG - Document all changes here
// ============================================================================
//
// 2026-01-29 v1.0.0 - Initial creation
//   - Added: submit_action, confirmation, submitted_by (required)
//   - Added: additional_notes, skip_notifications (optional)
//   - API Version: v1
//   - Based on: agent_profile_management_api_part1.yaml
//
// HOW TO ADD A NEW FIELD:
// 1. Add field in factory constructor (in appropriate section)
// 2. Add JSDoc comment with API field name and description
// 3. Add @JsonKey annotation if API name differs from Dart name
// 4. Run: flutter pub run build_runner build --delete-conflicting-outputs
// 5. Document the change in this log with date and reason
//
// HOW TO REMOVE A FIELD:
// 1. Delete or comment out the field
// 2. Run: flutter pub run build_runner build --delete-conflicting-outputs
// 3. Document the removal in this log
//
// HOW TO RENAME A FIELD:
// 1. Change the Dart field name
// 2. Update @JsonKey(name: 'api_field_name') if needed
// 3. Run: flutter pub run build_runner build --delete-conflicting-outputs
// 4. Document the change in this log
//
// EXAMPLE FUTURE CHANGE:
// 2026-02-15 v1.1.0 - Added workflow_override
//   - Added: workflow_override (optional)
//   - Reason: Support for admin workflow bypass
//   - API Version: v1.1
//
// ============================================================================

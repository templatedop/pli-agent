/// API Endpoints Configuration
///
/// IMPORTANT: All API endpoints are defined here in one place.
/// This makes it easy to update endpoints when API changes.
///
/// USAGE:
/// - Each endpoint is a constant string
/// - Use {param} for path parameters
/// - Use ApiEndpoints.replacePath() to replace parameters
///
/// EXAMPLE:
/// ```dart
/// final url = ApiEndpoints.agentProfile;  // '/agents/{agent_id}'
/// final finalUrl = ApiEndpoints.replacePath(url, {'agent_id': 'AGT-2026-000001'});
/// // Result: '/agents/AGT-2026-000001'
/// ```

class ApiEndpoints {
  // Prevent instantiation
  ApiEndpoints._();

  // ============================================================================
  // CONFIGURATION - UPDATE THESE IF ENDPOINT STRUCTURE CHANGES
  // ============================================================================

  /// API Version prefix (usually '/v1', '/v2', etc.)
  /// Change this if API version changes
  static const String versionPrefix = '/v1';

  /// Whether to include version prefix in endpoints
  /// Set to false if version is already in base URL
  static const bool includeVersionPrefix = false;

  // ============================================================================
  // HELPER METHODS
  // ============================================================================

  /// Replace path parameters in endpoint URL
  /// Example: replacePath('/agents/{agent_id}', {'agent_id': 'AGT-001'})
  /// Returns: '/agents/AGT-001'
  static String replacePath(String path, Map<String, String> params) {
    String result = path;
    params.forEach((key, value) {
      result = result.replaceAll('{$key}', value);
    });
    return result;
  }

  /// Get endpoint with version prefix if configured
  static String _endpoint(String path) {
    if (includeVersionPrefix) {
      return '$versionPrefix$path';
    }
    return path;
  }

  // ============================================================================
  // UJ-001: AGENT PROFILE CREATION (21 APIs)
  // ============================================================================
  // Journey: Create new agent profile with validation and HRMS integration

  /// Step 1: Initiate profile creation session
  /// POST /agent-profiles/initiate
  /// Request: {agent_type, initiated_by}
  /// Response: {session_id, workflow_state}
  static String get initiateProfile => _endpoint('/agent-profiles/initiate');

  /// Step 2: Fetch employee data from HRMS
  /// POST /agent-profiles/{session_id}/fetch-hrms
  /// Request: {employee_id, fetch_mode}
  /// Response: {employee_data, auto_populated_fields}
  static String get fetchHrms => _endpoint('/agent-profiles/{session_id}/fetch-hrms');

  /// Step 3: Get list of active advisor coordinators
  /// GET /advisor-coordinators
  /// Query: status, circle_id, division_id, page, limit
  /// Response: {coordinators[], pagination}
  static String get advisorCoordinators => _endpoint('/advisor-coordinators');

  /// Step 3: Link advisor to coordinator
  /// POST /agent-profiles/{session_id}/link-coordinator
  /// Request: {coordinator_id, linkage_effective_date}
  /// Response: {coordinator_linkage, workflow_state}
  static String get linkCoordinator => _endpoint('/agent-profiles/{session_id}/link-coordinator');

  /// Step 4: Validate profile details
  /// POST /agent-profiles/{session_id}/validate-basic
  /// Request: {personal_info, addresses, contacts, emails, office_association}
  /// Response: {validation_status, validation_results}
  static String get validateProfile => _endpoint('/agent-profiles/{session_id}/validate-basic');

  /// Step 5: Submit profile for creation
  /// POST /agent-profiles/{session_id}/submit
  /// Request: {submit_action, confirmation, submitted_by}
  /// Response: {agent_profile, workflow_state, notifications_sent}
  static String get submitProfile => _endpoint('/agent-profiles/{session_id}/submit');

  // Lookup APIs for UJ-001
  static String get agentTypes => _endpoint('/agent-types');
  static String get categories => _endpoint('/categories');
  static String get designations => _endpoint('/designations');
  static String get officeTypes => _endpoint('/office-types');
  static String get states => _endpoint('/states');

  // Validation APIs for UJ-001
  static String get validatePan => _endpoint('/validations/pan/check-uniqueness');
  static String get validateEmployeeId => _endpoint('/validations/hrms/employee-id');
  static String get validateIfsc => _endpoint('/validations/bank/ifsc');
  static String get validateOffice => _endpoint('/validations/office/{office_code}');

  // Workflow Management APIs for UJ-001
  static String get sessionStatus => _endpoint('/agent-profiles/sessions/{session_id}/status');
  static String get saveSession => _endpoint('/agent-profiles/sessions/{session_id}/save');
  static String get resumeSession => _endpoint('/agent-profiles/sessions/{session_id}/resume');
  static String get cancelSession => _endpoint('/agent-profiles/sessions/{session_id}');

  // Status & Notification APIs for UJ-001
  static String get creationStatus => _endpoint('/agent-profiles/creation-status/{agent_id}');
  static String get resendWelcome => _endpoint('/agents/{agent_id}/notifications/resend-welcome');

  // ============================================================================
  // UJ-002: AGENT PROFILE UPDATE (7 APIs)
  // ============================================================================
  // Journey: Update existing agent profile with approval workflow

  /// Search agents by multiple criteria
  /// GET /agents/search
  /// Query: agent_id, name, pan_number, mobile_number, status, office_code, page, limit
  /// Response: {results[], pagination}
  static String get searchAgents => _endpoint('/agents/search');

  /// Get complete agent profile
  /// GET /agents/{agent_id}
  /// Response: {agent_profile, workflow_state}
  static String get agentProfile => _endpoint('/agents/{agent_id}');

  /// Get update form for specific section
  /// GET /agents/{agent_id}/update-form?section={section}
  /// Query: section (personal_info, address, contact, license, bank_details, status)
  /// Response: {section, current_data, editable_fields}
  static String get updateForm => _endpoint('/agents/{agent_id}/update-form');

  /// Update profile section
  /// PUT /agents/{agent_id}/sections/{section}
  /// Request: {changes, update_reason, updated_by}
  /// Response: {update_status, changes_applied, pending_approval}
  static String get updateSection => _endpoint('/agents/{agent_id}/sections/{section}');

  /// Approve profile update
  /// PUT /approvals/{approval_request_id}/approve
  /// Request: {action, approver_comments, approved_by, attachments}
  /// Response: {approval_status, changes_applied, audit_logs}
  static String get approveUpdate => _endpoint('/approvals/{approval_request_id}/approve');

  /// Reject profile update
  /// PUT /approvals/{approval_request_id}/reject
  /// Request: {action, rejector_comments, rejected_by}
  /// Response: {approval_status, rejected_at}
  static String get rejectUpdate => _endpoint('/approvals/{approval_request_id}/reject');

  /// Get audit history
  /// GET /agents/{agent_id}/audit-history
  /// Query: from_date, to_date, page, limit
  /// Response: {audit_logs[], pagination}
  static String get auditHistory => _endpoint('/agents/{agent_id}/audit-history');

  // ============================================================================
  // UJ-003: LICENSE MANAGEMENT (10 APIs)
  // ============================================================================
  // Journey: Manage agent licenses with renewal tracking

  /// Get all licenses for an agent
  /// GET /agents/{agent_id}/licenses
  /// Query: status (Active, Expired, Pending_Renewal)
  /// Response: {licenses[], workflow_state, sla_tracking}
  static String get agentLicenses => _endpoint('/agents/{agent_id}/licenses');

  /// Add new license
  /// POST /agents/{agent_id}/licenses
  /// Request: {license_line, license_type, license_number, resident_status, license_date, authority_date}
  /// Response: {license, renewal_calculation, reminders_scheduled}
  static String get addLicense => _endpoint('/agents/{agent_id}/licenses');

  /// Get specific license details
  /// GET /agents/{agent_id}/licenses/{license_id}
  /// Response: {license, renewal_history, reminder_schedule}
  static String get licenseDetails => _endpoint('/agents/{agent_id}/licenses/{license_id}');

  /// Update license
  /// PUT /agents/{agent_id}/licenses/{license_id}
  /// Request: {license_type, license_number, resident_status, authority_date}
  /// Response: {license, audit_log}
  static String get updateLicense => _endpoint('/agents/{agent_id}/licenses/{license_id}');

  /// Delete license
  /// DELETE /agents/{agent_id}/licenses/{license_id}
  /// Response: 204 No Content
  static String get deleteLicense => _endpoint('/agents/{agent_id}/licenses/{license_id}');

  /// Renew license
  /// PUT /agents/{agent_id}/licenses/{license_id}/renew
  /// Request: {renewal_type, licentiate_exam_passed, exam_date, exam_certificate_number}
  /// Response: {license, renewal_calculation, reminders_scheduled}
  static String get renewLicense => _endpoint('/agents/{agent_id}/licenses/{license_id}/renew');

  /// Get license types dropdown
  /// GET /license-types
  /// Response: {license_types[]}
  static String get licenseTypes => _endpoint('/license-types');

  /// Get licenses expiring soon
  /// GET /licenses/expiring
  /// Query: days (default 30), office_code, page, limit
  /// Response: {licenses[], pagination}
  static String get expiringLicenses => _endpoint('/licenses/expiring');

  /// Get reminder schedule for license
  /// GET /licenses/{license_id}/reminders
  /// Response: {reminder_schedule[]}
  static String get licenseReminders => _endpoint('/licenses/{license_id}/reminders');

  /// Trigger batch deactivation for expired licenses
  /// POST /licenses/expired
  /// Request: {batch_date, dry_run}
  /// Response: {batch_id, agents_deactivated, agent_ids[], notifications_sent}
  static String get expiredLicenses => _endpoint('/licenses/expired');

  // ============================================================================
  // UJ-004: AGENT TERMINATION (3 APIs)
  // ============================================================================
  // Journey: Terminate agent with document generation

  /// Terminate agent
  /// POST /agents/{agent_id}/terminate
  /// Request: {termination_reason, termination_reason_code, effective_date, terminated_by}
  /// Response: {termination_status, actions_performed, termination_letter_url}
  static String get terminateAgent => _endpoint('/agents/{agent_id}/terminate');

  /// Get termination letter
  /// GET /agents/{agent_id}/termination-letter
  /// Query: format (PDF, HTML)
  /// Response: Binary file (PDF)
  static String get terminationLetter => _endpoint('/agents/{agent_id}/termination-letter');

  /// Get termination reasons dropdown
  /// GET /termination/reasons
  /// Response: {termination_reasons[]}
  static String get terminationReasons => _endpoint('/termination/reasons');

  // ============================================================================
  // UJ-005: PORTAL AUTHENTICATION (5 APIs)
  // ============================================================================
  // Journey: Agent portal login with OTP

  /// Initiate portal login
  /// POST /agents/{agent_id}/portal/login
  /// Request: {password, device_info}
  /// Response: {transaction_id, otp_sent_to, otp_expires_at}
  static String get portalLogin => _endpoint('/agents/{agent_id}/portal/login');

  /// Verify OTP
  /// POST /agents/{agent_id}/portal/otp/verify
  /// Request: {otp, transaction_id}
  /// Response: {session_id, access_token, refresh_token, agent_profile}
  static String get verifyOtp => _endpoint('/agents/{agent_id}/portal/otp/verify');

  /// Portal logout
  /// POST /agents/{agent_id}/portal/logout
  /// Request: {session_id}
  /// Response: {message, logged_out_at}
  static String get portalLogout => _endpoint('/agents/{agent_id}/portal/logout');

  /// Unlock account (admin)
  /// POST /agents/{agent_id}/portal/unlock
  /// Request: {unlock_reason, unlocked_by, notify_agent}
  /// Response: {account_status, unlocked_at}
  static String get unlockAccount => _endpoint('/agents/{agent_id}/portal/unlock');

  /// Get session status
  /// GET /agents/portal/session/status
  /// Query: session_id
  /// Response: {session_id, is_valid, session_remaining_minutes}
  static String get portalSessionStatus => _endpoint('/agents/portal/session/status');

  // ============================================================================
  // UJ-006: SELF-SERVICE PROFILE UPDATE (3 APIs)
  // ============================================================================
  // Journey: Agent updates own profile

  /// Get self-service profile (read)
  /// GET /agents/{agent_id}/portal/profile
  /// Response: {agent_profile, workflow_state}
  static String get selfServiceProfile => _endpoint('/agents/{agent_id}/portal/profile');

  /// Update self-service profile (write)
  /// PUT /agents/{agent_id}/portal/profile
  /// Request: {update_type, otp, transaction_id, address_changes, contact_changes, bank_details_changes}
  /// Response: {update_status, changes_applied, audit_log}
  static String get updateSelfServiceProfile => _endpoint('/agents/{agent_id}/portal/profile');

  /// Request OTP for profile update
  /// POST /agents/{agent_id}/portal/profile/otp
  /// Request: {update_type}
  /// Response: {transaction_id, otp_sent_to, expires_at}
  static String get requestProfileOtp => _endpoint('/agents/{agent_id}/portal/profile/otp');

  /// Get agent dashboard
  /// GET /dashboard/agent/{agent_id}
  /// Response: {profile_summary, performance_metrics, pending_tasks, notifications}
  static String get agentDashboard => _endpoint('/dashboard/agent/{agent_id}');

  // ============================================================================
  // UJ-007: BANK DETAILS MANAGEMENT (5 APIs)
  // ============================================================================
  // Journey: Manage bank account details

  /// Get bank details (portal - masked)
  /// GET /agents/{agent_id}/portal/bank-details
  /// Response: {bank_details[] (masked)}
  static String get portalBankDetails => _endpoint('/agents/{agent_id}/portal/bank-details');

  /// Add bank details (portal)
  /// POST /agents/{agent_id}/portal/bank-details
  /// Request: {account_type, account_number, ifsc_code, bank_name, branch_name, otp}
  /// Response: {bank_details_id, account_number_masked, verified}
  static String get addPortalBankDetails => _endpoint('/agents/{agent_id}/portal/bank-details');

  /// Get bank details (admin - unmasked)
  /// GET /agents/{agent_id}/bank-details
  /// Response: {bank_details[]}
  static String get agentBankDetails => _endpoint('/agents/{agent_id}/bank-details');

  /// Add bank details (admin)
  /// POST /agents/{agent_id}/bank-details
  /// Request: {account_type, account_number, ifsc_code, bank_name, branch_name}
  /// Response: {bank_details_id, verified}
  static String get addAgentBankDetails => _endpoint('/agents/{agent_id}/bank-details');

  /// Update bank details
  /// PUT /agents/{agent_id}/bank-details
  /// Request: {account_number, ifsc_code, bank_name, branch_name}
  /// Response: {bank_details_id, updated}
  static String get updateBankDetails => _endpoint('/agents/{agent_id}/bank-details');

  // ============================================================================
  // UJ-008: GOAL SETTING (5 APIs)
  // ============================================================================
  // Journey: Set and track performance goals

  /// Get agent goals
  /// GET /agents/{agent_id}/goals
  /// Query: goal_period, status (ACTIVE, COMPLETED, CANCELLED)
  /// Response: {goals[], workflow_state}
  static String get agentGoals => _endpoint('/agents/{agent_id}/goals');

  /// Set agent goals
  /// POST /agents/{agent_id}/goals
  /// Request: {goal_period, targets, set_by}
  /// Response: {goal_id, goals_set}
  static String get setAgentGoals => _endpoint('/agents/{agent_id}/goals');

  /// Update goals
  /// PUT /agents/{agent_id}/goals/{goal_id}
  /// Request: {goal_period, targets}
  /// Response: {goal_id, updated}
  static String get updateGoals => _endpoint('/agents/{agent_id}/goals/{goal_id}');

  /// Delete goals
  /// DELETE /agents/{agent_id}/goals/{goal_id}
  /// Response: 204 No Content
  static String get deleteGoals => _endpoint('/agents/{agent_id}/goals/{goal_id}');

  /// Get goal progress
  /// GET /agents/{agent_id}/goals/progress
  /// Query: goal_id (optional)
  /// Response: {goals_progress[]}
  static String get goalProgress => _endpoint('/agents/{agent_id}/goals/progress');

  /// Get goal templates
  /// GET /goals/templates
  /// Response: {templates[]}
  static String get goalTemplates => _endpoint('/goals/templates');

  // ============================================================================
  // UJ-009: STATUS REINSTATEMENT (6 APIs)
  // ============================================================================
  // Journey: Reinstate terminated/deactivated agent

  /// Create reinstatement request
  /// POST /reinstatement/request
  /// Request: {agent_id, reinstatement_reason, effective_date, requested_by}
  /// Response: {request_id, reinstatement_status}
  static String get reinstatementRequest => _endpoint('/reinstatement/request');

  /// Approve reinstatement
  /// PUT /reinstatement/{request_id}/approve
  /// Request: {action, approver_comments, approved_by}
  /// Response: {approval_status, agent_id, reinstated_at}
  static String get approveReinstatement => _endpoint('/reinstatement/{request_id}/approve');

  /// Reject reinstatement
  /// PUT /reinstatement/{request_id}/reject
  /// Request: {action, rejector_comments, rejected_by}
  /// Response: {approval_status, rejected_at}
  static String get rejectReinstatement => _endpoint('/reinstatement/{request_id}/reject');

  /// Upload reinstatement documents
  /// POST /reinstatement/{request_id}/documents
  /// Request: Multipart form data (documents[])
  /// Response: {uploaded_document_ids[]}
  static String get uploadReinstatementDocs => _endpoint('/reinstatement/{request_id}/documents');

  /// Get reinstatement reasons dropdown
  /// GET /reinstatement/reasons
  /// Response: {reinstatement_reasons[]}
  static String get reinstatementReasons => _endpoint('/reinstatement/reasons');

  /// Get status types dropdown
  /// GET /status-types
  /// Response: {status_types[]}
  static String get statusTypes => _endpoint('/status-types');

  /// Reinstate agent (direct)
  /// POST /agents/{agent_id}/reinstate
  /// Request: {reinstatement_reason, effective_date, reinstated_by}
  /// Response: {request_id, reinstatement_status}
  static String get reinstateAgent => _endpoint('/agents/{agent_id}/reinstate');

  // ============================================================================
  // UJ-010: SEARCH & EXPORT (4 APIs)
  // ============================================================================
  // Journey: Advanced search and data export

  /// Configure export
  /// POST /agents/export/configure
  /// Request: {export_name, filters, fields, output_format}
  /// Response: {export_config_id}
  static String get configureExport => _endpoint('/agents/export/configure');

  /// Execute export
  /// POST /agents/export/execute
  /// Request: {export_config_id, requested_by}
  /// Response: {export_id, status}
  static String get executeExport => _endpoint('/agents/export/execute');

  /// Get export status
  /// GET /agents/export/{export_id}/status
  /// Response: {export_id, status, progress_percentage, file_url}
  static String get exportStatus => _endpoint('/agents/export/{export_id}/status');

  /// Download export file
  /// GET /agents/export/{export_id}/download
  /// Response: Binary file (Excel/PDF)
  static String get downloadExport => _endpoint('/agents/export/{export_id}/download');

  // ============================================================================
  // ADDITIONAL APIS
  // ============================================================================

  /// Get agent hierarchy
  /// GET /agents/{agent_id}/hierarchy
  /// Response: {hierarchy_chain[]}
  static String get agentHierarchy => _endpoint('/agents/{agent_id}/hierarchy');

  /// Get product authorization
  /// GET /agents/{agent_id}/product-authorization
  /// Response: {product_authorizations[]}
  static String get productAuthorization => _endpoint('/agents/{agent_id}/product-authorization');

  /// Grant product authorization
  /// POST /agents/{agent_id}/product-authorization
  /// Request: {product_class, effective_date, granted_by}
  /// Response: {authorization_id}
  static String get grantProductAuthorization => _endpoint('/agents/{agent_id}/product-authorization');

  /// Get agent timeline
  /// GET /agents/{agent_id}/timeline
  /// Query: from_date, to_date, activity_type, page, limit
  /// Response: {timeline[], pagination}
  static String get agentTimeline => _endpoint('/agents/{agent_id}/timeline');

  /// Get agent notifications
  /// GET /agents/{agent_id}/notifications
  /// Query: from_date, to_date, notification_type, page, limit
  /// Response: {notifications[], pagination}
  static String get agentNotifications => _endpoint('/agents/{agent_id}/notifications');

  /// HRMS employee update webhook
  /// POST /webhooks/hrms/employee-update
  /// Request: {event_id, event_type, timestamp, employee_data}
  /// Response: {status, event_id}
  static String get hrmsWebhook => _endpoint('/webhooks/hrms/employee-update');

  // ============================================================================
  // LOOKUP ENDPOINT FOR V1 AGENT TYPES
  // ============================================================================

  /// Fetch Agent Types with v1 prefix (special case)
  /// GET /v1/agents/lookup/agent-types
  /// Response: {agentTypes[]}
  static String get agentTypesV1 => '/v1/agents/lookup/agent-types';
}

// ============================================================================
// QUICK REFERENCE GUIDE
// ============================================================================
//
// HOW TO USE ENDPOINTS:
// 1. Import: import 'package:pli_agent_management/core/constants/api_endpoints_config.dart';
// 2. Use: final url = ApiEndpoints.agentProfile;
// 3. Replace params: final finalUrl = ApiEndpoints.replacePath(url, {'agent_id': 'AGT-001'});
//
// HOW TO UPDATE ENDPOINTS:
// 1. Find the endpoint in this file
// 2. Update the path string
// 3. Save - changes apply to entire app
//
// HOW TO ADD NEW ENDPOINT:
// 1. Add a new static getter in appropriate section
// 2. Add JSDoc comment with HTTP method, request, response
// 3. Use _endpoint() wrapper to apply version prefix
//
// EXAMPLE:
// /// Get agent details
// /// GET /agents/{agent_id}/details
// /// Response: {agent_details}
// static String get agentDetails => _endpoint('/agents/{agent_id}/details');
//
// ============================================================================

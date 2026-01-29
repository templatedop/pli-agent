/// API Endpoints Constants
/// Organized by User Journey
class ApiEndpoints {
  // Base
  static const String apiVersion = '/v1';

  // ============================================================================
  // UJ-001: Agent Profile Creation
  // ============================================================================
  static const String initiateProfile = '/agent-profiles/initiate';
  static const String fetchHrms = '/agent-profiles/{session_id}/fetch-hrms';
  static const String advisorCoordinators = '/advisor-coordinators';
  static const String linkCoordinator = '/agent-profiles/{session_id}/link-coordinator';
  static const String validateProfile = '/agent-profiles/{session_id}/validate-basic';
  static const String submitProfile = '/agent-profiles/{session_id}/submit';

  // ============================================================================
  // UJ-002: Agent Profile Update
  // ============================================================================
  static const String searchAgents = '/agents/search';
  static const String agentProfile = '/agents/{agent_id}';
  static const String updateForm = '/agents/{agent_id}/update-form';
  static const String updateSection = '/agents/{agent_id}/sections/{section}';
  static const String approveUpdate = '/approvals/{approval_request_id}/approve';
  static const String rejectUpdate = '/approvals/{approval_request_id}/reject';
  static const String auditHistory = '/agents/{agent_id}/audit-history';

  // ============================================================================
  // UJ-003: License Management
  // ============================================================================
  static const String agentLicenses = '/agents/{agent_id}/licenses';
  static const String licenseDetails = '/agents/{agent_id}/licenses/{license_id}';
  static const String renewLicense = '/agents/{agent_id}/licenses/{license_id}/renew';
  static const String licenseTypes = '/license-types';
  static const String expiringLicenses = '/licenses/expiring';
  static const String licenseReminders = '/licenses/{license_id}/reminders';
  static const String expiredLicenses = '/licenses/expired';

  // ============================================================================
  // UJ-004: Agent Termination
  // ============================================================================
  static const String terminateAgent = '/agents/{agent_id}/terminate';
  static const String terminationLetter = '/agents/{agent_id}/termination-letter';
  static const String terminationReasons = '/termination/reasons';

  // ============================================================================
  // UJ-005: Portal Authentication
  // ============================================================================
  static const String portalLogin = '/agents/{agent_id}/portal/login';
  static const String verifyOtp = '/agents/{agent_id}/portal/otp/verify';
  static const String portalLogout = '/agents/{agent_id}/portal/logout';
  static const String unlockAccount = '/agents/{agent_id}/portal/unlock';
  static const String sessionStatus = '/agents/portal/session/status';

  // ============================================================================
  // UJ-006: Self-Service Profile Update
  // ============================================================================
  static const String selfServiceProfile = '/agents/{agent_id}/portal/profile';
  static const String requestProfileOtp = '/agents/{agent_id}/portal/profile/otp';
  static const String agentDashboard = '/dashboard/agent/{agent_id}';

  // ============================================================================
  // UJ-007: Bank Details Management
  // ============================================================================
  static const String portalBankDetails = '/agents/{agent_id}/portal/bank-details';
  static const String agentBankDetails = '/agents/{agent_id}/bank-details';

  // ============================================================================
  // UJ-008: Goal Setting
  // ============================================================================
  static const String agentGoals = '/agents/{agent_id}/goals';
  static const String goalDetails = '/agents/{agent_id}/goals/{goal_id}';
  static const String goalProgress = '/agents/{agent_id}/goals/progress';
  static const String goalTemplates = '/goals/templates';

  // ============================================================================
  // UJ-009: Status Reinstatement
  // ============================================================================
  static const String reinstatementRequest = '/reinstatement/request';
  static const String approveReinstatement = '/reinstatement/{request_id}/approve';
  static const String rejectReinstatement = '/reinstatement/{request_id}/reject';
  static const String uploadReinstatementDocs = '/reinstatement/{request_id}/documents';
  static const String reinstatementReasons = '/reinstatement/reasons';
  static const String reinstateAgent = '/agents/{agent_id}/reinstate';

  // ============================================================================
  // UJ-010: Search & Export
  // ============================================================================
  static const String configureExport = '/agents/export/configure';
  static const String executeExport = '/agents/export/execute';
  static const String exportStatus = '/agents/export/{export_id}/status';
  static const String downloadExport = '/agents/export/{export_id}/download';

  // ============================================================================
  // Lookup APIs (Dropdowns)
  // ============================================================================
  static const String agentTypes = '/v1/agents/lookup/agent-types';
  static const String categories = '/categories';
  static const String designations = '/designations';
  static const String officeTypes = '/office-types';
  static const String states = '/states';
  static const String statusTypes = '/status-types';
  static const String agentHierarchy = '/agents/{agent_id}/hierarchy';

  // ============================================================================
  // Validation APIs
  // ============================================================================
  static const String validatePan = '/validations/pan/check-uniqueness';
  static const String validateEmployeeId = '/validations/hrms/employee-id';
  static const String validateIfsc = '/validations/bank/ifsc';
  static const String validateOffice = '/validations/office/{office_code}';

  // ============================================================================
  // Workflow APIs
  // ============================================================================
  static const String sessionStatusById = '/agent-profiles/sessions/{session_id}/status';
  static const String saveSession = '/agent-profiles/sessions/{session_id}/save';
  static const String resumeSession = '/agent-profiles/sessions/{session_id}/resume';
  static const String cancelSession = '/agent-profiles/sessions/{session_id}';

  // ============================================================================
  // Notification APIs
  // ============================================================================
  static const String resendWelcome = '/agents/{agent_id}/notifications/resend-welcome';
  static const String agentNotifications = '/agents/{agent_id}/notifications';

  // ============================================================================
  // Product Authorization
  // ============================================================================
  static const String productAuthorization = '/agents/{agent_id}/product-authorization';

  // ============================================================================
  // Timeline & Audit
  // ============================================================================
  static const String agentTimeline = '/agents/{agent_id}/timeline';
  static const String creationStatus = '/agent-profiles/creation-status/{agent_id}';

  // ============================================================================
  // Webhooks
  // ============================================================================
  static const String hrmsWebhook = '/webhooks/hrms/employee-update';

  // Helper method to replace path parameters
  static String replacePath(String path, Map<String, String> params) {
    String result = path;
    params.forEach((key, value) {
      result = result.replaceAll('{$key}', value);
    });
    return result;
  }
}

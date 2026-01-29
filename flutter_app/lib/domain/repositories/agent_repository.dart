/// Domain Repository Interface: Agent Repository
///
/// This is a contract (interface) that defines what operations are available
/// for agent data access. The actual implementation is in the data layer.
///
/// **Clean Architecture Layer**: Domain
/// **Dependencies**: Only domain entities and Either type
/// **Implementation**: See data/repositories/agent_repository_impl.dart
///
/// **Why Interface?**
/// - Decouples business logic from data source implementation
/// - Makes testing easy (mock implementations)
/// - Allows swapping data sources without changing domain logic

import 'package:dartz/dartz.dart';

import '../../core/errors/failures.dart';
import '../entities/agent_profile.dart';

/// Agent repository interface
///
/// Defines all operations related to agent profile management.
/// Returns Either<Failure, Success> for proper error handling.
abstract class AgentRepository {
  // ============================================================================
  // UJ-001: AGENT PROFILE CREATION
  // ============================================================================

  /// Initiate profile creation session
  ///
  /// **API**: POST /agent-profiles/initiate
  /// **Returns**: Either<Failure, SessionId>
  ///
  /// **Usage**:
  /// ```dart
  /// final result = await repository.initiateProfileCreation(
  ///   agentType: AgentType.advisor,
  ///   initiatedBy: 'admin_123',
  /// );
  ///
  /// result.fold(
  ///   (failure) => // Handle error,
  ///   (sessionId) => // Proceed with session,
  /// );
  /// ```
  Future<Either<Failure, String>> initiateProfileCreation({
    required AgentType agentType,
    required String initiatedBy,
  });

  /// Fetch employee data from HRMS
  ///
  /// **API**: POST /agent-profiles/{session_id}/fetch-hrms
  /// **Returns**: Either<Failure, HrmsData>
  Future<Either<Failure, Map<String, dynamic>>> fetchHrmsData({
    required String sessionId,
    required String employeeId,
  });

  /// Link advisor to coordinator
  ///
  /// **API**: POST /agent-profiles/{session_id}/link-coordinator
  /// **Returns**: Either<Failure, bool> (true if linked successfully)
  Future<Either<Failure, bool>> linkCoordinator({
    required String sessionId,
    required String coordinatorId,
    required DateTime effectiveDate,
  });

  /// Validate profile details
  ///
  /// **API**: POST /agent-profiles/{session_id}/validate-basic
  /// **Returns**: Either<Failure, ValidationResult>
  Future<Either<Failure, Map<String, dynamic>>> validateProfile({
    required String sessionId,
    required Map<String, dynamic> profileData,
  });

  /// Submit profile for creation
  ///
  /// **API**: POST /agent-profiles/{session_id}/submit
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> submitProfile({
    required String sessionId,
    required String submittedBy,
  });

  // ============================================================================
  // UJ-002: AGENT PROFILE UPDATE & SEARCH
  // ============================================================================

  /// Search agents by criteria
  ///
  /// **API**: GET /agents/search
  /// **Returns**: Either<Failure, List<AgentProfile>>
  ///
  /// **Usage**:
  /// ```dart
  /// final result = await repository.searchAgents(
  ///   criteria: {'status': 'ACTIVE', 'office_code': 'HO-001'},
  ///   page: 1,
  ///   limit: 20,
  /// );
  /// ```
  Future<Either<Failure, List<AgentProfile>>> searchAgents({
    Map<String, dynamic>? criteria,
    int page = 1,
    int limit = 20,
  });

  /// Get agent profile by ID
  ///
  /// **API**: GET /agents/{agent_id}
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> getAgentProfile(String agentId);

  /// Update agent profile section
  ///
  /// **API**: PUT /agents/{agent_id}/sections/{section}
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> updateAgentSection({
    required String agentId,
    required String section,
    required Map<String, dynamic> changes,
    required String updatedBy,
    String? updateReason,
  });

  // ============================================================================
  // UJ-004: AGENT TERMINATION
  // ============================================================================

  /// Terminate agent
  ///
  /// **API**: POST /agents/{agent_id}/terminate
  /// **Returns**: Either<Failure, bool>
  Future<Either<Failure, bool>> terminateAgent({
    required String agentId,
    required String terminationReason,
    required String terminationReasonCode,
    required DateTime effectiveDate,
    required String terminatedBy,
  });

  /// Get termination letter
  ///
  /// **API**: GET /agents/{agent_id}/termination-letter
  /// **Returns**: Either<Failure, bytes> (PDF file)
  Future<Either<Failure, List<int>>> getTerminationLetter(String agentId);

  // ============================================================================
  // UJ-009: STATUS REINSTATEMENT
  // ============================================================================

  /// Create reinstatement request
  ///
  /// **API**: POST /reinstatement/request
  /// **Returns**: Either<Failure, String> (request ID)
  Future<Either<Failure, String>> createReinstatementRequest({
    required String agentId,
    required String reinstatementReason,
    required DateTime effectiveDate,
    required String requestedBy,
  });

  /// Approve reinstatement
  ///
  /// **API**: PUT /reinstatement/{request_id}/approve
  /// **Returns**: Either<Failure, bool>
  Future<Either<Failure, bool>> approveReinstatement({
    required String requestId,
    required String approvedBy,
    String? comments,
  });

  // ============================================================================
  // AUDIT & TIMELINE
  // ============================================================================

  /// Get agent audit history
  ///
  /// **API**: GET /agents/{agent_id}/audit-history
  /// **Returns**: Either<Failure, List<AuditLog>>
  Future<Either<Failure, List<Map<String, dynamic>>>> getAuditHistory({
    required String agentId,
    DateTime? fromDate,
    DateTime? toDate,
    int page = 1,
    int limit = 20,
  });

  /// Get agent timeline
  ///
  /// **API**: GET /agents/{agent_id}/timeline
  /// **Returns**: Either<Failure, List<TimelineEvent>>
  Future<Either<Failure, List<Map<String, dynamic>>>> getAgentTimeline({
    required String agentId,
    DateTime? fromDate,
    DateTime? toDate,
    String? activityType,
  });

  // ============================================================================
  // LOOKUPS & DROPDOWNS
  // ============================================================================

  /// Get agent types dropdown
  ///
  /// **API**: GET /agent-types
  /// **Returns**: Either<Failure, List<AgentType>>
  Future<Either<Failure, List<Map<String, String>>>> getAgentTypes();

  /// Get status types dropdown
  ///
  /// **API**: GET /status-types
  /// **Returns**: Either<Failure, List<StatusType>>
  Future<Either<Failure, List<Map<String, String>>>> getStatusTypes();

  /// Get termination reasons dropdown
  ///
  /// **API**: GET /termination/reasons
  /// **Returns**: Either<Failure, List<TerminationReason>>
  Future<Either<Failure, List<Map<String, String>>>> getTerminationReasons();

  /// Get reinstatement reasons dropdown
  ///
  /// **API**: GET /reinstatement/reasons
  /// **Returns**: Either<Failure, List<ReinstatementReason>>
  Future<Either<Failure, List<Map<String, String>>>> getReinstatementReasons();

  /// Get categories dropdown
  ///
  /// **API**: GET /categories
  /// **Returns**: Either<Failure, List<Category>>
  Future<Either<Failure, List<Map<String, String>>>> getCategories();

  /// Get states dropdown
  ///
  /// **API**: GET /states
  /// **Returns**: Either<Failure, List<State>>
  Future<Either<Failure, List<Map<String, String>>>> getStates();

  /// Get office types dropdown
  ///
  /// **API**: GET /office-types
  /// **Returns**: Either<Failure, List<OfficeType>>
  Future<Either<Failure, List<Map<String, String>>>> getOfficeTypes();

  /// Get designations dropdown
  ///
  /// **API**: GET /designations
  /// **Returns**: Either<Failure, List<Designation>>
  Future<Either<Failure, List<Map<String, String>>>> getDesignations();

  // ============================================================================
  // VALIDATION
  // ============================================================================

  /// Validate PAN uniqueness
  ///
  /// **API**: POST /validations/pan/check-uniqueness
  /// **Returns**: Either<Failure, bool> (true if unique)
  Future<Either<Failure, bool>> validatePanUniqueness({
    required String panNumber,
    String? excludeAgentId,
  });

  /// Validate employee ID in HRMS
  ///
  /// **API**: POST /validations/hrms/employee-id
  /// **Returns**: Either<Failure, bool> (true if valid)
  Future<Either<Failure, bool>> validateEmployeeId(String employeeId);

  /// Validate IFSC code
  ///
  /// **API**: POST /validations/bank/ifsc
  /// **Returns**: Either<Failure, Map> (bank details if valid)
  Future<Either<Failure, Map<String, String>>> validateIfscCode(
    String ifscCode,
  );

  /// Validate office code
  ///
  /// **API**: GET /validations/office/{office_code}
  /// **Returns**: Either<Failure, Map> (office details if valid)
  Future<Either<Failure, Map<String, dynamic>>> validateOfficeCode(
    String officeCode,
  );

  // ============================================================================
  // COORDINATOR MANAGEMENT
  // ============================================================================

  /// Get list of active advisor coordinators
  ///
  /// **API**: GET /advisor-coordinators
  /// **Returns**: Either<Failure, List<Coordinator>>
  Future<Either<Failure, List<Map<String, dynamic>>>> getAdvisorCoordinators({
    String? status,
    String? circleId,
    String? divisionId,
    int page = 1,
    int limit = 20,
  });

  /// Get agent hierarchy
  ///
  /// **API**: GET /agents/{agent_id}/hierarchy
  /// **Returns**: Either<Failure, Map> (hierarchy chain)
  Future<Either<Failure, Map<String, dynamic>>> getAgentHierarchy(
    String agentId,
  );

  // ============================================================================
  // SESSION MANAGEMENT (for workflow)
  // ============================================================================

  /// Get session status
  ///
  /// **API**: GET /agent-profiles/sessions/{session_id}/status
  /// **Returns**: Either<Failure, Map> (session status)
  Future<Either<Failure, Map<String, dynamic>>> getSessionStatus(
    String sessionId,
  );

  /// Save session data (checkpoint)
  ///
  /// **API**: POST /agent-profiles/sessions/{session_id}/save
  /// **Returns**: Either<Failure, bool>
  Future<Either<Failure, bool>> saveSession({
    required String sessionId,
    required Map<String, dynamic> formData,
  });

  /// Resume session
  ///
  /// **API**: GET /agent-profiles/sessions/{session_id}/resume
  /// **Returns**: Either<Failure, Map> (saved session data)
  Future<Either<Failure, Map<String, dynamic>>> resumeSession(
    String sessionId,
  );

  /// Cancel session
  ///
  /// **API**: DELETE /agent-profiles/sessions/{session_id}
  /// **Returns**: Either<Failure, bool>
  Future<Either<Failure, bool>> cancelSession(String sessionId);
}

// ============================================================================
// NOTES ON USAGE
// ============================================================================
//
// This interface uses the Either type from dartz package for functional
// error handling:
//
// - Left side: Failure (error case)
// - Right side: Success value
//
// Example usage in a use case:
// ```dart
// final result = await repository.getAgentProfile('AGT-001');
//
// return result.fold(
//   (failure) => Left(failure),           // Pass failure up
//   (agentProfile) {
//     // Do business logic
//     return Right(agentProfile);         // Return success
//   },
// );
// ```
//
// Benefits of this approach:
// 1. Forces explicit error handling
// 2. Type-safe (no exceptions)
// 3. Composable with functional operators
// 4. Clear separation of success/failure paths
//
// ============================================================================

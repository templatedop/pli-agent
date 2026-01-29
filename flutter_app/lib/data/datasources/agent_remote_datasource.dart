/// Remote Data Source: Agent Profile
///
/// This class handles all API calls related to agent profiles.
/// It uses the ApiClient for HTTP communication and returns data models.
///
/// **Responsibilities**:
/// - Make HTTP requests to agent profile APIs
/// - Handle API responses
/// - Throw appropriate exceptions on errors
/// - Return data models (not domain entities)
///
/// **Usage**:
/// ```dart
/// final dataSource = AgentRemoteDataSource(apiClient);
/// final agentModel = await dataSource.getAgentProfile('AGT-2026-000001');
/// ```

import 'package:dio/dio.dart';
import '../../core/constants/api_endpoints_config.dart';
import '../../core/errors/exceptions.dart';
import '../../core/network/api_client.dart';
import '../models/agent_model.dart';

/// Remote data source for agent profile operations
class AgentRemoteDataSource {
  final ApiClient apiClient;

  AgentRemoteDataSource(this.apiClient);

  // ==========================================================================
  // UJ-001: AGENT PROFILE CREATION
  // ==========================================================================

  /// Step 1: Initiate profile creation session
  /// POST /agent-profiles/initiate
  Future<Map<String, dynamic>> initiateProfileCreation({
    required String agentType,
    required String initiatedBy,
  }) async {
    try {
      final response = await apiClient.post(
        ApiEndpoints.initiateProfile,
        data: {
          'agent_type': agentType,
          'initiated_by': initiatedBy,
        },
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to initiate profile creation',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Step 2: Fetch employee data from HRMS
  /// POST /agent-profiles/{session_id}/fetch-hrms
  Future<Map<String, dynamic>> fetchHrmsData({
    required String sessionId,
    required String employeeId,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.fetchHrms,
        {'session_id': sessionId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {'employee_id': employeeId},
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to fetch HRMS data',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Step 3: Link coordinator
  /// POST /agent-profiles/{session_id}/link-coordinator
  Future<Map<String, dynamic>> linkCoordinator({
    required String sessionId,
    required String coordinatorId,
    required DateTime effectiveDate,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.linkCoordinator,
        {'session_id': sessionId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {
          'coordinator_id': coordinatorId,
          'effective_date': effectiveDate.toIso8601String(),
        },
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to link coordinator',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Step 4: Validate profile data
  /// POST /agent-profiles/{session_id}/validate
  Future<Map<String, dynamic>> validateProfile({
    required String sessionId,
    required Map<String, dynamic> profileData,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.validateProfile,
        {'session_id': sessionId},
      );

      final response = await apiClient.post(
        endpoint,
        data: profileData,
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Profile validation failed',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Step 5: Submit profile for creation
  /// POST /agent-profiles/{session_id}/submit
  Future<AgentModel> submitProfile({
    required String sessionId,
    required String submittedBy,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.submitProfile,
        {'session_id': sessionId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {'submitted_by': submittedBy},
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return AgentModel.fromJson(response.data as Map<String, dynamic>);
      } else {
        throw ServerException(
          message: 'Failed to submit profile',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // UJ-002: AGENT PROFILE SEARCH & RETRIEVAL
  // ==========================================================================

  /// Search agents with criteria
  /// GET /agents/search
  Future<List<AgentModel>> searchAgents({
    Map<String, dynamic>? criteria,
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final queryParams = {
        'page': page.toString(),
        'limit': limit.toString(),
        if (criteria != null) ...criteria,
      };

      final response = await apiClient.get(
        ApiEndpoints.searchAgents,
        queryParameters: queryParams,
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final agents = data['agents'] as List;
        return agents
            .map((json) => AgentModel.fromJson(json as Map<String, dynamic>))
            .toList();
      } else {
        throw ServerException(
          message: 'Failed to search agents',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get agent profile by ID
  /// GET /agents/{agent_id}
  Future<AgentModel> getAgentProfile(String agentId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.agentProfile,
        {'agent_id': agentId},
      );

      final response = await apiClient.get(endpoint);

      if (response.statusCode == 200) {
        return AgentModel.fromJson(response.data as Map<String, dynamic>);
      } else if (response.statusCode == 404) {
        throw NotFoundException(message: 'Agent not found');
      } else {
        throw ServerException(
          message: 'Failed to get agent profile',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get agent hierarchy
  /// GET /agents/{agent_id}/hierarchy
  Future<Map<String, dynamic>> getAgentHierarchy(String agentId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.agentHierarchy,
        {'agent_id': agentId},
      );

      final response = await apiClient.get(endpoint);

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to get agent hierarchy',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // UJ-002: AGENT PROFILE UPDATES
  // ==========================================================================

  /// Update agent profile section
  /// PATCH /agents/{agent_id}/{section}
  Future<AgentModel> updateAgentSection({
    required String agentId,
    required String section,
    required Map<String, dynamic> changes,
    required String updatedBy,
    String? updateReason,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.updateAgentSection,
        {'agent_id': agentId, 'section': section},
      );

      final response = await apiClient.patch(
        endpoint,
        data: {
          'changes': changes,
          'updated_by': updatedBy,
          if (updateReason != null) 'update_reason': updateReason,
        },
      );

      if (response.statusCode == 200) {
        return AgentModel.fromJson(response.data as Map<String, dynamic>);
      } else {
        throw ServerException(
          message: 'Failed to update agent section',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // UJ-004: AGENT TERMINATION
  // ==========================================================================

  /// Terminate agent
  /// POST /agents/{agent_id}/terminate
  Future<Map<String, dynamic>> terminateAgent({
    required String agentId,
    required String terminationReason,
    required DateTime effectiveDate,
    required String terminatedBy,
    String? remarks,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.terminateAgent,
        {'agent_id': agentId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {
          'termination_reason': terminationReason,
          'effective_date': effectiveDate.toIso8601String(),
          'terminated_by': terminatedBy,
          if (remarks != null) 'remarks': remarks,
        },
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to terminate agent',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get termination letter
  /// GET /agents/{agent_id}/termination-letter
  Future<List<int>> getTerminationLetter(String agentId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.terminationLetter,
        {'agent_id': agentId},
      );

      final response = await apiClient.get(
        endpoint,
        options: Options(responseType: ResponseType.bytes),
      );

      if (response.statusCode == 200) {
        return response.data as List<int>;
      } else {
        throw ServerException(
          message: 'Failed to get termination letter',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // UJ-009: AGENT REINSTATEMENT
  // ==========================================================================

  /// Create reinstatement request
  /// POST /agents/{agent_id}/reinstatement-requests
  Future<Map<String, dynamic>> createReinstatementRequest({
    required String agentId,
    required String requestedBy,
    required String justification,
    DateTime? effectiveDate,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.reinstatementRequest,
        {'agent_id': agentId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {
          'requested_by': requestedBy,
          'justification': justification,
          if (effectiveDate != null)
            'effective_date': effectiveDate.toIso8601String(),
        },
      );

      if (response.statusCode == 201) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to create reinstatement request',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Approve reinstatement
  /// POST /agents/{agent_id}/reinstatement-requests/{request_id}/approve
  Future<Map<String, dynamic>> approveReinstatement({
    required String agentId,
    required String requestId,
    required String approvedBy,
    String? remarks,
  }) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.approveReinstatement,
        {'agent_id': agentId, 'request_id': requestId},
      );

      final response = await apiClient.post(
        endpoint,
        data: {
          'approved_by': approvedBy,
          if (remarks != null) 'remarks': remarks,
        },
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to approve reinstatement',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // LOOKUP APIs
  // ==========================================================================

  /// Get agent types dropdown
  /// GET /lookups/agent-types
  Future<List<Map<String, dynamic>>> getAgentTypes() async {
    try {
      final response = await apiClient.get(ApiEndpoints.agentTypes);

      if (response.statusCode == 200) {
        return (response.data as List)
            .map((e) => e as Map<String, dynamic>)
            .toList();
      } else {
        throw ServerException(
          message: 'Failed to get agent types',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get status types dropdown
  /// GET /lookups/status-types
  Future<List<Map<String, dynamic>>> getStatusTypes() async {
    try {
      final response = await apiClient.get(ApiEndpoints.statusTypes);

      if (response.statusCode == 200) {
        return (response.data as List)
            .map((e) => e as Map<String, dynamic>)
            .toList();
      } else {
        throw ServerException(
          message: 'Failed to get status types',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // VALIDATION APIs
  // ==========================================================================

  /// Validate PAN uniqueness
  /// POST /validations/pan-uniqueness
  Future<Map<String, dynamic>> validatePanUniqueness(String panNumber) async {
    try {
      final response = await apiClient.post(
        ApiEndpoints.validatePan,
        data: {'pan_number': panNumber},
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'PAN validation failed',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Validate employee ID (HRMS)
  /// POST /validations/employee-id
  Future<Map<String, dynamic>> validateEmployeeId(String employeeId) async {
    try {
      final response = await apiClient.post(
        ApiEndpoints.validateEmployeeId,
        data: {'employee_id': employeeId},
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Employee ID validation failed',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Validate IFSC code
  /// POST /validations/ifsc-code
  Future<Map<String, dynamic>> validateIfscCode(String ifscCode) async {
    try {
      final response = await apiClient.post(
        ApiEndpoints.validateIfsc,
        data: {'ifsc_code': ifscCode},
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'IFSC validation failed',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Validate office code
  /// POST /validations/office-code
  Future<Map<String, dynamic>> validateOfficeCode(String officeCode) async {
    try {
      final response = await apiClient.post(
        ApiEndpoints.validateOfficeCode,
        data: {'office_code': officeCode},
      );

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Office code validation failed',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // SESSION MANAGEMENT
  // ==========================================================================

  /// Get session status
  /// GET /agent-profiles/sessions/{session_id}/status
  Future<Map<String, dynamic>> getSessionStatus(String sessionId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.sessionStatus,
        {'session_id': sessionId},
      );

      final response = await apiClient.get(endpoint);

      if (response.statusCode == 200) {
        return response.data as Map<String, dynamic>;
      } else {
        throw ServerException(
          message: 'Failed to get session status',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Cancel session
  /// POST /agent-profiles/sessions/{session_id}/cancel
  Future<void> cancelSession(String sessionId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.cancelSession,
        {'session_id': sessionId},
      );

      final response = await apiClient.post(endpoint);

      if (response.statusCode != 200) {
        throw ServerException(
          message: 'Failed to cancel session',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  // ==========================================================================
  // ERROR HANDLING
  // ==========================================================================

  /// Handle Dio errors and convert to appropriate exceptions
  Exception _handleDioError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return NetworkException(message: 'Connection timeout');

      case DioExceptionType.connectionError:
        return NetworkException(message: 'No internet connection');

      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final message = error.response?.data?['message'] ?? 'Server error';

        if (statusCode == 401) {
          return UnauthorizedException(message: 'Unauthorized');
        } else if (statusCode == 404) {
          return NotFoundException(message: 'Resource not found');
        } else if (statusCode == 409) {
          return ConflictException(
            message: message,
            data: error.response?.data,
          );
        } else if (statusCode == 422) {
          return ValidationException(
            message: message,
            data: error.response?.data,
          );
        } else {
          return ServerException(
            message: message,
            statusCode: statusCode,
          );
        }

      case DioExceptionType.cancel:
        return ServerException(message: 'Request cancelled');

      default:
        return ServerException(message: 'Unknown error occurred');
    }
  }
}

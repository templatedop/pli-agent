/// Repository Implementation: Agent Profile
///
/// This class implements the AgentRepository interface from the domain layer.
/// It connects the domain layer with the data layer.
///
/// **Responsibilities**:
/// - Implement all repository methods
/// - Call remote data source
/// - Convert models to entities
/// - Handle exceptions and return Either<Failure, Success>
/// - Add caching logic (future enhancement)
///
/// **Usage**:
/// ```dart
/// final repository = AgentRepositoryImpl(remoteDataSource: dataSource);
/// final result = await repository.getAgentProfile('AGT-2026-000001');
///
/// result.fold(
///   (failure) => // Handle error,
///   (agentProfile) => // Use entity,
/// );
/// ```

import 'package:dartz/dartz.dart';
import '../../core/errors/exceptions.dart';
import '../../core/errors/failures.dart';
import '../../domain/entities/agent_profile.dart';
import '../../domain/repositories/agent_repository.dart';
import '../datasources/agent_remote_datasource.dart';

/// Implementation of AgentRepository
class AgentRepositoryImpl implements AgentRepository {
  final AgentRemoteDataSource remoteDataSource;

  AgentRepositoryImpl({required this.remoteDataSource});

  // ==========================================================================
  // UJ-001: AGENT PROFILE CREATION
  // ==========================================================================

  @override
  Future<Either<Failure, String>> initiateProfileCreation({
    required AgentType agentType,
    required String initiatedBy,
  }) async {
    try {
      final result = await remoteDataSource.initiateProfileCreation(
        agentType: _agentTypeToString(agentType),
        initiatedBy: initiatedBy,
      );

      final sessionId = result['session_id'] as String;
      return Right(sessionId);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> fetchHrmsData({
    required String sessionId,
    required String employeeId,
  }) async {
    try {
      final result = await remoteDataSource.fetchHrmsData(
        sessionId: sessionId,
        employeeId: employeeId,
      );

      return Right(result);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, bool>> linkCoordinator({
    required String sessionId,
    required String coordinatorId,
    required DateTime effectiveDate,
  }) async {
    try {
      await remoteDataSource.linkCoordinator(
        sessionId: sessionId,
        coordinatorId: coordinatorId,
        effectiveDate: effectiveDate,
      );

      return const Right(true);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> validateProfile({
    required String sessionId,
    required Map<String, dynamic> profileData,
  }) async {
    try {
      final result = await remoteDataSource.validateProfile(
        sessionId: sessionId,
        profileData: profileData,
      );

      return Right(result);
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, AgentProfile>> submitProfile({
    required String sessionId,
    required String submittedBy,
  }) async {
    try {
      final model = await remoteDataSource.submitProfile(
        sessionId: sessionId,
        submittedBy: submittedBy,
      );

      return Right(model.toEntity());
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // UJ-002: AGENT PROFILE SEARCH & RETRIEVAL
  // ==========================================================================

  @override
  Future<Either<Failure, List<AgentProfile>>> searchAgents({
    Map<String, dynamic>? criteria,
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final models = await remoteDataSource.searchAgents(
        criteria: criteria,
        page: page,
        limit: limit,
      );

      final entities = models.map((model) => model.toEntity()).toList();
      return Right(entities);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, AgentProfile>> getAgentProfile(String agentId) async {
    try {
      final model = await remoteDataSource.getAgentProfile(agentId);
      return Right(model.toEntity());
    } on NotFoundException catch (e) {
      return Left(NotFoundFailure(message: e.message));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> getAgentHierarchy(
    String agentId,
  ) async {
    try {
      final result = await remoteDataSource.getAgentHierarchy(agentId);
      return Right(result);
    } on NotFoundException catch (e) {
      return Left(NotFoundFailure(message: e.message));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // UJ-002: AGENT PROFILE UPDATES
  // ==========================================================================

  @override
  Future<Either<Failure, AgentProfile>> updateAgentSection({
    required String agentId,
    required String section,
    required Map<String, dynamic> changes,
    required String updatedBy,
    String? updateReason,
  }) async {
    try {
      final model = await remoteDataSource.updateAgentSection(
        agentId: agentId,
        section: section,
        changes: changes,
        updatedBy: updatedBy,
        updateReason: updateReason,
      );

      return Right(model.toEntity());
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // UJ-004: AGENT TERMINATION
  // ==========================================================================

  @override
  Future<Either<Failure, bool>> terminateAgent({
    required String agentId,
    required String terminationReason,
    required DateTime effectiveDate,
    required String terminatedBy,
    String? remarks,
  }) async {
    try {
      await remoteDataSource.terminateAgent(
        agentId: agentId,
        terminationReason: terminationReason,
        effectiveDate: effectiveDate,
        terminatedBy: terminatedBy,
        remarks: remarks,
      );

      return const Right(true);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, List<int>>> getTerminationLetter(
    String agentId,
  ) async {
    try {
      final bytes = await remoteDataSource.getTerminationLetter(agentId);
      return Right(bytes);
    } on NotFoundException catch (e) {
      return Left(NotFoundFailure(message: e.message));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // UJ-009: AGENT REINSTATEMENT
  // ==========================================================================

  @override
  Future<Either<Failure, String>> createReinstatementRequest({
    required String agentId,
    required String requestedBy,
    required String justification,
    DateTime? effectiveDate,
  }) async {
    try {
      final result = await remoteDataSource.createReinstatementRequest(
        agentId: agentId,
        requestedBy: requestedBy,
        justification: justification,
        effectiveDate: effectiveDate,
      );

      final requestId = result['request_id'] as String;
      return Right(requestId);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, bool>> approveReinstatement({
    required String agentId,
    required String requestId,
    required String approvedBy,
    String? remarks,
  }) async {
    try {
      await remoteDataSource.approveReinstatement(
        agentId: agentId,
        requestId: requestId,
        approvedBy: approvedBy,
        remarks: remarks,
      );

      return const Right(true);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // LOOKUP APIs
  // ==========================================================================

  @override
  Future<Either<Failure, List<Map<String, dynamic>>>> getAgentTypes() async {
    try {
      final result = await remoteDataSource.getAgentTypes();
      return Right(result);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, List<Map<String, dynamic>>>> getStatusTypes() async {
    try {
      final result = await remoteDataSource.getStatusTypes();
      return Right(result);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // VALIDATION APIs
  // ==========================================================================

  @override
  Future<Either<Failure, bool>> validatePanUniqueness(String panNumber) async {
    try {
      final result = await remoteDataSource.validatePanUniqueness(panNumber);
      final isUnique = result['is_unique'] as bool;
      return Right(isUnique);
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on ConflictException catch (e) {
      return Left(ConflictFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> validateEmployeeId(
    String employeeId,
  ) async {
    try {
      final result = await remoteDataSource.validateEmployeeId(employeeId);
      return Right(result);
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> validateIfscCode(
    String ifscCode,
  ) async {
    try {
      final result = await remoteDataSource.validateIfscCode(ifscCode);
      return Right(result);
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> validateOfficeCode(
    String officeCode,
  ) async {
    try {
      final result = await remoteDataSource.validateOfficeCode(officeCode);
      return Right(result);
    } on ValidationException catch (e) {
      return Left(ValidationFailure(
        message: e.message,
        data: e.data,
      ));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // SESSION MANAGEMENT
  // ==========================================================================

  @override
  Future<Either<Failure, Map<String, dynamic>>> getSessionStatus(
    String sessionId,
  ) async {
    try {
      final result = await remoteDataSource.getSessionStatus(sessionId);
      return Right(result);
    } on NotFoundException catch (e) {
      return Left(NotFoundFailure(message: e.message));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  @override
  Future<Either<Failure, void>> cancelSession(String sessionId) async {
    try {
      await remoteDataSource.cancelSession(sessionId);
      return const Right(null);
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }

  // ==========================================================================
  // HELPER METHODS
  // ==========================================================================

  /// Convert AgentType enum to API string
  String _agentTypeToString(AgentType type) {
    switch (type) {
      case AgentType.advisor:
        return 'advisor';
      case AgentType.advisorCoordinator:
        return 'advisor_coordinator';
      case AgentType.departmentalEmployee:
        return 'departmental_employee';
      case AgentType.fieldOfficer:
        return 'field_officer';
      case AgentType.directAgent:
        return 'direct_agent';
      case AgentType.gds:
        return 'gds';
    }
  }
}

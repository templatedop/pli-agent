/// Use Case: Get Agent Profile
///
/// Retrieves a complete agent profile by agent ID.
/// Simple use case that delegates to repository.
///
/// **Usage**:
/// ```dart
/// final useCase = GetAgentProfileUseCase(repository);
/// final result = await useCase('AGT-2026-000001');
///
/// result.fold(
///   (failure) => // Handle error,
///   (agentProfile) => // Display profile,
/// );
/// ```

import 'package:dartz/dartz.dart';

import '../../core/errors/failures.dart';
import '../entities/agent_profile.dart';
import '../repositories/agent_repository.dart';

/// Use case for retrieving agent profile
class GetAgentProfileUseCase {
  final AgentRepository repository;

  GetAgentProfileUseCase(this.repository);

  /// Execute the use case
  ///
  /// **Parameters**: agentId (e.g., 'AGT-2026-000001')
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> call(String agentId) async {
    // Validate agent ID format
    if (!_isValidAgentId(agentId)) {
      return Left(
        ValidationFailure(
          message: 'Invalid agent ID format',
          data: {'agentId': agentId},
        ),
      );
    }

    // Fetch from repository
    final result = await repository.getAgentProfile(agentId);

    return result.fold(
      (failure) => Left(failure),
      (agentProfile) {
        // Business logic: Check if profile is accessible
        // (Additional checks can be added here)

        return Right(agentProfile);
      },
    );
  }

  /// Validate agent ID format (AGT-YYYY-NNNNNN)
  bool _isValidAgentId(String agentId) {
    final regex = RegExp(r'^AGT-\d{4}-\d{6}$');
    return regex.hasMatch(agentId);
  }
}

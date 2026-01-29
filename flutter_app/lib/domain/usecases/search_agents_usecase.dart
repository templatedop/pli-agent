/// Use Case: Search Agents
///
/// Searches for agents based on multiple criteria with pagination.
///
/// **Usage**:
/// ```dart
/// final useCase = SearchAgentsUseCase(repository);
///
/// final result = await useCase(SearchAgentsParams(
///   criteria: {
///     'status': 'ACTIVE',
///     'office_code': 'HO-001',
///   },
///   page: 1,
///   limit: 20,
/// ));
/// ```

import 'package:dartz/dartz.dart';
import 'package:equatable/equatable.dart';

import '../../core/errors/failures.dart';
import '../entities/agent_profile.dart';
import '../repositories/agent_repository.dart';

/// Use case for searching agents
class SearchAgentsUseCase {
  final AgentRepository repository;

  SearchAgentsUseCase(this.repository);

  /// Execute the use case
  ///
  /// **Parameters**: SearchAgentsParams
  /// **Returns**: Either<Failure, List<AgentProfile>>
  Future<Either<Failure, List<AgentProfile>>> call(
    SearchAgentsParams params,
  ) async {
    // Validate pagination parameters
    if (params.page < 1) {
      return Left(
        const ValidationFailure(message: 'Page must be >= 1'),
      );
    }

    if (params.limit < 1 || params.limit > 100) {
      return Left(
        const ValidationFailure(message: 'Limit must be between 1 and 100'),
      );
    }

    // Fetch from repository
    final result = await repository.searchAgents(
      criteria: params.criteria,
      page: params.page,
      limit: params.limit,
    );

    return result.fold(
      (failure) => Left(failure),
      (agents) {
        // Business logic: Filter out inaccessible profiles if needed
        // (Additional filtering can be added here based on user role/permissions)

        return Right(agents);
      },
    );
  }
}

/// Parameters for SearchAgentsUseCase
class SearchAgentsParams extends Equatable {
  /// Search criteria (e.g., {'status': 'ACTIVE', 'name': 'John'})
  final Map<String, dynamic>? criteria;

  /// Page number (1-indexed)
  final int page;

  /// Items per page (max 100)
  final int limit;

  const SearchAgentsParams({
    this.criteria,
    this.page = 1,
    this.limit = 20,
  });

  @override
  List<Object?> get props => [criteria, page, limit];
}

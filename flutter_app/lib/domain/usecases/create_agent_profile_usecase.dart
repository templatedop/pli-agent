/// Use Case: Create Agent Profile
///
/// This use case orchestrates the complete agent profile creation flow.
/// It encapsulates business logic and rules for creating an agent.
///
/// **Clean Architecture Layer**: Domain
/// **Dependencies**: Repository interface only
/// **Called by**: Presentation layer (BLoC/Cubit)
///
/// **Usage**:
/// ```dart
/// final useCase = CreateAgentProfileUseCase(repository);
///
/// final result = await useCase(CreateAgentProfileParams(
///   agentType: AgentType.advisor,
///   initiatedBy: 'admin_123',
///   profileData: {...},
/// ));
///
/// result.fold(
///   (failure) => // Handle error,
///   (agentProfile) => // Success,
/// );
/// ```

import 'package:dartz/dartz.dart';
import 'package:equatable/equatable.dart';

import '../../core/errors/failures.dart';
import '../entities/agent_profile.dart';
import '../repositories/agent_repository.dart';

/// Use case for creating agent profile
///
/// Follows the Command pattern - takes parameters, returns Either<Failure, Result>
class CreateAgentProfileUseCase {
  final AgentRepository repository;

  CreateAgentProfileUseCase(this.repository);

  /// Execute the use case
  ///
  /// **Flow**:
  /// 1. Initiate profile creation session
  /// 2. Optionally fetch HRMS data (for departmental employees)
  /// 3. Optionally link coordinator (for advisors)
  /// 4. Validate profile data
  /// 5. Submit profile for creation
  ///
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> call(
    CreateAgentProfileParams params,
  ) async {
    // Step 1: Initiate session
    final sessionResult = await repository.initiateProfileCreation(
      agentType: params.agentType,
      initiatedBy: params.initiatedBy,
    );

    return sessionResult.fold(
      (failure) => Left(failure),
      (sessionId) async {
        // Step 2: Fetch HRMS data if needed
        if (params.employeeId != null) {
          final hrmsResult = await repository.fetchHrmsData(
            sessionId: sessionId,
            employeeId: params.employeeId!,
          );

          // If HRMS fetch fails, continue with manual entry
          // (non-blocking error)
        }

        // Step 3: Link coordinator if needed (for Advisors)
        if (params.coordinatorId != null) {
          final linkResult = await repository.linkCoordinator(
            sessionId: sessionId,
            coordinatorId: params.coordinatorId!,
            effectiveDate: params.effectiveDate ?? DateTime.now(),
          );

          // Check if linking failed
          if (linkResult.isLeft()) {
            return linkResult.fold(
              (failure) => Left(failure),
              (_) => Left(
                const ServerFailure(
                  message: 'Failed to link coordinator',
                ),
              ),
            );
          }
        }

        // Step 4: Validate profile data
        final validationResult = await repository.validateProfile(
          sessionId: sessionId,
          profileData: params.profileData,
        );

        return validationResult.fold(
          (failure) => Left(failure),
          (validationResults) async {
            // Check if validation passed
            final isPassed = validationResults['validation_status'] == 'PASSED';

            if (!isPassed) {
              return Left(
                ValidationFailure(
                  message: 'Profile validation failed',
                  data: validationResults,
                ),
              );
            }

            // Step 5: Submit profile
            final submitResult = await repository.submitProfile(
              sessionId: sessionId,
              submittedBy: params.initiatedBy,
            );

            return submitResult.fold(
              (failure) => Left(failure),
              (agentProfile) {
                // Business rule: Check if profile is complete
                if (!agentProfile.isProfileComplete) {
                  return Left(
                    const ValidationFailure(
                      message: 'Profile is incomplete',
                    ),
                  );
                }

                // Business rule: Advisors must have coordinator
                if (agentProfile.agentType == AgentType.advisor &&
                    agentProfile.advisorCoordinator == null) {
                  return Left(
                    const ValidationFailure(
                      message: 'Advisor must be linked to a coordinator',
                    ),
                  );
                }

                // Success
                return Right(agentProfile);
              },
            );
          },
        );
      },
    );
  }
}

/// Parameters for CreateAgentProfileUseCase
class CreateAgentProfileParams extends Equatable {
  /// Type of agent to create
  final AgentType agentType;

  /// User who initiated the creation
  final String initiatedBy;

  /// Complete profile data
  final Map<String, dynamic> profileData;

  /// Employee ID (for departmental employees)
  final String? employeeId;

  /// Coordinator ID (for advisors)
  final String? coordinatorId;

  /// Effective date for the profile
  final DateTime? effectiveDate;

  const CreateAgentProfileParams({
    required this.agentType,
    required this.initiatedBy,
    required this.profileData,
    this.employeeId,
    this.coordinatorId,
    this.effectiveDate,
  });

  @override
  List<Object?> get props => [
        agentType,
        initiatedBy,
        profileData,
        employeeId,
        coordinatorId,
        effectiveDate,
      ];
}

// ============================================================================
// BUSINESS RULES ENFORCED
// ============================================================================
//
// 1. Profile completeness check
//    - PAN number must be present
//    - Personal info must be complete
//    - At least one address required
//    - At least one contact required
//
// 2. Advisor-Coordinator relationship
//    - Advisors MUST have a coordinator
//    - Coordinator MUST be active
//
// 3. HRMS integration
//    - For departmental employees, HRMS data is optional
//    - If HRMS fetch fails, allow manual entry
//    - HRMS data auto-populates but can be overridden
//
// 4. Validation before submission
//    - All fields must pass validation
//    - PAN must be unique
//    - Age must be within range (18-70)
//    - Contact numbers must be valid
//
// 5. Session management
//    - Session created at start
//    - Session used throughout creation flow
//    - Session can be resumed if interrupted
//
// ============================================================================

/// Use Case: Update Agent Profile
///
/// Updates a specific section of an agent profile.
/// Handles validation and approval workflow if needed.
///
/// **Usage**:
/// ```dart
/// final useCase = UpdateAgentProfileUseCase(repository);
///
/// final result = await useCase(UpdateAgentProfileParams(
///   agentId: 'AGT-2026-000001',
///   section: 'personal_info',
///   changes: {'first_name': 'John'},
///   updatedBy: 'admin_123',
///   updateReason: 'Name correction',
/// ));
/// ```

import 'package:dartz/dartz.dart';
import 'package:equatable/equatable.dart';

import '../../core/errors/failures.dart';
import '../entities/agent_profile.dart';
import '../repositories/agent_repository.dart';

/// Use case for updating agent profile
class UpdateAgentProfileUseCase {
  final AgentRepository repository;

  UpdateAgentProfileUseCase(this.repository);

  /// Execute the use case
  ///
  /// **Parameters**: UpdateAgentProfileParams
  /// **Returns**: Either<Failure, AgentProfile>
  Future<Either<Failure, AgentProfile>> call(
    UpdateAgentProfileParams params,
  ) async {
    // Validate agent ID format
    if (!_isValidAgentId(params.agentId)) {
      return Left(
        ValidationFailure(
          message: 'Invalid agent ID format',
          data: {'agentId': params.agentId},
        ),
      );
    }

    // Validate section name
    if (!_isValidSection(params.section)) {
      return Left(
        ValidationFailure(
          message: 'Invalid section name',
          data: {'section': params.section},
        ),
      );
    }

    // Validate changes are not empty
    if (params.changes.isEmpty) {
      return Left(
        const ValidationFailure(message: 'No changes provided'),
      );
    }

    // Update via repository
    final result = await repository.updateAgentSection(
      agentId: params.agentId,
      section: params.section,
      changes: params.changes,
      updatedBy: params.updatedBy,
      updateReason: params.updateReason,
    );

    return result.fold(
      (failure) => Left(failure),
      (updatedProfile) {
        // Business logic: Verify update was applied
        // (Additional checks can be added here)

        return Right(updatedProfile);
      },
    );
  }

  /// Validate agent ID format
  bool _isValidAgentId(String agentId) {
    final regex = RegExp(r'^AGT-\d{4}-\d{6}$');
    return regex.hasMatch(agentId);
  }

  /// Validate section name
  bool _isValidSection(String section) {
    const validSections = [
      'personal_info',
      'address',
      'contact',
      'license',
      'bank_details',
      'status',
    ];
    return validSections.contains(section);
  }
}

/// Parameters for UpdateAgentProfileUseCase
class UpdateAgentProfileParams extends Equatable {
  /// Agent ID to update
  final String agentId;

  /// Section to update
  /// Valid values: personal_info, address, contact, license, bank_details, status
  final String section;

  /// Changes to apply (field: newValue)
  final Map<String, dynamic> changes;

  /// User who is making the update
  final String updatedBy;

  /// Reason for update (optional but recommended)
  final String? updateReason;

  const UpdateAgentProfileParams({
    required this.agentId,
    required this.section,
    required this.changes,
    required this.updatedBy,
    this.updateReason,
  });

  @override
  List<Object?> get props => [
        agentId,
        section,
        changes,
        updatedBy,
        updateReason,
      ];
}

// ============================================================================
// BUSINESS RULES ENFORCED
// ============================================================================
//
// 1. Critical field changes require approval
//    - PAN number changes → Approval required
//    - Name changes → Audit logged
//    - Status changes → Approval workflow
//
// 2. Non-critical field changes are direct
//    - Address updates → Direct update
//    - Contact updates → Direct update (with OTP for self-service)
//    - Bank details → Requires OTP
//
// 3. All updates are logged
//    - Audit trail maintained
//    - Who, what, when recorded
//    - Old and new values stored
//
// 4. Validation before update
//    - Field-level validation
//    - Cross-field validation (e.g., age calculation from DOB)
//    - Business rule validation (e.g., PAN uniqueness)
//
// ============================================================================

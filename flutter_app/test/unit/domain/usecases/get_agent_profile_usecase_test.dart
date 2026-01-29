import 'package:dartz/dartz.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:pli_agent_management/core/errors/failures.dart';
import 'package:pli_agent_management/domain/entities/agent_profile.dart';
import 'package:pli_agent_management/domain/repositories/agent_repository.dart';
import 'package:pli_agent_management/domain/usecases/get_agent_profile_usecase.dart';

// Mock repository
class MockAgentRepository extends Mock implements AgentRepository {}

void main() {
  late GetAgentProfileUseCase useCase;
  late MockAgentRepository mockRepository;

  setUp(() {
    mockRepository = MockAgentRepository();
    useCase = GetAgentProfileUseCase(mockRepository);
  });

  const tAgentId = 'AGT-2026-000001';
  final tAgentProfile = AgentProfile(
    agentId: tAgentId,
    profileId: 'PROF-001',
    agentType: AgentType.advisor,
    fullName: 'John Doe',
    status: AgentStatus.active,
  );

  group('GetAgentProfileUseCase', () {
    test('should return AgentProfile when repository call succeeds', () async {
      // Arrange
      when(() => mockRepository.getAgentProfile(any()))
          .thenAnswer((_) async => Right(tAgentProfile));

      // Act
      final result = await useCase(tAgentId);

      // Assert
      expect(result, Right(tAgentProfile));
      verify(() => mockRepository.getAgentProfile(tAgentId)).called(1);
      verifyNoMoreInteractions(mockRepository);
    });

    test('should return ValidationFailure when agent ID is invalid', () async {
      // Arrange
      const invalidAgentId = 'INVALID-ID';

      // Act
      final result = await useCase(invalidAgentId);

      // Assert
      expect(result.isLeft(), true);
      result.fold(
        (failure) {
          expect(failure, isA<ValidationFailure>());
          expect(failure.message, contains('Invalid agent ID format'));
        },
        (_) => fail('Should return failure'),
      );
      verifyNever(() => mockRepository.getAgentProfile(any()));
    });

    test('should return ServerFailure when repository call fails', () async {
      // Arrange
      const tServerFailure = ServerFailure(message: 'Server error');
      when(() => mockRepository.getAgentProfile(any()))
          .thenAnswer((_) async => const Left(tServerFailure));

      // Act
      final result = await useCase(tAgentId);

      // Assert
      expect(result, const Left(tServerFailure));
      verify(() => mockRepository.getAgentProfile(tAgentId)).called(1);
    });

    test('should return NetworkFailure when no internet connection', () async {
      // Arrange
      const tNetworkFailure = NetworkFailure(message: 'No internet connection');
      when(() => mockRepository.getAgentProfile(any()))
          .thenAnswer((_) async => const Left(tNetworkFailure));

      // Act
      final result = await useCase(tAgentId);

      // Assert
      expect(result, const Left(tNetworkFailure));
      verify(() => mockRepository.getAgentProfile(tAgentId)).called(1);
    });

    test('should return NotFoundFailure when agent does not exist', () async {
      // Arrange
      const tNotFoundFailure = NotFoundFailure(message: 'Agent not found');
      when(() => mockRepository.getAgentProfile(any()))
          .thenAnswer((_) async => const Left(tNotFoundFailure));

      // Act
      final result = await useCase(tAgentId);

      // Assert
      expect(result, const Left(tNotFoundFailure));
      verify(() => mockRepository.getAgentProfile(tAgentId)).called(1);
    });

    test('should validate agent ID format correctly', () async {
      // Valid formats
      expect(await _callUseCaseAndCheckValidation('AGT-2026-000001'), true);
      expect(await _callUseCaseAndCheckValidation('AGT-2025-999999'), true);

      // Invalid formats
      expect(await _callUseCaseAndCheckValidation(''), false);
      expect(await _callUseCaseAndCheckValidation('AGT-2026'), false);
      expect(await _callUseCaseAndCheckValidation('2026-000001'), false);
      expect(await _callUseCaseAndCheckValidation('AGT-26-000001'), false);
    });
  });

  // Helper function to check validation
  Future<bool> _callUseCaseAndCheckValidation(String agentId) async {
    when(() => mockRepository.getAgentProfile(any()))
        .thenAnswer((_) async => Right(tAgentProfile));

    final result = await useCase(agentId);

    return result.fold(
      (failure) => failure is! ValidationFailure,
      (_) => true,
    );
  }
}

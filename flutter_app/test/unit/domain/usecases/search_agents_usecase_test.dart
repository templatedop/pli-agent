import 'package:dartz/dartz.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:pli_agent_management/core/errors/failures.dart';
import 'package:pli_agent_management/domain/entities/agent_profile.dart';
import 'package:pli_agent_management/domain/repositories/agent_repository.dart';
import 'package:pli_agent_management/domain/usecases/search_agents_usecase.dart';

// Mock repository
class MockAgentRepository extends Mock implements AgentRepository {}

void main() {
  late SearchAgentsUseCase useCase;
  late MockAgentRepository mockRepository;

  setUp(() {
    mockRepository = MockAgentRepository();
    useCase = SearchAgentsUseCase(mockRepository);
  });

  final tAgentsList = [
    const AgentProfile(
      agentId: 'AGT-2026-000001',
      profileId: 'PROF-001',
      agentType: AgentType.advisor,
      fullName: 'John Doe',
      status: AgentStatus.active,
    ),
    const AgentProfile(
      agentId: 'AGT-2026-000002',
      profileId: 'PROF-002',
      agentType: AgentType.advisorCoordinator,
      fullName: 'Jane Smith',
      status: AgentStatus.active,
    ),
  ];

  group('SearchAgentsUseCase', () {
    test('should return list of agents when repository call succeeds', () async {
      // Arrange
      const tParams = SearchAgentsParams(
        criteria: {'status': 'active'},
        page: 1,
        limit: 20,
      );
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => Right(tAgentsList));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, Right(tAgentsList));
      verify(() => mockRepository.searchAgents(
            criteria: tParams.criteria,
            page: tParams.page,
            limit: tParams.limit,
          )).called(1);
    });

    test('should return empty list when no agents match criteria', () async {
      // Arrange
      const tParams = SearchAgentsParams(
        criteria: {'status': 'inactive'},
        page: 1,
        limit: 20,
      );
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => const Right([]));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, const Right([]));
      result.fold(
        (_) => fail('Should return empty list'),
        (agents) => expect(agents.isEmpty, true),
      );
    });

    test('should handle search with no criteria', () async {
      // Arrange
      const tParams = SearchAgentsParams(page: 1, limit: 20);
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => Right(tAgentsList));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, Right(tAgentsList));
      verify(() => mockRepository.searchAgents(
            criteria: null,
            page: 1,
            limit: 20,
          )).called(1);
    });

    test('should handle pagination correctly', () async {
      // Arrange
      const tParams = SearchAgentsParams(
        criteria: {'status': 'active'},
        page: 2,
        limit: 10,
      );
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => Right(tAgentsList));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, Right(tAgentsList));
      verify(() => mockRepository.searchAgents(
            criteria: tParams.criteria,
            page: 2,
            limit: 10,
          )).called(1);
    });

    test('should return ServerFailure when repository call fails', () async {
      // Arrange
      const tParams = SearchAgentsParams(page: 1, limit: 20);
      const tServerFailure = ServerFailure(message: 'Server error');
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => const Left(tServerFailure));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, const Left(tServerFailure));
    });

    test('should return NetworkFailure when no internet connection', () async {
      // Arrange
      const tParams = SearchAgentsParams(page: 1, limit: 20);
      const tNetworkFailure = NetworkFailure(message: 'No internet connection');
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => const Left(tNetworkFailure));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, const Left(tNetworkFailure));
    });

    test('should search with multiple criteria', () async {
      // Arrange
      const tParams = SearchAgentsParams(
        criteria: {
          'status': 'active',
          'agent_type': 'advisor',
          'full_name': 'John',
        },
        page: 1,
        limit: 20,
      );
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => Right(tAgentsList));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, Right(tAgentsList));
      verify(() => mockRepository.searchAgents(
            criteria: tParams.criteria,
            page: 1,
            limit: 20,
          )).called(1);
    });

    test('should use default values for page and limit if not provided', () async {
      // Arrange
      const tParams = SearchAgentsParams();
      when(() => mockRepository.searchAgents(
            criteria: any(named: 'criteria'),
            page: any(named: 'page'),
            limit: any(named: 'limit'),
          )).thenAnswer((_) async => Right(tAgentsList));

      // Act
      final result = await useCase(tParams);

      // Assert
      expect(result, Right(tAgentsList));
      verify(() => mockRepository.searchAgents(
            criteria: null,
            page: 1,
            limit: 20,
          )).called(1);
    });
  });
}

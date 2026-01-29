import 'package:bloc_test/bloc_test.dart';
import 'package:dartz/dartz.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:pli_agent_management/core/errors/failures.dart';
import 'package:pli_agent_management/domain/entities/agent_profile.dart';
import 'package:pli_agent_management/domain/usecases/create_agent_profile_usecase.dart';
import 'package:pli_agent_management/domain/usecases/get_agent_profile_usecase.dart';
import 'package:pli_agent_management/domain/usecases/search_agents_usecase.dart';
import 'package:pli_agent_management/domain/usecases/update_agent_profile_usecase.dart';
import 'package:pli_agent_management/presentation/bloc/agent_profile/agent_profile_barrel.dart';

// Mock use cases
class MockGetAgentProfileUseCase extends Mock implements GetAgentProfileUseCase {}
class MockSearchAgentsUseCase extends Mock implements SearchAgentsUseCase {}
class MockCreateAgentProfileUseCase extends Mock implements CreateAgentProfileUseCase {}
class MockUpdateAgentProfileUseCase extends Mock implements UpdateAgentProfileUseCase {}

// Register fallback values for mocktail
class FakeSearchAgentsParams extends Fake implements SearchAgentsParams {}

void main() {
  late AgentProfileBloc bloc;
  late MockGetAgentProfileUseCase mockGetAgentProfileUseCase;
  late MockSearchAgentsUseCase mockSearchAgentsUseCase;
  late MockCreateAgentProfileUseCase mockCreateAgentProfileUseCase;
  late MockUpdateAgentProfileUseCase mockUpdateAgentProfileUseCase;

  setUp(() {
    mockGetAgentProfileUseCase = MockGetAgentProfileUseCase();
    mockSearchAgentsUseCase = MockSearchAgentsUseCase();
    mockCreateAgentProfileUseCase = MockCreateAgentProfileUseCase();
    mockUpdateAgentProfileUseCase = MockUpdateAgentProfileUseCase();

    bloc = AgentProfileBloc(
      getAgentProfileUseCase: mockGetAgentProfileUseCase,
      searchAgentsUseCase: mockSearchAgentsUseCase,
      createAgentProfileUseCase: mockCreateAgentProfileUseCase,
      updateAgentProfileUseCase: mockUpdateAgentProfileUseCase,
    );

    // Register fallback values
    registerFallbackValue(FakeSearchAgentsParams());
  });

  tearDown(() {
    bloc.close();
  });

  const tAgentId = 'AGT-2026-000001';
  const tAgentProfile = AgentProfile(
    agentId: tAgentId,
    profileId: 'PROF-001',
    agentType: AgentType.advisor,
    fullName: 'John Doe',
    status: AgentStatus.active,
  );

  final tAgentsList = [tAgentProfile];

  group('GetAgentProfileEvent', () {
    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Loaded] when GetAgentProfileEvent succeeds',
      build: () {
        when(() => mockGetAgentProfileUseCase(any()))
            .thenAnswer((_) async => const Right(tAgentProfile));
        return bloc;
      },
      act: (bloc) => bloc.add(const GetAgentProfileEvent(tAgentId)),
      expect: () => [
        const AgentProfileLoading(),
        const AgentProfileLoaded(tAgentProfile),
      ],
      verify: (_) {
        verify(() => mockGetAgentProfileUseCase(tAgentId)).called(1);
      },
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Error] when GetAgentProfileEvent fails',
      build: () {
        when(() => mockGetAgentProfileUseCase(any()))
            .thenAnswer((_) async => const Left(ServerFailure(message: 'Server error')));
        return bloc;
      },
      act: (bloc) => bloc.add(const GetAgentProfileEvent(tAgentId)),
      expect: () => [
        const AgentProfileLoading(),
        const AgentProfileError(
          message: 'Server error',
          errorCode: 'SERVER_ERROR',
        ),
      ],
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, NetworkError] when network fails',
      build: () {
        when(() => mockGetAgentProfileUseCase(any()))
            .thenAnswer((_) async => const Left(NetworkFailure(message: 'No connection')));
        return bloc;
      },
      act: (bloc) => bloc.add(const GetAgentProfileEvent(tAgentId)),
      expect: () => [
        const AgentProfileLoading(),
        const AgentProfileNetworkError(message: 'No connection'),
      ],
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Error] with NOT_FOUND when agent not found',
      build: () {
        when(() => mockGetAgentProfileUseCase(any()))
            .thenAnswer((_) async => const Left(NotFoundFailure(message: 'Agent not found')));
        return bloc;
      },
      act: (bloc) => bloc.add(const GetAgentProfileEvent(tAgentId)),
      expect: () => [
        const AgentProfileLoading(),
        const AgentProfileError(
          message: 'Agent not found',
          errorCode: 'NOT_FOUND',
        ),
      ],
    );
  });

  group('SearchAgentsEvent', () {
    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Loaded] when SearchAgentsEvent succeeds',
      build: () {
        when(() => mockSearchAgentsUseCase(any()))
            .thenAnswer((_) async => Right(tAgentsList));
        return bloc;
      },
      act: (bloc) => bloc.add(const SearchAgentsEvent()),
      expect: () => [
        const AgentSearchLoading(),
        AgentSearchLoaded(
          agents: tAgentsList,
          currentPage: 1,
          totalPages: 1,
          totalCount: 1,
        ),
      ],
      verify: (_) {
        verify(() => mockSearchAgentsUseCase(any())).called(1);
      },
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Empty] when no agents found',
      build: () {
        when(() => mockSearchAgentsUseCase(any()))
            .thenAnswer((_) async => const Right([]));
        return bloc;
      },
      act: (bloc) => bloc.add(const SearchAgentsEvent()),
      expect: () => [
        const AgentSearchLoading(),
        const AgentSearchEmpty(),
      ],
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Loading, Error] when search fails',
      build: () {
        when(() => mockSearchAgentsUseCase(any()))
            .thenAnswer((_) async => const Left(ServerFailure(message: 'Search failed')));
        return bloc;
      },
      act: (bloc) => bloc.add(const SearchAgentsEvent()),
      expect: () => [
        const AgentSearchLoading(),
        const AgentProfileError(
          message: 'Search failed',
          errorCode: 'SERVER_ERROR',
        ),
      ],
    );

    blocTest<AgentProfileBloc, AgentProfileState>(
      'passes search criteria correctly',
      build: () {
        when(() => mockSearchAgentsUseCase(any()))
            .thenAnswer((_) async => Right(tAgentsList));
        return bloc;
      },
      act: (bloc) => bloc.add(const SearchAgentsEvent(
        criteria: {'status': 'active'},
        page: 2,
        limit: 10,
      )),
      verify: (_) {
        verify(() => mockSearchAgentsUseCase(
          const SearchAgentsParams(
            criteria: {'status': 'active'},
            page: 2,
            limit: 10,
          ),
        )).called(1);
      },
    );
  });

  group('ClearSearchResultsEvent', () {
    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Initial] when ClearSearchResultsEvent is added',
      build: () => bloc,
      act: (bloc) => bloc.add(const ClearSearchResultsEvent()),
      expect: () => [const AgentProfileInitial()],
    );
  });

  group('ResetAgentProfileEvent', () {
    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Initial] when ResetAgentProfileEvent is added',
      build: () => bloc,
      act: (bloc) => bloc.add(const ResetAgentProfileEvent()),
      expect: () => [const AgentProfileInitial()],
    );
  });

  group('InitiateProfileCreationEvent', () {
    blocTest<AgentProfileBloc, AgentProfileState>(
      'emits [Creating, SessionInitiated] when initiation succeeds',
      build: () => bloc,
      act: (bloc) => bloc.add(const InitiateProfileCreationEvent(
        agentType: AgentType.advisor,
        initiatedBy: 'USER-001',
      )),
      expect: () => [
        const AgentProfileCreating(message: 'Initiating profile creation session...'),
        isA<ProfileCreationSessionInitiated>(),
      ],
    );
  });

  group('Error Mapping', () {
    test('ValidationFailure maps to AgentProfileValidationError', () async {
      // Arrange
      when(() => mockGetAgentProfileUseCase(any()))
          .thenAnswer((_) async => const Left(ValidationFailure(
                message: 'Validation failed',
                data: {'field': 'error'},
              )));

      // Act
      bloc.add(const GetAgentProfileEvent(tAgentId));

      // Assert
      await expectLater(
        bloc.stream,
        emitsInOrder([
          const AgentProfileLoading(),
          const AgentProfileValidationError(
            message: 'Validation failed',
            validationErrors: {'field': 'error'},
          ),
        ]),
      );
    });

    test('UnauthorizedFailure maps to AgentProfileError with UNAUTHORIZED', () async {
      // Arrange
      when(() => mockGetAgentProfileUseCase(any()))
          .thenAnswer((_) async => const Left(UnauthorizedFailure(message: 'Unauthorized')));

      // Act
      bloc.add(const GetAgentProfileEvent(tAgentId));

      // Assert
      await expectLater(
        bloc.stream,
        emitsInOrder([
          const AgentProfileLoading(),
          const AgentProfileError(
            message: 'Unauthorized',
            errorCode: 'UNAUTHORIZED',
          ),
        ]),
      );
    });

    test('ConflictFailure maps to AgentProfileError with CONFLICT', () async {
      // Arrange
      when(() => mockGetAgentProfileUseCase(any()))
          .thenAnswer((_) async => const Left(ConflictFailure(
                message: 'Conflict',
                data: {'reason': 'duplicate'},
              )));

      // Act
      bloc.add(const GetAgentProfileEvent(tAgentId));

      // Assert
      await expectLater(
        bloc.stream,
        emitsInOrder([
          const AgentProfileLoading(),
          const AgentProfileError(
            message: 'Conflict',
            errorCode: 'CONFLICT',
            errorData: {'reason': 'duplicate'},
          ),
        ]),
      );
    });
  });
}

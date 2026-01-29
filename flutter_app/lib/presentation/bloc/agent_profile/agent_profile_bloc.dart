/// Agent Profile BLoC
///
/// Manages agent profile state and handles business logic events.
/// Connects UI with domain layer use cases.
///
/// **Usage**:
/// ```dart
/// BlocProvider(
///   create: (_) => sl<AgentProfileBloc>(),
///   child: AgentProfileScreen(),
/// )
/// ```

import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../core/errors/failures.dart';
import '../../../domain/usecases/create_agent_profile_usecase.dart';
import '../../../domain/usecases/get_agent_profile_usecase.dart';
import '../../../domain/usecases/search_agents_usecase.dart';
import '../../../domain/usecases/update_agent_profile_usecase.dart';
import 'agent_profile_event.dart';
import 'agent_profile_state.dart';

/// BLoC for managing agent profile operations
class AgentProfileBloc extends Bloc<AgentProfileEvent, AgentProfileState> {
  // Use cases
  final GetAgentProfileUseCase getAgentProfileUseCase;
  final SearchAgentsUseCase searchAgentsUseCase;
  final CreateAgentProfileUseCase createAgentProfileUseCase;
  final UpdateAgentProfileUseCase updateAgentProfileUseCase;

  AgentProfileBloc({
    required this.getAgentProfileUseCase,
    required this.searchAgentsUseCase,
    required this.createAgentProfileUseCase,
    required this.updateAgentProfileUseCase,
  }) : super(const AgentProfileInitial()) {
    // Register event handlers
    on<GetAgentProfileEvent>(_onGetAgentProfile);
    on<SearchAgentsEvent>(_onSearchAgents);
    on<ClearSearchResultsEvent>(_onClearSearchResults);
    on<InitiateProfileCreationEvent>(_onInitiateProfileCreation);
    on<CreateAgentProfileEvent>(_onCreateAgentProfile);
    on<UpdateAgentProfileEvent>(_onUpdateAgentProfile);
    on<ResetAgentProfileEvent>(_onResetAgentProfile);
  }

  // ===========================================================================
  // GET AGENT PROFILE
  // ===========================================================================

  Future<void> _onGetAgentProfile(
    GetAgentProfileEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileLoading());

    final result = await getAgentProfileUseCase(event.agentId);

    result.fold(
      (failure) => emit(_mapFailureToState(failure)),
      (agentProfile) => emit(AgentProfileLoaded(agentProfile)),
    );
  }

  // ===========================================================================
  // SEARCH AGENTS
  // ===========================================================================

  Future<void> _onSearchAgents(
    SearchAgentsEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentSearchLoading());

    final result = await searchAgentsUseCase(
      SearchAgentsParams(
        criteria: event.criteria,
        page: event.page,
        limit: event.limit,
      ),
    );

    result.fold(
      (failure) => emit(_mapFailureToState(failure)),
      (agents) {
        if (agents.isEmpty) {
          emit(const AgentSearchEmpty());
        } else {
          emit(AgentSearchLoaded(
            agents: agents,
            currentPage: event.page,
            totalPages: 1, // TODO: Get from API response
            totalCount: agents.length,
          ));
        }
      },
    );
  }

  Future<void> _onClearSearchResults(
    ClearSearchResultsEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileInitial());
  }

  // ===========================================================================
  // INITIATE PROFILE CREATION
  // ===========================================================================

  Future<void> _onInitiateProfileCreation(
    InitiateProfileCreationEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileCreating(
      message: 'Initiating profile creation session...',
    ));

    // For now, just emit a session initiated state
    // This will be enhanced when we have the full creation flow
    emit(ProfileCreationSessionInitiated(
      sessionId: 'temp-session-${DateTime.now().millisecondsSinceEpoch}',
      message: 'Profile creation session initiated. Ready to proceed.',
    ));
  }

  // ===========================================================================
  // CREATE AGENT PROFILE
  // ===========================================================================

  Future<void> _onCreateAgentProfile(
    CreateAgentProfileEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileCreating(
      message: 'Creating agent profile...',
    ));

    final result = await createAgentProfileUseCase(event.params);

    result.fold(
      (failure) => emit(_mapFailureToState(failure)),
      (agentProfile) => emit(AgentProfileCreated(
        agentProfile: agentProfile,
        message: 'Agent profile created successfully!',
      )),
    );
  }

  // ===========================================================================
  // UPDATE AGENT PROFILE
  // ===========================================================================

  Future<void> _onUpdateAgentProfile(
    UpdateAgentProfileEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileUpdating(
      message: 'Updating agent profile...',
    ));

    final result = await updateAgentProfileUseCase(event.params);

    result.fold(
      (failure) => emit(_mapFailureToState(failure)),
      (agentProfile) => emit(AgentProfileUpdated(
        agentProfile: agentProfile,
        message: 'Agent profile updated successfully!',
      )),
    );
  }

  // ===========================================================================
  // RESET
  // ===========================================================================

  Future<void> _onResetAgentProfile(
    ResetAgentProfileEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileInitial());
  }

  // ===========================================================================
  // HELPER METHODS
  // ===========================================================================

  /// Map failure to appropriate error state
  AgentProfileState _mapFailureToState(Failure failure) {
    if (failure is NetworkFailure) {
      return AgentProfileNetworkError(message: failure.message);
    } else if (failure is ValidationFailure) {
      return AgentProfileValidationError(
        message: failure.message,
        validationErrors: failure.data,
      );
    } else if (failure is NotFoundFailure) {
      return AgentProfileError(
        message: failure.message,
        errorCode: 'NOT_FOUND',
      );
    } else if (failure is UnauthorizedFailure) {
      return AgentProfileError(
        message: failure.message,
        errorCode: 'UNAUTHORIZED',
      );
    } else if (failure is ConflictFailure) {
      return AgentProfileError(
        message: failure.message,
        errorCode: 'CONFLICT',
        errorData: failure.data,
      );
    } else if (failure is ServerFailure) {
      return AgentProfileError(
        message: failure.message,
        errorCode: 'SERVER_ERROR',
      );
    } else {
      return const AgentProfileError(
        message: 'An unexpected error occurred',
        errorCode: 'UNKNOWN',
      );
    }
  }
}

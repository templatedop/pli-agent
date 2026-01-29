/// Agent Profile Events
///
/// All events that can trigger state changes in AgentProfileBloc

import 'package:equatable/equatable.dart';
import '../../../domain/entities/agent_profile.dart';
import '../../../domain/usecases/create_agent_profile_usecase.dart';
import '../../../domain/usecases/update_agent_profile_usecase.dart';

abstract class AgentProfileEvent extends Equatable {
  const AgentProfileEvent();

  @override
  List<Object?> get props => [];
}

// =============================================================================
// AGENT PROFILE RETRIEVAL
// =============================================================================

/// Event to get agent profile by ID
class GetAgentProfileEvent extends AgentProfileEvent {
  final String agentId;

  const GetAgentProfileEvent(this.agentId);

  @override
  List<Object?> get props => [agentId];
}

// =============================================================================
// AGENT PROFILE SEARCH
// =============================================================================

/// Event to search agents with criteria
class SearchAgentsEvent extends AgentProfileEvent {
  final Map<String, dynamic>? criteria;
  final int page;
  final int limit;

  const SearchAgentsEvent({
    this.criteria,
    this.page = 1,
    this.limit = 20,
  });

  @override
  List<Object?> get props => [criteria, page, limit];
}

/// Event to clear search results
class ClearSearchResultsEvent extends AgentProfileEvent {
  const ClearSearchResultsEvent();
}

// =============================================================================
// AGENT PROFILE CREATION
// =============================================================================

/// Event to initiate profile creation
class InitiateProfileCreationEvent extends AgentProfileEvent {
  final AgentType agentType;
  final String initiatedBy;

  const InitiateProfileCreationEvent({
    required this.agentType,
    required this.initiatedBy,
  });

  @override
  List<Object?> get props => [agentType, initiatedBy];
}

/// Event to create agent profile
class CreateAgentProfileEvent extends AgentProfileEvent {
  final CreateAgentProfileParams params;

  const CreateAgentProfileEvent(this.params);

  @override
  List<Object?> get props => [params];
}

// =============================================================================
// AGENT PROFILE UPDATE
// =============================================================================

/// Event to update agent profile section
class UpdateAgentProfileEvent extends AgentProfileEvent {
  final UpdateAgentProfileParams params;

  const UpdateAgentProfileEvent(this.params);

  @override
  List<Object?> get props => [params];
}

// =============================================================================
// STATE RESET
// =============================================================================

/// Event to reset BLoC to initial state
class ResetAgentProfileEvent extends AgentProfileEvent {
  const ResetAgentProfileEvent();
}

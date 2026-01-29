/// Agent Profile States
///
/// All possible states for AgentProfileBloc

import 'package:equatable/equatable.dart';
import '../../../domain/entities/agent_profile.dart';

abstract class AgentProfileState extends Equatable {
  const AgentProfileState();

  @override
  List<Object?> get props => [];
}

// =============================================================================
// INITIAL STATE
// =============================================================================

/// Initial state when BLoC is created
class AgentProfileInitial extends AgentProfileState {
  const AgentProfileInitial();
}

// =============================================================================
// LOADING STATES
// =============================================================================

/// Loading agent profile
class AgentProfileLoading extends AgentProfileState {
  const AgentProfileLoading();
}

/// Loading search results
class AgentSearchLoading extends AgentProfileState {
  const AgentSearchLoading();
}

/// Creating agent profile
class AgentProfileCreating extends AgentProfileState {
  final String? message;

  const AgentProfileCreating({this.message});

  @override
  List<Object?> get props => [message];
}

/// Updating agent profile
class AgentProfileUpdating extends AgentProfileState {
  final String? message;

  const AgentProfileUpdating({this.message});

  @override
  List<Object?> get props => [message];
}

// =============================================================================
// SUCCESS STATES
// =============================================================================

/// Agent profile loaded successfully
class AgentProfileLoaded extends AgentProfileState {
  final AgentProfile agentProfile;

  const AgentProfileLoaded(this.agentProfile);

  @override
  List<Object?> get props => [agentProfile];
}

/// Agent search results loaded
class AgentSearchLoaded extends AgentProfileState {
  final List<AgentProfile> agents;
  final int currentPage;
  final int totalPages;
  final int totalCount;

  const AgentSearchLoaded({
    required this.agents,
    required this.currentPage,
    this.totalPages = 1,
    this.totalCount = 0,
  });

  @override
  List<Object?> get props => [agents, currentPage, totalPages, totalCount];

  /// Check if there are more pages to load
  bool get hasMorePages => currentPage < totalPages;
}

/// Profile creation session initiated
class ProfileCreationSessionInitiated extends AgentProfileState {
  final String sessionId;
  final String message;

  const ProfileCreationSessionInitiated({
    required this.sessionId,
    required this.message,
  });

  @override
  List<Object?> get props => [sessionId, message];
}

/// Agent profile created successfully
class AgentProfileCreated extends AgentProfileState {
  final AgentProfile agentProfile;
  final String message;

  const AgentProfileCreated({
    required this.agentProfile,
    required this.message,
  });

  @override
  List<Object?> get props => [agentProfile, message];
}

/// Agent profile updated successfully
class AgentProfileUpdated extends AgentProfileState {
  final AgentProfile agentProfile;
  final String message;

  const AgentProfileUpdated({
    required this.agentProfile,
    required this.message,
  });

  @override
  List<Object?> get props => [agentProfile, message];
}

// =============================================================================
// ERROR STATES
// =============================================================================

/// Error occurred
class AgentProfileError extends AgentProfileState {
  final String message;
  final String? errorCode;
  final Map<String, dynamic>? errorData;

  const AgentProfileError({
    required this.message,
    this.errorCode,
    this.errorData,
  });

  @override
  List<Object?> get props => [message, errorCode, errorData];
}

/// Validation error
class AgentProfileValidationError extends AgentProfileState {
  final String message;
  final Map<String, dynamic>? validationErrors;

  const AgentProfileValidationError({
    required this.message,
    this.validationErrors,
  });

  @override
  List<Object?> get props => [message, validationErrors];
}

/// Network error
class AgentProfileNetworkError extends AgentProfileState {
  final String message;

  const AgentProfileNetworkError({
    this.message = 'No internet connection',
  });

  @override
  List<Object?> get props => [message];
}

// =============================================================================
// EMPTY STATES
// =============================================================================

/// No agents found in search
class AgentSearchEmpty extends AgentProfileState {
  final String message;

  const AgentSearchEmpty({
    this.message = 'No agents found',
  });

  @override
  List<Object?> get props => [message];
}

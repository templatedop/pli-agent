import 'package:equatable/equatable.dart';

abstract class Failure extends Equatable {
  final String message;
  final int? statusCode;
  final dynamic data;

  const Failure({
    required this.message,
    this.statusCode,
    this.data,
  });

  @override
  List<Object?> get props => [message, statusCode, data];
}

// Server Failure
class ServerFailure extends Failure {
  const ServerFailure({
    required super.message,
    super.statusCode,
    super.data,
  });
}

// Network Failure
class NetworkFailure extends Failure {
  const NetworkFailure({
    required super.message,
  }) : super(statusCode: null);
}

// Validation Failure
class ValidationFailure extends Failure {
  const ValidationFailure({
    required super.message,
    super.data,
  }) : super(statusCode: 400);
}

// Cache Failure
class CacheFailure extends Failure {
  const CacheFailure({
    required super.message,
  }) : super(statusCode: null);
}

// Unauthorized Failure
class UnauthorizedFailure extends Failure {
  const UnauthorizedFailure({
    required super.message,
  }) : super(statusCode: 401);
}

// Not Found Failure
class NotFoundFailure extends Failure {
  const NotFoundFailure({
    required super.message,
  }) : super(statusCode: 404);
}

// Timeout Failure
class TimeoutFailure extends Failure {
  const TimeoutFailure({
    required super.message,
  }) : super(statusCode: 408);
}

// Conflict Failure (e.g., duplicate PAN)
class ConflictFailure extends Failure {
  const ConflictFailure({
    required super.message,
    super.data,
  }) : super(statusCode: 409);
}

// Generic Failure
class GenericFailure extends Failure {
  const GenericFailure({
    required super.message,
    super.statusCode,
  });
}

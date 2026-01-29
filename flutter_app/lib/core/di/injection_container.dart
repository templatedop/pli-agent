/// Dependency Injection Container
///
/// This file sets up all dependencies using GetIt service locator.
/// It registers all layers: data sources, repositories, use cases, BLoCs.
///
/// **Usage**:
/// ```dart
/// void main() async {
///   WidgetsFlutterBinding.ensureInitialized();
///
///   // Initialize dependency injection
///   await initializeDependencies();
///
///   runApp(MyApp());
/// }
/// ```
///
/// **Getting Dependencies**:
/// ```dart
/// // In your code
/// final useCase = sl<GetAgentProfileUseCase>();
/// final repository = sl<AgentRepository>();
/// ```

import 'package:dio/dio.dart';
import 'package:get_it/get_it.dart';

import '../../data/datasources/agent_remote_datasource.dart';
import '../../data/repositories/agent_repository_impl.dart';
import '../../domain/repositories/agent_repository.dart';
import '../../domain/usecases/create_agent_profile_usecase.dart';
import '../../domain/usecases/get_agent_profile_usecase.dart';
import '../../domain/usecases/search_agents_usecase.dart';
import '../../domain/usecases/update_agent_profile_usecase.dart';
import '../../presentation/bloc/agent_profile/agent_profile_bloc.dart';
import '../config/api_config.dart';
import '../network/api_client.dart';

/// Service locator instance
/// Use this to get dependencies throughout the app
final sl = GetIt.instance;

/// Initialize all dependencies
///
/// Call this in main() before runApp()
Future<void> initializeDependencies() async {
  // ==========================================================================
  // CORE LAYER
  // ==========================================================================

  // Dio instance for HTTP calls
  sl.registerLazySingleton(() => Dio());

  // API Client (wraps Dio with interceptors and configuration)
  sl.registerLazySingleton(
    () => ApiClient(
      dio: sl(),
      baseUrl: ApiConfig.baseUrl,
    ),
  );

  // ==========================================================================
  // DATA LAYER
  // ==========================================================================

  // Remote Data Sources
  sl.registerLazySingleton(
    () => AgentRemoteDataSource(sl()),
  );

  // Repositories (implementations)
  sl.registerLazySingleton<AgentRepository>(
    () => AgentRepositoryImpl(
      remoteDataSource: sl(),
    ),
  );

  // ==========================================================================
  // DOMAIN LAYER
  // ==========================================================================

  // Use Cases - Agent Profile Creation (UJ-001)
  sl.registerLazySingleton(
    () => CreateAgentProfileUseCase(sl()),
  );

  // Use Cases - Agent Profile Retrieval (UJ-002)
  sl.registerLazySingleton(
    () => GetAgentProfileUseCase(sl()),
  );

  // Use Cases - Agent Search (UJ-002)
  sl.registerLazySingleton(
    () => SearchAgentsUseCase(sl()),
  );

  // Use Cases - Agent Profile Update (UJ-002)
  sl.registerLazySingleton(
    () => UpdateAgentProfileUseCase(sl()),
  );

  // ==========================================================================
  // PRESENTATION LAYER (BLoCs/Cubits)
  // ==========================================================================

  // Agent Profile BLoC (Factory - new instance per screen)
  sl.registerFactory(
    () => AgentProfileBloc(
      getAgentProfileUseCase: sl(),
      searchAgentsUseCase: sl(),
      createAgentProfileUseCase: sl(),
      updateAgentProfileUseCase: sl(),
    ),
  );
}

/// Reset all dependencies (useful for testing)
Future<void> resetDependencies() async {
  await sl.reset();
}

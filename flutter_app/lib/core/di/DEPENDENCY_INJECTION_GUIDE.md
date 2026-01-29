# Dependency Injection Guide - GetIt Service Locator

## 📋 Overview

We use **GetIt** as our service locator for dependency injection. This provides:
- **Single source of truth** for all dependencies
- **Lazy initialization** for efficient memory usage
- **Easy testing** with mock dependencies
- **Inversion of control** following SOLID principles

**Key File**: `lib/core/di/injection_container.dart`

---

## 🎯 What is Dependency Injection?

Dependency Injection (DI) is a design pattern where:
- Objects don't create their dependencies
- Dependencies are provided from outside
- Makes code more testable and maintainable

### Without DI (Bad):
```dart
class AgentProfileScreen {
  // Creates dependencies directly
  final repository = AgentRepositoryImpl(
    remoteDataSource: AgentRemoteDataSource(
      apiClient: ApiClient(dio: Dio()),
    ),
  );
}
```

### With DI (Good):
```dart
class AgentProfileScreen {
  // Dependencies injected
  final GetAgentProfileUseCase getAgentProfileUseCase;

  AgentProfileScreen({required this.getAgentProfileUseCase});
}

// In DI container:
sl.registerLazySingleton(() => GetAgentProfileUseCase(sl()));
```

---

## 📁 Current Setup

### Service Locator Instance

```dart
/// Global service locator
final sl = GetIt.instance;
```

Use `sl` to get any registered dependency:

```dart
// Get a use case
final useCase = sl<GetAgentProfileUseCase>();

// Get a repository
final repository = sl<AgentRepository>();

// Get API client
final apiClient = sl<ApiClient>();
```

### Initialization

In `main.dart`:

```dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Hive for local storage
  await Hive.initFlutter();

  // Initialize dependency injection
  await di.initializeDependencies();

  runApp(const PLIAgentManagementApp());
}
```

---

## 🏗️ Registration Types

### 1. registerLazySingleton

**Use for**: Services that should only have one instance

```dart
sl.registerLazySingleton(() => ApiClient(dio: sl()));
```

- Creates instance on **first access**
- **Reuses same instance** on every call
- Good for: Repositories, data sources, API clients

### 2. registerSingleton

**Use for**: Pre-initialized singletons

```dart
sl.registerSingleton(ApiClient(dio: sl()));
```

- Creates instance **immediately**
- Reuses same instance on every call
- Good for: Configuration objects, pre-built services

### 3. registerFactory

**Use for**: New instance every time

```dart
sl.registerFactory(() => AgentProfileBloc(useCase: sl()));
```

- Creates **new instance** on every call
- Good for: BLoCs, Cubits, temporary objects

---

## 📊 Current Dependencies

### Core Layer

```dart
// Dio for HTTP
sl.registerLazySingleton(() => Dio());

// API Client (wraps Dio)
sl.registerLazySingleton(() => ApiClient(
  dio: sl(),
  baseUrl: ApiConfig.baseUrl,
));
```

### Data Layer

```dart
// Remote Data Sources
sl.registerLazySingleton(() => AgentRemoteDataSource(sl()));

// Repository Implementations
sl.registerLazySingleton<AgentRepository>(
  () => AgentRepositoryImpl(remoteDataSource: sl()),
);
```

**Note**: Repositories are registered with **interface type** (`AgentRepository`), not implementation type. This allows easy swapping of implementations.

### Domain Layer

```dart
// Use Cases
sl.registerLazySingleton(() => CreateAgentProfileUseCase(sl()));
sl.registerLazySingleton(() => GetAgentProfileUseCase(sl()));
sl.registerLazySingleton(() => SearchAgentsUseCase(sl()));
sl.registerLazySingleton(() => UpdateAgentProfileUseCase(sl()));
```

### Presentation Layer

```dart
// BLoCs (will be registered as we create them)
sl.registerFactory(() => AgentProfileBloc(
  getAgentProfileUseCase: sl(),
  createAgentProfileUseCase: sl(),
  updateAgentProfileUseCase: sl(),
));
```

**Note**: BLoCs are registered as **factories** because we want a new instance for each screen.

---

## 🔧 How to Add New Dependencies

### Step 1: Create the Dependency

Example: Adding a new use case

```dart
// lib/domain/usecases/delete_agent_profile_usecase.dart
class DeleteAgentProfileUseCase {
  final AgentRepository repository;

  DeleteAgentProfileUseCase(this.repository);

  Future<Either<Failure, bool>> call(String agentId) async {
    return await repository.deleteAgentProfile(agentId);
  }
}
```

### Step 2: Register in injection_container.dart

```dart
// Add to Domain Layer section
sl.registerLazySingleton(
  () => DeleteAgentProfileUseCase(sl()),
);
```

### Step 3: Use in Code

```dart
// In your BLoC or screen
class AgentProfileBloc extends Bloc<AgentProfileEvent, AgentProfileState> {
  final DeleteAgentProfileUseCase deleteAgentProfileUseCase;

  AgentProfileBloc({
    required this.deleteAgentProfileUseCase,
  }) : super(AgentProfileInitial());
}

// Register BLoC
sl.registerFactory(
  () => AgentProfileBloc(
    deleteAgentProfileUseCase: sl(),  // Automatically resolves
  ),
);
```

---

## 🎨 Using Dependencies

### In Screens (with BLoC)

```dart
class AgentProfileScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AgentProfileBloc>(),  // Get BLoC from DI
      child: AgentProfileView(),
    );
  }
}
```

### In Widgets (Direct Access)

```dart
class AgentSearchWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    // Get use case directly (not recommended for production)
    final searchUseCase = sl<SearchAgentsUseCase>();

    return SearchButton(
      onPressed: () async {
        final result = await searchUseCase(criteria: {...});
        // Handle result
      },
    );
  }
}
```

**Note**: Prefer using BLoC pattern over direct access for better separation.

---

## 🧪 Testing with DI

### Mock Dependencies

```dart
// In your test file
class MockAgentRepository extends Mock implements AgentRepository {}

void main() {
  late MockAgentRepository mockRepository;
  late GetAgentProfileUseCase useCase;

  setUp(() {
    mockRepository = MockAgentRepository();
    useCase = GetAgentProfileUseCase(mockRepository);
  });

  test('should return AgentProfile when repository succeeds', () async {
    // Arrange
    when(() => mockRepository.getAgentProfile(any()))
        .thenAnswer((_) async => Right(tAgentProfile));

    // Act
    final result = await useCase('AGT-001');

    // Assert
    expect(result, Right(tAgentProfile));
  });
}
```

### Reset Dependencies (for tests)

```dart
setUp(() async {
  // Reset and reinitialize for each test
  await di.resetDependencies();
  await di.initializeDependencies();
});
```

---

## 📊 Dependency Graph

Our current dependency graph:

```
Screen/Widget
    ↓
BLoC (Factory)
    ↓
Use Case (LazySingleton)
    ↓
Repository Interface (LazySingleton)
    ↓
Repository Implementation (LazySingleton)
    ↓
Data Source (LazySingleton)
    ↓
API Client (LazySingleton)
    ↓
Dio (LazySingleton)
```

Each layer depends on the layer below it.

---

## 🎯 Best Practices

### 1. Use Interfaces, Not Implementations

```dart
// ✅ Good: Register with interface
sl.registerLazySingleton<AgentRepository>(
  () => AgentRepositoryImpl(remoteDataSource: sl()),
);

// ❌ Bad: Register with implementation
sl.registerLazySingleton(() => AgentRepositoryImpl(...));
```

### 2. Lazy Initialization

```dart
// ✅ Good: Lazy singleton (created on first use)
sl.registerLazySingleton(() => ApiClient(dio: sl()));

// ❌ Bad: Eager singleton (created immediately)
sl.registerSingleton(ApiClient(dio: sl()));
```

### 3. Factory for BLoCs

```dart
// ✅ Good: New BLoC per screen
sl.registerFactory(() => AgentProfileBloc(...));

// ❌ Bad: Shared BLoC across screens
sl.registerLazySingleton(() => AgentProfileBloc(...));
```

### 4. Constructor Injection

```dart
// ✅ Good: Dependencies in constructor
class MyBloc {
  final UseCase useCase;
  MyBloc({required this.useCase});
}

// ❌ Bad: Getting from service locator
class MyBloc {
  final useCase = sl<UseCase>();
}
```

---

## 🔍 Debugging Dependencies

### Check if Registered

```dart
if (sl.isRegistered<AgentRepository>()) {
  print('Repository is registered');
}
```

### Get Registration Info

```dart
// Check all registrations
print(sl.allReady());  // Are all dependencies ready?
```

### Common Errors

**Error**: `GetIt: Object/factory with type X is not registered`

**Solution**: Make sure you registered the dependency in `injection_container.dart`

**Error**: `Circular dependency detected`

**Solution**: Check your dependency chain - object A depends on B, and B depends on A

---

## 🚀 Adding New Layers

### Example: Adding Analytics Layer

```dart
// 1. Create service
class AnalyticsService {
  void logEvent(String event) {
    // Log to analytics
  }
}

// 2. Register in injection_container.dart
sl.registerLazySingleton(() => AnalyticsService());

// 3. Use in use cases or BLoCs
class CreateAgentProfileUseCase {
  final AgentRepository repository;
  final AnalyticsService analytics;

  CreateAgentProfileUseCase(this.repository, this.analytics);

  Future<Either<Failure, AgentProfile>> call(params) async {
    analytics.logEvent('create_agent_profile_started');
    final result = await repository.submitProfile(...);
    analytics.logEvent('create_agent_profile_completed');
    return result;
  }
}

// 4. Update registration
sl.registerLazySingleton(
  () => CreateAgentProfileUseCase(sl(), sl()),  // Both dependencies resolved
);
```

---

## 📚 Further Reading

- [GetIt Package Documentation](https://pub.dev/packages/get_it)
- [Dependency Injection Principles](https://martinfowler.com/articles/injection.html)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Clean Architecture DI](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

---

## 📝 Summary

**Current Setup**:
- ✅ GetIt service locator configured
- ✅ All layers registered (Core, Data, Domain)
- ✅ Use cases ready to use
- ✅ Easy to add new dependencies

**Registration Pattern**:
- `registerLazySingleton`: Repositories, data sources, use cases
- `registerFactory`: BLoCs, cubits
- `registerSingleton`: Pre-built config objects

**Usage Pattern**:
```dart
// Register
sl.registerLazySingleton(() => MyService());

// Use
final service = sl<MyService>();
```

**Next**: Create BLoCs using the registered use cases!

---

**Created**: 2026-01-29
**Phase**: 4 - Dependency Injection
**Progress**: DI Setup Complete ✅

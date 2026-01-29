# Domain Layer Guide - Clean Architecture

## 📋 Overview

The **domain layer** is the innermost layer in Clean Architecture. It contains:
- **Entities**: Core business objects
- **Repositories**: Interfaces (contracts) for data access
- **Use Cases**: Business logic and orchestration

**Key Principle**: The domain layer has **ZERO dependencies** on external frameworks. It's pure Dart code representing business rules.

---

## 📁 Structure

```
lib/domain/
├── entities/
│   └── agent_profile.dart         # Core business entity
│
├── repositories/
│   └── agent_repository.dart      # Repository interface (contract)
│
├── usecases/
│   ├── create_agent_profile_usecase.dart
│   ├── get_agent_profile_usecase.dart
│   ├── search_agents_usecase.dart
│   └── update_agent_profile_usecase.dart
│
└── DOMAIN_LAYER_GUIDE.md          # This file
```

---

## 🎯 Entities

### What are Entities?

Entities are **pure business objects** that represent core domain concepts. They contain:
- **Data fields** (properties)
- **Business logic** (methods)
- **Validation rules**
- **No framework dependencies**

### Example: AgentProfile Entity

```dart
class AgentProfile extends Equatable {
  final String agentId;
  final String fullName;
  final AgentStatus status;
  final AgentType agentType;

  const AgentProfile({
    required this.agentId,
    required this.fullName,
    required this.status,
    required this.agentType,
  });

  // Business logic methods
  bool get isActive => status == AgentStatus.active;
  bool get canPerformTransactions => isActive;
  bool get hasValidLicense => licenses?.any((l) => l.isValid) ?? false;
}
```

### Key Features

✅ **Immutable**: All fields are final
✅ **Equatable**: Automatic equality comparison
✅ **Business Logic**: Methods like `isActive`, `canPerformTransactions`
✅ **No Dependencies**: Pure Dart, no Flutter or external packages
✅ **Value Objects**: Nested objects like `PersonalInfo`, `Address`

### Enums

```dart
enum AgentStatus {
  active,
  inactive,
  suspended,
  terminated,
  deactivated;

  bool get allowsTransactions => this == AgentStatus.active;
}

enum AgentType {
  advisor,
  advisorCoordinator,
  departmentalEmployee,
  fieldOfficer,
  directAgent,
  gds;
}
```

---

## 🔗 Repositories (Interfaces)

### What are Repository Interfaces?

Repositories define **contracts** for data access. They:
- **Don't implement** data fetching (that's in the data layer)
- **Define operations** needed by use cases
- **Return Either<Failure, Success>** for error handling
- **Abstract data sources** (API, database, cache)

### Example: AgentRepository Interface

```dart
abstract class AgentRepository {
  /// Get agent profile by ID
  Future<Either<Failure, AgentProfile>> getAgentProfile(String agentId);

  /// Search agents
  Future<Either<Failure, List<AgentProfile>>> searchAgents({
    Map<String, dynamic>? criteria,
    int page = 1,
    int limit = 20,
  });

  /// Update agent section
  Future<Either<Failure, AgentProfile>> updateAgentSection({
    required String agentId,
    required String section,
    required Map<String, dynamic> changes,
    required String updatedBy,
  });
}
```

### Key Operations (78 APIs mapped)

**Profile Creation (UJ-001)**:
- `initiateProfileCreation()` - Start session
- `fetchHrmsData()` - Get HRMS data
- `linkCoordinator()` - Link to coordinator
- `validateProfile()` - Validate fields
- `submitProfile()` - Create profile

**Profile Management (UJ-002)**:
- `searchAgents()` - Search with criteria
- `getAgentProfile()` - Get by ID
- `updateAgentSection()` - Update section

**Termination (UJ-004)**:
- `terminateAgent()` - Terminate
- `getTerminationLetter()` - Download letter

**Reinstatement (UJ-009)**:
- `createReinstatementRequest()` - Request reinstatement
- `approveReinstatement()` - Approve request

**Lookups**:
- `getAgentTypes()` - Get types dropdown
- `getStatusTypes()` - Get status dropdown
- `getTerminationReasons()` - Get reasons
- And many more...

**Validation**:
- `validatePanUniqueness()` - Check PAN
- `validateEmployeeId()` - Check HRMS
- `validateIfscCode()` - Validate bank
- `validateOfficeCode()` - Check office

---

## ⚙️ Use Cases

### What are Use Cases?

Use cases contain **business logic**. They:
- **Orchestrate** data flow
- **Enforce** business rules
- **Call** repository methods
- **Return** Either<Failure, Success>

### Example: GetAgentProfileUseCase

```dart
class GetAgentProfileUseCase {
  final AgentRepository repository;

  GetAgentProfileUseCase(this.repository);

  Future<Either<Failure, AgentProfile>> call(String agentId) async {
    // Validate input
    if (!_isValidAgentId(agentId)) {
      return Left(ValidationFailure(message: 'Invalid agent ID'));
    }

    // Fetch from repository
    final result = await repository.getAgentProfile(agentId);

    // Apply business rules
    return result.fold(
      (failure) => Left(failure),
      (agentProfile) => Right(agentProfile),
    );
  }
}
```

### Use Case Pattern

All use cases follow this pattern:

```dart
class XyzUseCase {
  final Repository repository;

  XyzUseCase(this.repository);

  Future<Either<Failure, Result>> call(Params params) async {
    // 1. Validate input
    // 2. Call repository
    // 3. Apply business rules
    // 4. Return result
  }
}
```

### Created Use Cases

1. **CreateAgentProfileUseCase** - Complete profile creation flow
   - Initiates session
   - Fetches HRMS data (optional)
   - Links coordinator (for advisors)
   - Validates profile
   - Submits for creation
   - Enforces business rules

2. **GetAgentProfileUseCase** - Retrieve profile
   - Validates agent ID format
   - Fetches profile
   - Returns result

3. **SearchAgentsUseCase** - Search with pagination
   - Validates pagination params
   - Applies search criteria
   - Returns filtered results

4. **UpdateAgentProfileUseCase** - Update profile section
   - Validates agent ID and section
   - Applies changes
   - Handles approval workflow
   - Logs audit trail

---

## 🔄 Either Type (Functional Error Handling)

### Why Either?

The domain layer uses `Either<Failure, Success>` from the **dartz** package for error handling:

- **Left side**: Failure (error case)
- **Right side**: Success value
- **Type-safe**: No exceptions
- **Explicit**: Forces error handling

### Usage Pattern

```dart
final result = await useCase(params);

result.fold(
  (failure) {
    // Handle error
    if (failure is NetworkFailure) {
      // Show network error
    } else if (failure is ValidationFailure) {
      // Show validation error
    }
  },
  (success) {
    // Handle success
    // Use the result
  },
);
```

### Failure Types

```dart
// Network errors
NetworkFailure(message: 'No internet connection')

// Server errors
ServerFailure(message: 'Server error', statusCode: 500)

// Validation errors
ValidationFailure(message: 'Invalid PAN', data: {...})

// Not found errors
NotFoundFailure(message: 'Agent not found')

// Unauthorized errors
UnauthorizedFailure(message: 'Not authorized')

// Conflict errors (e.g., duplicate PAN)
ConflictFailure(message: 'PAN already exists', data: {...})
```

---

## 📊 Business Rules Enforced

### Profile Creation

1. **Profile Completeness**
   - PAN number required
   - Personal info required
   - At least one address
   - At least one contact

2. **Advisor-Coordinator Relationship**
   - Advisors MUST have coordinator
   - Coordinator MUST be active

3. **HRMS Integration**
   - Departmental employees can use HRMS
   - HRMS failure allows manual entry
   - HRMS data can be overridden

4. **Validation**
   - PAN must be unique
   - Age must be 18-70
   - Contact numbers must be valid
   - Email must be valid

### Profile Update

1. **Critical Field Changes**
   - PAN changes require approval
   - Name changes are audited
   - Status changes require approval

2. **Non-Critical Changes**
   - Address updates are direct
   - Contact updates are direct
   - Bank details require OTP

3. **Audit Trail**
   - All changes logged
   - Who, what, when recorded
   - Old and new values stored

---

## 🎨 How to Use

### In Presentation Layer (BLoC)

```dart
class AgentProfileBloc extends Bloc<AgentProfileEvent, AgentProfileState> {
  final GetAgentProfileUseCase getAgentProfileUseCase;

  AgentProfileBloc({required this.getAgentProfileUseCase})
      : super(AgentProfileInitial());

  @override
  Stream<AgentProfileState> mapEventToState(
    AgentProfileEvent event,
  ) async* {
    if (event is GetAgentProfileEvent) {
      yield AgentProfileLoading();

      // Call use case
      final result = await getAgentProfileUseCase(event.agentId);

      // Handle result
      yield result.fold(
        (failure) => AgentProfileError(message: failure.message),
        (agentProfile) => AgentProfileLoaded(agentProfile: agentProfile),
      );
    }
  }
}
```

### In Dependency Injection

```dart
// Register use cases
sl.registerLazySingleton(
  () => GetAgentProfileUseCase(sl()),
);

sl.registerLazySingleton(
  () => CreateAgentProfileUseCase(sl()),
);

sl.registerLazySingleton(
  () => SearchAgentsUseCase(sl()),
);

sl.registerLazySingleton(
  () => UpdateAgentProfileUseCase(sl()),
);

// Register repository (data layer will implement this)
sl.registerLazySingleton<AgentRepository>(
  () => AgentRepositoryImpl(
    remoteDataSource: sl(),
  ),
);
```

---

## ✅ Benefits of This Architecture

### 1. Testability
```dart
// Easy to test use cases with mock repository
test('should return AgentProfile when repository returns success', () async {
  // Arrange
  when(mockRepository.getAgentProfile(any))
      .thenAnswer((_) async => Right(tAgentProfile));

  // Act
  final result = await useCase('AGT-001');

  // Assert
  expect(result, Right(tAgentProfile));
});
```

### 2. Independence
- Domain layer doesn't know about Flutter
- Can be used in Flutter, CLI, web, etc.
- Easy to reuse in different platforms

### 3. Maintainability
- Business rules in one place
- Easy to find and modify
- Clear separation of concerns

### 4. Flexibility
- Change data source without changing domain
- Swap API for local database
- Add caching layer transparently

---

## 🚀 Next Steps

1. **Implement Data Layer** (in progress)
   - Create models that map to entities
   - Implement repository interface
   - Create remote data sources
   - Add data mappers

2. **Set Up Dependency Injection**
   - Register use cases
   - Register repositories
   - Register data sources

3. **Create BLoCs/Cubits**
   - Use use cases in BLoCs
   - Manage state
   - Handle events

4. **Build UI**
   - Consume BLoC state
   - Display entities
   - Handle user input

---

## 📚 Further Reading

- [Clean Architecture by Robert Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Flutter Clean Architecture Guide](https://resocoder.com/flutter-clean-architecture-tdd/)
- [Dartz Package Documentation](https://pub.dev/packages/dartz)
- [Equatable Package Documentation](https://pub.dev/packages/equatable)

---

## 📝 Summary

The domain layer is now complete with:

✅ **1 Entity**: AgentProfile (with nested value objects)
✅ **1 Repository Interface**: AgentRepository (78 API methods)
✅ **4 Use Cases**: Create, Get, Search, Update

**Key Achievements**:
- Zero framework dependencies
- Type-safe error handling with Either
- Business rules enforced in use cases
- Clear contracts via repository interfaces
- Immutable entities with Equatable
- Ready for testing

**Next**: Implement data layer to connect domain to APIs!

---

**Created**: 2026-01-29
**Phase**: 2 - Domain Layer
**Progress**: Domain Layer Complete ✅

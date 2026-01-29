# Data Layer Guide - Clean Architecture

## 📋 Overview

The **data layer** is the outer layer in Clean Architecture. It contains:
- **Models**: Data structures that map to API responses/requests
- **Data Sources**: Classes that make API calls
- **Repository Implementations**: Concrete implementations of domain repository interfaces
- **Mappers**: Convert between models and domain entities

**Key Principle**: The data layer implements the contracts defined in the domain layer and handles all external data operations.

---

## 📁 Structure

```
lib/data/
├── models/
│   ├── agent_model.dart              # Main agent model
│   ├── agent_model.freezed.dart      # Generated (immutability)
│   ├── agent_model.g.dart            # Generated (JSON)
│   ├── requests/                     # Request models (if needed)
│   └── responses/                    # Response models (if needed)
│
├── datasources/
│   └── agent_remote_datasource.dart  # API calls
│
├── repositories/
│   └── agent_repository_impl.dart    # Repository implementation
│
└── DATA_LAYER_GUIDE.md               # This file
```

---

## 🎯 Data Models

### What are Data Models?

Data models are **DTOs (Data Transfer Objects)** that:
- **Map to API structure** (JSON)
- **Use Freezed** for immutability
- **Handle JSON serialization** with json_annotation
- **Convert to domain entities** via toEntity() methods
- **Match API field names** exactly

### Example: AgentModel

```dart
@freezed
class AgentModel with _$AgentModel {
  const AgentModel._();

  const factory AgentModel({
    @JsonKey(name: 'agent_id') required String agentId,
    @JsonKey(name: 'full_name') required String fullName,
    required String status,
    // ... other fields
  }) = _AgentModel;

  factory AgentModel.fromJson(Map<String, dynamic> json) =>
      _$AgentModelFromJson(json);

  /// Convert to Domain Entity
  AgentProfile toEntity() {
    return AgentProfile(
      agentId: agentId,
      fullName: fullName,
      status: _parseAgentStatus(status),
      // ... map all fields
    );
  }
}
```

### Key Features

✅ **Freezed**: Immutable, copyWith, equality
✅ **JSON Serialization**: Automatic with build_runner
✅ **@JsonKey**: Maps Dart names to API names
✅ **toEntity()**: Converts model to domain entity
✅ **Parsing**: Converts API strings to enums

### Nested Models

All nested structures have their own models:

```dart
@freezed
class PersonalInfoModel with _$PersonalInfoModel {
  const factory PersonalInfoModel({
    @JsonKey(name: 'date_of_birth') String? dateOfBirth,
    String? gender,
    // ... other fields
  }) = _PersonalInfoModel;

  factory PersonalInfoModel.fromJson(Map<String, dynamic> json) =>
      _$PersonalInfoModelFromJson(json);

  PersonalInfo toEntity() {
    return PersonalInfo(
      dateOfBirth: dateOfBirth != null ? DateTime.parse(dateOfBirth!) : null,
      gender: gender,
      // ... map fields
    );
  }
}
```

### Why Separate Models from Entities?

1. **API Independence**: API changes don't affect business logic
2. **JSON Handling**: Models handle serialization, entities don't
3. **Transformation**: Can adapt API structure to domain needs
4. **Type Safety**: Parse API strings to proper types
5. **Flexibility**: Multiple APIs can map to same entity

---

## 🌐 Remote Data Sources

### What are Data Sources?

Data sources are classes that:
- **Make HTTP requests** using ApiClient
- **Handle responses** and extract data
- **Throw exceptions** on errors
- **Return models** (not entities)
- **Don't do business logic**

### Example: AgentRemoteDataSource

```dart
class AgentRemoteDataSource {
  final ApiClient apiClient;

  AgentRemoteDataSource(this.apiClient);

  /// Get agent profile by ID
  Future<AgentModel> getAgentProfile(String agentId) async {
    try {
      final endpoint = ApiEndpoints.replacePath(
        ApiEndpoints.agentProfile,
        {'agent_id': agentId},
      );

      final response = await apiClient.get(endpoint);

      if (response.statusCode == 200) {
        return AgentModel.fromJson(response.data as Map<String, dynamic>);
      } else if (response.statusCode == 404) {
        throw NotFoundException(message: 'Agent not found');
      } else {
        throw ServerException(
          message: 'Failed to get agent profile',
          statusCode: response.statusCode,
        );
      }
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Handle Dio errors
  Exception _handleDioError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
        return NetworkException(message: 'Connection timeout');
      case DioExceptionType.connectionError:
        return NetworkException(message: 'No internet connection');
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        if (statusCode == 401) {
          return UnauthorizedException(message: 'Unauthorized');
        } else if (statusCode == 404) {
          return NotFoundException(message: 'Resource not found');
        } else {
          return ServerException(
            message: error.response?.data?['message'] ?? 'Server error',
            statusCode: statusCode,
          );
        }
      default:
        return ServerException(message: 'Unknown error occurred');
    }
  }
}
```

### Data Source Methods

The AgentRemoteDataSource implements all 78 API operations:

**Profile Creation (UJ-001)**:
- `initiateProfileCreation()` - POST /agent-profiles/initiate
- `fetchHrmsData()` - POST /agent-profiles/{session_id}/fetch-hrms
- `linkCoordinator()` - POST /agent-profiles/{session_id}/link-coordinator
- `validateProfile()` - POST /agent-profiles/{session_id}/validate
- `submitProfile()` - POST /agent-profiles/{session_id}/submit

**Search & Retrieval (UJ-002)**:
- `searchAgents()` - GET /agents/search
- `getAgentProfile()` - GET /agents/{agent_id}
- `getAgentHierarchy()` - GET /agents/{agent_id}/hierarchy

**Updates (UJ-002)**:
- `updateAgentSection()` - PATCH /agents/{agent_id}/{section}

**Termination (UJ-004)**:
- `terminateAgent()` - POST /agents/{agent_id}/terminate
- `getTerminationLetter()` - GET /agents/{agent_id}/termination-letter

**Reinstatement (UJ-009)**:
- `createReinstatementRequest()` - POST /agents/{agent_id}/reinstatement-requests
- `approveReinstatement()` - POST /agents/{agent_id}/reinstatement-requests/{request_id}/approve

**Lookups**:
- `getAgentTypes()` - GET /lookups/agent-types
- `getStatusTypes()` - GET /lookups/status-types

**Validation**:
- `validatePanUniqueness()` - POST /validations/pan-uniqueness
- `validateEmployeeId()` - POST /validations/employee-id
- `validateIfscCode()` - POST /validations/ifsc-code
- `validateOfficeCode()` - POST /validations/office-code

**Session Management**:
- `getSessionStatus()` - GET /agent-profiles/sessions/{session_id}/status
- `cancelSession()` - POST /agent-profiles/sessions/{session_id}/cancel

---

## 🔄 Repository Implementation

### What is Repository Implementation?

Repository implementation:
- **Implements** the domain repository interface
- **Uses** data sources to fetch data
- **Converts** models to entities
- **Handles** exceptions and returns Either<Failure, Success>
- **Adds** caching logic (future enhancement)

### Example: AgentRepositoryImpl

```dart
class AgentRepositoryImpl implements AgentRepository {
  final AgentRemoteDataSource remoteDataSource;

  AgentRepositoryImpl({required this.remoteDataSource});

  @override
  Future<Either<Failure, AgentProfile>> getAgentProfile(String agentId) async {
    try {
      // 1. Call data source
      final model = await remoteDataSource.getAgentProfile(agentId);

      // 2. Convert model to entity
      final entity = model.toEntity();

      // 3. Return success
      return Right(entity);
    } on NotFoundException catch (e) {
      return Left(NotFoundFailure(message: e.message));
    } on NetworkException catch (e) {
      return Left(NetworkFailure(message: e.message));
    } on ServerException catch (e) {
      return Left(ServerFailure(
        message: e.message,
        statusCode: e.statusCode,
      ));
    } catch (e) {
      return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
    }
  }
}
```

### Error Handling Pattern

All repository methods follow this pattern:

```dart
try {
  // Call data source
  final result = await dataSource.someMethod();

  // Convert to entity if needed
  final entity = result.toEntity();

  // Return success
  return Right(entity);
} on SpecificException catch (e) {
  // Convert to specific failure
  return Left(SpecificFailure(message: e.message));
} on NetworkException catch (e) {
  return Left(NetworkFailure(message: e.message));
} on ServerException catch (e) {
  return Left(ServerFailure(message: e.message, statusCode: e.statusCode));
} catch (e) {
  // Catch-all for unknown errors
  return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
}
```

### Exception to Failure Mapping

| Exception | Failure |
|-----------|---------|
| NetworkException | NetworkFailure |
| ServerException | ServerFailure |
| ValidationException | ValidationFailure |
| NotFoundException | NotFoundFailure |
| UnauthorizedException | UnauthorizedFailure |
| ConflictException | ConflictFailure |

---

## 🔧 How to Add New API Operations

### Step 1: Update Data Source

Add the API method to `AgentRemoteDataSource`:

```dart
/// New API operation
/// POST /agents/{agent_id}/new-operation
Future<Map<String, dynamic>> newOperation({
  required String agentId,
  required String someParam,
}) async {
  try {
    final endpoint = ApiEndpoints.replacePath(
      ApiEndpoints.newOperation,  // Add to api_endpoints_config.dart first!
      {'agent_id': agentId},
    );

    final response = await apiClient.post(
      endpoint,
      data: {'some_param': someParam},
    );

    if (response.statusCode == 200) {
      return response.data as Map<String, dynamic>;
    } else {
      throw ServerException(
        message: 'Failed to perform new operation',
        statusCode: response.statusCode,
      );
    }
  } on DioException catch (e) {
    throw _handleDioError(e);
  }
}
```

### Step 2: Update Repository Interface

Add method to `AgentRepository` interface in domain layer:

```dart
abstract class AgentRepository {
  // ... existing methods

  /// New operation
  Future<Either<Failure, bool>> newOperation({
    required String agentId,
    required String someParam,
  });
}
```

### Step 3: Implement in Repository

Add implementation to `AgentRepositoryImpl`:

```dart
@override
Future<Either<Failure, bool>> newOperation({
  required String agentId,
  required String someParam,
}) async {
  try {
    await remoteDataSource.newOperation(
      agentId: agentId,
      someParam: someParam,
    );

    return const Right(true);
  } on NetworkException catch (e) {
    return Left(NetworkFailure(message: e.message));
  } on ServerException catch (e) {
    return Left(ServerFailure(
      message: e.message,
      statusCode: e.statusCode,
    ));
  } catch (e) {
    return Left(ServerFailure(message: 'Unknown error: ${e.toString()}'));
  }
}
```

---

## 📊 Model-Entity Mapping

### Why Convert?

**API Response** → **Model** → **Entity** → **Business Logic**

1. **API changes**: Only affect models, not entities
2. **Type safety**: Models parse strings to enums
3. **Date handling**: Models parse ISO strings to DateTime
4. **Field mapping**: API uses snake_case, Dart uses camelCase

### Example Mapping

**API Response**:
```json
{
  "agent_id": "AGT-2026-000001",
  "full_name": "John Doe",
  "status": "active",
  "date_of_birth": "1990-01-15"
}
```

**Model** (AgentModel):
```dart
AgentModel(
  agentId: "AGT-2026-000001",
  fullName: "John Doe",
  status: "active",               // String
  dateOfBirth: "1990-01-15",      // String
)
```

**Entity** (AgentProfile):
```dart
AgentProfile(
  agentId: "AGT-2026-000001",
  fullName: "John Doe",
  status: AgentStatus.active,     // Enum
  dateOfBirth: DateTime(1990, 1, 15),  // DateTime
)
```

---

## 🎨 How to Use

### In Dependency Injection

```dart
// Register data source
sl.registerLazySingleton(
  () => AgentRemoteDataSource(sl()),
);

// Register repository implementation
sl.registerLazySingleton<AgentRepository>(
  () => AgentRepositoryImpl(
    remoteDataSource: sl(),
  ),
);
```

### In Use Cases

```dart
class GetAgentProfileUseCase {
  final AgentRepository repository;  // Uses interface, not implementation

  GetAgentProfileUseCase(this.repository);

  Future<Either<Failure, AgentProfile>> call(String agentId) async {
    // Repository handles all data layer complexity
    return await repository.getAgentProfile(agentId);
  }
}
```

---

## ✅ Benefits of This Architecture

### 1. Separation of Concerns
- Data sources: API calls only
- Models: JSON handling only
- Repository: Orchestration and error handling
- Entities: Business logic only

### 2. Testability
```dart
// Easy to mock data source
test('should return AgentProfile when API call succeeds', () async {
  // Arrange
  when(mockDataSource.getAgentProfile(any))
      .thenAnswer((_) async => tAgentModel);

  // Act
  final result = await repository.getAgentProfile('AGT-001');

  // Assert
  expect(result, Right(tAgentProfile));
  verify(mockDataSource.getAgentProfile('AGT-001'));
});
```

### 3. API Independence
- Change API structure without affecting domain
- Switch between different APIs
- Add caching transparently

### 4. Error Handling
- Consistent error handling across all operations
- Type-safe failures
- Easy to handle different error scenarios

---

## 🚀 Next Steps

1. **Set Up Dependency Injection** ✅ (next)
   - Register data sources
   - Register repositories
   - Register use cases

2. **Create BLoCs/Cubits**
   - Use use cases in BLoCs
   - Manage UI state
   - Handle user events

3. **Build UI**
   - Consume BLoC state
   - Display entities
   - Handle user input

4. **Add Caching** (future enhancement)
   - Local data source
   - Cache strategy
   - Offline support

---

## 🧪 Testing

### Testing Data Sources

```dart
test('should return AgentModel when API call succeeds', () async {
  // Arrange
  when(mockApiClient.get(any))
      .thenAnswer((_) async => Response(
        data: tAgentJson,
        statusCode: 200,
      ));

  // Act
  final result = await dataSource.getAgentProfile('AGT-001');

  // Assert
  expect(result, equals(tAgentModel));
});
```

### Testing Repository

```dart
test('should return AgentProfile when data source succeeds', () async {
  // Arrange
  when(mockDataSource.getAgentProfile(any))
      .thenAnswer((_) async => tAgentModel);

  // Act
  final result = await repository.getAgentProfile('AGT-001');

  // Assert
  expect(result, equals(Right(tAgentProfile)));
});

test('should return NetworkFailure when network error occurs', () async {
  // Arrange
  when(mockDataSource.getAgentProfile(any))
      .thenThrow(NetworkException(message: 'No connection'));

  // Act
  final result = await repository.getAgentProfile('AGT-001');

  // Assert
  expect(result, equals(Left(NetworkFailure(message: 'No connection'))));
});
```

---

## 📚 Further Reading

- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [Data Mapper Pattern](https://martinfowler.com/eaaCatalog/dataMapper.html)
- [Freezed Package](https://pub.dev/packages/freezed)
- [JSON Serialization](https://docs.flutter.dev/data-and-backend/json)

---

## 📝 Summary

The data layer is now complete with:

✅ **AgentModel**: Complete data model with nested models
✅ **AgentRemoteDataSource**: All 78 API operations
✅ **AgentRepositoryImpl**: Repository implementation with error handling
✅ **Type-safe**: Freezed models with JSON serialization
✅ **Error Handling**: Comprehensive exception to failure conversion
✅ **Mapping**: Model to entity conversion

**Key Achievements**:
- All API operations implemented
- Consistent error handling pattern
- Type-safe JSON serialization
- Clean separation from domain layer
- Ready for dependency injection
- Ready for testing

**Next**: Set up dependency injection to wire everything together!

---

**Created**: 2026-01-29
**Phase**: 3 - Data Layer
**Progress**: Data Layer Complete ✅

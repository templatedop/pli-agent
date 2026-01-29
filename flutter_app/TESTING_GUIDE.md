# Testing Guide - Flutter App

## 📋 Overview

This guide covers both manual and automated testing for the PLI Agent Management app.

---

## 🏃 Running the App (Manual Testing)

### Step 1: Install Dependencies

```bash
cd /home/user/pli-agent/flutter_app
flutter pub get
```

### Step 2: Generate Code (CRITICAL!)

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

This generates the Freezed files for models (*.freezed.dart and *.g.dart).

### Step 3: Verify Setup

```bash
flutter doctor
```

Make sure everything shows checkmarks or at least no critical errors.

### Step 4: Run the App

```bash
# Run on Chrome (recommended for development)
flutter run -d chrome --dart-define=ENVIRONMENT=development

# Or run on any available device
flutter run
```

### Step 5: Manual Testing Checklist

Once the app is running:

**Home Screen**:
- [ ] App loads successfully
- [ ] Status cards display correctly
- [ ] "Search Agents" button is visible
- [ ] "Create Agent Profile" button is visible
- [ ] Clicking "Search Agents" navigates to search screen

**Agent Search Screen**:
- [ ] Search form displays with all fields
- [ ] Can expand/collapse search form
- [ ] Agent Type dropdown shows all types
- [ ] Status dropdown shows all statuses
- [ ] Clear button resets form
- [ ] Search button triggers search (will show error without API)
- [ ] Empty state shows when no search performed
- [ ] Error state shows when search fails
- [ ] Loading state shows during search

---

## 🧪 Automated Testing

### Test Structure

```
test/
├── unit/
│   ├── domain/
│   │   ├── usecases/
│   │   │   ├── get_agent_profile_usecase_test.dart
│   │   │   ├── search_agents_usecase_test.dart
│   │   │   └── create_agent_profile_usecase_test.dart
│   │   └── entities/
│   │       └── agent_profile_test.dart
│   │
│   └── presentation/
│       └── bloc/
│           └── agent_profile_bloc_test.dart
│
└── widget/
    ├── widgets/
    │   ├── custom_button_test.dart
    │   ├── custom_text_field_test.dart
    │   └── agent_profile_card_test.dart
    └── screens/
        └── agent_search_screen_test.dart
```

### Running Tests

```bash
# Run all tests
flutter test

# Run specific test file
flutter test test/unit/domain/usecases/get_agent_profile_usecase_test.dart

# Run with coverage
flutter test --coverage

# View coverage report (requires lcov)
genhtml coverage/lcov.info -o coverage/html
open coverage/html/index.html
```

---

## 📝 Test Examples

### 1. Use Case Unit Test

```dart
void main() {
  late GetAgentProfileUseCase useCase;
  late MockAgentRepository mockRepository;

  setUp(() {
    mockRepository = MockAgentRepository();
    useCase = GetAgentProfileUseCase(mockRepository);
  });

  test('should return AgentProfile when repository succeeds', () async {
    // Arrange
    when(() => mockRepository.getAgentProfile(any()))
        .thenAnswer((_) async => Right(tAgentProfile));

    // Act
    final result = await useCase('AGT-2026-000001');

    // Assert
    expect(result, Right(tAgentProfile));
    verify(() => mockRepository.getAgentProfile('AGT-2026-000001')).called(1);
  });

  test('should return Failure when repository fails', () async {
    // Arrange
    when(() => mockRepository.getAgentProfile(any()))
        .thenAnswer((_) async => Left(ServerFailure(message: 'Error')));

    // Act
    final result = await useCase('AGT-2026-000001');

    // Assert
    expect(result, Left(ServerFailure(message: 'Error')));
  });
}
```

### 2. BLoC Unit Test

```dart
void main() {
  late AgentProfileBloc bloc;
  late MockGetAgentProfileUseCase mockGetAgentProfileUseCase;

  setUp(() {
    mockGetAgentProfileUseCase = MockGetAgentProfileUseCase();
    bloc = AgentProfileBloc(
      getAgentProfileUseCase: mockGetAgentProfileUseCase,
      // ... other use cases
    );
  });

  tearDown(() {
    bloc.close();
  });

  blocTest<AgentProfileBloc, AgentProfileState>(
    'emits [Loading, Loaded] when GetAgentProfileEvent succeeds',
    build: () {
      when(() => mockGetAgentProfileUseCase(any()))
          .thenAnswer((_) async => Right(tAgentProfile));
      return bloc;
    },
    act: (bloc) => bloc.add(GetAgentProfileEvent('AGT-001')),
    expect: () => [
      const AgentProfileLoading(),
      AgentProfileLoaded(tAgentProfile),
    ],
  );
}
```

### 3. Widget Test

```dart
void main() {
  testWidgets('CustomButton displays text and responds to tap', (tester) async {
    bool tapped = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: CustomButton(
            text: 'Test Button',
            onPressed: () => tapped = true,
          ),
        ),
      ),
    );

    // Verify button displays text
    expect(find.text('Test Button'), findsOneWidget);

    // Tap button
    await tester.tap(find.byType(CustomButton));
    await tester.pump();

    // Verify callback was called
    expect(tapped, true);
  });
}
```

---

## 🐛 Common Issues

### Issue: "MissingPluginException"

**Cause**: Running on web without proper setup

**Solution**:
```bash
flutter clean
flutter pub get
flutter run -d chrome
```

### Issue: "part of" errors

**Cause**: Freezed files not generated

**Solution**:
```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### Issue: Tests fail with "No implementation found"

**Cause**: Missing mock implementations

**Solution**: Add mocks in test file:
```dart
class MockAgentRepository extends Mock implements AgentRepository {}
```

### Issue: "DioException" in tests

**Cause**: Real HTTP calls being made

**Solution**: Mock the ApiClient:
```dart
when(() => mockApiClient.get(any()))
    .thenAnswer((_) async => Response(data: {...}, statusCode: 200));
```

---

## ✅ Testing Best Practices

### 1. Test Naming

Use clear, descriptive names:
```dart
// ✅ Good
test('should return AgentProfile when repository succeeds', () {});

// ❌ Bad
test('test1', () {});
```

### 2. AAA Pattern

Arrange-Act-Assert:
```dart
test('description', () async {
  // Arrange: Set up test data and mocks
  when(() => mock.method()).thenAnswer(...);

  // Act: Execute the code under test
  final result = await useCase();

  // Assert: Verify the results
  expect(result, expectedValue);
});
```

### 3. Test Independence

Each test should be independent:
```dart
setUp(() {
  // Fresh instances for each test
  mock = MockRepository();
  useCase = UseCase(mock);
});

tearDown(() {
  // Clean up after each test
  bloc.close();
});
```

### 4. Mock External Dependencies

Always mock:
- Repositories (in use case tests)
- Use cases (in BLoC tests)
- API clients (in data source tests)

Never mock:
- Entities (use real instances)
- Value objects (use real instances)

### 5. Test Edge Cases

Don't just test the happy path:
```dart
test('should handle empty list', () {});
test('should handle null values', () {});
test('should handle network errors', () {});
test('should validate invalid input', () {});
```

---

## 📊 Coverage Goals

**Targets**:
- Domain layer: 90%+ (critical business logic)
- Data layer: 80%+ (API integration)
- Presentation layer: 70%+ (BLoC)
- Widgets: 60%+ (UI components)

**Priority**:
1. Use cases (highest priority)
2. Repository implementations
3. BLoCs
4. Complex widgets
5. Simple widgets (lowest priority)

---

## 🚀 Next Steps

### Before Committing Code:

1. **Run all tests**: `flutter test`
2. **Check coverage**: `flutter test --coverage`
3. **Run analyzer**: `flutter analyze`
4. **Format code**: `dart format .`

### Continuous Testing:

```bash
# Watch mode - auto-run tests on file changes
flutter test --watch
```

---

## 📚 Testing Tools

### Required Packages (already in pubspec.yaml):

```yaml
dev_dependencies:
  flutter_test:
    sdk: flutter
  bloc_test: ^9.1.0        # BLoC testing
  mocktail: ^1.0.0         # Mocking
```

### Useful Commands:

```bash
# Test with verbose output
flutter test --verbose

# Test specific file
flutter test test/unit/domain/usecases/

# Test with coverage
flutter test --coverage

# Analyze code
flutter analyze

# Format code
dart format lib/ test/
```

---

## 📝 Summary

**Manual Testing**:
- ✅ Run app with `flutter run -d chrome`
- ✅ Test all user flows
- ✅ Check error handling
- ✅ Verify UI responsiveness

**Automated Testing**:
- ✅ Unit tests for use cases
- ✅ Unit tests for BLoCs
- ✅ Widget tests for components
- ✅ Integration tests for user flows

**Coverage**:
- ✅ Domain layer: 90%+
- ✅ Data layer: 80%+
- ✅ Presentation: 70%+

**Before Commit**:
- ✅ All tests pass
- ✅ No analyzer errors
- ✅ Code formatted

---

**Created**: 2026-01-29
**Phase**: Testing
**Status**: Ready to Test! ✅

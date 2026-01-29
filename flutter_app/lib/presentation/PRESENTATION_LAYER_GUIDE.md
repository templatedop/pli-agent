## 📋 Overview

The **presentation layer** is the outermost layer in Clean Architecture. It contains:
- **BLoCs/Cubits**: State management and business logic orchestration
- **Screens**: Page-level widgets
- **Widgets**: Reusable UI components
- **Routes**: Navigation configuration

**Key Principle**: The presentation layer depends on the domain layer (use cases) but knows nothing about the data layer.

---

## 📁 Structure

```
lib/presentation/
├── bloc/
│   └── agent_profile/
│       ├── agent_profile_bloc.dart      # BLoC class
│       ├── agent_profile_event.dart     # Events
│       ├── agent_profile_state.dart     # States
│       └── agent_profile_barrel.dart    # Exports
│
├── screens/
│   ├── home_screen.dart                 # Home/dashboard
│   └── agent/                           # Agent-related screens
│       ├── agent_profile_screen.dart
│       ├── agent_create_screen.dart
│       └── agent_search_screen.dart
│
├── widgets/
│   ├── common/                          # Shared widgets
│   │   ├── custom_button.dart
│   │   ├── custom_text_field.dart
│   │   └── loading_indicator.dart
│   └── agent/                           # Agent-specific widgets
│       ├── agent_profile_card.dart
│       └── agent_search_bar.dart
│
└── PRESENTATION_LAYER_GUIDE.md          # This file
```

---

## 🎯 BLoC Pattern

### What is BLoC?

**BLoC** (Business Logic Component) is a state management pattern that:
- **Separates** business logic from UI
- **Uses streams** for reactive programming
- **Makes testing easier** (test logic without UI)
- **Reuses logic** across multiple widgets

### BLoC Architecture

```
Widget (UI)
    ↓ (adds event)
Event
    ↓
BLoC
    ↓ (calls)
Use Case
    ↓
Repository
    ↓ (returns)
BLoC
    ↓ (emits)
State
    ↓ (rebuilds)
Widget (UI)
```

---

## 🔄 Agent Profile BLoC

### Events

Events are user actions or triggers:

```dart
// Get agent profile
add(GetAgentProfileEvent('AGT-2026-000001'));

// Search agents
add(SearchAgentsEvent(
  criteria: {'status': 'active'},
  page: 1,
  limit: 20,
));

// Create agent profile
add(CreateAgentProfileEvent(params));

// Update agent profile
add(UpdateAgentProfileEvent(params));

// Reset to initial state
add(ResetAgentProfileEvent());
```

### States

States represent the current status:

**Loading States**:
- `AgentProfileLoading` - Loading single profile
- `AgentSearchLoading` - Loading search results
- `AgentProfileCreating` - Creating profile
- `AgentProfileUpdating` - Updating profile

**Success States**:
- `AgentProfileLoaded` - Profile loaded
- `AgentSearchLoaded` - Search results ready
- `AgentProfileCreated` - Profile created successfully
- `AgentProfileUpdated` - Profile updated successfully

**Error States**:
- `AgentProfileError` - General error
- `AgentProfileValidationError` - Validation failed
- `AgentProfileNetworkError` - No internet connection

**Empty States**:
- `AgentSearchEmpty` - No results found

### BLoC Class

```dart
class AgentProfileBloc extends Bloc<AgentProfileEvent, AgentProfileState> {
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
    on<GetAgentProfileEvent>(_onGetAgentProfile);
    on<SearchAgentsEvent>(_onSearchAgents);
    on<CreateAgentProfileEvent>(_onCreateAgentProfile);
    on<UpdateAgentProfileEvent>(_onUpdateAgentProfile);
  }

  Future<void> _onGetAgentProfile(
    GetAgentProfileEvent event,
    Emitter<AgentProfileState> emit,
  ) async {
    emit(const AgentProfileLoading());

    final result = await getAgentProfileUseCase(event.agentId);

    result.fold(
      (failure) => emit(AgentProfileError(message: failure.message)),
      (agentProfile) => emit(AgentProfileLoaded(agentProfile)),
    );
  }
}
```

---

## 🎨 Using BLoC in UI

### Step 1: Provide BLoC

Wrap your screen with `BlocProvider`:

```dart
class AgentProfileScreen extends StatelessWidget {
  final String agentId;

  const AgentProfileScreen({required this.agentId});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AgentProfileBloc>()
        ..add(GetAgentProfileEvent(agentId)),  // Load on init
      child: const AgentProfileView(),
    );
  }
}
```

### Step 2: Listen to State

Use `BlocBuilder` to rebuild on state changes:

```dart
class AgentProfileView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Agent Profile')),
      body: BlocBuilder<AgentProfileBloc, AgentProfileState>(
        builder: (context, state) {
          // Loading state
          if (state is AgentProfileLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          // Error state
          if (state is AgentProfileError) {
            return Center(
              child: Text(
                state.message,
                style: const TextStyle(color: Colors.red),
              ),
            );
          }

          // Success state
          if (state is AgentProfileLoaded) {
            return AgentProfileDetails(agentProfile: state.agentProfile);
          }

          // Initial state
          return const Center(child: Text('No data'));
        },
      ),
    );
  }
}
```

### Step 3: Dispatch Events

Use `context.read<BLoC>().add(event)`:

```dart
ElevatedButton(
  onPressed: () {
    context.read<AgentProfileBloc>().add(
      UpdateAgentProfileEvent(params),
    );
  },
  child: const Text('Update Profile'),
)
```

---

## 🔧 Advanced BLoC Features

### BlocListener

Listen to state without rebuilding:

```dart
BlocListener<AgentProfileBloc, AgentProfileState>(
  listener: (context, state) {
    // Show snackbar on success
    if (state is AgentProfileCreated) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(state.message)),
      );
    }

    // Navigate on success
    if (state is AgentProfileCreated) {
      Navigator.of(context).pop();
    }

    // Show error dialog
    if (state is AgentProfileError) {
      showDialog(
        context: context,
        builder: (_) => AlertDialog(
          title: const Text('Error'),
          content: Text(state.message),
        ),
      );
    }
  },
  child: YourWidget(),
)
```

### BlocConsumer

Combine builder and listener:

```dart
BlocConsumer<AgentProfileBloc, AgentProfileState>(
  listener: (context, state) {
    // Side effects (snackbars, navigation)
    if (state is AgentProfileCreated) {
      Navigator.pop(context);
    }
  },
  builder: (context, state) {
    // UI updates
    if (state is AgentProfileLoading) {
      return const CircularProgressIndicator();
    }
    return YourWidget();
  },
)
```

### BlocSelector

Rebuild only when specific data changes:

```dart
BlocSelector<AgentProfileBloc, AgentProfileState, String>(
  selector: (state) {
    if (state is AgentProfileLoaded) {
      return state.agentProfile.fullName;
    }
    return '';
  },
  builder: (context, fullName) {
    return Text(fullName);  // Rebuilds only when name changes
  },
)
```

---

## 📊 State Management Best Practices

### 1. Keep BLoCs Focused

**Good**: One BLoC per feature
```dart
AgentProfileBloc  // Handles agent profile operations
SearchBloc        // Handles search operations
```

**Bad**: One BLoC for everything
```dart
AppBloc  // Too broad
```

### 2. Use Immutable States

**Good**: Equatable for equality comparison
```dart
class AgentProfileLoaded extends AgentProfileState {
  final AgentProfile agentProfile;

  const AgentProfileLoaded(this.agentProfile);

  @override
  List<Object?> get props => [agentProfile];
}
```

**Bad**: Mutable state
```dart
class AgentProfileState {
  AgentProfile? agentProfile;  // Can be changed
}
```

### 3. Handle All States in UI

```dart
BlocBuilder<AgentProfileBloc, AgentProfileState>(
  builder: (context, state) {
    if (state is AgentProfileLoading) return LoadingWidget();
    if (state is AgentProfileError) return ErrorWidget(state.message);
    if (state is AgentProfileLoaded) return ProfileWidget(state.agentProfile);

    // Don't forget initial/empty states!
    return EmptyWidget();
  },
)
```

### 4. Dispose BLoCs Properly

BLoCs are automatically disposed when removed from widget tree with `BlocProvider`.

For manual management:
```dart
@override
void dispose() {
  bloc.close();  // Close streams
  super.dispose();
}
```

---

## 🧪 Testing BLoCs

### Unit Testing

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

  test('should emit [Loading, Loaded] when GetAgentProfileEvent is added', () {
    // Arrange
    when(() => mockGetAgentProfileUseCase(any()))
        .thenAnswer((_) async => Right(tAgentProfile));

    // Assert later
    expectLater(
      bloc.stream,
      emitsInOrder([
        const AgentProfileLoading(),
        AgentProfileLoaded(tAgentProfile),
      ]),
    );

    // Act
    bloc.add(GetAgentProfileEvent('AGT-001'));
  });

  test('should emit [Loading, Error] when use case fails', () {
    // Arrange
    when(() => mockGetAgentProfileUseCase(any()))
        .thenAnswer((_) async => Left(ServerFailure(message: 'Error')));

    // Assert later
    expectLater(
      bloc.stream,
      emitsInOrder([
        const AgentProfileLoading(),
        isA<AgentProfileError>(),
      ]),
    );

    // Act
    bloc.add(GetAgentProfileEvent('AGT-001'));
  });
}
```

### Widget Testing

```dart
testWidgets('should display loading indicator when state is loading', (tester) async {
  await tester.pumpWidget(
    MaterialApp(
      home: BlocProvider(
        create: (_) => mockBloc,
        child: AgentProfileScreen(agentId: 'AGT-001'),
      ),
    ),
  );

  // Emit loading state
  whenListen(
    mockBloc,
    Stream.fromIterable([const AgentProfileLoading()]),
    initialState: const AgentProfileInitial(),
  );

  await tester.pump();

  // Verify
  expect(find.byType(CircularProgressIndicator), findsOneWidget);
});
```

---

## 📚 Common Patterns

### Loading with Progress

```dart
class AgentProfileLoading extends AgentProfileState {
  final double progress;

  const AgentProfileLoading({this.progress = 0.0});
}

// In UI
if (state is AgentProfileLoading) {
  return LinearProgressIndicator(value: state.progress);
}
```

### Pagination

```dart
class AgentSearchLoaded extends AgentProfileState {
  final List<AgentProfile> agents;
  final int currentPage;
  final bool hasMorePages;

  bool get canLoadMore => hasMorePages;
}

// Load more
if (state.canLoadMore) {
  context.read<AgentProfileBloc>().add(
    SearchAgentsEvent(page: state.currentPage + 1),
  );
}
```

### Optimistic Updates

```dart
Future<void> _onUpdateAgentProfile(...) async {
  // Emit success immediately (optimistic)
  emit(AgentProfileUpdated(updatedProfile, 'Updating...'));

  final result = await updateAgentProfileUseCase(params);

  result.fold(
    (failure) {
      // Revert on failure
      emit(AgentProfileError(message: failure.message));
    },
    (agentProfile) {
      // Confirm success
      emit(AgentProfileUpdated(agentProfile, 'Updated successfully'));
    },
  );
}
```

---

## ✅ Benefits of BLoC Pattern

### 1. Separation of Concerns
- Business logic separate from UI
- Easy to understand and maintain
- Clear responsibilities

### 2. Testability
- Test business logic without UI
- Mock dependencies easily
- Fast unit tests

### 3. Reusability
- Same BLoC in multiple widgets
- Share state across screens
- Reduce code duplication

### 4. Reactive
- UI automatically updates on state changes
- No manual setState() calls
- Stream-based for async operations

---

## 🚀 Next Steps

1. **Create Screens** (in progress)
   - Agent profile detail screen
   - Agent creation form
   - Agent search screen

2. **Create Widgets**
   - Profile cards
   - Form fields
   - Loading indicators
   - Error displays

3. **Add Navigation**
   - Set up go_router
   - Define routes
   - Handle deep linking

4. **Polish UI**
   - Add animations
   - Improve UX
   - Add accessibility

---

## 📚 Further Reading

- [Flutter BLoC Documentation](https://bloclibrary.dev/)
- [BLoC Pattern Guide](https://www.didierboelens.com/2018/08/reactive-programming-streams-bloc/)
- [Flutter State Management](https://docs.flutter.dev/data-and-backend/state-mgmt/options)
- [Testing BLoCs](https://bloclibrary.dev/#/testing)

---

## 📝 Summary

The presentation layer is now set up with:

✅ **AgentProfileBloc**: Complete BLoC with all operations
✅ **Events**: Get, Search, Create, Update, Reset
✅ **States**: Loading, Success, Error, Empty
✅ **Dependency Injection**: BLoC registered in DI container
✅ **Documentation**: Complete guide for BLoC pattern

**Key Achievements**:
- Clean separation of UI and business logic
- Type-safe state management
- Ready for widget integration
- Easy to test
- Follows Flutter best practices

**Next**: Create UI screens to consume the BLoC!

---

**Created**: 2026-01-29
**Phase**: 5 - Presentation Layer (BLoC)
**Progress**: BLoC Complete ✅

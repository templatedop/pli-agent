import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:pli_agent_management/presentation/widgets/common/custom_button.dart';

void main() {
  group('CustomButton', () {
    testWidgets('displays text correctly', (tester) async {
      // Arrange
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
            ),
          ),
        ),
      );

      // Assert
      expect(find.text('Test Button'), findsOneWidget);
    });

    testWidgets('calls onPressed when tapped', (tester) async {
      // Arrange
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

      // Act
      await tester.tap(find.byType(CustomButton));
      await tester.pump();

      // Assert
      expect(tapped, true);
    });

    testWidgets('does not call onPressed when disabled', (tester) async {
      // Arrange
      bool tapped = false;

      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              onPressed: null,  // Disabled
            ),
          ),
        ),
      );

      // Act
      await tester.tap(find.byType(CustomButton));
      await tester.pump();

      // Assert
      expect(tapped, false);
    });

    testWidgets('shows loading indicator when isLoading is true', (tester) async {
      // Arrange
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              isLoading: true,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Test Button'), findsNothing);
    });

    testWidgets('displays icon when provided', (tester) async {
      // Arrange
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              icon: Icons.add,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      expect(find.byIcon(Icons.add), findsOneWidget);
      expect(find.text('Test Button'), findsOneWidget);
    });

    testWidgets('respects custom width', (tester) async {
      // Arrange
      const customWidth = 200.0;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              width: customWidth,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      final sizedBox = tester.widget<SizedBox>(
        find.ancestor(
          of: find.byType(ElevatedButton),
          matching: find.byType(SizedBox),
        ).first,
      );
      expect(sizedBox.width, customWidth);
    });

    testWidgets('renders as ElevatedButton for primary type', (tester) async {
      // Arrange
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              type: ButtonType.primary,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      expect(find.byType(ElevatedButton), findsOneWidget);
    });

    testWidgets('renders as OutlinedButton for outlined type', (tester) async {
      // Arrange
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              type: ButtonType.outlined,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      expect(find.byType(OutlinedButton), findsOneWidget);
    });

    testWidgets('renders as TextButton for text type', (tester) async {
      // Arrange
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              type: ButtonType.text,
              onPressed: () {},
            ),
          ),
        ),
      );

      // Assert
      expect(find.byType(TextButton), findsOneWidget);
    });

    testWidgets('handles tap with icon correctly', (tester) async {
      // Arrange
      bool tapped = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              icon: Icons.search,
              onPressed: () => tapped = true,
            ),
          ),
        ),
      );

      // Act
      await tester.tap(find.byType(CustomButton));
      await tester.pump();

      // Assert
      expect(tapped, true);
      expect(find.byIcon(Icons.search), findsOneWidget);
    });

    testWidgets('button is disabled when isLoading and onPressed is not null', (tester) async {
      // Arrange
      bool tapped = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              isLoading: true,
              onPressed: () => tapped = true,
            ),
          ),
        ),
      );

      // Act - Try to tap
      await tester.tap(find.byType(CustomButton));
      await tester.pump();

      // Assert - Should not be tapped because loading
      expect(tapped, false);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });
  });
}

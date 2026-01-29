#!/bin/bash

# Flutter App Setup Script
# Run this script after installing Flutter SDK

set -e  # Exit on error

echo "========================================="
echo "PLI Agent Management - Flutter Setup"
echo "========================================="
echo ""

# Check if Flutter is installed
if ! command -v flutter &> /dev/null
then
    echo "❌ Flutter is not installed!"
    echo ""
    echo "Please install Flutter first:"
    echo "  macOS/Linux: https://docs.flutter.dev/get-started/install"
    echo "  Windows: https://docs.flutter.dev/get-started/install/windows"
    echo ""
    exit 1
fi

echo "✅ Flutter found: $(flutter --version | head -n 1)"
echo ""

# Navigate to flutter_app directory
cd "$(dirname "$0")"

# Step 1: Install dependencies
echo "📦 Step 1: Installing dependencies..."
flutter pub get

if [ $? -eq 0 ]; then
    echo "✅ Dependencies installed successfully"
else
    echo "❌ Failed to install dependencies"
    exit 1
fi

echo ""

# Step 2: Generate code
echo "🔨 Step 2: Generating code for Freezed models..."
flutter pub run build_runner build --delete-conflicting-outputs

if [ $? -eq 0 ]; then
    echo "✅ Code generation completed"
else
    echo "❌ Code generation failed"
    exit 1
fi

echo ""

# Step 3: Verify setup
echo "🔍 Step 3: Verifying Flutter setup..."
flutter doctor

echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Run the app:"
echo "     flutter run -d chrome --dart-define=ENVIRONMENT=development"
echo ""
echo "  2. Or use watch mode for development:"
echo "     flutter pub run build_runner watch"
echo ""
echo "  3. Read CONFIGURATION_GUIDE.md for configuration"
echo ""
echo "Happy coding! 🚀"
echo ""

@echo off
REM Flutter App Setup Script for Windows
REM Run this script after installing Flutter SDK

echo =========================================
echo PLI Agent Management - Flutter Setup
echo =========================================
echo.

REM Check if Flutter is installed
where flutter >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [X] Flutter is not installed!
    echo.
    echo Please install Flutter first:
    echo   https://docs.flutter.dev/get-started/install/windows
    echo.
    pause
    exit /b 1
)

flutter --version | findstr /C:"Flutter"
echo.

REM Navigate to flutter_app directory
cd /d "%~dp0"

REM Step 1: Install dependencies
echo [*] Step 1: Installing dependencies...
call flutter pub get

if %ERRORLEVEL% NEQ 0 (
    echo [X] Failed to install dependencies
    pause
    exit /b 1
)

echo [OK] Dependencies installed successfully
echo.

REM Step 2: Generate code
echo [*] Step 2: Generating code for Freezed models...
call flutter pub run build_runner build --delete-conflicting-outputs

if %ERRORLEVEL% NEQ 0 (
    echo [X] Code generation failed
    pause
    exit /b 1
)

echo [OK] Code generation completed
echo.

REM Step 3: Verify setup
echo [*] Step 3: Verifying Flutter setup...
call flutter doctor

echo.
echo =========================================
echo [OK] Setup Complete!
echo =========================================
echo.
echo Next steps:
echo   1. Run the app:
echo      flutter run -d chrome --dart-define=ENVIRONMENT=development
echo.
echo   2. Or use watch mode for development:
echo      flutter pub run build_runner watch
echo.
echo   3. Read CONFIGURATION_GUIDE.md for configuration
echo.
echo Happy coding! 🚀
echo.
pause

import 'package:flutter/material.dart';

class AppColors {
  // Primary Colors (PLI Brand - Indian Postal Colors)
  static const Color primary = Color(0xFF1565C0); // Deep Blue
  static const Color primaryLight = Color(0xFFE3F2FD);
  static const Color primaryDark = Color(0xFF0D47A1);

  // Secondary Colors
  static const Color secondary = Color(0xFFFFC107); // Amber/Gold
  static const Color secondaryLight = Color(0xFFFFF8E1);
  static const Color secondaryDark = Color(0xFFFFA000);

  // Status Colors
  static const Color success = Color(0xFF4CAF50); // Green
  static const Color successLight = Color(0xFFE8F5E9);
  static const Color warning = Color(0xFFFF9800); // Orange
  static const Color warningLight = Color(0xFFFFF3E0);
  static const Color error = Color(0xFFD32F2F); // Red
  static const Color errorLight = Color(0xFFFFEBEE);
  static const Color info = Color(0xFF2196F3); // Blue
  static const Color infoLight = Color(0xFFE3F2FD);

  // Agent Status Colors
  static const Color statusActive = Color(0xFF4CAF50);
  static const Color statusInactive = Color(0xFF9E9E9E);
  static const Color statusSuspended = Color(0xFFFF9800);
  static const Color statusTerminated = Color(0xFFD32F2F);
  static const Color statusDeactivated = Color(0xFF757575);

  // License Status Colors
  static const Color licenseActive = Color(0xFF4CAF50);
  static const Color licenseExpired = Color(0xFFD32F2F);
  static const Color licenseExpiringSoon = Color(0xFFFF9800);

  // SLA Status Colors
  static const Color slaGreen = Color(0xFF4CAF50);
  static const Color slaYellow = Color(0xFFFFC107);
  static const Color slaRed = Color(0xFFD32F2F);

  // Neutral Colors
  static const Color background = Color(0xFFF5F5F5);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color divider = Color(0xFFE0E0E0);

  // Text Colors
  static const Color textPrimary = Color(0xFF212121);
  static const Color textSecondary = Color(0xFF757575);
  static const Color textDisabled = Color(0xFFBDBDBD);
  static const Color textHint = Color(0xFF9E9E9E);

  // Workflow State Colors
  static const Color workflowInitiated = Color(0xFF2196F3);
  static const Color workflowInProgress = Color(0xFFFF9800);
  static const Color workflowCompleted = Color(0xFF4CAF50);
  static const Color workflowPending = Color(0xFFFFC107);
  static const Color workflowRejected = Color(0xFFD32F2F);

  // Chart Colors (for dashboard)
  static const List<Color> chartColors = [
    Color(0xFF1565C0),
    Color(0xFF4CAF50),
    Color(0xFFFFC107),
    Color(0xFFD32F2F),
    Color(0xFF9C27B0),
    Color(0xFF00BCD4),
    Color(0xFFFF5722),
    Color(0xFF607D8B),
  ];

  // Gradient
  static const LinearGradient primaryGradient = LinearGradient(
    colors: [primary, primaryDark],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // Get status color by status string
  static Color getStatusColor(String status) {
    switch (status.toUpperCase()) {
      case 'ACTIVE':
        return statusActive;
      case 'INACTIVE':
        return statusInactive;
      case 'SUSPENDED':
        return statusSuspended;
      case 'TERMINATED':
        return statusTerminated;
      case 'DEACTIVATED':
        return statusDeactivated;
      default:
        return textSecondary;
    }
  }

  // Get SLA status color
  static Color getSlaColor(String sla) {
    switch (sla.toUpperCase()) {
      case 'GREEN':
        return slaGreen;
      case 'YELLOW':
        return slaYellow;
      case 'RED':
        return slaRed;
      default:
        return textSecondary;
    }
  }

  // Get workflow state color
  static Color getWorkflowColor(String state) {
    switch (state.toUpperCase()) {
      case 'INITIATED':
        return workflowInitiated;
      case 'IN_PROGRESS':
      case 'PENDING':
        return workflowInProgress;
      case 'COMPLETED':
        return workflowCompleted;
      case 'REJECTED':
        return workflowRejected;
      default:
        return textSecondary;
    }
  }
}

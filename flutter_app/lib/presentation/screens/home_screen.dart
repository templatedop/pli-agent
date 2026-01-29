/// Home Screen - Placeholder
///
/// This is a temporary placeholder screen while we build out the full UI.
/// It will be replaced with the actual dashboard/home screen.

import 'package:flutter/material.dart';
import '../../core/theme/app_colors.dart';
import '../widgets/common/custom_button.dart';
import 'agent/agent_search_screen.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('PLI Agent Management System'),
        centerTitle: true,
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // PLI Logo placeholder
              Container(
                width: 120,
                height: 120,
                decoration: BoxDecoration(
                  color: AppColors.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(60),
                ),
                child: const Icon(
                  Icons.business,
                  size: 60,
                  color: AppColors.primary,
                ),
              ),

              const SizedBox(height: 32),

              // Welcome text
              Text(
                'Welcome to PLI Agent Management',
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                textAlign: TextAlign.center,
              ),

              const SizedBox(height: 16),

              Text(
                'Department of Posts - India Post',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: AppColors.textSecondary,
                    ),
                textAlign: TextAlign.center,
              ),

              const SizedBox(height: 48),

              // Status cards
              _buildStatusCard(
                context,
                title: 'Domain Layer',
                status: 'Complete',
                icon: Icons.check_circle,
                color: AppColors.success,
              ),

              const SizedBox(height: 16),

              _buildStatusCard(
                context,
                title: 'Data Layer',
                status: 'Complete',
                icon: Icons.check_circle,
                color: AppColors.success,
              ),

              const SizedBox(height: 16),

              _buildStatusCard(
                context,
                title: 'Dependency Injection',
                status: 'Complete',
                icon: Icons.check_circle,
                color: AppColors.success,
              ),

              const SizedBox(height: 16),

              _buildStatusCard(
                context,
                title: 'UI Layer',
                status: 'In Progress',
                icon: Icons.hourglass_empty,
                color: AppColors.warning,
              ),

              const SizedBox(height: 48),

              // Action buttons
              CustomButton(
                text: 'Search Agents',
                icon: Icons.search,
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => const AgentSearchScreen(),
                    ),
                  );
                },
                width: double.infinity,
              ),

              const SizedBox(height: 16),

              CustomButton(
                text: 'Create Agent Profile',
                icon: Icons.person_add,
                onPressed: () {
                  // TODO: Navigate to create agent screen
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Agent creation screen coming soon!'),
                    ),
                  );
                },
                type: ButtonType.secondary,
                width: double.infinity,
              ),

              const SizedBox(height: 48),

              // Info text
              Text(
                'Architecture: Clean Architecture with BLoC',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                      fontStyle: FontStyle.italic,
                    ),
              ),

              const SizedBox(height: 8),

              Text(
                '78 API endpoints • 4 use cases • Type-safe error handling',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.textSecondary,
                    ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatusCard(
    BuildContext context, {
    required String title,
    required String status,
    required IconData icon,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: color.withOpacity(0.3),
          width: 1,
        ),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 32),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  status,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: color,
                        fontWeight: FontWeight.w500,
                      ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Agent Profile Card Widget
///
/// Displays agent profile summary in a card format

import 'package:flutter/material.dart';
import '../../../core/theme/app_colors.dart';
import '../../../domain/entities/agent_profile.dart';

class AgentProfileCard extends StatelessWidget {
  final AgentProfile agent;
  final VoidCallback? onTap;

  const AgentProfileCard({
    super.key,
    required this.agent,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: AppColors.border.withOpacity(0.3),
        ),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header row
              Row(
                children: [
                  // Avatar
                  CircleAvatar(
                    radius: 24,
                    backgroundColor: _getStatusColor().withOpacity(0.2),
                    child: Text(
                      _getInitials(),
                      style: TextStyle(
                        color: _getStatusColor(),
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),

                  // Name and ID
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          agent.fullName,
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          agent.agentId,
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: AppColors.textSecondary,
                              ),
                        ),
                      ],
                    ),
                  ),

                  // Status badge
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: _getStatusColor().withOpacity(0.1),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: _getStatusColor().withOpacity(0.3),
                      ),
                    ),
                    child: Text(
                      _getStatusText(),
                      style: TextStyle(
                        color: _getStatusColor(),
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 16),
              const Divider(height: 1),
              const SizedBox(height: 12),

              // Details grid
              Row(
                children: [
                  Expanded(
                    child: _buildDetailItem(
                      context,
                      icon: Icons.business_center,
                      label: 'Type',
                      value: _getAgentTypeText(),
                    ),
                  ),
                  Expanded(
                    child: _buildDetailItem(
                      context,
                      icon: Icons.location_on_outlined,
                      label: 'Office',
                      value: agent.office?.officeName ?? 'N/A',
                    ),
                  ),
                ],
              ),

              if (agent.agentType == AgentType.advisor && agent.advisorCoordinator != null) ...[
                const SizedBox(height: 12),
                _buildDetailItem(
                  context,
                  icon: Icons.person_outline,
                  label: 'Coordinator',
                  value: agent.advisorCoordinator!.coordinatorName,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDetailItem(
    BuildContext context, {
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Row(
      children: [
        Icon(
          icon,
          size: 16,
          color: AppColors.textSecondary,
        ),
        const SizedBox(width: 6),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.textSecondary,
                      fontSize: 11,
                    ),
              ),
              Text(
                value,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
      ],
    );
  }

  String _getInitials() {
    final names = agent.fullName.split(' ');
    if (names.length >= 2) {
      return '${names[0][0]}${names[1][0]}'.toUpperCase();
    }
    return agent.fullName.substring(0, 2).toUpperCase();
  }

  Color _getStatusColor() {
    switch (agent.status) {
      case AgentStatus.active:
        return AppColors.success;
      case AgentStatus.inactive:
        return AppColors.textSecondary;
      case AgentStatus.suspended:
        return AppColors.warning;
      case AgentStatus.terminated:
        return AppColors.error;
      case AgentStatus.deactivated:
        return AppColors.textSecondary;
    }
  }

  String _getStatusText() {
    switch (agent.status) {
      case AgentStatus.active:
        return 'Active';
      case AgentStatus.inactive:
        return 'Inactive';
      case AgentStatus.suspended:
        return 'Suspended';
      case AgentStatus.terminated:
        return 'Terminated';
      case AgentStatus.deactivated:
        return 'Deactivated';
    }
  }

  String _getAgentTypeText() {
    switch (agent.agentType) {
      case AgentType.advisor:
        return 'Advisor';
      case AgentType.advisorCoordinator:
        return 'Coordinator';
      case AgentType.departmentalEmployee:
        return 'Employee';
      case AgentType.fieldOfficer:
        return 'Field Officer';
      case AgentType.directAgent:
        return 'Direct Agent';
      case AgentType.gds:
        return 'GDS';
    }
  }
}

/// Agent Search Screen
///
/// Allows searching for agents with various criteria (UJ-002)

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../core/di/injection_container.dart';
import '../../../core/theme/app_colors.dart';
import '../../../domain/entities/agent_profile.dart';
import '../../bloc/agent_profile/agent_profile_barrel.dart';
import '../../widgets/agent/agent_profile_card.dart';
import '../../widgets/common/custom_button.dart';
import '../../widgets/common/custom_dropdown.dart';
import '../../widgets/common/custom_text_field.dart';
import '../../widgets/common/error_display.dart';
import '../../widgets/common/loading_indicator.dart';

class AgentSearchScreen extends StatelessWidget {
  const AgentSearchScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AgentProfileBloc>(),
      child: const AgentSearchView(),
    );
  }
}

class AgentSearchView extends StatefulWidget {
  const AgentSearchView({super.key});

  @override
  State<AgentSearchView> createState() => _AgentSearchViewState();
}

class _AgentSearchViewState extends State<AgentSearchView> {
  final _formKey = GlobalKey<FormState>();
  final _agentIdController = TextEditingController();
  final _nameController = TextEditingController();
  final _panController = TextEditingController();

  AgentType? _selectedAgentType;
  AgentStatus? _selectedStatus;
  bool _isSearchExpanded = true;

  @override
  void dispose() {
    _agentIdController.dispose();
    _nameController.dispose();
    _panController.dispose();
    super.dispose();
  }

  void _performSearch() {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final criteria = <String, dynamic>{};

    if (_agentIdController.text.isNotEmpty) {
      criteria['agent_id'] = _agentIdController.text;
    }
    if (_nameController.text.isNotEmpty) {
      criteria['full_name'] = _nameController.text;
    }
    if (_panController.text.isNotEmpty) {
      criteria['pan_number'] = _panController.text;
    }
    if (_selectedAgentType != null) {
      criteria['agent_type'] = _getAgentTypeString(_selectedAgentType!);
    }
    if (_selectedStatus != null) {
      criteria['status'] = _getStatusString(_selectedStatus!);
    }

    context.read<AgentProfileBloc>().add(
      SearchAgentsEvent(criteria: criteria),
    );

    setState(() {
      _isSearchExpanded = false;
    });
  }

  void _clearSearch() {
    _agentIdController.clear();
    _nameController.clear();
    _panController.clear();
    setState(() {
      _selectedAgentType = null;
      _selectedStatus = null;
      _isSearchExpanded = true;
    });
    context.read<AgentProfileBloc>().add(const ClearSearchResultsEvent());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Search Agents'),
        actions: [
          IconButton(
            icon: Icon(_isSearchExpanded ? Icons.expand_less : Icons.expand_more),
            onPressed: () {
              setState(() {
                _isSearchExpanded = !_isSearchExpanded;
              });
            },
            tooltip: _isSearchExpanded ? 'Collapse' : 'Expand',
          ),
        ],
      ),
      body: Column(
        children: [
          // Search form
          if (_isSearchExpanded) _buildSearchForm(),

          // Results
          Expanded(
            child: BlocBuilder<AgentProfileBloc, AgentProfileState>(
              builder: (context, state) {
                if (state is AgentProfileInitial) {
                  return const EmptyDisplay(
                    title: 'Search Agents',
                    message: 'Enter search criteria above to find agents',
                    icon: Icons.search,
                  );
                }

                if (state is AgentSearchLoading) {
                  return const LoadingIndicator(
                    message: 'Searching agents...',
                  );
                }

                if (state is AgentProfileError) {
                  return ErrorDisplay(
                    message: state.message,
                    onRetry: _performSearch,
                  );
                }

                if (state is AgentSearchEmpty) {
                  return EmptyDisplay(
                    title: 'No Results',
                    message: state.message,
                    icon: Icons.search_off,
                    onAction: _clearSearch,
                    actionText: 'Clear Search',
                  );
                }

                if (state is AgentSearchLoaded) {
                  return _buildResultsList(state.agents);
                }

                return const SizedBox.shrink();
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSearchForm() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Form(
        key: _formKey,
        child: Column(
          children: [
            // Agent ID
            CustomTextField(
              label: 'Agent ID',
              hint: 'e.g., AGT-2026-000001',
              controller: _agentIdController,
              keyboardType: TextInputType.text,
            ),
            const SizedBox(height: 16),

            // Name
            CustomTextField(
              label: 'Full Name',
              hint: 'Enter agent name',
              controller: _nameController,
              keyboardType: TextInputType.name,
            ),
            const SizedBox(height: 16),

            // PAN Number
            CustomTextField(
              label: 'PAN Number',
              hint: 'Enter PAN number',
              controller: _panController,
              keyboardType: TextInputType.text,
            ),
            const SizedBox(height: 16),

            // Agent Type
            CustomDropdown<AgentType>(
              label: 'Agent Type',
              hint: 'Select agent type',
              value: _selectedAgentType,
              items: AgentType.values.map((type) {
                return DropdownMenuItem(
                  value: type,
                  child: Text(_getAgentTypeText(type)),
                );
              }).toList(),
              onChanged: (value) {
                setState(() {
                  _selectedAgentType = value;
                });
              },
            ),
            const SizedBox(height: 16),

            // Status
            CustomDropdown<AgentStatus>(
              label: 'Status',
              hint: 'Select status',
              value: _selectedStatus,
              items: AgentStatus.values.map((status) {
                return DropdownMenuItem(
                  value: status,
                  child: Text(_getStatusText(status)),
                );
              }).toList(),
              onChanged: (value) {
                setState(() {
                  _selectedStatus = value;
                });
              },
            ),
            const SizedBox(height: 24),

            // Action buttons
            Row(
              children: [
                Expanded(
                  child: CustomButton(
                    text: 'Clear',
                    onPressed: _clearSearch,
                    type: ButtonType.outlined,
                    icon: Icons.clear,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: CustomButton(
                    text: 'Search',
                    onPressed: _performSearch,
                    icon: Icons.search,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildResultsList(List<AgentProfile> agents) {
    return Column(
      children: [
        // Results header
        Container(
          padding: const EdgeInsets.all(16),
          color: AppColors.backgroundSecondary,
          child: Row(
            children: [
              Text(
                '${agents.length} result${agents.length != 1 ? 's' : ''} found',
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
              ),
            ],
          ),
        ),

        // Results list
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: agents.length,
            itemBuilder: (context, index) {
              return AgentProfileCard(
                agent: agents[index],
                onTap: () {
                  // Navigate to profile detail
                  // TODO: Implement navigation
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('View profile: ${agents[index].fullName}'),
                    ),
                  );
                },
              );
            },
          ),
        ),
      ],
    );
  }

  String _getAgentTypeText(AgentType type) {
    switch (type) {
      case AgentType.advisor:
        return 'Advisor';
      case AgentType.advisorCoordinator:
        return 'Advisor Coordinator';
      case AgentType.departmentalEmployee:
        return 'Departmental Employee';
      case AgentType.fieldOfficer:
        return 'Field Officer';
      case AgentType.directAgent:
        return 'Direct Agent';
      case AgentType.gds:
        return 'GDS';
    }
  }

  String _getAgentTypeString(AgentType type) {
    switch (type) {
      case AgentType.advisor:
        return 'advisor';
      case AgentType.advisorCoordinator:
        return 'advisor_coordinator';
      case AgentType.departmentalEmployee:
        return 'departmental_employee';
      case AgentType.fieldOfficer:
        return 'field_officer';
      case AgentType.directAgent:
        return 'direct_agent';
      case AgentType.gds:
        return 'gds';
    }
  }

  String _getStatusText(AgentStatus status) {
    switch (status) {
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

  String _getStatusString(AgentStatus status) {
    switch (status) {
      case AgentStatus.active:
        return 'active';
      case AgentStatus.inactive:
        return 'inactive';
      case AgentStatus.suspended:
        return 'suspended';
      case AgentStatus.terminated:
        return 'terminated';
      case AgentStatus.deactivated:
        return 'deactivated';
    }
  }
}

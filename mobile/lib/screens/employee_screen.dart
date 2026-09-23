import 'package:flutter/material.dart';
import '../models/resource_models.dart';
import '../services/resource_api_client.dart';

class EmployeeScreen extends StatefulWidget {
  final String tenantId;
  final ResourceApiClient? apiClient;

  const EmployeeScreen({
    super.key,
    required this.tenantId,
    this.apiClient,
  });

  @override
  State<EmployeeScreen> createState() => _EmployeeScreenState();
}

class _EmployeeScreenState extends State<EmployeeScreen> {
  late final ResourceApiClient _apiClient;
  List<EmployeeModel> _employees = [];
  bool _isLoading = true;
  String? _errorMessage;
  String _selectedFilter = 'ALL'; // ALL, AVAILABLE_DRIVERS, DRIVER, OPERATOR, DISPATCHER

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ResourceApiClient();
    _fetchEmployees();
  }

  Future<void> _fetchEmployees() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      List<EmployeeModel> list;
      if (_selectedFilter == 'AVAILABLE_DRIVERS') {
        list = await _apiClient.getAvailableDrivers(widget.tenantId);
      } else if (_selectedFilter == 'ALL') {
        list = await _apiClient.getEmployees(widget.tenantId);
      } else {
        list = await _apiClient.getEmployees(
          widget.tenantId,
          operationalRole: _selectedFilter,
        );
      }

      setState(() {
        _employees = list;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Color _getAvailabilityColor(String status) {
    switch (status.toUpperCase()) {
      case 'AVAILABLE':
        return const Color(0xFF10B981);
      case 'BUSY':
        return const Color(0xFF3B82F6);
      case 'OFF_DUTY':
        return Colors.amber;
      case 'UNAVAILABLE':
        return const Color(0xFFF43F5E);
      default:
        return Colors.grey;
    }
  }

  Color _getRoleColor(String role) {
    switch (role.toUpperCase()) {
      case 'DRIVER':
      case 'DELIVERY_EXECUTIVE':
        return const Color(0xFF38BDF8);
      case 'OPERATOR':
      case 'WAREHOUSE_OPERATOR':
        return const Color(0xFFA855F7);
      case 'DISPATCHER':
        return const Color(0xFFF59E0B);
      case 'SUPERVISOR':
      case 'MANAGER':
      case 'BRANCH_MANAGER':
        return const Color(0xFF10B981);
      default:
        return const Color(0xFF94A3B8);
    }
  }

  Future<void> _showStatusDialog(EmployeeModel employee) async {
    String selectedStatus = employee.availabilityStatus;

    final updated = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              backgroundColor: const Color(0xFF1E293B),
              title: Text(
                'Update ${employee.fullName}',
                style: const TextStyle(color: Colors.white, fontSize: 16),
              ),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Operational Availability',
                    style: TextStyle(color: Colors.white70, fontSize: 13),
                  ),
                  const SizedBox(height: 12),
                  ...['AVAILABLE', 'BUSY', 'OFF_DUTY', 'UNAVAILABLE'].map((s) {
                    final color = _getAvailabilityColor(s);
                    return RadioListTile<String>(
                      value: s,
                      groupValue: selectedStatus,
                      activeColor: color,
                      title: Text(
                        s.replaceAll('_', ' '),
                        style: TextStyle(
                          color: selectedStatus == s ? color : Colors.white70,
                          fontWeight: selectedStatus == s ? FontWeight.bold : FontWeight.normal,
                          fontSize: 14,
                        ),
                      ),
                      onChanged: (val) {
                        if (val != null) {
                          setDialogState(() {
                            selectedStatus = val;
                          });
                        }
                      },
                    );
                  }),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(false),
                  child: const Text('Cancel', style: TextStyle(color: Colors.white54)),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF2563EB),
                    foregroundColor: Colors.white,
                  ),
                  onPressed: () async {
                    try {
                      await _apiClient.updateEmployeeStatus(
                        widget.tenantId,
                        employee.id,
                        availabilityStatus: selectedStatus,
                      );
                      if (ctx.mounted) Navigator.of(ctx).pop(true);
                    } catch (e) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(content: Text('Failed to update status: $e')),
                      );
                    }
                  },
                  child: const Text('Update'),
                ),
              ],
            );
          },
        );
      },
    );

    if (updated == true) {
      _fetchEmployees();
    }
  }

  Future<void> _showMyProfileSheet() async {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF0F172A),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) {
        return FutureBuilder<EmployeeMeModel>(
          future: _apiClient.getMyProfile(widget.tenantId),
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const SizedBox(
                height: 250,
                child: Center(
                  child: CircularProgressIndicator(color: Color(0xFF38BDF8)),
                ),
              );
            }
            if (snapshot.hasError) {
              return Container(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.error_outline, color: Color(0xFFF43F5E), size: 40),
                    const SizedBox(height: 12),
                    const Text(
                      'Unable to load your profile',
                      style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      snapshot.error.toString(),
                      style: const TextStyle(color: Colors.white70, fontSize: 12),
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
              );
            }

            final me = snapshot.data!;
            final emp = me.employee;
            final veh = me.assignedVehicle;

            return Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: const Color(0xFF38BDF8).withOpacity(0.15),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(Icons.badge_outlined, color: Color(0xFF38BDF8), size: 24),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              emp.fullName,
                              style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18),
                            ),
                            Text(
                              '${emp.employeeCode} • ${emp.designation}',
                              style: const TextStyle(color: Colors.white70, fontSize: 13),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFF1E293B),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Operational Role:', style: TextStyle(color: Colors.white54, fontSize: 13)),
                            Text(emp.operationalRole, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 13)),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('System Role:', style: TextStyle(color: Colors.white54, fontSize: 13)),
                            Text(me.systemRole, style: const TextStyle(color: Color(0xFF38BDF8), fontWeight: FontWeight.w600, fontSize: 13)),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Availability:', style: TextStyle(color: Colors.white54, fontSize: 13)),
                            Text(emp.availabilityStatus, style: TextStyle(color: _getAvailabilityColor(emp.availabilityStatus), fontWeight: FontWeight.w600, fontSize: 13)),
                          ],
                        ),
                      ],
                    ),
                  ),
                  if (veh != null) ...[
                    const SizedBox(height: 14),
                    const Text('Assigned Vehicle', style: TextStyle(color: Colors.white70, fontWeight: FontWeight.w600, fontSize: 13)),
                    const SizedBox(height: 6),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFF1E293B),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: const Color(0xFF38BDF8).withOpacity(0.3)),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.local_shipping_outlined, color: Color(0xFF38BDF8), size: 22),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(veh.registrationNumber, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14)),
                                Text('${veh.vehicleType}${veh.makeModel != null ? " • ${veh.makeModel}" : ""}', style: const TextStyle(color: Colors.white60, fontSize: 12)),
                              ],
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                            decoration: BoxDecoration(
                              color: const Color(0xFF10B981).withOpacity(0.15),
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: Text(veh.status, style: const TextStyle(color: Color(0xFF10B981), fontSize: 11, fontWeight: FontWeight.bold)),
                          ),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Staff & Drivers', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: const Color(0xFF0F172A),
        actions: [
          IconButton(
            icon: const Icon(Icons.account_circle_outlined),
            onPressed: _showMyProfileSheet,
            tooltip: 'My Staff Profile',
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchEmployees,
            tooltip: 'Refresh',
          ),
        ],
      ),
      backgroundColor: const Color(0xFF020617),
      body: Column(
        children: [
          // Filter Bar
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                _buildFilterChip('ALL', 'All Staff'),
                const SizedBox(width: 8),
                _buildFilterChip('AVAILABLE_DRIVERS', 'Available Drivers'),
                const SizedBox(width: 8),
                _buildFilterChip('DRIVER', 'Drivers'),
                const SizedBox(width: 8),
                _buildFilterChip('OPERATOR', 'Operators'),
                const SizedBox(width: 8),
                _buildFilterChip('DISPATCHER', 'Dispatchers'),
              ],
            ),
          ),
          // Content
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator(color: Color(0xFF38BDF8)))
                : _errorMessage != null
                    ? Center(
                        child: Padding(
                          padding: const EdgeInsets.all(24.0),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.error_outline, color: Color(0xFFF43F5E), size: 48),
                              const SizedBox(height: 16),
                              const Text(
                                'Unable to load employees',
                                style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                              ),
                              const SizedBox(height: 8),
                              Text(
                                _errorMessage!,
                                textAlign: TextAlign.center,
                                style: const TextStyle(color: Colors.white70, fontSize: 13),
                              ),
                              const SizedBox(height: 20),
                              ElevatedButton.icon(
                                onPressed: _fetchEmployees,
                                icon: const Icon(Icons.refresh),
                                label: const Text('Try Again'),
                              ),
                            ],
                          ),
                        ),
                      )
                    : _employees.isEmpty
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(Icons.people_outline, color: Colors.white30, size: 56),
                                const SizedBox(height: 16),
                                Text(
                                  _selectedFilter == 'AVAILABLE_DRIVERS'
                                      ? 'No available drivers found'
                                      : 'No employees found',
                                  style: const TextStyle(color: Colors.white70, fontSize: 16),
                                ),
                              ],
                            ),
                          )
                        : RefreshIndicator(
                            onRefresh: _fetchEmployees,
                            color: const Color(0xFF38BDF8),
                            child: ListView.builder(
                              padding: const EdgeInsets.all(16.0),
                              itemCount: _employees.length,
                              itemBuilder: (context, index) {
                                final emp = _employees[index];
                                final availColor = _getAvailabilityColor(emp.availabilityStatus);
                                final roleColor = _getRoleColor(emp.operationalRole);

                                return Card(
                                  color: const Color(0xFF1E293B),
                                  margin: const EdgeInsets.only(bottom: 12),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                                  child: Padding(
                                    padding: const EdgeInsets.all(16.0),
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                          children: [
                                            Expanded(
                                              child: Column(
                                                crossAxisAlignment: CrossAxisAlignment.start,
                                                children: [
                                                  Row(
                                                    children: [
                                                      Text(
                                                        emp.fullName,
                                                        style: const TextStyle(
                                                          fontSize: 16,
                                                          fontWeight: FontWeight.bold,
                                                          color: Colors.white,
                                                        ),
                                                      ),
                                                      const SizedBox(width: 8),
                                                      if (emp.verificationStatus == 'VERIFIED')
                                                        const Icon(
                                                          Icons.verified,
                                                          size: 16,
                                                          color: Color(0xFF10B981),
                                                        ),
                                                    ],
                                                  ),
                                                  const SizedBox(height: 4),
                                                  Text(
                                                    emp.employeeCode,
                                                    style: const TextStyle(
                                                      color: Color(0xFF38BDF8),
                                                      fontSize: 12,
                                                      fontWeight: FontWeight.bold,
                                                      fontFamily: 'monospace',
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                            GestureDetector(
                                              onTap: () => _showStatusDialog(emp),
                                              child: Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                                decoration: BoxDecoration(
                                                  color: availColor.withOpacity(0.15),
                                                  borderRadius: BorderRadius.circular(12),
                                                  border: Border.all(color: availColor),
                                                ),
                                                child: Row(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    Container(
                                                      width: 6,
                                                      height: 6,
                                                      decoration: BoxDecoration(
                                                        color: availColor,
                                                        shape: BoxShape.circle,
                                                      ),
                                                    ),
                                                    const SizedBox(width: 6),
                                                    Text(
                                                      emp.availabilityStatus.replaceAll('_', ' '),
                                                      style: TextStyle(
                                                        color: availColor,
                                                        fontSize: 11,
                                                        fontWeight: FontWeight.bold,
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                        const SizedBox(height: 12),
                                        Row(
                                          children: [
                                            Container(
                                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                              decoration: BoxDecoration(
                                                color: roleColor.withOpacity(0.15),
                                                borderRadius: BorderRadius.circular(6),
                                                border: Border.all(color: roleColor.withOpacity(0.3)),
                                              ),
                                              child: Text(
                                                emp.operationalRole,
                                                style: TextStyle(
                                                  color: roleColor,
                                                  fontSize: 11,
                                                  fontWeight: FontWeight.bold,
                                                ),
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            if (emp.branchName != null) ...[
                                              const Icon(Icons.hub_outlined, size: 14, color: Colors.white54),
                                              const SizedBox(width: 4),
                                              Text(
                                                emp.branchName!,
                                                style: const TextStyle(color: Colors.white70, fontSize: 12),
                                              ),
                                            ],
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                );
                              },
                            ),
                          ),
          ),
        ],
      ),
    );
  }

  Widget _buildFilterChip(String value, String label) {
    final isSelected = _selectedFilter == value;
    return ChoiceChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          setState(() {
            _selectedFilter = value;
          });
          _fetchEmployees();
        }
      },
      selectedColor: const Color(0xFF2563EB),
      backgroundColor: const Color(0xFF1E293B),
      labelStyle: TextStyle(
        color: isSelected ? Colors.white : Colors.white70,
        fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
        fontSize: 12,
      ),
    );
  }
}

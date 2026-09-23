import 'package:flutter/material.dart';
import '../models/resource_models.dart';
import '../services/resource_api_client.dart';

class VehicleScreen extends StatefulWidget {
  final String tenantId;
  final ResourceApiClient? apiClient;

  const VehicleScreen({
    super.key,
    required this.tenantId,
    this.apiClient,
  });

  @override
  State<VehicleScreen> createState() => _VehicleScreenState();
}

class _VehicleScreenState extends State<VehicleScreen> {
  late final ResourceApiClient _apiClient;
  List<VehicleModel> _vehicles = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ResourceApiClient();
    _fetchVehicles();
  }

  Future<void> _fetchVehicles() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final vehicles = await _apiClient.getVehicles(widget.tenantId);
      setState(() {
        _vehicles = vehicles;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'AVAILABLE':
        return const Color(0xFF10B981);
      case 'ASSIGNED':
        return const Color(0xFF3B82F6);
      case 'IN_TRANSIT':
        return Colors.amber;
      case 'MAINTENANCE':
        return const Color(0xFFF43F5E);
      case 'DECOMMISSIONED':
        return const Color(0xFF6B7280);
      default:
        return Colors.grey;
    }
  }

  Color _getAvailabilityColor(String status) {
    switch (status) {
      case 'AVAILABLE':
        return const Color(0xFF10B981);
      case 'BUSY':
        return const Color(0xFF3B82F6);
      case 'MAINTENANCE':
        return const Color(0xFFF59E0B);
      case 'OUT_OF_SERVICE':
        return const Color(0xFFF43F5E);
      default:
        return Colors.grey;
    }
  }

  Future<void> _handleUnassign(VehicleModel vehicle) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF1E293B),
        title: const Text('Unassign Driver', style: TextStyle(color: Colors.white)),
        content: Text(
          'Are you sure you want to unassign driver from ${vehicle.registrationNumber}? Both will be marked AVAILABLE.',
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancel', style: TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFEF4444),
              foregroundColor: Colors.white,
            ),
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Unassign'),
          ),
        ],
      ),
    );

    if (confirm == true) {
      try {
        await _apiClient.unassignVehicle(widget.tenantId, vehicle.id);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Driver unassigned successfully')),
        );
        _fetchVehicles();
      } catch (e) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Unassignment failed: $e')),
        );
      }
    }
  }

  Future<void> _handleAssign(VehicleModel vehicle) async {
    try {
      final availableDrivers = await _apiClient.getAvailableDrivers(
        widget.tenantId,
        branchId: vehicle.assignedBranchId,
      );

      if (!mounted) return;

      if (availableDrivers.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('No available drivers found in this branch')),
        );
        return;
      }

      final selectedDriver = await showModalBottomSheet<EmployeeModel>(
        context: context,
        backgroundColor: const Color(0xFF1E293B),
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
        ),
        builder: (ctx) {
          return SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Assign Driver to ${vehicle.registrationNumber}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: ListView.builder(
                      itemCount: availableDrivers.length,
                      itemBuilder: (context, idx) {
                        final driver = availableDrivers[idx];
                        return ListTile(
                          leading: const CircleAvatar(
                            backgroundColor: Color(0xFF0F172A),
                            child: Icon(Icons.person, color: Color(0xFF38BDF8)),
                          ),
                          title: Text(
                            driver.fullName,
                            style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                          ),
                          subtitle: Text(
                            '${driver.employeeCode} • ${driver.branchName ?? 'Branch'}',
                            style: const TextStyle(color: Colors.white70),
                          ),
                          trailing: const Icon(Icons.chevron_right, color: Colors.white54),
                          onTap: () => Navigator.of(ctx).pop(driver),
                        );
                      },
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      );

      if (selectedDriver != null) {
        await _apiClient.assignVehicle(widget.tenantId, vehicle.id, selectedDriver.id);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Assigned ${selectedDriver.fullName} to ${vehicle.registrationNumber}')),
        );
        _fetchVehicles();
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Assignment failed: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Fleet Vehicles', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: const Color(0xFF0F172A),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchVehicles,
            tooltip: 'Refresh',
          ),
        ],
      ),
      backgroundColor: const Color(0xFF020617),
      body: _isLoading
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
                          'Unable to load vehicles',
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
                          onPressed: _fetchVehicles,
                          icon: const Icon(Icons.refresh),
                          label: const Text('Try Again'),
                        ),
                      ],
                    ),
                  ),
                )
              : _vehicles.isEmpty
                  ? const Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.local_shipping, color: Colors.white30, size: 56),
                          SizedBox(height: 16),
                          Text(
                            'No vehicles registered yet',
                            style: TextStyle(color: Colors.white70, fontSize: 16),
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _fetchVehicles,
                      color: const Color(0xFF38BDF8),
                      child: ListView.builder(
                        padding: const EdgeInsets.all(16.0),
                        itemCount: _vehicles.length,
                        itemBuilder: (context, index) {
                          final v = _vehicles[index];
                          final statusColor = _getStatusColor(v.status);
                          final availColor = _getAvailabilityColor(v.availabilityStatus);

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
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFF38BDF8).withOpacity(0.15),
                                          borderRadius: BorderRadius.circular(6),
                                          border: Border.all(color: const Color(0xFF38BDF8).withOpacity(0.3)),
                                        ),
                                        child: Text(
                                          v.registrationNumber,
                                          style: const TextStyle(
                                            color: Color(0xFF38BDF8),
                                            fontSize: 13,
                                            fontWeight: FontWeight.bold,
                                          ),
                                        ),
                                      ),
                                      Row(
                                        children: [
                                          Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                            decoration: BoxDecoration(
                                              color: availColor.withOpacity(0.15),
                                              borderRadius: BorderRadius.circular(12),
                                              border: Border.all(color: availColor),
                                            ),
                                            child: Text(
                                              v.availabilityStatus.replaceAll('_', ' '),
                                              style: TextStyle(
                                                color: availColor,
                                                fontSize: 11,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ),
                                          const SizedBox(width: 6),
                                          Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                            decoration: BoxDecoration(
                                              color: statusColor.withOpacity(0.2),
                                              borderRadius: BorderRadius.circular(12),
                                              border: Border.all(color: statusColor),
                                            ),
                                            child: Text(
                                              v.status,
                                              style: TextStyle(
                                                color: statusColor,
                                                fontSize: 11,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    v.makeModel ?? v.vehicleType.replaceAll('_', ' '),
                                    style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                                  ),
                                  if (v.isElectric) ...[
                                    const SizedBox(height: 4),
                                    const Row(
                                      children: [
                                        Icon(Icons.bolt, size: 14, color: Color(0xFF38BDF8)),
                                        SizedBox(width: 4),
                                        Text(
                                          'Electric Zero-Emission Vehicle',
                                          style: TextStyle(color: Color(0xFF38BDF8), fontSize: 12),
                                        ),
                                      ],
                                    ),
                                  ],
                                  const SizedBox(height: 12),
                                  Container(
                                    padding: const EdgeInsets.all(10),
                                    decoration: BoxDecoration(
                                      color: Colors.black26,
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                                      children: [
                                        Column(
                                          children: [
                                            const Text('Max Payload', style: TextStyle(color: Colors.white54, fontSize: 11)),
                                            const SizedBox(height: 2),
                                            Text('${v.maxWeightKg} kg', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
                                          ],
                                        ),
                                        Column(
                                          children: [
                                            const Text('Volume', style: TextStyle(color: Colors.white54, fontSize: 11)),
                                            const SizedBox(height: 2),
                                            Text('${v.maxVolumeCbm} m³', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
                                          ],
                                        ),
                                        Column(
                                          children: [
                                            const Text('Active Driver', style: TextStyle(color: Colors.white54, fontSize: 11)),
                                            const SizedBox(height: 2),
                                            Text(
                                              v.currentDriverName ?? 'Unassigned',
                                              style: TextStyle(
                                                color: v.currentDriverName != null ? const Color(0xFF60A5FA) : Colors.white38,
                                                fontWeight: FontWeight.bold,
                                                fontSize: 13,
                                              ),
                                            ),
                                            if (v.currentDriverCode != null)
                                              Text(
                                                v.currentDriverCode!,
                                                style: const TextStyle(color: Colors.white54, fontSize: 10, fontFamily: 'monospace'),
                                              ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  Row(
                                    mainAxisAlignment: MainAxisAlignment.end,
                                    children: [
                                      if (v.isAssigned)
                                        OutlinedButton.icon(
                                          icon: const Icon(Icons.person_remove, size: 16),
                                          label: const Text('Unassign'),
                                          style: OutlinedButton.styleFrom(
                                            foregroundColor: const Color(0xFFF43F5E),
                                            side: const BorderSide(color: Color(0xFFF43F5E)),
                                          ),
                                          onPressed: () => _handleUnassign(v),
                                        )
                                      else if (v.isAvailable)
                                        ElevatedButton.icon(
                                          icon: const Icon(Icons.person_add, size: 16),
                                          label: const Text('Assign Driver'),
                                          style: ElevatedButton.styleFrom(
                                            backgroundColor: const Color(0xFF2563EB),
                                            foregroundColor: Colors.white,
                                          ),
                                          onPressed: () => _handleAssign(v),
                                        ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
    );
  }
}

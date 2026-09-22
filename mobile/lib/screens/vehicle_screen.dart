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
      default:
        return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Fleet Vehicles'),
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
                  : ListView.builder(
                      padding: const EdgeInsets.all(16.0),
                      itemCount: _vehicles.length,
                      itemBuilder: (context, index) {
                        final v = _vehicles[index];
                        final statusColor = _getStatusColor(v.status);

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
                                const SizedBox(height: 8),
                                Text(
                                  v.makeModel ?? v.vehicleType,
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
                                            v.currentDriverName ?? 'None',
                                            style: TextStyle(
                                              color: v.currentDriverName != null ? const Color(0xFF60A5FA) : Colors.white38,
                                              fontWeight: FontWeight.bold,
                                              fontSize: 13,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
    );
  }
}

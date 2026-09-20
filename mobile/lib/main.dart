import 'package:flutter/material.dart';

void main() {
  runApp(const LogiFlowsApp());
}

class LogiFlowsApp extends StatelessWidget {
  const LogiFlowsApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'LogiFlows Driver & Custody',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF3B82F6),
          brightness: Brightness.dark,
          surface: const Color(0xFF0F172A),
        ),
        scaffoldBackgroundColor: const Color(0xFF020617),
      ),
      home: const DriverDashboardScreen(),
    );
  }
}

class DriverDashboardScreen extends StatefulWidget {
  const DriverDashboardScreen({super.key});

  @override
  State<DriverDashboardScreen> createState() => _DriverDashboardScreenState();
}

class _DriverDashboardScreenState extends State<DriverDashboardScreen> {
  bool _isConnecting = false;
  String _apiStatus = 'Ready to Connect';
  Color _statusColor = Colors.grey;

  void _checkBackendStatus() async {
    setState(() {
      _isConnecting = true;
      _apiStatus = 'Checking Core Backend...';
      _statusColor = Colors.amber;
    });

    // Simulated check for initial scaffold
    await Future.delayed(const Duration(milliseconds: 800));

    setState(() {
      _isConnecting = false;
      _apiStatus = 'Backend Operational (8080)';
      _statusColor = Colors.emerald;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Row(
          children: [
            Icon(Icons.local_shipping, color: Color(0xFF38BDF8)),
            SizedBox(width: 10),
            Text('LogiFlows Driver Pro', style: TextStyle(fontWeight: FontWeight.bold)),
          ],
        ),
        backgroundColor: const Color(0xFF0F172A),
        elevation: 0,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Card(
              color: const Color(0xFF1E293B),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'System Connectivity',
                          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Colors.white),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: _statusColor.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.Border.all(color: _statusColor),
                          ),
                          child: Text(
                            _apiStatus,
                            style: TextStyle(color: _statusColor, fontSize: 12, fontWeight: FontWeight.bold),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton.icon(
                      onPressed: _isConnecting ? null : _checkBackendStatus,
                      icon: const Icon(Icons.sync),
                      label: const Text('Probe LogiFlows Network'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF2563EB),
                        foregroundColor: Colors.white,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Active Custody Assignments',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
            ),
            const SizedBox(height: 10),
            Expanded(
              child: ListView(
                children: const [
                  _ParcelCard(
                    trackingNumber: 'TRK-2026-9041',
                    destination: 'North Hub -> Central Branch',
                    status: 'In Transit',
                    eta: '14:30 IST',
                  ),
                  _ParcelCard(
                    trackingNumber: 'TRK-2026-9042',
                    destination: 'Central Branch -> Airport Station',
                    status: 'Out for Delivery',
                    eta: '15:15 IST',
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ParcelCard extends StatelessWidget {
  final String trackingNumber;
  final String destination;
  final String status;
  final String eta;

  const _ParcelCard({
    required this.trackingNumber,
    required this.destination,
    required this.status,
    required this.eta,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      color: const Color(0xFF1E293B),
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: const CircleAvatar(
          backgroundColor: Color(0xFF0F172A),
          child: Icon(Icons.qr_code_scanner, color: Color(0xFF38BDF8)),
        ),
        title: Text(trackingNumber, style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white)),
        subtitle: Text(destination, style: const TextStyle(color: Colors.white70)),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(status, style: const TextStyle(color: Colors.emeraldAccent, fontWeight: FontWeight.bold)),
            Text('ETA: $eta', style: const TextStyle(color: Colors.white54, fontSize: 11)),
          ],
        ),
      ),
    );
  }
}

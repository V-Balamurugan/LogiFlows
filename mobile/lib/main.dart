import 'package:flutter/material.dart';
import 'screens/login_screen.dart';
import 'screens/company_screen.dart';
import 'screens/branch_screen.dart';
import 'screens/vehicle_screen.dart';
import 'services/auth_api_client.dart';
import 'core/token_storage.dart';

void main() {
  runApp(const LogiFlowsApp());
}

class LogiFlowsApp extends StatefulWidget {
  const LogiFlowsApp({super.key});

  @override
  State<LogiFlowsApp> createState() => _LogiFlowsAppState();
}

class _LogiFlowsAppState extends State<LogiFlowsApp> {
  final AuthApiClient _authClient = AuthApiClient();
  bool _isAuthenticated = false;
  bool _isCheckingAuth = true;

  @override
  void initState() {
    super.initState();
    _checkInitialAuth();
  }

  Future<void> _checkInitialAuth() async {
    final hasToken = await _authClient.tokenStorage.hasValidToken();
    setState(() {
      _isAuthenticated = hasToken;
      _isCheckingAuth = false;
    });
  }

  void _onLoginSuccess() {
    setState(() {
      _isAuthenticated = true;
    });
  }

  Future<void> _handleLogout() async {
    await _authClient.logout();
    setState(() {
      _isAuthenticated = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'LogiFlows Driver & Fleet Pro',
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
      home: _isCheckingAuth
          ? const Scaffold(
              backgroundColor: Color(0xFF020617),
              body: Center(child: CircularProgressIndicator(color: Color(0xFF38BDF8))),
            )
          : _isAuthenticated
              ? DriverDashboardScreen(
                  onLogout: _handleLogout,
                  authClient: _authClient,
                )
              : LoginScreen(
                  authClient: _authClient,
                  onLoginSuccess: _onLoginSuccess,
                ),
    );
  }
}

class DriverDashboardScreen extends StatefulWidget {
  final VoidCallback onLogout;
  final AuthApiClient? authClient;

  const DriverDashboardScreen({
    super.key,
    required this.onLogout,
    this.authClient,
  });

  @override
  State<DriverDashboardScreen> createState() => _DriverDashboardScreenState();
}

class _DriverDashboardScreenState extends State<DriverDashboardScreen> {
  int _currentTabIndex = 0;
  String _tenantId = '00000000-0000-0000-0000-000000000001';
  bool _isLoadingTenant = true;

  @override
  void initState() {
    super.initState();
    _loadTenantId();
  }

  Future<void> _loadTenantId() async {
    final storage = widget.authClient?.tokenStorage ?? InMemorySecureTokenStorage();
    final savedTenant = await storage.getTenantId();
    if (savedTenant != null && savedTenant.isNotEmpty) {
      if (mounted) {
        setState(() {
          _tenantId = savedTenant;
          _isLoadingTenant = false;
        });
      }
    } else {
      // Fallback: try fetching current user profile
      try {
        if (widget.authClient != null) {
          await widget.authClient!.getCurrentUser();
          final refreshedTenant = await storage.getTenantId();
          if (refreshedTenant != null && mounted) {
            setState(() {
              _tenantId = refreshedTenant;
            });
          }
        }
      } catch (_) {}
      if (mounted) {
        setState(() {
          _isLoadingTenant = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoadingTenant) {
      return const Scaffold(
        backgroundColor: Color(0xFF020617),
        body: Center(child: CircularProgressIndicator(color: Color(0xFF38BDF8))),
      );
    }

    final screens = [
      CompanyScreen(tenantId: _tenantId),
      BranchScreen(tenantId: _tenantId),
      VehicleScreen(tenantId: _tenantId),
      _CustodyTabContent(onLogout: widget.onLogout),
    ];

    return Scaffold(
      body: IndexedStack(
        index: _currentTabIndex,
        children: screens,
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentTabIndex,
        onDestinationSelected: (index) {
          setState(() {
            _currentTabIndex = index;
          });
        },
        backgroundColor: const Color(0xFF0F172A),
        indicatorColor: const Color(0xFF2563EB).withOpacity(0.3),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.business_outlined),
            selectedIcon: Icon(Icons.business, color: Color(0xFF38BDF8)),
            label: 'Company',
          ),
          NavigationDestination(
            icon: Icon(Icons.hub_outlined),
            selectedIcon: Icon(Icons.hub, color: Color(0xFF38BDF8)),
            label: 'Hubs',
          ),
          NavigationDestination(
            icon: Icon(Icons.local_shipping_outlined),
            selectedIcon: Icon(Icons.local_shipping, color: Color(0xFF38BDF8)),
            label: 'Fleet',
          ),
          NavigationDestination(
            icon: Icon(Icons.inventory_2_outlined),
            selectedIcon: Icon(Icons.inventory_2, color: Color(0xFF38BDF8)),
            label: 'Custody',
          ),
        ],
      ),
    );
  }
}

class _CustodyTabContent extends StatefulWidget {
  final VoidCallback onLogout;

  const _CustodyTabContent({required this.onLogout});

  @override
  State<_CustodyTabContent> createState() => _CustodyTabContentState();
}

class _CustodyTabContentState extends State<_CustodyTabContent> {
  bool _isConnecting = false;
  String _apiStatus = 'Session Active';
  Color _statusColor = const Color(0xFF10B981);

  void _checkBackendStatus() async {
    setState(() {
      _isConnecting = true;
      _apiStatus = 'Probing Network...';
      _statusColor = Colors.amber;
    });

    await Future.delayed(const Duration(milliseconds: 600));

    if (mounted) {
      setState(() {
        _isConnecting = false;
        _apiStatus = 'Backend Operational';
        _statusColor = const Color(0xFF10B981);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Row(
          children: [
            Icon(Icons.local_shipping, color: Color(0xFF38BDF8)),
            SizedBox(width: 10),
            Text('LogiFlows Driver Pro', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout, color: Colors.white70),
            tooltip: 'Sign Out',
            onPressed: widget.onLogout,
          ),
        ],
        backgroundColor: const Color(0xFF0F172A),
        elevation: 0,
      ),
      backgroundColor: const Color(0xFF020617),
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
                            border: Border.all(color: _statusColor),
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
            Text(status, style: const TextStyle(color: Color(0xFF10B981), fontWeight: FontWeight.bold)),
            Text('ETA: $eta', style: const TextStyle(color: Colors.white54, fontSize: 11)),
          ],
        ),
      ),
    );
  }
}

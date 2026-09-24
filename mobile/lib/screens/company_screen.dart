import 'package:flutter/material.dart';
import '../models/resource_models.dart';
import '../services/resource_api_client.dart';

class CompanyScreen extends StatefulWidget {
  final String tenantId;
  final ResourceApiClient? apiClient;

  const CompanyScreen({
    super.key,
    required this.tenantId,
    this.apiClient,
  });

  @override
  State<CompanyScreen> createState() => _CompanyScreenState();
}

class _CompanyScreenState extends State<CompanyScreen> {
  late final ResourceApiClient _apiClient;
  CompanyModel? _company;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ResourceApiClient();
    _fetchCompany();
  }

  Future<void> _fetchCompany() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final company = await _apiClient.getCompanyDetails(widget.tenantId);
      if (mounted) {
        setState(() {
          _company = company;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Row(
          children: [
            Icon(Icons.business, color: Color(0xFF38BDF8), size: 22),
            SizedBox(width: 8),
            Text('Company Profile', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
          ],
        ),
        backgroundColor: const Color(0xFF0F172A),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchCompany,
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
                          'Unable to load organization details',
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
                          onPressed: _fetchCompany,
                          icon: const Icon(Icons.refresh),
                          label: const Text('Try Again'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF2563EB),
                            foregroundColor: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ),
                )
              : _company == null
                  ? const Center(
                      child: Text(
                        'No company profile available',
                        style: TextStyle(color: Colors.white70),
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _fetchCompany,
                      color: const Color(0xFF38BDF8),
                      child: ListView(
                        padding: const EdgeInsets.all(16.0),
                        children: [
                          // Company Header Card
                          Card(
                            color: const Color(0xFF1E293B),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                            child: Padding(
                              padding: const EdgeInsets.all(20.0),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                    children: [
                                      Expanded(
                                        child: Text(
                                          _company!.name,
                                          style: const TextStyle(
                                            fontSize: 20,
                                            fontWeight: FontWeight.bold,
                                            color: Colors.white,
                                          ),
                                        ),
                                      ),
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                        decoration: BoxDecoration(
                                          color: _company!.status == 'ACTIVE'
                                              ? const Color(0xFF10B981).withOpacity(0.2)
                                              : Colors.grey.withOpacity(0.2),
                                          borderRadius: BorderRadius.circular(20),
                                          border: Border.all(
                                            color: _company!.status == 'ACTIVE'
                                                ? const Color(0xFF10B981)
                                                : Colors.grey,
                                          ),
                                        ),
                                        child: Text(
                                          _company!.status,
                                          style: TextStyle(
                                            color: _company!.status == 'ACTIVE'
                                                ? const Color(0xFF34D399)
                                                : Colors.grey,
                                            fontSize: 12,
                                            fontWeight: FontWeight.bold,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  Row(
                                    children: [
                                      const Icon(Icons.tag, size: 16, color: Color(0xFF38BDF8)),
                                      const SizedBox(width: 4),
                                      Text(
                                        _company!.slug,
                                        style: const TextStyle(
                                          fontFamily: 'monospace',
                                          color: Color(0xFF38BDF8),
                                          fontWeight: FontWeight.w600,
                                          fontSize: 14,
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(height: 16),

                          // Organization Details Card
                          Card(
                            color: const Color(0xFF1E293B),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                            child: Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Text(
                                    'Enterprise Identifiers',
                                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white70),
                                  ),
                                  const Divider(color: Colors.white12, height: 20),
                                  _InfoRow(
                                    icon: Icons.fingerprint,
                                    label: 'Company UUID',
                                    value: _company!.id,
                                    isMono: true,
                                  ),
                                  const SizedBox(height: 12),
                                  _InfoRow(
                                    icon: Icons.email_outlined,
                                    label: 'Contact Email',
                                    value: _company!.contactEmail ?? 'No registered email',
                                  ),
                                  if (_company!.createdAt != null) ...[
                                    const SizedBox(height: 12),
                                    _InfoRow(
                                      icon: Icons.calendar_today_outlined,
                                      label: 'Registration Date',
                                      value: _company!.createdAt!.split('T').first,
                                    ),
                                  ],
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(height: 16),

                          // Security and Isolation Banner
                          Container(
                            padding: const EdgeInsets.all(16.0),
                            decoration: BoxDecoration(
                              color: const Color(0xFF0F172A),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: const Color(0xFF1E293B)),
                            ),
                            child: const Row(
                              children: [
                                Icon(Icons.shield, color: Color(0xFF10B981), size: 28),
                                SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        'Multi-Tenant Data Isolation',
                                        style: TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 13),
                                      ),
                                      SizedBox(height: 2),
                                      Text(
                                        'All branch hubs and delivery assets are strictly isolated to this organization.',
                                        style: TextStyle(color: Colors.white54, fontSize: 11),
                                      ),
                                    ],
                                  ),
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

class _InfoRow extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final bool isMono;

  const _InfoRow({
    required this.icon,
    required this.label,
    required this.value,
    this.isMono = false,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 16, color: Colors.white54),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: const TextStyle(fontSize: 11, color: Colors.white38)),
              const SizedBox(height: 2),
              Text(
                value,
                style: TextStyle(
                  fontSize: 13,
                  color: Colors.white,
                  fontFamily: isMono ? 'monospace' : null,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

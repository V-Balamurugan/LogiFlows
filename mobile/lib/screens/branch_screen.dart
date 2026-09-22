import 'package:flutter/material.dart';
import '../models/resource_models.dart';
import '../services/resource_api_client.dart';

class BranchScreen extends StatefulWidget {
  final String tenantId;
  final ResourceApiClient? apiClient;

  const BranchScreen({
    super.key,
    required this.tenantId,
    this.apiClient,
  });

  @override
  State<BranchScreen> createState() => _BranchScreenState();
}

class _BranchScreenState extends State<BranchScreen> {
  late final ResourceApiClient _apiClient;
  List<BranchModel> _branches = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ResourceApiClient();
    _fetchBranches();
  }

  Future<void> _fetchBranches() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final branches = await _apiClient.getBranches(widget.tenantId);
      setState(() {
        _branches = branches;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hubs & Branches'),
        backgroundColor: const Color(0xFF0F172A),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchBranches,
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
                        Text(
                          'Unable to load branches',
                          style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _errorMessage!,
                          textAlign: TextAlign.center,
                          style: const TextStyle(color: Colors.white70, fontSize: 13),
                        ),
                        const SizedBox(height: 20),
                        ElevatedButton.icon(
                          onPressed: _fetchBranches,
                          icon: const Icon(Icons.refresh),
                          label: const Text('Try Again'),
                        ),
                      ],
                    ),
                  ),
                )
              : _branches.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.location_city, color: Colors.white30, size: 56),
                          const SizedBox(height: 16),
                          const Text(
                            'No branches registered yet',
                            style: TextStyle(color: Colors.white70, fontSize: 16),
                          ),
                        ],
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(16.0),
                      itemCount: _branches.length,
                      itemBuilder: (context, index) {
                        final b = _branches[index];
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
                                        color: const Color(0xFF0284C7).withOpacity(0.2),
                                        borderRadius: BorderRadius.circular(6),
                                        border: Border.all(color: const Color(0xFF0284C7)),
                                      ),
                                      child: Text(
                                        b.branchCode,
                                        style: const TextStyle(
                                          color: Color(0xFF38BDF8),
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                      decoration: BoxDecoration(
                                        color: b.isActive
                                            ? const Color(0xFF10B981).withOpacity(0.2)
                                            : Colors.grey.withOpacity(0.2),
                                        borderRadius: BorderRadius.circular(12),
                                      ),
                                      child: Text(
                                        b.status,
                                        style: TextStyle(
                                          color: b.isActive ? const Color(0xFF34D399) : Colors.grey,
                                          fontSize: 11,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 10),
                                Text(
                                  b.name,
                                  style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  '${b.address}, ${b.city}',
                                  style: const TextStyle(fontSize: 13, color: Colors.white70),
                                ),
                                const SizedBox(height: 10),
                                Row(
                                  children: [
                                    const Icon(Icons.explore, size: 14, color: Color(0xFF38BDF8)),
                                    const SizedBox(width: 4),
                                    Text(
                                      '${b.latitude.toStringAsFixed(4)}, ${b.longitude.toStringAsFixed(4)}',
                                      style: const TextStyle(fontSize: 12, color: Colors.white54, fontFamily: 'monospace'),
                                    ),
                                    const Spacer(),
                                    const Icon(Icons.radar, size: 14, color: Colors.amber),
                                    const SizedBox(width: 4),
                                    Text(
                                      '${b.coverageRadiusKm} km radius',
                                      style: const TextStyle(fontSize: 12, color: Colors.white54),
                                    ),
                                  ],
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

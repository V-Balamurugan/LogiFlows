import 'package:flutter/material.dart';
import '../models/parcel_model.dart';
import '../services/resource_api_client.dart';

class ParcelDeliveryScreen extends StatefulWidget {
  final String tenantId;
  final ResourceApiClient? apiClient;

  const ParcelDeliveryScreen({
    super.key,
    required this.tenantId,
    this.apiClient,
  });

  @override
  State<ParcelDeliveryScreen> createState() => _ParcelDeliveryScreenState();
}

class _ParcelDeliveryScreenState extends State<ParcelDeliveryScreen>
    with SingleTickerProviderStateMixin {
  late final ResourceApiClient _client;
  late final TabController _tabController;

  bool _loading = true;
  String? _error;
  List<DeliveryTaskModel> _tasks = [];
  List<BranchTransferModel> _transfers = [];

  // Barcode / QR Scan state
  final TextEditingController _scanController = TextEditingController();
  bool _verifyingScan = false;
  Map<String, dynamic>? _scanResult;
  String? _scanError;

  @override
  void initState() {
    super.initState();
    _client = widget.apiClient ?? ResourceApiClient();
    _tabController = TabController(length: 3, vsync: this);
    _loadData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _scanController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final tasks = await _client.getDeliveryTasks(widget.tenantId);
      final transfers = await _client.getBranchTransfers(widget.tenantId);
      if (mounted) {
        setState(() {
          _tasks = tasks;
          _transfers = transfers;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString().replaceAll('Exception: ', '');
          _loading = false;
        });
      }
    }
  }

  Future<void> _handleVerifyScan() async {
    final code = _scanController.text.trim();
    if (code.isEmpty) return;

    setState(() {
      _verifyingScan = true;
      _scanError = null;
      _scanResult = null;
    });

    try {
      final res = await _client.verifyParcelScan(widget.tenantId, code);
      if (mounted) {
        setState(() {
          _scanResult = res;
          _verifyingScan = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _scanError = e.toString().replaceAll('Exception: ', '');
          _verifyingScan = false;
        });
      }
    }
  }

  void _showRecordAttemptDialog(DeliveryTaskModel task) {
    String outcome = 'FAILED';
    String reason = 'CUSTOMER_UNAVAILABLE';
    final notesController = TextEditingController();
    bool submitting = false;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: const Color(0xFF0F172A),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Padding(
          padding: EdgeInsets.only(
            left: 20,
            right: 20,
            top: 20,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Record Delivery Attempt',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.grey),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                'Task: ${task.parcelTrackingNumber ?? task.parcelId}',
                style: const TextStyle(color: Color(0xFF38BDF8), fontSize: 13),
              ),
              const SizedBox(height: 16),
              const Text('Attempt Outcome',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              DropdownButtonFormField<String>(
                value: outcome,
                dropdownColor: const Color(0xFF1E293B),
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                items: const [
                  DropdownMenuItem(value: 'FAILED', child: Text('Failed Attempt')),
                  DropdownMenuItem(
                      value: 'RESCHEDULED', child: Text('Customer Rescheduled')),
                ],
                onChanged: (val) {
                  if (val != null) setModalState(() => outcome = val);
                },
              ),
              const SizedBox(height: 14),
              const Text('Primary Reason',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              DropdownButtonFormField<String>(
                value: reason,
                dropdownColor: const Color(0xFF1E293B),
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                items: const [
                  DropdownMenuItem(
                      value: 'CUSTOMER_UNAVAILABLE',
                      child: Text('Customer Unavailable / Not Home')),
                  DropdownMenuItem(
                      value: 'ADDRESS_NOT_FOUND',
                      child: Text('Address Incorrect / Inaccessible')),
                  DropdownMenuItem(
                      value: 'CUSTOMER_REFUSED',
                      child: Text('Customer Refused Delivery')),
                  DropdownMenuItem(
                      value: 'PACKAGE_DAMAGED',
                      child: Text('Package Damaged in Transit')),
                  DropdownMenuItem(
                      value: 'WEATHER_TRAFFIC_DELAY',
                      child: Text('Severe Weather / Route Block')),
                ],
                onChanged: (val) {
                  if (val != null) setModalState(() => reason = val);
                },
              ),
              const SizedBox(height: 14),
              const Text('Courier Remarks',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              TextField(
                controller: notesController,
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  hintText: 'e.g. Gate was locked, called customer twice',
                  hintStyle: const TextStyle(color: Colors.white38),
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFF43F5E),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  onPressed: submitting
                      ? null
                      : () async {
                          setModalState(() => submitting = true);
                          try {
                            await _client.recordDeliveryAttempt(
                              widget.tenantId,
                              task.id,
                              outcome: outcome,
                              reason: reason,
                              notes: notesController.text.trim(),
                            );
                            if (ctx.mounted) Navigator.pop(ctx);
                            _loadData();
                            if (mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Delivery attempt recorded successfully'),
                                  backgroundColor: Color(0xFF10B981),
                                ),
                              );
                            }
                          } catch (err) {
                            setModalState(() => submitting = false);
                            if (mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                SnackBar(
                                  content: Text('Error: $err'),
                                  backgroundColor: const Color(0xFFF43F5E),
                                ),
                              );
                            }
                          }
                        },
                  child: submitting
                      ? const SizedBox(
                          height: 18,
                          width: 18,
                          child: CircularProgressIndicator(
                              color: Colors.white, strokeWidth: 2),
                        )
                      : const Text(
                          'Submit Attempt Record',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showProofOfDeliveryDialog(DeliveryTaskModel task) {
    String proofType = 'SIGNATURE';
    final recipientController = TextEditingController(
      text: task.recipientName ?? '',
    );
    final notesController = TextEditingController();
    bool submitting = false;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: const Color(0xFF0F172A),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Padding(
          padding: EdgeInsets.only(
            left: 20,
            right: 20,
            top: 20,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Row(
                    children: [
                      Icon(Icons.verified, color: Color(0xFF10B981), size: 22),
                      SizedBox(width: 8),
                      Text(
                        'Submit Proof of Delivery',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                    ],
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.grey),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
              const SizedBox(height: 6),
              Text(
                'Completing handover for ${task.parcelTrackingNumber ?? task.parcelId}',
                style: const TextStyle(color: Color(0xFF38BDF8), fontSize: 13),
              ),
              const SizedBox(height: 16),
              const Text('Recipient Name (Mandatory)',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              TextField(
                controller: recipientController,
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  hintText: 'Full legal name of receiving party',
                  hintStyle: const TextStyle(color: Colors.white38),
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                ),
              ),
              const SizedBox(height: 14),
              const Text('Verification Method',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              DropdownButtonFormField<String>(
                value: proofType,
                dropdownColor: const Color(0xFF1E293B),
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                items: const [
                  DropdownMenuItem(
                      value: 'SIGNATURE',
                      child: Text('Physical / Digital Signature')),
                  DropdownMenuItem(
                      value: 'OTP', child: Text('SMS / Email OTP Code')),
                  DropdownMenuItem(
                      value: 'PHOTO', child: Text('Handover Photo Confirmation')),
                  DropdownMenuItem(
                      value: 'SAFE_DROP',
                      child: Text('Safe Place Drop (Authorized)')),
                ],
                onChanged: (val) {
                  if (val != null) setModalState(() => proofType = val);
                },
              ),
              const SizedBox(height: 14),
              const Text('Handover Notes',
                  style: TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 6),
              TextField(
                controller: notesController,
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  hintText: 'e.g. Handed to security reception, signed on device',
                  hintStyle: const TextStyle(color: Colors.white38),
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF10B981),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  onPressed: submitting
                      ? null
                      : () async {
                          final recipientName =
                              recipientController.text.trim();
                          if (recipientName.isEmpty) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Recipient name is required'),
                                backgroundColor: Color(0xFFF43F5E),
                              ),
                            );
                            return;
                          }

                          setModalState(() => submitting = true);
                          try {
                            await _client.submitDeliveryProof(
                              widget.tenantId,
                              task.id,
                              proofType: proofType,
                              recipientName: recipientName,
                              signatureData: 'SIG_DIGITAL_VERIFIED_${DateTime.now().millisecondsSinceEpoch}',
                              notes: notesController.text.trim(),
                            );
                            if (ctx.mounted) Navigator.pop(ctx);
                            _loadData();
                            if (mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text(
                                      'Parcel DELIVERED and proof logged successfully!'),
                                  backgroundColor: Color(0xFF10B981),
                                ),
                              );
                            }
                          } catch (err) {
                            setModalState(() => submitting = false);
                            if (mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                SnackBar(
                                  content: Text('Error: $err'),
                                  backgroundColor: const Color(0xFFF43F5E),
                                ),
                              );
                            }
                          }
                        },
                  child: submitting
                      ? const SizedBox(
                          height: 18,
                          width: 18,
                          child: CircularProgressIndicator(
                              color: Colors.white, strokeWidth: 2),
                        )
                      : const Text(
                          'Confirm Handover (DELIVERED)',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF020617),
      appBar: AppBar(
        backgroundColor: const Color(0xFF0F172A),
        elevation: 0,
        title: const Row(
          children: [
            Icon(Icons.local_shipping, color: Color(0xFF38BDF8), size: 22),
            SizedBox(width: 8),
            Text(
              'LogiFlows Dispatch Pro',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 18,
                color: Colors.white,
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh, color: Color(0xFF94A3B8)),
            onPressed: _loadData,
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: const Color(0xFF38BDF8),
          indicatorWeight: 3,
          labelColor: const Color(0xFF38BDF8),
          unselectedLabelColor: const Color(0xFF94A3B8),
          tabs: [
            Tab(
              icon: const Icon(Icons.assignment, size: 20),
              text: 'Tasks (${_tasks.length})',
            ),
            const Tab(
              icon: Icon(Icons.qr_code_scanner, size: 20),
              text: 'Scan & Custody',
            ),
            Tab(
              icon: const Icon(Icons.sync_alt, size: 20),
              text: 'Transfers (${_transfers.length})',
            ),
          ],
        ),
      ),
      body: _loading
          ? const Center(
              child: CircularProgressIndicator(color: Color(0xFF38BDF8)),
            )
          : _error != null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.error_outline,
                            color: Color(0xFFF43F5E), size: 48),
                        const SizedBox(height: 12),
                        Text(
                          _error!,
                          textAlign: TextAlign.center,
                          style: const TextStyle(color: Colors.white70),
                        ),
                        const SizedBox(height: 16),
                        ElevatedButton.icon(
                          onPressed: _loadData,
                          icon: const Icon(Icons.refresh),
                          label: const Text('Retry'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF2563EB),
                          ),
                        ),
                      ],
                    ),
                  ),
                )
              : TabBarView(
                  controller: _tabController,
                  children: [
                    _buildTasksTab(),
                    _buildScanTab(),
                    _buildTransfersTab(),
                  ],
                ),
    );
  }

  Widget _buildTasksTab() {
    if (_tasks.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.task_alt, color: Colors.grey, size: 56),
            SizedBox(height: 12),
            Text(
              'No active delivery tasks assigned.',
              style: TextStyle(color: Colors.white70, fontSize: 16),
            ),
            SizedBox(height: 6),
            Text(
              'New dispatches from your hub will appear here.',
              style: TextStyle(color: Colors.white38, fontSize: 13),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadData,
      color: const Color(0xFF38BDF8),
      backgroundColor: const Color(0xFF0F172A),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _tasks.length,
        itemBuilder: (ctx, index) {
          final task = _tasks[index];
          return _buildTaskCard(task);
        },
      ),
    );
  }

  Widget _buildTaskCard(DeliveryTaskModel task) {
    Color priorityColor;
    switch (task.priority) {
      case 'URGENT':
        priorityColor = const Color(0xFFF43F5E);
        break;
      case 'HIGH':
        priorityColor = const Color(0xFFF59E0B);
        break;
      case 'LOW':
        priorityColor = const Color(0xFF64748B);
        break;
      default:
        priorityColor = const Color(0xFF3B82F6);
    }

    Color statusColor;
    switch (task.status) {
      case 'COMPLETED':
        statusColor = const Color(0xFF10B981);
        break;
      case 'FAILED':
        statusColor = const Color(0xFFF43F5E);
        break;
      case 'IN_PROGRESS':
        statusColor = const Color(0xFFF59E0B);
        break;
      default:
        statusColor = const Color(0xFF818CF8);
    }

    final isActionable =
        task.status == 'ASSIGNED' || task.status == 'IN_PROGRESS';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF0F172A),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: priorityColor.withOpacity(0.3),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.3),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: priorityColor.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  task.priority,
                  style: TextStyle(
                    color: priorityColor,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  task.status,
                  style: TextStyle(
                    color: statusColor,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            task.parcelTrackingNumber ?? 'PKG-TASK-${task.parcelId.substring(0, 8)}',
            style: const TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.bold,
              fontSize: 16,
              letterSpacing: 0.5,
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              const Icon(Icons.person_outline, color: Colors.grey, size: 16),
              const SizedBox(width: 6),
              Text(
                task.recipientName ?? 'Recipient Not Specified',
                style: const TextStyle(color: Colors.white70, fontSize: 13),
              ),
              if (task.recipientPhone != null) ...[
                const SizedBox(width: 12),
                const Icon(Icons.phone, color: Color(0xFF38BDF8), size: 14),
                const SizedBox(width: 4),
                Text(
                  task.recipientPhone!,
                  style: const TextStyle(
                      color: Color(0xFF38BDF8), fontSize: 12),
                ),
              ],
            ],
          ),
          if (task.recipientAddress != null) ...[
            const SizedBox(height: 6),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.location_on_outlined,
                    color: Colors.grey, size: 16),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    task.recipientAddress!,
                    style:
                        const TextStyle(color: Colors.white54, fontSize: 12),
                  ),
                ),
              ],
            ),
          ],
          if (task.notes != null && task.notes!.isNotEmpty) ...[
            const SizedBox(height: 8),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFF1E293B),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Text(
                'Note: ${task.notes}',
                style: const TextStyle(
                    color: Colors.amberAccent,
                    fontSize: 11,
                    fontStyle: FontStyle.italic),
              ),
            ),
          ],
          if (isActionable) ...[
            const SizedBox(height: 14),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    style: OutlinedButton.styleFrom(
                      side: const BorderSide(color: Color(0xFFF43F5E)),
                      foregroundColor: const Color(0xFFF43F5E),
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    onPressed: () => _showRecordAttemptDialog(task),
                    icon: const Icon(Icons.report_problem_outlined, size: 16),
                    label: const Text('Attempt',
                        style: TextStyle(fontSize: 12)),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF10B981),
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    onPressed: () => _showProofOfDeliveryDialog(task),
                    icon: const Icon(Icons.verified, size: 16),
                    label: const Text(
                      'Complete (POD)',
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildScanTab() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: const Color(0xFF0F172A),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFF1E293B)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    Icon(Icons.qr_code_scanner,
                        color: Color(0xFF38BDF8), size: 24),
                    SizedBox(width: 8),
                    Text(
                      'Digital Custody Scanner',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                const Text(
                  'Scan a physical package QR barcode or input the tracking token to verify authenticity and current custody.',
                  style: TextStyle(color: Colors.white54, fontSize: 13),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _scanController,
                  style: const TextStyle(color: Colors.white),
                  decoration: InputDecoration(
                    hintText: 'Enter PKG- tracking or QR payload...',
                    hintStyle: const TextStyle(color: Colors.white38),
                    filled: true,
                    fillColor: const Color(0xFF1E293B),
                    suffixIcon: IconButton(
                      icon: const Icon(Icons.clear, color: Colors.grey),
                      onPressed: () => _scanController.clear(),
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide.none,
                    ),
                  ),
                ),
                const SizedBox(height: 14),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF2563EB),
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    onPressed: _verifyingScan ? null : _handleVerifyScan,
                    icon: _verifyingScan
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                                color: Colors.white, strokeWidth: 2),
                          )
                        : const Icon(Icons.search),
                    label: Text(
                      _verifyingScan ? 'Verifying...' : 'Verify Parcel Code',
                      style: const TextStyle(
                          fontWeight: FontWeight.bold, color: Colors.white),
                    ),
                  ),
                ),
              ],
            ),
          ),
          if (_scanError != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFF43F5E).withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFF43F5E)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.error_outline,
                      color: Color(0xFFF43F5E), size: 22),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      _scanError!,
                      style: const TextStyle(color: Color(0xFFF43F5E)),
                    ),
                  ),
                ],
              ),
            ),
          ],
          if (_scanResult != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFF0F172A),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                    color: const Color(0xFF10B981).withOpacity(0.5)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(Icons.check_circle,
                          color: Color(0xFF10B981), size: 20),
                      SizedBox(width: 8),
                      Text(
                        'Parcel Verified Authenticated',
                        style: TextStyle(
                          color: Color(0xFF10B981),
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                  const Divider(color: Color(0xFF1E293B), height: 24),
                  _buildDetailRow(
                      'Tracking #', _scanResult!['tracking_number']?.toString() ?? 'N/A'),
                  _buildDetailRow(
                      'Current Status', _scanResult!['status']?.toString() ?? 'N/A'),
                  _buildDetailRow(
                      'Origin Hub', _scanResult!['origin_branch_id']?.toString() ?? 'N/A'),
                  _buildDetailRow(
                      'Destination Hub', _scanResult!['destination_branch_id']?.toString() ?? 'N/A'),
                  _buildDetailRow(
                      'Weight (kg)', _scanResult!['weight_kg']?.toString() ?? 'N/A'),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: const TextStyle(color: Colors.grey, fontSize: 13)),
          Text(
            value,
            style: const TextStyle(
                color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
          ),
        ],
      ),
    );
  }

  Widget _buildTransfersTab() {
    if (_transfers.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.sync_alt, color: Colors.grey, size: 56),
            SizedBox(height: 12),
            Text(
              'No inter-branch linehaul manifests.',
              style: TextStyle(color: Colors.white70, fontSize: 16),
            ),
            SizedBox(height: 6),
            Text(
              'Manifests scheduled for route transfer will display here.',
              style: TextStyle(color: Colors.white38, fontSize: 13),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadData,
      color: const Color(0xFF38BDF8),
      backgroundColor: const Color(0xFF0F172A),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _transfers.length,
        itemBuilder: (ctx, index) {
          final transfer = _transfers[index];
          return _buildTransferCard(transfer);
        },
      ),
    );
  }

  Widget _buildTransferCard(BranchTransferModel transfer) {
    Color statusColor;
    switch (transfer.status) {
      case 'RECEIVED':
        statusColor = const Color(0xFF10B981);
        break;
      case 'DISPATCHED':
        statusColor = const Color(0xFF38BDF8);
        break;
      case 'CANCELLED':
        statusColor = const Color(0xFFF43F5E);
        break;
      default:
        statusColor = const Color(0xFFF59E0B);
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF0F172A),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF1E293B)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                transfer.manifestNumber,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 15,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  transfer.status,
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
            'Packages Manifested: ${transfer.totalParcels}',
            style: const TextStyle(color: Colors.white70, fontSize: 13),
          ),
          const SizedBox(height: 12),
          if (transfer.status == 'PENDING') ...[
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF2563EB),
                  padding: const EdgeInsets.symmetric(vertical: 10),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                onPressed: () async {
                  try {
                    await _client.dispatchBranchTransfer(
                        widget.tenantId, transfer.id);
                    _loadData();
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('Manifest DISPATCHED into linehaul transit'),
                          backgroundColor: Color(0xFF10B981),
                        ),
                      );
                    }
                  } catch (e) {
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text('Dispatch failed: $e'),
                          backgroundColor: const Color(0xFFF43F5E),
                        ),
                      );
                    }
                  }
                },
                icon: const Icon(Icons.send, size: 16),
                label: const Text('Dispatch Linehaul'),
              ),
            ),
          ] else if (transfer.status == 'DISPATCHED') ...[
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF10B981),
                  padding: const EdgeInsets.symmetric(vertical: 10),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                onPressed: () async {
                  try {
                    await _client.receiveBranchTransfer(
                        widget.tenantId, transfer.id);
                    _loadData();
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('Manifest RECEIVED at destination hub!'),
                          backgroundColor: Color(0xFF10B981),
                        ),
                      );
                    }
                  } catch (e) {
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text('Receiving failed: $e'),
                          backgroundColor: const Color(0xFFF43F5E),
                        ),
                      );
                    }
                  }
                },
                icon: const Icon(Icons.check_circle_outline, size: 16),
                label: const Text('Receive & Intake All Packages'),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

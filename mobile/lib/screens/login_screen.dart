import 'package:flutter/material.dart';
import '../core/api_config.dart';
import '../services/auth_api_client.dart';
import 'register_screen.dart';

class LoginScreen extends StatefulWidget {
  final AuthApiClient authClient;
  final VoidCallback onLoginSuccess;

  const LoginScreen({
    super.key,
    required this.authClient,
    required this.onLoginSuccess,
  });

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _isLoading = false;
  bool _obscurePassword = true;
  String? _errorMessage;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _showServerConfigDialog() {
    final hostController = TextEditingController(text: ApiConfig.defaultHost);
    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            backgroundColor: const Color(0xFF0F172A),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
              side: const BorderSide(color: Color(0xFF334155)),
            ),
            title: const Row(
              children: [
                Icon(Icons.dns_outlined, color: Color(0xFF38BDF8), size: 22),
                SizedBox(width: 8),
                Text('Server Configuration', style: TextStyle(color: Colors.white, fontSize: 18)),
              ],
            ),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Select a preset or enter host IP:',
                    style: TextStyle(color: Colors.white70, fontSize: 13),
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      ActionChip(
                        label: const Text('localhost (Web/Desktop)'),
                        backgroundColor: const Color(0xFF1E293B),
                        labelStyle: const TextStyle(color: Color(0xFF38BDF8), fontSize: 12),
                        onPressed: () {
                          setDialogState(() {
                            hostController.text = 'localhost';
                          });
                        },
                      ),
                      ActionChip(
                        label: const Text('10.0.2.2 (Android Emulator)'),
                        backgroundColor: const Color(0xFF1E293B),
                        labelStyle: const TextStyle(color: Color(0xFF38BDF8), fontSize: 12),
                        onPressed: () {
                          setDialogState(() {
                            hostController.text = '10.0.2.2';
                          });
                        },
                      ),
                      ActionChip(
                        label: const Text('127.0.0.1'),
                        backgroundColor: const Color(0xFF1E293B),
                        labelStyle: const TextStyle(color: Color(0xFF38BDF8), fontSize: 12),
                        onPressed: () {
                          setDialogState(() {
                            hostController.text = '127.0.0.1';
                          });
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: hostController,
                    style: const TextStyle(color: Colors.white),
                    decoration: const InputDecoration(
                      labelText: 'Host IP or Domain',
                      labelStyle: TextStyle(color: Colors.white60),
                      border: OutlineInputBorder(),
                      prefixIcon: Icon(Icons.router_outlined, color: Color(0xFF38BDF8)),
                    ),
                    onChanged: (_) => setDialogState(() {}),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Target URL: http://${hostController.text.trim().isEmpty ? 'localhost' : hostController.text.trim()}:8080/api/v1',
                    style: const TextStyle(color: Colors.white54, fontSize: 11, fontFamily: 'monospace'),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: const Text('Cancel', style: TextStyle(color: Colors.white60)),
              ),
              ElevatedButton(
                style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF2563EB)),
                onPressed: () {
                  final newHost = hostController.text.trim();
                  if (newHost.isNotEmpty) {
                    ApiConfig.setCustomHost(newHost);
                  } else {
                    ApiConfig.setCustomHost(null);
                  }
                  Navigator.pop(ctx);
                  setState(() {
                    _errorMessage = null;
                  });
                },
                child: const Text('Save & Apply', style: TextStyle(color: Colors.white)),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      await widget.authClient.login(
        email: _emailController.text,
        password: _passwordController.text,
      );
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
        widget.onLoginSuccess();
      }
    } catch (e) {
      if (mounted) {
        String msg = e.toString();
        if (msg.startsWith('Exception: ')) {
          msg = msg.substring(11);
        }
        if (msg.contains('ClientException') || msg.contains('Failed to fetch') || msg.contains('SocketException')) {
          msg = 'Unable to connect to LogiFlows API at ${ApiConfig.baseUrl}. Please verify the backend is running or change host settings.';
        }
        setState(() {
          _errorMessage = msg;
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF020617),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Logo / Header
                  const Center(
                    child: CircleAvatar(
                      radius: 36,
                      backgroundColor: Color(0xFF1E293B),
                      child: Icon(Icons.local_shipping, size: 40, color: Color(0xFF38BDF8)),
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Text(
                    'LogiFlows Driver Pro',
                    textAlign: TextAlign.center,
                    style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.white),
                  ),
                  const SizedBox(height: 6),
                  const Text(
                    'Sign in to access assigned shipments & routes',
                    textAlign: TextAlign.center,
                    style: TextStyle(fontSize: 14, color: Colors.white70),
                  ),
                  const SizedBox(height: 32),

                  // Error Banner
                  if (_errorMessage != null) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFEF4444).withOpacity(0.15),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: const Color(0xFFEF4444).withOpacity(0.4)),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(Icons.error_outline, color: Color(0xFFF87171), size: 20),
                              const SizedBox(width: 10),
                              Expanded(
                                child: Text(
                                  _errorMessage!,
                                  style: const TextStyle(color: Color(0xFFFCA5A5), fontSize: 13),
                                ),
                              ),
                            ],
                          ),
                          if (_errorMessage!.contains('Unable to connect')) ...[
                            const SizedBox(height: 8),
                            Align(
                              alignment: Alignment.centerRight,
                              child: TextButton.icon(
                                style: TextButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                  backgroundColor: const Color(0xFF1E293B),
                                ),
                                onPressed: _showServerConfigDialog,
                                icon: const Icon(Icons.settings, size: 14, color: Color(0xFF38BDF8)),
                                label: const Text('Configure Server Host', style: TextStyle(color: Color(0xFF38BDF8), fontSize: 12)),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],

                  // Email Field
                  TextFormField(
                    controller: _emailController,
                    keyboardType: TextInputType.emailAddress,
                    style: const TextStyle(color: Colors.white),
                    decoration: InputDecoration(
                      labelText: 'Corporate Email',
                      labelStyle: const TextStyle(color: Colors.white60),
                      prefixIcon: const Icon(Icons.email_outlined, color: Color(0xFF38BDF8)),
                      filled: true,
                      fillColor: const Color(0xFF0F172A),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10),
                        borderSide: const BorderSide(color: Color(0xFF334155)),
                      ),
                    ),
                    validator: (val) {
                      if (val == null || val.trim().isEmpty) return 'Email is required';
                      if (!val.contains('@')) return 'Enter a valid email address';
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),

                  // Password Field
                  TextFormField(
                    controller: _passwordController,
                    obscureText: _obscurePassword,
                    style: const TextStyle(color: Colors.white),
                    decoration: InputDecoration(
                      labelText: 'Password',
                      labelStyle: const TextStyle(color: Colors.white60),
                      prefixIcon: const Icon(Icons.lock_outline, color: Color(0xFF38BDF8)),
                      suffixIcon: IconButton(
                        icon: Icon(
                          _obscurePassword ? Icons.visibility_off : Icons.visibility,
                          color: Colors.white60,
                        ),
                        onPressed: () {
                          setState(() {
                            _obscurePassword = !_obscurePassword;
                          });
                        },
                      ),
                      filled: true,
                      fillColor: const Color(0xFF0F172A),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10),
                        borderSide: const BorderSide(color: Color(0xFF334155)),
                      ),
                    ),
                    validator: (val) {
                      if (val == null || val.isEmpty) return 'Password is required';
                      if (val.length < 8) return 'Password must be at least 8 characters';
                      return null;
                    },
                  ),
                  const SizedBox(height: 24),

                  // Submit Button
                  ElevatedButton(
                    onPressed: _isLoading ? null : _handleLogin,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF2563EB),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                    ),
                    child: _isLoading
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          )
                        : const Text('Sign In', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  ),
                  const SizedBox(height: 16),

                  // Quick Demo Credentials Helper
                  Center(
                    child: TextButton.icon(
                      onPressed: () {
                        setState(() {
                          _emailController.text = 'admin1@gmail.com';
                          _passwordController.text = 'ComplexP@ssw0rd!2026';
                        });
                      },
                      icon: const Icon(Icons.auto_fix_high, size: 14, color: Color(0xFF38BDF8)),
                      label: const Text(
                        'Autofill Demo Credentials (admin1@gmail.com)',
                        style: TextStyle(color: Color(0xFF38BDF8), fontSize: 12),
                      ),
                    ),
                  ),

                  // Register link
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text("Don't have an account?", style: TextStyle(color: Colors.white60)),
                      TextButton(
                        onPressed: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) => RegisterScreen(
                                authClient: widget.authClient,
                                onRegisterSuccess: widget.onLoginSuccess,
                              ),
                            ),
                          );
                        },
                        child: const Text('Register Company', style: TextStyle(color: Color(0xFF38BDF8))),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  // Active Server Endpoint Indicator
                  Center(
                    child: InkWell(
                      onTap: _showServerConfigDialog,
                      borderRadius: BorderRadius.circular(20),
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                        decoration: BoxDecoration(
                          color: const Color(0xFF0F172A),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: const Color(0xFF1E293B)),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 8,
                              height: 8,
                              decoration: const BoxDecoration(
                                color: Color(0xFF10B981),
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'Server: ${ApiConfig.baseUrl}',
                              style: const TextStyle(color: Colors.white54, fontSize: 11, fontFamily: 'monospace'),
                            ),
                            const SizedBox(width: 6),
                            const Icon(Icons.settings, size: 13, color: Colors.white38),
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

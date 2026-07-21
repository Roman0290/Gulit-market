import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'auth_provider.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final user = auth.currentUser;

    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: user == null
          ? const SizedBox.shrink()
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                CircleAvatar(
                  radius: 32,
                  child: Text(user.name.isNotEmpty ? user.name[0].toUpperCase() : '?'),
                ),
                const SizedBox(height: 12),
                Text(user.name, style: Theme.of(context).textTheme.titleLarge),
                Text(user.email, style: Theme.of(context).textTheme.bodyMedium),
                const SizedBox(height: 4),
                Chip(label: Text(user.role)),
                const SizedBox(height: 24),
                ListTile(
                  leading: const Icon(Icons.phone_outlined),
                  title: Text(user.phone),
                  contentPadding: EdgeInsets.zero,
                ),
                const Divider(),
                ListTile(
                  leading: const Icon(Icons.logout),
                  title: const Text('Log out'),
                  contentPadding: EdgeInsets.zero,
                  onTap: () => context.read<AuthProvider>().logout(),
                ),
              ],
            ),
    );
  }
}

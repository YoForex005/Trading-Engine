import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  TouchableOpacity,
  Switch,
  Alert,
} from 'react-native';
import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '@/store';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { Button } from '@/components/Button';
import { setThemeMode, setHapticEnabled } from '@/store/slices/uiSlice';
import { logout } from '@/store/slices/authSlice';
import { biometricAuth, secureStorage } from '@/utils/security';
import { useLogoutMutation } from '@/services/api';

export default function SettingsScreen() {
  const theme = useTheme();
  const dispatch = useDispatch();
  const { themeMode, hapticEnabled } = useSelector((state: RootState) => state.ui);
  const { biometricEnabled } = useSelector((state: RootState) => state.auth);
  const [logoutMutation] = useLogoutMutation();

  const handleToggleBiometric = async () => {
    if (!biometricEnabled) {
      const isAvailable = await biometricAuth.isAvailable();
      if (!isAvailable) {
        Alert.alert('Error', 'Biometric authentication not available on this device');
        return;
      }

      const success = await biometricAuth.createKeys();
      if (success) {
        Alert.alert('Success', 'Biometric authentication enabled');
      }
    } else {
      await biometricAuth.deleteKeys();
      Alert.alert('Success', 'Biometric authentication disabled');
    }
  };

  const handleLogout = async () => {
    Alert.alert('Logout', 'Are you sure you want to logout?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Logout',
        style: 'destructive',
        onPress: async () => {
          try {
            await logoutMutation().unwrap();
          } catch (error) {
            // Continue logout even if API call fails
          }

          await secureStorage.removeItem('authToken');
          await secureStorage.removeItem('refreshToken');
          dispatch(logout());
        },
      },
    ]);
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <ScrollView contentContainerStyle={styles.scrollView}>
        <View style={styles.header}>
          <Text
            style={[
              styles.title,
              { color: theme.colors.text, fontSize: theme.typography.h2.fontSize },
            ]}
          >
            Settings
          </Text>
        </View>

        {/* Appearance */}
        <Card elevation="md" style={styles.card}>
          <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>Appearance</Text>

          <View style={styles.settingRow}>
            <Text style={[styles.settingLabel, { color: theme.colors.text }]}>Theme</Text>
            <View style={styles.themeButtons}>
              {(['light', 'dark', 'auto'] as const).map((mode) => (
                <TouchableOpacity
                  key={mode}
                  onPress={() => dispatch(setThemeMode(mode))}
                  style={[
                    styles.themeButton,
                    {
                      backgroundColor:
                        themeMode === mode ? theme.colors.primary : theme.colors.surface,
                      borderRadius: theme.borderRadius.sm,
                    },
                  ]}
                >
                  <Text
                    style={[
                      styles.themeButtonText,
                      { color: themeMode === mode ? '#FFF' : theme.colors.text },
                    ]}
                  >
                    {mode.charAt(0).toUpperCase() + mode.slice(1)}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>
        </Card>

        {/* Preferences */}
        <Card elevation="md" style={styles.card}>
          <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>Preferences</Text>

          <View style={styles.settingRow}>
            <Text style={[styles.settingLabel, { color: theme.colors.text }]}>
              Haptic Feedback
            </Text>
            <Switch
              value={hapticEnabled}
              onValueChange={(value) => dispatch(setHapticEnabled(value))}
            />
          </View>

          <View style={styles.settingRow}>
            <Text style={[styles.settingLabel, { color: theme.colors.text }]}>
              Biometric Authentication
            </Text>
            <Switch value={biometricEnabled} onValueChange={handleToggleBiometric} />
          </View>
        </Card>

        {/* Security */}
        <Card elevation="md" style={styles.card}>
          <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>Security</Text>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Change Password
            </Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Two-Factor Authentication
            </Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Active Sessions
            </Text>
          </TouchableOpacity>
        </Card>

        {/* About */}
        <Card elevation="md" style={styles.card}>
          <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>About</Text>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Terms of Service
            </Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Privacy Policy
            </Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.settingItem}>
            <Text style={[styles.settingItemText, { color: theme.colors.text }]}>
              Help & Support
            </Text>
          </TouchableOpacity>

          <View style={styles.versionRow}>
            <Text style={[styles.versionText, { color: theme.colors.textSecondary }]}>
              Version 1.0.0
            </Text>
          </View>
        </Card>

        <Button
          title="Logout"
          onPress={handleLogout}
          variant="error"
          fullWidth
          style={{ marginTop: 16 }}
        />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scrollView: {
    padding: 16,
  },
  header: {
    marginBottom: 16,
  },
  title: {
    fontWeight: '700',
  },
  card: {
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 16,
  },
  settingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  settingLabel: {
    fontSize: 16,
  },
  themeButtons: {
    flexDirection: 'row',
  },
  themeButton: {
    paddingVertical: 6,
    paddingHorizontal: 12,
    marginLeft: 8,
  },
  themeButtonText: {
    fontSize: 14,
    fontWeight: '600',
  },
  settingItem: {
    paddingVertical: 12,
  },
  settingItemText: {
    fontSize: 16,
  },
  versionRow: {
    paddingVertical: 12,
    alignItems: 'center',
  },
  versionText: {
    fontSize: 14,
  },
});

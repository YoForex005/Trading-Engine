import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Alert,
} from 'react-native';
import { useDispatch } from 'react-redux';
import type { StackNavigationProp } from '@react-navigation/stack';
import type { RootStackParamList } from '@/types';
import { useTheme } from '@/hooks/useTheme';
import { Input } from '@/components/Input';
import { Button } from '@/components/Button';
import { useLoginMutation } from '@/services/api';
import { setCredentials } from '@/store/slices/authSlice';
import { secureStorage, biometricAuth } from '@/utils/security';
import { validateEmail } from '@/utils/validation';

type LoginScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Login'>;
};

export default function LoginScreen({ navigation }: LoginScreenProps) {
  const theme = useTheme();
  const dispatch = useDispatch();
  const [login, { isLoading }] = useLoginMutation();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  const handleBiometricLogin = async () => {
    const isAvailable = await biometricAuth.isAvailable();
    if (!isAvailable) {
      Alert.alert('Error', 'Biometric authentication not available');
      return;
    }

    const authenticated = await biometricAuth.authenticate('Login to Trading Engine');
    if (authenticated) {
      const savedEmail = await secureStorage.getItem('userEmail');
      const savedPassword = await secureStorage.getItem('userPassword');

      if (savedEmail && savedPassword) {
        handleLogin(savedEmail, savedPassword);
      }
    }
  };

  const handleLogin = async (emailParam?: string, passwordParam?: string) => {
    const loginEmail = emailParam || email;
    const loginPassword = passwordParam || password;

    // Validation
    let isValid = true;

    if (!loginEmail) {
      setEmailError('Email is required');
      isValid = false;
    } else if (!validateEmail(loginEmail)) {
      setEmailError('Invalid email format');
      isValid = false;
    } else {
      setEmailError('');
    }

    if (!loginPassword) {
      setPasswordError('Password is required');
      isValid = false;
    } else {
      setPasswordError('');
    }

    if (!isValid) return;

    try {
      const result = await login({
        email: loginEmail,
        password: loginPassword,
      }).unwrap();

      // Store credentials
      await secureStorage.setItem('authToken', result.token);
      await secureStorage.setItem('refreshToken', result.refreshToken);
      await secureStorage.setItem('userEmail', loginEmail);

      // Store password for biometric login if enabled
      if (result.user.biometricEnabled) {
        await secureStorage.setItem('userPassword', loginPassword);
      }

      dispatch(setCredentials(result));
    } catch (error: any) {
      Alert.alert('Login Failed', error?.data?.message || 'Invalid credentials');
    }
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}
      >
        <ScrollView contentContainerStyle={styles.scrollView}>
          <View style={styles.content}>
            <Text
              style={[
                styles.title,
                {
                  color: theme.colors.text,
                  fontSize: theme.typography.h1.fontSize,
                  fontWeight: '700',
                },
              ]}
            >
              Welcome Back
            </Text>
            <Text
              style={[
                styles.subtitle,
                {
                  color: theme.colors.textSecondary,
                  fontSize: theme.typography.body1.fontSize,
                  marginBottom: theme.spacing.xl,
                },
              ]}
            >
              Sign in to continue trading
            </Text>

            <Input
              label="Email"
              placeholder="Enter your email"
              value={email}
              onChangeText={setEmail}
              keyboardType="email-address"
              autoCapitalize="none"
              error={emailError}
            />

            <Input
              label="Password"
              placeholder="Enter your password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              error={passwordError}
            />

            <Button
              title="Sign In"
              onPress={() => handleLogin()}
              loading={isLoading}
              fullWidth
              style={{ marginTop: theme.spacing.md }}
            />

            <Button
              title="Sign In with Biometrics"
              onPress={handleBiometricLogin}
              variant="outline"
              fullWidth
              style={{ marginTop: theme.spacing.md }}
            />

            <Button
              title="Create Account"
              onPress={() => navigation.navigate('Register')}
              variant="outline"
              fullWidth
              style={{ marginTop: theme.spacing.md }}
            />
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  keyboardView: {
    flex: 1,
  },
  scrollView: {
    flexGrow: 1,
    justifyContent: 'center',
  },
  content: {
    padding: 24,
  },
  title: {
    marginBottom: 8,
  },
  subtitle: {},
});

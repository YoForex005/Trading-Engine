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
import { useRegisterMutation } from '@/services/api';
import { setCredentials } from '@/store/slices/authSlice';
import { secureStorage } from '@/utils/security';
import { validateEmail, validatePassword } from '@/utils/validation';

type RegisterScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Register'>;
};

export default function RegisterScreen({ navigation }: RegisterScreenProps) {
  const theme = useTheme();
  const dispatch = useDispatch();
  const [register, { isLoading }] = useRegisterMutation();

  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleRegister = async () => {
    const newErrors: Record<string, string> = {};

    // Validation
    if (!firstName) newErrors.firstName = 'First name is required';
    if (!lastName) newErrors.lastName = 'Last name is required';

    if (!email) {
      newErrors.email = 'Email is required';
    } else if (!validateEmail(email)) {
      newErrors.email = 'Invalid email format';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else {
      const passwordValidation = validatePassword(password);
      if (!passwordValidation.valid) {
        newErrors.password = passwordValidation.errors[0];
      }
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);

    if (Object.keys(newErrors).length > 0) return;

    try {
      const result = await register({
        email,
        password,
        firstName,
        lastName,
        phoneNumber: phoneNumber || undefined,
      }).unwrap();

      // Store credentials
      await secureStorage.setItem('authToken', result.token);
      await secureStorage.setItem('refreshToken', result.refreshToken);
      await secureStorage.setItem('userEmail', email);

      dispatch(setCredentials(result));
    } catch (error: any) {
      Alert.alert('Registration Failed', error?.data?.message || 'Unable to create account');
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
              Create Account
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
              Start your trading journey
            </Text>

            <Input
              label="First Name"
              placeholder="Enter your first name"
              value={firstName}
              onChangeText={setFirstName}
              error={errors.firstName}
            />

            <Input
              label="Last Name"
              placeholder="Enter your last name"
              value={lastName}
              onChangeText={setLastName}
              error={errors.lastName}
            />

            <Input
              label="Email"
              placeholder="Enter your email"
              value={email}
              onChangeText={setEmail}
              keyboardType="email-address"
              autoCapitalize="none"
              error={errors.email}
            />

            <Input
              label="Phone Number (Optional)"
              placeholder="+1 234 567 8900"
              value={phoneNumber}
              onChangeText={setPhoneNumber}
              keyboardType="phone-pad"
              error={errors.phoneNumber}
            />

            <Input
              label="Password"
              placeholder="Create a password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              error={errors.password}
            />

            <Input
              label="Confirm Password"
              placeholder="Confirm your password"
              value={confirmPassword}
              onChangeText={setConfirmPassword}
              secureTextEntry
              error={errors.confirmPassword}
            />

            <Button
              title="Create Account"
              onPress={handleRegister}
              loading={isLoading}
              fullWidth
              style={{ marginTop: theme.spacing.md }}
            />

            <Button
              title="Already have an account? Sign In"
              onPress={() => navigation.goBack()}
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
  },
  content: {
    padding: 24,
  },
  title: {
    marginBottom: 8,
  },
  subtitle: {},
});

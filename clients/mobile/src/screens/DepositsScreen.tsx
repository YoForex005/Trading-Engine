import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  Alert,
  TouchableOpacity,
} from 'react';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { Input } from '@/components/Input';
import { Button } from '@/components/Button';
import { useGetAccountsQuery, useCreateDepositMutation } from '@/services/api';
import type { Currency } from '@/types';

export default function DepositsScreen() {
  const theme = useTheme();
  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const [createDeposit, { isLoading }] = useCreateDepositMutation();

  const [amount, setAmount] = useState('');
  const [method, setMethod] = useState<'BANK_TRANSFER' | 'CREDIT_CARD' | 'CRYPTO'>(
    'BANK_TRANSFER',
  );

  const handleDeposit = async () => {
    if (!account) {
      Alert.alert('Error', 'No account selected');
      return;
    }

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      Alert.alert('Error', 'Invalid amount');
      return;
    }

    try {
      await createDeposit({
        accountId: account.id,
        amount: amountNum,
        currency: account.currency,
        method,
      }).unwrap();

      Alert.alert('Success', 'Deposit request submitted');
      setAmount('');
    } catch (error: any) {
      Alert.alert('Error', error?.data?.message || 'Failed to create deposit');
    }
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
            Deposit Funds
          </Text>
        </View>

        <Card elevation="md" style={styles.card}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>
            Select Payment Method
          </Text>

          <View style={styles.methodSelector}>
            <TouchableOpacity
              onPress={() => setMethod('BANK_TRANSFER')}
              style={[
                styles.methodButton,
                {
                  backgroundColor:
                    method === 'BANK_TRANSFER' ? theme.colors.primary : theme.colors.surface,
                  borderRadius: theme.borderRadius.md,
                },
              ]}
            >
              <Text
                style={[
                  styles.methodText,
                  { color: method === 'BANK_TRANSFER' ? '#FFF' : theme.colors.text },
                ]}
              >
                Bank Transfer
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => setMethod('CREDIT_CARD')}
              style={[
                styles.methodButton,
                {
                  backgroundColor:
                    method === 'CREDIT_CARD' ? theme.colors.primary : theme.colors.surface,
                  borderRadius: theme.borderRadius.md,
                },
              ]}
            >
              <Text
                style={[
                  styles.methodText,
                  { color: method === 'CREDIT_CARD' ? '#FFF' : theme.colors.text },
                ]}
              >
                Credit Card
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => setMethod('CRYPTO')}
              style={[
                styles.methodButton,
                {
                  backgroundColor:
                    method === 'CRYPTO' ? theme.colors.primary : theme.colors.surface,
                  borderRadius: theme.borderRadius.md,
                },
              ]}
            >
              <Text
                style={[
                  styles.methodText,
                  { color: method === 'CRYPTO' ? '#FFF' : theme.colors.text },
                ]}
              >
                Crypto
              </Text>
            </TouchableOpacity>
          </View>

          <Input
            label="Amount"
            placeholder="Enter amount"
            value={amount}
            onChangeText={setAmount}
            keyboardType="decimal-pad"
          />

          <Button
            title="Submit Deposit"
            onPress={handleDeposit}
            loading={isLoading}
            fullWidth
          />
        </Card>

        <Card elevation="md" style={styles.infoCard}>
          <Text style={[styles.infoTitle, { color: theme.colors.text }]}>
            Deposit Information
          </Text>
          <Text style={[styles.infoText, { color: theme.colors.textSecondary }]}>
            • Bank transfers typically take 1-3 business days
            {'\n'}• Credit card deposits are instant but may incur fees
            {'\n'}• Crypto deposits require network confirmations
            {'\n'}• Minimum deposit: $100
            {'\n'}• Maximum deposit: $100,000 per transaction
          </Text>
        </Card>
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
  label: {
    fontSize: 12,
    marginBottom: 12,
  },
  methodSelector: {
    marginBottom: 16,
  },
  methodButton: {
    paddingVertical: 12,
    paddingHorizontal: 16,
    marginBottom: 8,
  },
  methodText: {
    fontSize: 16,
    fontWeight: '600',
    textAlign: 'center',
  },
  infoCard: {
    marginBottom: 16,
  },
  infoTitle: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 12,
  },
  infoText: {
    fontSize: 14,
    lineHeight: 22,
  },
});

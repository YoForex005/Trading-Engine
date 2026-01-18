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
import { useGetAccountsQuery, useCreateWithdrawalMutation } from '@/services/api';

export default function WithdrawalsScreen() {
  const theme = useTheme();
  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const [createWithdrawal, { isLoading }] = useCreateWithdrawalMutation();

  const [amount, setAmount] = useState('');
  const [method, setMethod] = useState<'BANK_TRANSFER' | 'CRYPTO'>('BANK_TRANSFER');

  const handleWithdrawal = async () => {
    if (!account) {
      Alert.alert('Error', 'No account selected');
      return;
    }

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      Alert.alert('Error', 'Invalid amount');
      return;
    }

    if (amountNum > account.freeMargin) {
      Alert.alert('Error', 'Insufficient free margin');
      return;
    }

    try {
      await createWithdrawal({
        accountId: account.id,
        amount: amountNum,
        currency: account.currency,
        method,
      }).unwrap();

      Alert.alert('Success', 'Withdrawal request submitted');
      setAmount('');
    } catch (error: any) {
      Alert.alert('Error', error?.data?.message || 'Failed to create withdrawal');
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
            Withdraw Funds
          </Text>
        </View>

        <Card elevation="md" style={styles.card}>
          <View style={styles.balanceInfo}>
            <Text style={[styles.balanceLabel, { color: theme.colors.textSecondary }]}>
              Available to Withdraw
            </Text>
            <Text
              style={[
                styles.balanceValue,
                { color: theme.colors.text, fontSize: 24, fontWeight: '700' },
              ]}
            >
              ${account?.freeMargin.toFixed(2)}
            </Text>
          </View>

          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>
            Select Withdrawal Method
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
            title="Submit Withdrawal"
            onPress={handleWithdrawal}
            loading={isLoading}
            fullWidth
          />
        </Card>

        <Card elevation="md" style={styles.infoCard}>
          <Text style={[styles.infoTitle, { color: theme.colors.text }]}>
            Withdrawal Information
          </Text>
          <Text style={[styles.infoText, { color: theme.colors.textSecondary }]}>
            • Withdrawals are processed within 24 hours
            {'\n'}• Bank transfers may take 2-5 business days
            {'\n'}• Crypto withdrawals require network confirmations
            {'\n'}• Minimum withdrawal: $50
            {'\n'}• You cannot withdraw margin used in open positions
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
  balanceInfo: {
    marginBottom: 24,
    alignItems: 'center',
  },
  balanceLabel: {
    fontSize: 14,
    marginBottom: 8,
  },
  balanceValue: {},
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

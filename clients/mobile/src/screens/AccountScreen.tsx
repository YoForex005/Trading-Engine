import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
} from 'react-native';
import type { StackNavigationProp } from '@react-navigation/stack';
import type { RootStackParamList } from '@/types';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { useGetAccountsQuery, useGetProfileQuery } from '@/services/api';
import { formatCurrency, formatPercent } from '@/utils/formatters';

type AccountScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Account'>;
};

export default function AccountScreen({ navigation }: AccountScreenProps) {
  const theme = useTheme();
  const { data: user, refetch: refetchUser } = useGetProfileQuery();
  const { data: accounts, isLoading, refetch: refetchAccounts } = useGetAccountsQuery();
  const account = accounts?.[0];

  const handleRefresh = () => {
    refetchUser();
    refetchAccounts();
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <ScrollView
        contentContainerStyle={styles.scrollView}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={handleRefresh} />}
      >
        <View style={styles.header}>
          <Text
            style={[
              styles.title,
              { color: theme.colors.text, fontSize: theme.typography.h2.fontSize },
            ]}
          >
            Account
          </Text>
        </View>

        {/* User Info */}
        {user && (
          <Card elevation="md" style={styles.card}>
            <Text style={[styles.cardTitle, { color: theme.colors.text }]}>Profile</Text>
            <View style={styles.infoRow}>
              <Text style={[styles.infoLabel, { color: theme.colors.textSecondary }]}>Name</Text>
              <Text style={[styles.infoValue, { color: theme.colors.text }]}>
                {user.firstName} {user.lastName}
              </Text>
            </View>
            <View style={styles.infoRow}>
              <Text style={[styles.infoLabel, { color: theme.colors.textSecondary }]}>Email</Text>
              <Text style={[styles.infoValue, { color: theme.colors.text }]}>{user.email}</Text>
            </View>
            <View style={styles.infoRow}>
              <Text style={[styles.infoLabel, { color: theme.colors.textSecondary }]}>
                KYC Status
              </Text>
              <Text
                style={[
                  styles.infoValue,
                  {
                    color:
                      user.kycStatus === 'APPROVED'
                        ? theme.colors.success
                        : user.kycStatus === 'REJECTED'
                        ? theme.colors.error
                        : theme.colors.warning,
                  },
                ]}
              >
                {user.kycStatus}
              </Text>
            </View>
          </Card>
        )}

        {/* Account Balance */}
        {account && (
          <Card elevation="md" style={styles.card}>
            <Text style={[styles.cardTitle, { color: theme.colors.text }]}>
              Account Balance
            </Text>
            <Text
              style={[
                styles.balance,
                { color: theme.colors.text, fontSize: 32, fontWeight: '700', marginBottom: 16 },
              ]}
            >
              {formatCurrency(account.balance)}
            </Text>

            <View style={styles.accountStats}>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Equity
                </Text>
                <Text style={[styles.statValue, { color: theme.colors.text }]}>
                  {formatCurrency(account.equity)}
                </Text>
              </View>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Margin
                </Text>
                <Text style={[styles.statValue, { color: theme.colors.text }]}>
                  {formatCurrency(account.margin)}
                </Text>
              </View>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Free Margin
                </Text>
                <Text style={[styles.statValue, { color: theme.colors.text }]}>
                  {formatCurrency(account.freeMargin)}
                </Text>
              </View>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Margin Level
                </Text>
                <Text style={[styles.statValue, { color: theme.colors.text }]}>
                  {formatPercent(account.marginLevel)}
                </Text>
              </View>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Leverage
                </Text>
                <Text style={[styles.statValue, { color: theme.colors.text }]}>
                  1:{account.leverage}
                </Text>
              </View>
              <View style={styles.statItem}>
                <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                  Profit
                </Text>
                <Text
                  style={[
                    styles.statValue,
                    {
                      color:
                        account.profit >= 0 ? theme.colors.success : theme.colors.error,
                    },
                  ]}
                >
                  {formatCurrency(account.profit)}
                </Text>
              </View>
            </View>
          </Card>
        )}

        {/* Quick Actions */}
        <Card elevation="md" style={styles.card}>
          <Text style={[styles.cardTitle, { color: theme.colors.text }]}>Quick Actions</Text>

          <TouchableOpacity
            style={[styles.actionItem, { borderBottomColor: theme.colors.border }]}
            onPress={() => navigation.navigate('Deposits')}
          >
            <Text style={{ fontSize: 24 }}>💰</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>
              Deposit Funds
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.actionItem, { borderBottomColor: theme.colors.border }]}
            onPress={() => navigation.navigate('Withdrawals')}
          >
            <Text style={{ fontSize: 24 }}>💸</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>
              Withdraw Funds
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.actionItem, { borderBottomColor: theme.colors.border }]}
            onPress={() => navigation.navigate('History')}
          >
            <Text style={{ fontSize: 24 }}>📊</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>
              Trade History
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.actionItem, { borderBottomWidth: 0 }]}
            onPress={() => navigation.navigate('Settings')}
          >
            <Text style={{ fontSize: 24 }}>⚙️</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>Settings</Text>
          </TouchableOpacity>
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
  cardTitle: {
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 16,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 12,
  },
  infoLabel: {
    fontSize: 14,
  },
  infoValue: {
    fontSize: 14,
    fontWeight: '600',
  },
  balance: {},
  accountStats: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  statItem: {
    width: '50%',
    marginBottom: 16,
  },
  statLabel: {
    fontSize: 12,
    marginBottom: 4,
  },
  statValue: {
    fontSize: 16,
    fontWeight: '600',
  },
  actionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
    borderBottomWidth: 1,
  },
  actionText: {
    fontSize: 16,
    marginLeft: 16,
    fontWeight: '500',
  },
});

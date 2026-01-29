import React, { useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import { useSelector } from 'react-redux';
import type { StackNavigationProp } from '@react-navigation/stack';
import type { RootStackParamList } from '@/types';
import type { RootState } from '@/store';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { PriceDisplay } from '@/components/PriceDisplay';
import {
  useGetAccountsQuery,
  useGetPositionsQuery,
  useGetTickersQuery,
} from '@/services/api';
import { formatCurrency, formatPercent } from '@/utils/formatters';

type DashboardScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Dashboard'>;
};

export default function DashboardScreen({ navigation }: DashboardScreenProps) {
  const theme = useTheme();
  const watchlist = useSelector((state: RootState) => state.market.watchlist);

  const { data: accounts, isLoading: accountsLoading, refetch: refetchAccounts } =
    useGetAccountsQuery();
  const { data: tickers, isLoading: tickersLoading, refetch: refetchTickers } =
    useGetTickersQuery(watchlist);

  const account = accounts?.[0];
  const { data: positions, isLoading: positionsLoading, refetch: refetchPositions } =
    useGetPositionsQuery(account?.id || '', {
      skip: !account,
    });

  const handleRefresh = () => {
    refetchAccounts();
    refetchTickers();
    refetchPositions();
  };

  const isRefreshing = accountsLoading || tickersLoading || positionsLoading;

  if (!account) {
    return (
      <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
        <View style={styles.loadingContainer}>
          <Text style={{ color: theme.colors.text }}>Loading account...</Text>
        </View>
      </SafeAreaView>
    );
  }

  const totalPositions = positions?.length || 0;
  const totalProfit = positions?.reduce((sum, p) => sum + p.profit, 0) || 0;

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <ScrollView
        contentContainerStyle={styles.scrollView}
        refreshControl={
          <RefreshControl refreshing={isRefreshing} onRefresh={handleRefresh} />
        }
      >
        <View style={styles.header}>
          <Text
            style={[
              styles.title,
              { color: theme.colors.text, fontSize: theme.typography.h2.fontSize },
            ]}
          >
            Dashboard
          </Text>
          <TouchableOpacity onPress={() => navigation.navigate('Notifications')}>
            <Text style={{ color: theme.colors.primary, fontSize: 24 }}>🔔</Text>
          </TouchableOpacity>
        </View>

        {/* Account Summary */}
        <Card elevation="md" style={styles.accountCard}>
          <Text
            style={[
              styles.accountType,
              { color: theme.colors.textSecondary, fontSize: 14 },
            ]}
          >
            {account.accountType} ACCOUNT
          </Text>
          <Text
            style={[
              styles.balance,
              { color: theme.colors.text, fontSize: 32, fontWeight: '700' },
            ]}
          >
            {formatCurrency(account.balance)}
          </Text>

          <View style={styles.accountStats}>
            <View style={styles.stat}>
              <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                Equity
              </Text>
              <Text style={[styles.statValue, { color: theme.colors.text }]}>
                {formatCurrency(account.equity)}
              </Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                Free Margin
              </Text>
              <Text style={[styles.statValue, { color: theme.colors.text }]}>
                {formatCurrency(account.freeMargin)}
              </Text>
            </View>
            <View style={styles.stat}>
              <Text style={[styles.statLabel, { color: theme.colors.textSecondary }]}>
                Margin Level
              </Text>
              <Text style={[styles.statValue, { color: theme.colors.text }]}>
                {formatPercent(account.marginLevel)}
              </Text>
            </View>
          </View>
        </Card>

        {/* Positions Summary */}
        <Card elevation="md" style={styles.positionsCard}>
          <View style={styles.positionsHeader}>
            <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>
              Open Positions
            </Text>
            <TouchableOpacity onPress={() => navigation.navigate('Positions')}>
              <Text style={{ color: theme.colors.primary }}>View All</Text>
            </TouchableOpacity>
          </View>

          <View style={styles.positionsStats}>
            <View style={styles.positionStat}>
              <Text style={[styles.positionStatLabel, { color: theme.colors.textSecondary }]}>
                Total Positions
              </Text>
              <Text
                style={[
                  styles.positionStatValue,
                  { color: theme.colors.text, fontSize: 24, fontWeight: '700' },
                ]}
              >
                {totalPositions}
              </Text>
            </View>
            <View style={styles.positionStat}>
              <Text style={[styles.positionStatLabel, { color: theme.colors.textSecondary }]}>
                Total P/L
              </Text>
              <Text
                style={[
                  styles.positionStatValue,
                  {
                    color: totalProfit >= 0 ? theme.colors.success : theme.colors.error,
                    fontSize: 24,
                    fontWeight: '700',
                  },
                ]}
              >
                {formatCurrency(totalProfit)}
              </Text>
            </View>
          </View>
        </Card>

        {/* Watchlist */}
        <Card elevation="md" style={styles.watchlistCard}>
          <View style={styles.watchlistHeader}>
            <Text style={[styles.sectionTitle, { color: theme.colors.text }]}>Watchlist</Text>
            <TouchableOpacity onPress={() => navigation.navigate('Trading')}>
              <Text style={{ color: theme.colors.primary }}>Trade</Text>
            </TouchableOpacity>
          </View>

          {tickers?.map((ticker) => (
            <TouchableOpacity
              key={ticker.symbol}
              onPress={() => navigation.navigate('Chart', { symbol: ticker.symbol })}
              style={[
                styles.tickerRow,
                { borderBottomColor: theme.colors.border, borderBottomWidth: 1 },
              ]}
            >
              <PriceDisplay
                symbol={ticker.symbol}
                price={ticker.last}
                change24h={ticker.change24h}
                changePercent24h={ticker.changePercent24h}
                size="sm"
              />
            </TouchableOpacity>
          ))}
        </Card>

        {/* Quick Actions */}
        <View style={styles.quickActions}>
          <TouchableOpacity
            style={[
              styles.actionButton,
              { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.md },
            ]}
            onPress={() => navigation.navigate('Deposits')}
          >
            <Text style={{ fontSize: 24 }}>💰</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>Deposit</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[
              styles.actionButton,
              { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.md },
            ]}
            onPress={() => navigation.navigate('Withdrawals')}
          >
            <Text style={{ fontSize: 24 }}>💸</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>Withdraw</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[
              styles.actionButton,
              { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.md },
            ]}
            onPress={() => navigation.navigate('History')}
          >
            <Text style={{ fontSize: 24 }}>📊</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>History</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[
              styles.actionButton,
              { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.md },
            ]}
            onPress={() => navigation.navigate('Settings')}
          >
            <Text style={{ fontSize: 24 }}>⚙️</Text>
            <Text style={[styles.actionText, { color: theme.colors.text }]}>Settings</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  scrollView: {
    padding: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  title: {
    fontWeight: '700',
  },
  accountCard: {
    marginBottom: 16,
  },
  accountType: {
    marginBottom: 8,
  },
  balance: {
    marginBottom: 16,
  },
  accountStats: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  stat: {},
  statLabel: {
    fontSize: 12,
    marginBottom: 4,
  },
  statValue: {
    fontSize: 16,
    fontWeight: '600',
  },
  positionsCard: {
    marginBottom: 16,
  },
  positionsHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
  },
  positionsStats: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  positionStat: {
    alignItems: 'center',
  },
  positionStatLabel: {
    fontSize: 12,
    marginBottom: 8,
  },
  positionStatValue: {},
  watchlistCard: {
    marginBottom: 16,
  },
  watchlistHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  tickerRow: {
    paddingVertical: 12,
  },
  quickActions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  actionButton: {
    width: '48%',
    padding: 16,
    alignItems: 'center',
    marginBottom: 16,
  },
  actionText: {
    marginTop: 8,
    fontSize: 14,
    fontWeight: '600',
  },
});

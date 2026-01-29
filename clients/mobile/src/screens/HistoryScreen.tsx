import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  FlatList,
  RefreshControl,
} from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { useGetAccountsQuery, useGetTradesQuery } from '@/services/api';
import { formatCurrency, formatDateTime } from '@/utils/formatters';
import type { Trade } from '@/types';

export default function HistoryScreen() {
  const theme = useTheme();
  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const { data: trades, isLoading, refetch } = useGetTradesQuery(account?.id || '', {
    skip: !account,
  });

  const renderTrade = ({ item }: { item: Trade }) => {
    const isProfit = item.profit ? item.profit >= 0 : false;

    return (
      <Card elevation="md" style={styles.tradeCard}>
        <View style={styles.tradeHeader}>
          <View>
            <Text
              style={[
                styles.symbol,
                { color: theme.colors.text, fontSize: 16, fontWeight: '700' },
              ]}
            >
              {item.symbol}
            </Text>
            <Text
              style={[
                styles.side,
                {
                  color: item.side === 'BUY' ? theme.colors.buy : theme.colors.sell,
                  fontSize: 14,
                },
              ]}
            >
              {item.side} {item.volume}
            </Text>
          </View>
          {item.profit !== undefined && (
            <Text
              style={[
                styles.profit,
                {
                  color: isProfit ? theme.colors.success : theme.colors.error,
                  fontSize: 18,
                  fontWeight: '700',
                },
              ]}
            >
              {formatCurrency(item.profit)}
            </Text>
          )}
        </View>

        <View style={styles.tradeDetails}>
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Price</Text>
            <Text style={[styles.value, { color: theme.colors.text }]}>
              {formatCurrency(item.price)}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Commission</Text>
            <Text style={[styles.value, { color: theme.colors.text }]}>
              {formatCurrency(item.commission)}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Time</Text>
            <Text style={[styles.value, { color: theme.colors.textSecondary, fontSize: 12 }]}>
              {formatDateTime(item.executedAt)}
            </Text>
          </View>
        </View>
      </Card>
    );
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <View style={styles.header}>
        <Text
          style={[
            styles.title,
            { color: theme.colors.text, fontSize: theme.typography.h2.fontSize },
          ]}
        >
          Trade History
        </Text>
      </View>

      <FlatList
        data={trades}
        keyExtractor={(item) => item.id}
        renderItem={renderTrade}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyText, { color: theme.colors.textSecondary }]}>
              No trade history
            </Text>
          </View>
        }
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    padding: 16,
  },
  title: {
    fontWeight: '700',
  },
  list: {
    padding: 16,
  },
  tradeCard: {
    marginBottom: 16,
  },
  tradeHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 12,
  },
  symbol: {},
  side: {
    marginTop: 4,
  },
  profit: {},
  tradeDetails: {},
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  label: {
    fontSize: 14,
  },
  value: {
    fontSize: 14,
    fontWeight: '600',
  },
  emptyContainer: {
    padding: 32,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
  },
});

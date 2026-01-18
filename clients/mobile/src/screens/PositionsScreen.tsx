import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  FlatList,
  RefreshControl,
  Alert,
} from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { PositionCard } from '@/components/PositionCard';
import {
  useGetAccountsQuery,
  useGetPositionsQuery,
  useClosePositionMutation,
} from '@/services/api';

export default function PositionsScreen() {
  const theme = useTheme();
  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const { data: positions, isLoading, refetch } = useGetPositionsQuery(account?.id || '', {
    skip: !account,
  });
  const [closePosition] = useClosePositionMutation();

  const handleClosePosition = (positionId: string) => {
    if (!account) return;

    Alert.alert(
      'Close Position',
      'Are you sure you want to close this position?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Close',
          style: 'destructive',
          onPress: async () => {
            try {
              await closePosition({ accountId: account.id, positionId }).unwrap();
              Alert.alert('Success', 'Position closed successfully');
            } catch (error: any) {
              Alert.alert('Error', error?.data?.message || 'Failed to close position');
            }
          },
        },
      ],
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
          Positions
        </Text>
      </View>

      <FlatList
        data={positions}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <PositionCard
            position={item}
            onClose={() => handleClosePosition(item.id)}
          />
        )}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyText, { color: theme.colors.textSecondary }]}>
              No open positions
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
  emptyContainer: {
    padding: 32,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
  },
});

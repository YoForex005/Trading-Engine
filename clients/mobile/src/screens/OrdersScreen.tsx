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
import { OrderCard } from '@/components/OrderCard';
import {
  useGetAccountsQuery,
  useGetOrdersQuery,
  useCancelOrderMutation,
} from '@/services/api';

export default function OrdersScreen() {
  const theme = useTheme();
  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const { data: orders, isLoading, refetch } = useGetOrdersQuery(account?.id || '', {
    skip: !account,
  });
  const [cancelOrder] = useCancelOrderMutation();

  const handleCancelOrder = (orderId: string) => {
    if (!account) return;

    Alert.alert(
      'Cancel Order',
      'Are you sure you want to cancel this order?',
      [
        { text: 'No', style: 'cancel' },
        {
          text: 'Yes',
          style: 'destructive',
          onPress: async () => {
            try {
              await cancelOrder({ accountId: account.id, orderId }).unwrap();
              Alert.alert('Success', 'Order cancelled successfully');
            } catch (error: any) {
              Alert.alert('Error', error?.data?.message || 'Failed to cancel order');
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
          Orders
        </Text>
      </View>

      <FlatList
        data={orders}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <OrderCard
            order={item}
            onCancel={() => handleCancelOrder(item.id)}
          />
        )}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyText, { color: theme.colors.textSecondary }]}>
              No orders
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

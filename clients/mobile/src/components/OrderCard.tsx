import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { Card } from './Card';
import { Button } from './Button';
import { Order } from '@/types';
import { formatCurrency, formatDateTime } from '@/utils/formatters';

interface OrderCardProps {
  order: Order;
  onCancel: () => void;
}

export const OrderCard: React.FC<OrderCardProps> = ({ order, onCancel }) => {
  const theme = useTheme();

  const getStatusColor = () => {
    switch (order.status) {
      case 'FILLED':
        return theme.colors.success;
      case 'PENDING':
        return theme.colors.warning;
      case 'CANCELLED':
      case 'REJECTED':
        return theme.colors.error;
      case 'PARTIAL':
        return theme.colors.secondary;
      default:
        return theme.colors.textSecondary;
    }
  };

  const canCancel = order.status === 'PENDING' || order.status === 'PARTIAL';

  return (
    <Card elevation="md" style={styles.card}>
      <View style={styles.header}>
        <View>
          <Text
            style={[
              styles.symbol,
              { color: theme.colors.text, fontSize: 18, fontWeight: '700' },
            ]}
          >
            {order.symbol}
          </Text>
          <Text
            style={[
              styles.type,
              {
                color: theme.colors.textSecondary,
                fontSize: 14,
              },
            ]}
          >
            {order.type} {order.side}
          </Text>
        </View>
        <View style={styles.statusContainer}>
          <Text
            style={[
              styles.status,
              { color: getStatusColor(), fontSize: 14, fontWeight: '600' },
            ]}
          >
            {order.status}
          </Text>
        </View>
      </View>

      <View style={styles.details}>
        <View style={styles.row}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Volume</Text>
          <Text style={[styles.value, { color: theme.colors.text }]}>
            {order.filledVolume} / {order.volume}
          </Text>
        </View>
        {order.price && (
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Price</Text>
            <Text style={[styles.value, { color: theme.colors.text }]}>
              {formatCurrency(order.price)}
            </Text>
          </View>
        )}
        {order.stopPrice && (
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Stop Price</Text>
            <Text style={[styles.value, { color: theme.colors.text }]}>
              {formatCurrency(order.stopPrice)}
            </Text>
          </View>
        )}
        <View style={styles.row}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Created</Text>
          <Text style={[styles.value, { color: theme.colors.textSecondary, fontSize: 12 }]}>
            {formatDateTime(order.createdAt)}
          </Text>
        </View>
      </View>

      {canCancel && (
        <Button
          title="Cancel Order"
          onPress={onCancel}
          variant="outline"
          size="sm"
          style={styles.cancelButton}
        />
      )}
    </Card>
  );
};

const styles = StyleSheet.create({
  card: {
    marginBottom: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  symbol: {},
  type: {
    marginTop: 4,
  },
  statusContainer: {
    alignItems: 'flex-end',
  },
  status: {},
  details: {
    marginBottom: 16,
  },
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
  cancelButton: {},
});

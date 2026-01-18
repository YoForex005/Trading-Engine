import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { Card } from './Card';
import { Button } from './Button';
import { Position } from '@/types';
import { formatCurrency, formatPercent, formatDateTime } from '@/utils/formatters';

interface PositionCardProps {
  position: Position;
  onClose: () => void;
  onEditSL?: () => void;
  onEditTP?: () => void;
}

export const PositionCard: React.FC<PositionCardProps> = ({
  position,
  onClose,
  onEditSL,
  onEditTP,
}) => {
  const theme = useTheme();
  const isProfit = position.profit >= 0;
  const profitColor = isProfit ? theme.colors.success : theme.colors.error;

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
            {position.symbol}
          </Text>
          <Text
            style={[
              styles.side,
              {
                color: position.side === 'BUY' ? theme.colors.buy : theme.colors.sell,
                fontSize: 14,
              },
            ]}
          >
            {position.side} {position.volume}
          </Text>
        </View>
        <View style={styles.profitContainer}>
          <Text style={[styles.profit, { color: profitColor, fontSize: 20, fontWeight: '700' }]}>
            {formatCurrency(position.profit)}
          </Text>
          <Text style={[styles.profitPercent, { color: profitColor, fontSize: 14 }]}>
            {formatPercent(position.profitPercent)}
          </Text>
        </View>
      </View>

      <View style={styles.details}>
        <View style={styles.row}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Open Price</Text>
          <Text style={[styles.value, { color: theme.colors.text }]}>
            {formatCurrency(position.openPrice)}
          </Text>
        </View>
        <View style={styles.row}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Current Price</Text>
          <Text style={[styles.value, { color: theme.colors.text }]}>
            {formatCurrency(position.currentPrice)}
          </Text>
        </View>
        {position.stopLoss && (
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Stop Loss</Text>
            <TouchableOpacity onPress={onEditSL}>
              <Text style={[styles.value, { color: theme.colors.error }]}>
                {formatCurrency(position.stopLoss)}
              </Text>
            </TouchableOpacity>
          </View>
        )}
        {position.takeProfit && (
          <View style={styles.row}>
            <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Take Profit</Text>
            <TouchableOpacity onPress={onEditTP}>
              <Text style={[styles.value, { color: theme.colors.success }]}>
                {formatCurrency(position.takeProfit)}
              </Text>
            </TouchableOpacity>
          </View>
        )}
        <View style={styles.row}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Opened</Text>
          <Text style={[styles.value, { color: theme.colors.textSecondary, fontSize: 12 }]}>
            {formatDateTime(position.openedAt)}
          </Text>
        </View>
      </View>

      <Button
        title="Close Position"
        onPress={onClose}
        variant="error"
        size="sm"
        style={styles.closeButton}
      />
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
  side: {
    marginTop: 4,
  },
  profitContainer: {
    alignItems: 'flex-end',
  },
  profit: {},
  profitPercent: {
    marginTop: 4,
  },
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
  closeButton: {},
});

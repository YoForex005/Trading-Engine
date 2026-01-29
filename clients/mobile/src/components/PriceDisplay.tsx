import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { formatPrice, formatPercent } from '@/utils/formatters';

interface PriceDisplayProps {
  symbol: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  size?: 'sm' | 'md' | 'lg';
}

export const PriceDisplay: React.FC<PriceDisplayProps> = ({
  symbol,
  price,
  change24h,
  changePercent24h,
  size = 'md',
}) => {
  const theme = useTheme();
  const isPositive = change24h >= 0;
  const color = isPositive ? theme.colors.success : theme.colors.error;

  const getFontSize = () => {
    switch (size) {
      case 'sm':
        return { price: 16, change: 12 };
      case 'md':
        return { price: 20, change: 14 };
      case 'lg':
        return { price: 28, change: 16 };
      default:
        return { price: 20, change: 14 };
    }
  };

  const fontSize = getFontSize();

  return (
    <View style={styles.container}>
      <Text
        style={[
          styles.symbol,
          {
            color: theme.colors.textSecondary,
            fontSize: fontSize.change,
          },
        ]}
      >
        {symbol}
      </Text>
      <Text
        style={[
          styles.price,
          {
            color: theme.colors.text,
            fontSize: fontSize.price,
            fontWeight: '700',
          },
        ]}
      >
        {formatPrice(price, symbol)}
      </Text>
      <View style={styles.changeContainer}>
        <Text
          style={[
            styles.change,
            {
              color,
              fontSize: fontSize.change,
              fontWeight: '600',
            },
          ]}
        >
          {formatPercent(changePercent24h)}
        </Text>
        <Text
          style={[
            styles.change,
            {
              color,
              fontSize: fontSize.change,
            },
          ]}
        >
          {' '}
          ({isPositive ? '+' : ''}
          {formatPrice(change24h, symbol)})
        </Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    alignItems: 'flex-start',
  },
  symbol: {
    marginBottom: 4,
  },
  price: {
    marginBottom: 4,
  },
  changeContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  change: {},
});

import React, { useState } from 'react';
import { View, Text, StyleSheet, SafeAreaView, TouchableOpacity, Dimensions } from 'react-native';
import type { RouteProp } from '@react-navigation/native';
import type { StackNavigationProp } from '@react-navigation/stack';
import type { RootStackParamList } from '@/types';
import { useTheme } from '@/hooks/useTheme';
import { Button } from '@/components/Button';

const { width } = Dimensions.get('window');

type ChartScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Chart'>;
  route: RouteProp<RootStackParamList, 'Chart'>;
};

export default function ChartScreen({ navigation, route }: ChartScreenProps) {
  const theme = useTheme();
  const { symbol } = route.params;
  const [timeframe, setTimeframe] = useState('1h');

  const timeframes = ['1m', '5m', '15m', '1h', '4h', '1d'];

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={{ color: theme.colors.primary, fontSize: 18 }}>← Back</Text>
        </TouchableOpacity>
        <Text
          style={[
            styles.symbol,
            { color: theme.colors.text, fontSize: theme.typography.h3.fontSize },
          ]}
        >
          {symbol}
        </Text>
        <View style={{ width: 60 }} />
      </View>

      {/* Timeframe Selector */}
      <View style={styles.timeframeSelector}>
        {timeframes.map((tf) => (
          <TouchableOpacity
            key={tf}
            onPress={() => setTimeframe(tf)}
            style={[
              styles.timeframeButton,
              {
                backgroundColor:
                  timeframe === tf ? theme.colors.primary : theme.colors.surface,
                borderRadius: theme.borderRadius.sm,
              },
            ]}
          >
            <Text
              style={[
                styles.timeframeText,
                { color: timeframe === tf ? '#FFF' : theme.colors.text },
              ]}
            >
              {tf}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Chart Placeholder */}
      <View
        style={[
          styles.chartContainer,
          { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.md },
        ]}
      >
        <Text style={{ color: theme.colors.textSecondary, textAlign: 'center' }}>
          Chart for {symbol} - {timeframe}
          {'\n\n'}
          (Integrate react-native-charts-wrapper or TradingView)
        </Text>
      </View>

      {/* Quick Trade Buttons */}
      <View style={styles.tradeButtons}>
        <Button
          title="BUY"
          onPress={() => navigation.navigate('Trading', { symbol })}
          variant="success"
          style={{ flex: 1, marginRight: 8 }}
        />
        <Button
          title="SELL"
          onPress={() => navigation.navigate('Trading', { symbol })}
          variant="error"
          style={{ flex: 1, marginLeft: 8 }}
        />
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
  },
  symbol: {
    fontWeight: '700',
  },
  timeframeSelector: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    marginBottom: 16,
  },
  timeframeButton: {
    flex: 1,
    paddingVertical: 8,
    alignItems: 'center',
    marginHorizontal: 4,
  },
  timeframeText: {
    fontSize: 12,
    fontWeight: '600',
  },
  chartContainer: {
    flex: 1,
    margin: 16,
    padding: 16,
    justifyContent: 'center',
    alignItems: 'center',
  },
  tradeButtons: {
    flexDirection: 'row',
    padding: 16,
  },
});

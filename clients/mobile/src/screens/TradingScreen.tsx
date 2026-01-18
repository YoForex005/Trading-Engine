import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { useSelector } from 'react-redux';
import type { StackNavigationProp } from '@react-navigation/stack';
import type { RootStackParamList, OrderSide, OrderType } from '@/types';
import type { RootState } from '@/store';
import { useTheme } from '@/hooks/useTheme';
import { useHaptic } from '@/hooks/useHaptic';
import { Card } from '@/components/Card';
import { Button } from '@/components/Button';
import { Input } from '@/components/Input';
import { useGetAccountsQuery, useGetTickerQuery, usePlaceOrderMutation } from '@/services/api';
import { formatPrice, formatCurrency } from '@/utils/formatters';

type TradingScreenProps = {
  navigation: StackNavigationProp<RootStackParamList, 'Trading'>;
  route: { params?: { symbol?: string } };
};

export default function TradingScreen({ navigation, route }: TradingScreenProps) {
  const theme = useTheme();
  const { trigger } = useHaptic();
  const watchlist = useSelector((state: RootState) => state.market.watchlist);

  const [selectedSymbol, setSelectedSymbol] = useState(route.params?.symbol || watchlist[0]);
  const [side, setSide] = useState<OrderSide>('BUY');
  const [orderType, setOrderType] = useState<OrderType>('MARKET');
  const [volume, setVolume] = useState('1.0');
  const [price, setPrice] = useState('');
  const [stopLoss, setStopLoss] = useState('');
  const [takeProfit, setTakeProfit] = useState('');

  const { data: accounts } = useGetAccountsQuery();
  const account = accounts?.[0];
  const { data: ticker } = useGetTickerQuery(selectedSymbol, {
    pollingInterval: 1000,
  });
  const [placeOrder, { isLoading }] = usePlaceOrderMutation();

  const handlePlaceOrder = async () => {
    if (!account) {
      Alert.alert('Error', 'No account selected');
      return;
    }

    const volumeNum = parseFloat(volume);
    if (isNaN(volumeNum) || volumeNum <= 0) {
      Alert.alert('Error', 'Invalid volume');
      return;
    }

    trigger('impactMedium');

    try {
      await placeOrder({
        accountId: account.id,
        symbol: selectedSymbol,
        side,
        type: orderType,
        volume: volumeNum,
        price: price ? parseFloat(price) : undefined,
        stopLoss: stopLoss ? parseFloat(stopLoss) : undefined,
        takeProfit: takeProfit ? parseFloat(takeProfit) : undefined,
      }).unwrap();

      Alert.alert('Success', 'Order placed successfully');
      trigger('notificationSuccess');

      // Reset form
      setVolume('1.0');
      setPrice('');
      setStopLoss('');
      setTakeProfit('');
    } catch (error: any) {
      Alert.alert('Error', error?.data?.message || 'Failed to place order');
      trigger('notificationError');
    }
  };

  const currentPrice = side === 'BUY' ? ticker?.ask : ticker?.bid;
  const estimatedCost = currentPrice && parseFloat(volume) ? currentPrice * parseFloat(volume) : 0;

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
            Trade
          </Text>
        </View>

        {/* Symbol Selection */}
        <Card elevation="md" style={styles.symbolCard}>
          <Text style={[styles.label, { color: theme.colors.textSecondary }]}>Select Symbol</Text>
          <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.symbolScroll}>
            {watchlist.map((symbol) => (
              <TouchableOpacity
                key={symbol}
                onPress={() => setSelectedSymbol(symbol)}
                style={[
                  styles.symbolButton,
                  {
                    backgroundColor:
                      selectedSymbol === symbol ? theme.colors.primary : theme.colors.surface,
                    borderRadius: theme.borderRadius.md,
                  },
                ]}
              >
                <Text
                  style={[
                    styles.symbolText,
                    { color: selectedSymbol === symbol ? '#FFF' : theme.colors.text },
                  ]}
                >
                  {symbol}
                </Text>
              </TouchableOpacity>
            ))}
          </ScrollView>
        </Card>

        {/* Price Display */}
        {ticker && (
          <Card elevation="md" style={styles.priceCard}>
            <View style={styles.priceRow}>
              <View style={styles.priceColumn}>
                <Text style={[styles.priceLabel, { color: theme.colors.textSecondary }]}>BID</Text>
                <Text
                  style={[
                    styles.priceValue,
                    { color: theme.colors.sell, fontSize: 24, fontWeight: '700' },
                  ]}
                >
                  {formatPrice(ticker.bid, selectedSymbol)}
                </Text>
              </View>
              <View style={styles.priceColumn}>
                <Text style={[styles.priceLabel, { color: theme.colors.textSecondary }]}>ASK</Text>
                <Text
                  style={[
                    styles.priceValue,
                    { color: theme.colors.buy, fontSize: 24, fontWeight: '700' },
                  ]}
                >
                  {formatPrice(ticker.ask, selectedSymbol)}
                </Text>
              </View>
            </View>
          </Card>
        )}

        {/* Order Form */}
        <Card elevation="md" style={styles.orderCard}>
          {/* Side Selection */}
          <View style={styles.sideSelector}>
            <TouchableOpacity
              onPress={() => setSide('BUY')}
              style={[
                styles.sideButton,
                {
                  backgroundColor: side === 'BUY' ? theme.colors.buy : theme.colors.surface,
                  borderRadius: theme.borderRadius.md,
                },
              ]}
            >
              <Text style={[styles.sideText, { color: side === 'BUY' ? '#FFF' : theme.colors.text }]}>
                BUY
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              onPress={() => setSide('SELL')}
              style={[
                styles.sideButton,
                {
                  backgroundColor: side === 'SELL' ? theme.colors.sell : theme.colors.surface,
                  borderRadius: theme.borderRadius.md,
                },
              ]}
            >
              <Text style={[styles.sideText, { color: side === 'SELL' ? '#FFF' : theme.colors.text }]}>
                SELL
              </Text>
            </TouchableOpacity>
          </View>

          {/* Order Type */}
          <View style={styles.typeSelector}>
            {(['MARKET', 'LIMIT', 'STOP'] as OrderType[]).map((type) => (
              <TouchableOpacity
                key={type}
                onPress={() => setOrderType(type)}
                style={[
                  styles.typeButton,
                  {
                    backgroundColor:
                      orderType === type ? theme.colors.primary : theme.colors.surface,
                    borderRadius: theme.borderRadius.sm,
                  },
                ]}
              >
                <Text
                  style={[
                    styles.typeText,
                    { color: orderType === type ? '#FFF' : theme.colors.text },
                  ]}
                >
                  {type}
                </Text>
              </TouchableOpacity>
            ))}
          </View>

          <Input
            label="Volume"
            placeholder="1.0"
            value={volume}
            onChangeText={setVolume}
            keyboardType="decimal-pad"
          />

          {orderType !== 'MARKET' && (
            <Input
              label="Price"
              placeholder="Enter price"
              value={price}
              onChangeText={setPrice}
              keyboardType="decimal-pad"
            />
          )}

          <Input
            label="Stop Loss (Optional)"
            placeholder="Enter stop loss"
            value={stopLoss}
            onChangeText={setStopLoss}
            keyboardType="decimal-pad"
          />

          <Input
            label="Take Profit (Optional)"
            placeholder="Enter take profit"
            value={takeProfit}
            onChangeText={setTakeProfit}
            keyboardType="decimal-pad"
          />

          {/* Estimated Cost */}
          <View style={styles.estimatedCost}>
            <Text style={[styles.costLabel, { color: theme.colors.textSecondary }]}>
              Estimated Cost
            </Text>
            <Text style={[styles.costValue, { color: theme.colors.text, fontSize: 18, fontWeight: '600' }]}>
              {formatCurrency(estimatedCost)}
            </Text>
          </View>

          <Button
            title={`${side} ${selectedSymbol}`}
            onPress={handlePlaceOrder}
            variant={side === 'BUY' ? 'success' : 'error'}
            loading={isLoading}
            fullWidth
          />
        </Card>

        {/* Chart Button */}
        <Button
          title="View Chart"
          onPress={() => navigation.navigate('Chart', { symbol: selectedSymbol })}
          variant="outline"
          fullWidth
          style={{ marginTop: 16 }}
        />
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
  symbolCard: {
    marginBottom: 16,
  },
  label: {
    fontSize: 12,
    marginBottom: 8,
  },
  symbolScroll: {
    marginTop: 8,
  },
  symbolButton: {
    paddingVertical: 8,
    paddingHorizontal: 16,
    marginRight: 8,
  },
  symbolText: {
    fontSize: 14,
    fontWeight: '600',
  },
  priceCard: {
    marginBottom: 16,
  },
  priceRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  priceColumn: {
    alignItems: 'center',
  },
  priceLabel: {
    fontSize: 12,
    marginBottom: 8,
  },
  priceValue: {},
  orderCard: {
    marginBottom: 16,
  },
  sideSelector: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  sideButton: {
    flex: 1,
    paddingVertical: 12,
    alignItems: 'center',
    marginHorizontal: 4,
  },
  sideText: {
    fontSize: 16,
    fontWeight: '700',
  },
  typeSelector: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  typeButton: {
    flex: 1,
    paddingVertical: 8,
    alignItems: 'center',
    marginHorizontal: 4,
  },
  typeText: {
    fontSize: 14,
    fontWeight: '600',
  },
  estimatedCost: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
    paddingVertical: 12,
  },
  costLabel: {
    fontSize: 14,
  },
  costValue: {},
});

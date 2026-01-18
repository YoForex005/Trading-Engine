import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  FlatList,
  TouchableOpacity,
  Alert,
  Modal,
} from 'react-native';
import { useSelector } from 'react-redux';
import type { RootState } from '@/store';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { Button } from '@/components/Button';
import { Input } from '@/components/Input';
import {
  useGetAlertsQuery,
  useCreateAlertMutation,
  useDeleteAlertMutation,
} from '@/services/api';
import { formatPrice } from '@/utils/formatters';

export default function AlertsScreen() {
  const theme = useTheme();
  const watchlist = useSelector((state: RootState) => state.market.watchlist);
  const { data: alerts, refetch } = useGetAlertsQuery();
  const [createAlert] = useCreateAlertMutation();
  const [deleteAlert] = useDeleteAlertMutation();

  const [showModal, setShowModal] = useState(false);
  const [symbol, setSymbol] = useState(watchlist[0]);
  const [condition, setCondition] = useState<'ABOVE' | 'BELOW'>('ABOVE');
  const [price, setPrice] = useState('');

  const handleCreateAlert = async () => {
    const priceNum = parseFloat(price);
    if (isNaN(priceNum) || priceNum <= 0) {
      Alert.alert('Error', 'Invalid price');
      return;
    }

    try {
      await createAlert({
        symbol,
        condition,
        price: priceNum,
        active: true,
      }).unwrap();

      Alert.alert('Success', 'Price alert created');
      setShowModal(false);
      setPrice('');
    } catch (error: any) {
      Alert.alert('Error', error?.data?.message || 'Failed to create alert');
    }
  };

  const handleDeleteAlert = (id: string) => {
    Alert.alert('Delete Alert', 'Are you sure you want to delete this alert?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: async () => {
          try {
            await deleteAlert(id).unwrap();
            Alert.alert('Success', 'Alert deleted');
          } catch (error: any) {
            Alert.alert('Error', error?.data?.message || 'Failed to delete alert');
          }
        },
      },
    ]);
  };

  const renderAlert = ({ item }: { item: any }) => (
    <Card elevation="md" style={styles.alertCard}>
      <View style={styles.alertHeader}>
        <View>
          <Text style={[styles.symbol, { color: theme.colors.text, fontSize: 18, fontWeight: '700' }]}>
            {item.symbol}
          </Text>
          <Text style={[styles.condition, { color: theme.colors.textSecondary }]}>
            {item.condition} {formatPrice(item.price, item.symbol)}
          </Text>
        </View>
        <TouchableOpacity onPress={() => handleDeleteAlert(item.id)}>
          <Text style={{ fontSize: 24 }}>🗑️</Text>
        </TouchableOpacity>
      </View>
      {item.triggered && (
        <Text style={[styles.triggered, { color: theme.colors.success }]}>✓ Triggered</Text>
      )}
    </Card>
  );

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.colors.background }]}>
      <View style={styles.header}>
        <Text
          style={[
            styles.title,
            { color: theme.colors.text, fontSize: theme.typography.h2.fontSize },
          ]}
        >
          Price Alerts
        </Text>
        <TouchableOpacity onPress={() => setShowModal(true)}>
          <Text style={{ color: theme.colors.primary, fontSize: 28 }}>+</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={alerts}
        keyExtractor={(item) => item.id}
        renderItem={renderAlert}
        contentContainerStyle={styles.list}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyText, { color: theme.colors.textSecondary }]}>
              No price alerts
            </Text>
          </View>
        }
      />

      {/* Create Alert Modal */}
      <Modal visible={showModal} animationType="slide" transparent>
        <View style={styles.modalOverlay}>
          <View
            style={[
              styles.modalContent,
              { backgroundColor: theme.colors.surface, borderRadius: theme.borderRadius.lg },
            ]}
          >
            <Text
              style={[
                styles.modalTitle,
                { color: theme.colors.text, fontSize: theme.typography.h3.fontSize },
              ]}
            >
              Create Price Alert
            </Text>

            <Input label="Symbol" value={symbol} onChangeText={setSymbol} />

            <View style={styles.conditionSelector}>
              <TouchableOpacity
                onPress={() => setCondition('ABOVE')}
                style={[
                  styles.conditionButton,
                  {
                    backgroundColor:
                      condition === 'ABOVE' ? theme.colors.primary : theme.colors.surfaceVariant,
                    borderRadius: theme.borderRadius.md,
                  },
                ]}
              >
                <Text
                  style={[
                    styles.conditionText,
                    { color: condition === 'ABOVE' ? '#FFF' : theme.colors.text },
                  ]}
                >
                  Above
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={() => setCondition('BELOW')}
                style={[
                  styles.conditionButton,
                  {
                    backgroundColor:
                      condition === 'BELOW' ? theme.colors.primary : theme.colors.surfaceVariant,
                    borderRadius: theme.borderRadius.md,
                  },
                ]}
              >
                <Text
                  style={[
                    styles.conditionText,
                    { color: condition === 'BELOW' ? '#FFF' : theme.colors.text },
                  ]}
                >
                  Below
                </Text>
              </TouchableOpacity>
            </View>

            <Input
              label="Price"
              placeholder="Enter price"
              value={price}
              onChangeText={setPrice}
              keyboardType="decimal-pad"
            />

            <View style={styles.modalButtons}>
              <Button
                title="Cancel"
                onPress={() => setShowModal(false)}
                variant="outline"
                style={{ flex: 1, marginRight: 8 }}
              />
              <Button
                title="Create"
                onPress={handleCreateAlert}
                style={{ flex: 1, marginLeft: 8 }}
              />
            </View>
          </View>
        </View>
      </Modal>
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
  title: {
    fontWeight: '700',
  },
  list: {
    padding: 16,
  },
  alertCard: {
    marginBottom: 12,
  },
  alertHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  symbol: {},
  condition: {
    marginTop: 4,
    fontSize: 14,
  },
  triggered: {
    marginTop: 8,
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
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    padding: 16,
  },
  modalContent: {
    padding: 24,
  },
  modalTitle: {
    fontWeight: '700',
    marginBottom: 24,
  },
  conditionSelector: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  conditionButton: {
    flex: 1,
    paddingVertical: 12,
    alignItems: 'center',
    marginHorizontal: 4,
  },
  conditionText: {
    fontSize: 16,
    fontWeight: '600',
  },
  modalButtons: {
    flexDirection: 'row',
    marginTop: 16,
  },
});

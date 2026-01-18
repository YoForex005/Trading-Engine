import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  FlatList,
  TouchableOpacity,
  RefreshControl,
} from 'react-native';
import { useTheme } from '@/hooks/useTheme';
import { Card } from '@/components/Card';
import { useGetNotificationsQuery, useMarkNotificationReadMutation } from '@/services/api';
import { formatRelativeTime } from '@/utils/formatters';
import type { Notification } from '@/types';

export default function NotificationsScreen() {
  const theme = useTheme();
  const { data: notifications, isLoading, refetch } = useGetNotificationsQuery();
  const [markRead] = useMarkNotificationReadMutation();

  const handleNotificationPress = async (notification: Notification) => {
    if (!notification.read) {
      await markRead(notification.id);
    }
  };

  const renderNotification = ({ item }: { item: Notification }) => {
    const getIcon = () => {
      switch (item.type) {
        case 'ORDER_FILLED':
          return '✅';
        case 'MARGIN_CALL':
          return '⚠️';
        case 'ALERT_TRIGGERED':
          return '🔔';
        case 'NEWS':
          return '📰';
        case 'SYSTEM':
          return 'ℹ️';
        default:
          return '📬';
      }
    };

    return (
      <TouchableOpacity onPress={() => handleNotificationPress(item)}>
        <Card
          elevation="md"
          style={[
            styles.notificationCard,
            !item.read && { backgroundColor: theme.colors.surfaceVariant },
          ]}
        >
          <View style={styles.notificationHeader}>
            <Text style={{ fontSize: 24 }}>{getIcon()}</Text>
            <Text style={[styles.timestamp, { color: theme.colors.textSecondary }]}>
              {formatRelativeTime(item.createdAt)}
            </Text>
          </View>
          <Text
            style={[
              styles.notificationTitle,
              { color: theme.colors.text, fontWeight: item.read ? '400' : '700' },
            ]}
          >
            {item.title}
          </Text>
          <Text style={[styles.notificationMessage, { color: theme.colors.textSecondary }]}>
            {item.message}
          </Text>
        </Card>
      </TouchableOpacity>
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
          Notifications
        </Text>
      </View>

      <FlatList
        data={notifications}
        keyExtractor={(item) => item.id}
        renderItem={renderNotification}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyText, { color: theme.colors.textSecondary }]}>
              No notifications
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
  notificationCard: {
    marginBottom: 12,
  },
  notificationHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  timestamp: {
    fontSize: 12,
  },
  notificationTitle: {
    fontSize: 16,
    marginBottom: 4,
  },
  notificationMessage: {
    fontSize: 14,
    lineHeight: 20,
  },
  emptyContainer: {
    padding: 32,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
  },
});

import messaging from '@react-native-firebase/messaging';
import { Platform } from 'react-native';

class NotificationService {
  private fcmToken: string | null = null;

  async initialize(): Promise<void> {
    // Request permission
    const authStatus = await messaging().requestPermission();
    const enabled =
      authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
      authStatus === messaging.AuthorizationStatus.PROVISIONAL;

    if (!enabled) {
      console.log('Push notification permission denied');
      return;
    }

    // Get FCM token
    this.fcmToken = await messaging().getToken();
    console.log('FCM Token:', this.fcmToken);

    // Listen for token refresh
    messaging().onTokenRefresh(token => {
      this.fcmToken = token;
      console.log('FCM Token refreshed:', token);
      // Send new token to backend
      this.updateTokenOnServer(token);
    });

    // Handle foreground messages
    messaging().onMessage(async remoteMessage => {
      console.log('Foreground message:', remoteMessage);
      this.handleNotification(remoteMessage);
    });

    // Handle background messages
    messaging().setBackgroundMessageHandler(async remoteMessage => {
      console.log('Background message:', remoteMessage);
    });
  }

  async getToken(): Promise<string | null> {
    if (!this.fcmToken) {
      this.fcmToken = await messaging().getToken();
    }
    return this.fcmToken;
  }

  private async updateTokenOnServer(token: string): Promise<void> {
    // Send token to your backend
    try {
      // await api.post('/user/fcm-token', { token });
    } catch (error) {
      console.error('Error updating FCM token:', error);
    }
  }

  private handleNotification(remoteMessage: any): void {
    const { notification, data } = remoteMessage;

    // Handle different notification types
    if (data?.type === 'ORDER_FILLED') {
      // Handle order filled notification
    } else if (data?.type === 'MARGIN_CALL') {
      // Handle margin call notification
    } else if (data?.type === 'ALERT_TRIGGERED') {
      // Handle price alert notification
    }
  }

  async subscribeToTopic(topic: string): Promise<void> {
    await messaging().subscribeToTopic(topic);
  }

  async unsubscribeFromTopic(topic: string): Promise<void> {
    await messaging().unsubscribeFromTopic(topic);
  }
}

export const notificationService = new NotificationService();

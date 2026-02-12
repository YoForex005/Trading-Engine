/**
 * Notifications Hook
 * Connects to WebSocket for real-time notifications and manages toast display
 */

import { useEffect, useState, useCallback } from 'react';
import { useNotificationStore } from '../store/useNotificationStore';
import type { NotificationSeverity, NotificationType } from '../store/useNotificationStore';
import { API_ENDPOINTS } from '../config/api';

interface UseNotificationsProps {
  wsConnection: WebSocket | null;
}

export const useNotifications = ({ wsConnection }: UseNotificationsProps) => {
  const { addNotification } = useNotificationStore();
  const [activeToasts, setActiveToasts] = useState<any[]>([]);

  // Play sound for critical notifications (optional)
  const playNotificationSound = useCallback(() => {
    // Browser notification sound
    const audio = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBTGH0fPTgjMGHm7A7+OZUQ0PVKvn77BfGAg+ltvy0H0pBSl+zPLaizsIGGS56+OZUQ4OUKXi8LaHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsIGWO56+OaUA0OUKXi8LWHMwU2jdXzz38qBSl9y/Lai0YIDVmw6+6cUw8LTqPf87acPQU0itbyynspBCt7y+/aijsI');
    audio.volume = 0.3;
    audio.play().catch(() => {
      // Ignore errors if browser blocks audio
    });
  }, []);

  // Show toast notification
  const showToast = useCallback((notification: any) => {
    const toastId = notification.id;

    setActiveToasts(prev => [...prev, {
      id: toastId,
      severity: notification.severity,
      title: notification.title,
      message: notification.message,
      onDismiss: () => {
        setActiveToasts(current => current.filter(t => t.id !== toastId));
      }
    }]);

    // Play sound for critical notifications
    if (notification.severity === 'critical') {
      playNotificationSound();
    }
  }, [playNotificationSound]);

  // WebSocket message handler
  useEffect(() => {
    if (!wsConnection) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);

        // Handle notification messages
        if (data.type === 'notification') {
          const notification = {
            type: data.notificationType || 'system',
            severity: data.severity || 'info',
            title: data.title,
            message: data.message,
            data: data.data,
          };

          addNotification(notification);
          showToast({ ...notification, id: Date.now().toString() });
        }
      } catch (error) {
        console.error('[Notifications] Failed to parse WS message:', error);
      }
    };

    wsConnection.addEventListener('message', handleMessage);

    return () => {
      wsConnection.removeEventListener('message', handleMessage);
    };
  }, [wsConnection, addNotification, showToast]);

  // Fetch initial unread count on mount
  useEffect(() => {
    const fetchUnreadCount = async () => {
      try {
        // This endpoint would return { count: number }
        const response = await fetch(API_ENDPOINTS.notifications.unreadCount);
        if (response.ok) {
          const data = await response.json();
          console.log('[Notifications] Initial unread count:', data.count);
        }
      } catch (error) {
        console.log('[Notifications] Failed to fetch unread count:', error);
      }
    };

    fetchUnreadCount();
  }, []);

  return {
    activeToasts,
  };
};

/**
 * Generate Mock Notifications for Testing
 */
export const generateMockNotifications = () => {
  const mockNotifications: Array<{
    type: NotificationType;
    severity: NotificationSeverity;
    title: string;
    message: string;
  }> = [
    // Trading notifications
    {
      type: 'trading',
      severity: 'info',
      title: 'Order Filled',
      message: 'BUY 0.5 EURUSD @ 1.0850 executed successfully',
    },
    {
      type: 'trading',
      severity: 'info',
      title: 'Take Profit Hit',
      message: 'Position #12345 closed at TP: +$125.50',
    },
    {
      type: 'trading',
      severity: 'warning',
      title: 'Margin Warning',
      message: 'Margin level at 85%. Consider reducing position sizes.',
    },
    {
      type: 'trading',
      severity: 'error',
      title: 'Stop Loss Triggered',
      message: 'SELL 1.0 GBPUSD position closed at SL: -$45.20',
    },
    {
      type: 'trading',
      severity: 'critical',
      title: 'Margin Call',
      message: 'Urgent: Margin level below 50%. Deposit funds or close positions immediately.',
    },

    // Account notifications
    {
      type: 'account',
      severity: 'info',
      title: 'Deposit Confirmed',
      message: 'Your deposit of $5,000 has been successfully processed',
    },
    {
      type: 'account',
      severity: 'info',
      title: 'Withdrawal Processed',
      message: 'Withdrawal of $1,200 completed. Funds will arrive in 2-3 business days.',
    },
    {
      type: 'account',
      severity: 'warning',
      title: 'Account Verification',
      message: 'Your account verification documents are about to expire. Please update.',
    },

    // Security notifications
    {
      type: 'security',
      severity: 'warning',
      title: 'New Login Detected',
      message: 'New login from Chrome on Windows 11 - IP: 192.168.1.100',
    },
    {
      type: 'security',
      severity: 'info',
      title: 'Password Changed',
      message: 'Your password was successfully updated 5 minutes ago',
    },
    {
      type: 'security',
      severity: 'critical',
      title: 'Suspicious Activity',
      message: 'Multiple failed login attempts detected. Account temporarily locked.',
    },

    // System notifications
    {
      type: 'system',
      severity: 'error',
      title: 'LP OANDA Disconnected',
      message: 'Lost connection to OANDA liquidity provider. Reconnecting...',
    },
    {
      type: 'system',
      severity: 'info',
      title: 'Platform Update',
      message: 'RTX5 platform updated to version 2.1.4. New features available.',
    },
    {
      type: 'system',
      severity: 'warning',
      title: 'Scheduled Maintenance',
      message: 'Platform maintenance scheduled for tonight 2:00-4:00 AM UTC',
    },
    {
      type: 'system',
      severity: 'info',
      title: 'Market Holiday',
      message: 'US markets closed tomorrow for Independence Day',
    },
  ];

  return mockNotifications;
};

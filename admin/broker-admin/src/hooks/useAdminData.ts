/**
 * useAdminData Hook
 * Real-time admin data management (MT5 Manager parity)
 * Replaces polling with WebSocket subscriptions
 * PERFORMANCE FIX: WebSocket updates instead of 1-second polling
 */

import { useState, useEffect, useCallback } from 'react';
import {
  getAdminWebSocket,
  type Account,
  type Order,
  type AdminEvent,
} from '../services/adminWebSocket';
import { api } from '../services/apiClient';
import { API_CONFIG } from '../config/api';

const WS_URL = API_CONFIG.ADMIN_WS_URL;

interface UseAdminDataOptions {
  autoConnect?: boolean;
  fallbackToPolling?: boolean;
  pollingInterval?: number;
}

interface AdminData {
  accounts: Account[];
  orders: Order[];
  connectionState: 'connecting' | 'connected' | 'disconnected' | 'error';
  lastUpdate: number;
}

/**
 * Hook for real-time admin data
 * @param options - Configuration options
 * @returns Admin data and connection state
 */
export function useAdminData(options: UseAdminDataOptions = {}) {
  const {
    autoConnect = true,
    fallbackToPolling = true,
    pollingInterval = 2000,
  } = options;

  const [data, setData] = useState<AdminData>({
    accounts: [],
    orders: [],
    connectionState: 'disconnected',
    lastUpdate: Date.now(),
  });

  // Initialize WebSocket connection
  useEffect(() => {
    if (!autoConnect) return;

    const ws = getAdminWebSocket(WS_URL);
    ws.connect();

    // Subscribe to connection state
    const unsubState = ws.onStateChange((state) => {
      setData((prev) => ({ ...prev, connectionState: state }));
    });

    // Subscribe to account updates
    const unsubAccounts = ws.on('account_update', (event: AdminEvent) => {
      setData((prev) => {
        const updatedAccount = event.data as Account;
        const existingIndex = prev.accounts.findIndex(
          (a) => a.id === updatedAccount.id || a.login === updatedAccount.login
        );

        let newAccounts;
        if (existingIndex >= 0) {
          // Update existing account
          newAccounts = [...prev.accounts];
          newAccounts[existingIndex] = { ...newAccounts[existingIndex], ...updatedAccount };
        } else {
          // Add new account
          newAccounts = [...prev.accounts, updatedAccount];
        }

        return {
          ...prev,
          accounts: newAccounts,
          lastUpdate: Date.now(),
        };
      });
    });

    // Subscribe to order updates
    const unsubOrders = ws.on('order_new', (event: AdminEvent) => {
      setData((prev) => ({
        ...prev,
        orders: [...prev.orders, event.data as Order],
        lastUpdate: Date.now(),
      }));
    });

    const unsubOrderModify = ws.on('order_modify', (event: AdminEvent) => {
      setData((prev) => {
        const updatedOrder = event.data as Order;
        const newOrders = prev.orders.map((o) =>
          o.id === updatedOrder.id ? { ...o, ...updatedOrder } : o
        );
        return {
          ...prev,
          orders: newOrders,
          lastUpdate: Date.now(),
        };
      });
    });

    const unsubOrderClose = ws.on('order_close', (event: AdminEvent) => {
      setData((prev) => ({
        ...prev,
        orders: prev.orders.filter((o) => o.id !== event.data.id),
        lastUpdate: Date.now(),
      }));
    });

    // Cleanup
    return () => {
      unsubState();
      unsubAccounts();
      unsubOrders();
      unsubOrderModify();
      unsubOrderClose();
      ws.disconnect();
    };
  }, [autoConnect]);

  // Fallback polling if WebSocket disconnected
  useEffect(() => {
    if (!fallbackToPolling) return;
    if (data.connectionState === 'connected') return; // WebSocket active, no need to poll

    const fetchData = async () => {
      try {
        // Fetch accounts with authentication
        const accounts = await api.get<Account[]>(API_CONFIG.ADMIN_ACCOUNTS);
        setData((prev) => ({
          ...prev,
          accounts: accounts || [],
          lastUpdate: Date.now(),
        }));

        // Fetch orders with authentication
        const orders = await api.get<Order[]>(API_CONFIG.ADMIN_ORDERS);
        setData((prev) => ({
          ...prev,
          orders: orders || [],
          lastUpdate: Date.now(),
        }));
      } catch (error) {
        console.error('[useAdminData] Polling error:', error);
      }
    };

    // Initial fetch
    fetchData();

    // Set up polling interval
    const interval = setInterval(fetchData, pollingInterval);
    return () => clearInterval(interval);
  }, [fallbackToPolling, data.connectionState, pollingInterval]);

  // Manual refresh function
  const refresh = useCallback(async () => {
    try {
      const [accounts, orders] = await Promise.all([
        api.get<Account[]>(API_CONFIG.ADMIN_ACCOUNTS).catch(() => []),
        api.get<Order[]>(API_CONFIG.ADMIN_ORDERS).catch(() => []),
      ]);

      setData((prev) => ({
        ...prev,
        accounts,
        orders,
        lastUpdate: Date.now(),
      }));
    } catch (error) {
      console.error('[useAdminData] Refresh error:', error);
    }
  }, []);

  return {
    ...data,
    refresh,
    isConnected: data.connectionState === 'connected',
    isLoading: data.connectionState === 'connecting',
  };
}

/**
 * Hook for connection status indicator
 */
export function useAdminConnectionStatus() {
  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>(
    'disconnected'
  );

  useEffect(() => {
    try {
      const ws = getAdminWebSocket();
      return ws.onStateChange(setStatus);
    } catch {
      // WebSocket not initialized yet
      return () => {};
    }
  }, []);

  return status;
}

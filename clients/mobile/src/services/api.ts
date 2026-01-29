import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
  User,
  Account,
  Position,
  Order,
  Ticker,
  Trade,
  Alert,
  Notification,
  NewsItem,
  Deposit,
  Withdrawal,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  PlaceOrderRequest,
  ApiResponse,
} from '@/types';
import { secureStorage } from '@/utils/security';

const API_BASE_URL = __DEV__
  ? 'http://localhost:8080/api/v1'
  : 'https://api.tradingengine.com/api/v1';

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: API_BASE_URL,
    prepareHeaders: async (headers) => {
      const token = await secureStorage.getItem('authToken');
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      headers.set('Content-Type', 'application/json');
      return headers;
    },
  }),
  tagTypes: [
    'User',
    'Account',
    'Positions',
    'Orders',
    'Trades',
    'Alerts',
    'Notifications',
    'News',
    'Deposits',
    'Withdrawals',
  ],
  endpoints: (builder) => ({
    // Auth
    login: builder.mutation<LoginResponse, LoginRequest>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      invalidatesTags: ['User'],
    }),
    register: builder.mutation<LoginResponse, RegisterRequest>({
      query: (userData) => ({
        url: '/auth/register',
        method: 'POST',
        body: userData,
      }),
    }),
    logout: builder.mutation<void, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
      invalidatesTags: ['User', 'Account', 'Positions', 'Orders'],
    }),

    // User
    getProfile: builder.query<User, void>({
      query: () => '/user/profile',
      providesTags: ['User'],
    }),
    updateProfile: builder.mutation<User, Partial<User>>({
      query: (updates) => ({
        url: '/user/profile',
        method: 'PATCH',
        body: updates,
      }),
      invalidatesTags: ['User'],
    }),

    // Account
    getAccounts: builder.query<Account[], void>({
      query: () => '/accounts',
      providesTags: ['Account'],
    }),
    getAccount: builder.query<Account, string>({
      query: (id) => `/accounts/${id}`,
      providesTags: ['Account'],
    }),

    // Positions
    getPositions: builder.query<Position[], string>({
      query: (accountId) => `/accounts/${accountId}/positions`,
      providesTags: ['Positions'],
    }),
    closePosition: builder.mutation<void, { accountId: string; positionId: string }>({
      query: ({ accountId, positionId }) => ({
        url: `/accounts/${accountId}/positions/${positionId}/close`,
        method: 'POST',
      }),
      invalidatesTags: ['Positions', 'Account'],
    }),
    updatePositionSL: builder.mutation<void, {
      accountId: string;
      positionId: string;
      stopLoss: number;
    }>({
      query: ({ accountId, positionId, stopLoss }) => ({
        url: `/accounts/${accountId}/positions/${positionId}/sl`,
        method: 'PATCH',
        body: { stopLoss },
      }),
      invalidatesTags: ['Positions'],
    }),
    updatePositionTP: builder.mutation<void, {
      accountId: string;
      positionId: string;
      takeProfit: number;
    }>({
      query: ({ accountId, positionId, takeProfit }) => ({
        url: `/accounts/${accountId}/positions/${positionId}/tp`,
        method: 'PATCH',
        body: { takeProfit },
      }),
      invalidatesTags: ['Positions'],
    }),

    // Orders
    getOrders: builder.query<Order[], string>({
      query: (accountId) => `/accounts/${accountId}/orders`,
      providesTags: ['Orders'],
    }),
    placeOrder: builder.mutation<Order, PlaceOrderRequest & { accountId: string }>({
      query: ({ accountId, ...order }) => ({
        url: `/accounts/${accountId}/orders`,
        method: 'POST',
        body: order,
      }),
      invalidatesTags: ['Orders', 'Account'],
    }),
    cancelOrder: builder.mutation<void, { accountId: string; orderId: string }>({
      query: ({ accountId, orderId }) => ({
        url: `/accounts/${accountId}/orders/${orderId}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Orders'],
    }),

    // Market Data
    getTicker: builder.query<Ticker, string>({
      query: (symbol) => `/market/ticker/${symbol}`,
    }),
    getTickers: builder.query<Ticker[], string[]>({
      query: (symbols) => `/market/tickers?symbols=${symbols.join(',')}`,
    }),

    // Trades
    getTrades: builder.query<Trade[], string>({
      query: (accountId) => `/accounts/${accountId}/trades`,
      providesTags: ['Trades'],
    }),

    // Alerts
    getAlerts: builder.query<Alert[], void>({
      query: () => '/alerts',
      providesTags: ['Alerts'],
    }),
    createAlert: builder.mutation<Alert, Omit<Alert, 'id' | 'createdAt' | 'triggered'>>({
      query: (alert) => ({
        url: '/alerts',
        method: 'POST',
        body: alert,
      }),
      invalidatesTags: ['Alerts'],
    }),
    deleteAlert: builder.mutation<void, string>({
      query: (id) => ({
        url: `/alerts/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Alerts'],
    }),

    // Notifications
    getNotifications: builder.query<Notification[], void>({
      query: () => '/notifications',
      providesTags: ['Notifications'],
    }),
    markNotificationRead: builder.mutation<void, string>({
      query: (id) => ({
        url: `/notifications/${id}/read`,
        method: 'POST',
      }),
      invalidatesTags: ['Notifications'],
    }),

    // News
    getNews: builder.query<NewsItem[], { category?: string; limit?: number }>({
      query: ({ category, limit = 20 }) => {
        const params = new URLSearchParams();
        if (category) params.append('category', category);
        params.append('limit', limit.toString());
        return `/news?${params.toString()}`;
      },
      providesTags: ['News'],
    }),

    // Deposits
    getDeposits: builder.query<Deposit[], string>({
      query: (accountId) => `/accounts/${accountId}/deposits`,
      providesTags: ['Deposits'],
    }),
    createDeposit: builder.mutation<Deposit, Omit<Deposit, 'id' | 'createdAt' | 'status'>>({
      query: (deposit) => ({
        url: `/accounts/${deposit.accountId}/deposits`,
        method: 'POST',
        body: deposit,
      }),
      invalidatesTags: ['Deposits'],
    }),

    // Withdrawals
    getWithdrawals: builder.query<Withdrawal[], string>({
      query: (accountId) => `/accounts/${accountId}/withdrawals`,
      providesTags: ['Withdrawals'],
    }),
    createWithdrawal: builder.mutation<Withdrawal, Omit<Withdrawal, 'id' | 'createdAt' | 'status'>>({
      query: (withdrawal) => ({
        url: `/accounts/${withdrawal.accountId}/withdrawals`,
        method: 'POST',
        body: withdrawal,
      }),
      invalidatesTags: ['Withdrawals'],
    }),
  }),
});

export const {
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useGetProfileQuery,
  useUpdateProfileMutation,
  useGetAccountsQuery,
  useGetAccountQuery,
  useGetPositionsQuery,
  useClosePositionMutation,
  useUpdatePositionSLMutation,
  useUpdatePositionTPMutation,
  useGetOrdersQuery,
  usePlaceOrderMutation,
  useCancelOrderMutation,
  useGetTickerQuery,
  useGetTickersQuery,
  useGetTradesQuery,
  useGetAlertsQuery,
  useCreateAlertMutation,
  useDeleteAlertMutation,
  useGetNotificationsQuery,
  useMarkNotificationReadMutation,
  useGetNewsQuery,
  useGetDepositsQuery,
  useCreateDepositMutation,
  useGetWithdrawalsQuery,
  useCreateWithdrawalMutation,
} = api;

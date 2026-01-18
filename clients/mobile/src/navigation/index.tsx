import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { useSelector } from 'react-redux';
import type { RootState } from '@/store';
import type { RootStackParamList } from '@/types';

// Screens
import LoginScreen from '@/screens/LoginScreen';
import RegisterScreen from '@/screens/RegisterScreen';
import DashboardScreen from '@/screens/DashboardScreen';
import TradingScreen from '@/screens/TradingScreen';
import ChartScreen from '@/screens/ChartScreen';
import PositionsScreen from '@/screens/PositionsScreen';
import OrdersScreen from '@/screens/OrdersScreen';
import HistoryScreen from '@/screens/HistoryScreen';
import AccountScreen from '@/screens/AccountScreen';
import DepositsScreen from '@/screens/DepositsScreen';
import WithdrawalsScreen from '@/screens/WithdrawalsScreen';
import SettingsScreen from '@/screens/SettingsScreen';
import NotificationsScreen from '@/screens/NotificationsScreen';
import AlertsScreen from '@/screens/AlertsScreen';

const Stack = createStackNavigator<RootStackParamList>();
const Tab = createBottomTabNavigator();

function AuthStack() {
  return (
    <Stack.Navigator
      screenOptions={{
        headerShown: false,
      }}
    >
      <Stack.Screen name="Login" component={LoginScreen} />
      <Stack.Screen name="Register" component={RegisterScreen} />
    </Stack.Navigator>
  );
}

function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarStyle: {
          borderTopWidth: 1,
        },
      }}
    >
      <Tab.Screen
        name="Dashboard"
        component={DashboardScreen}
        options={{
          tabBarLabel: 'Home',
        }}
      />
      <Tab.Screen
        name="Trading"
        component={TradingScreen}
        options={{
          tabBarLabel: 'Trade',
        }}
      />
      <Tab.Screen
        name="Positions"
        component={PositionsScreen}
        options={{
          tabBarLabel: 'Positions',
        }}
      />
      <Tab.Screen
        name="Account"
        component={AccountScreen}
        options={{
          tabBarLabel: 'Account',
        }}
      />
    </Tab.Navigator>
  );
}

export function AppNavigator() {
  const isAuthenticated = useSelector((state: RootState) => state.auth.isAuthenticated);

  return (
    <NavigationContainer>
      <Stack.Navigator
        screenOptions={{
          headerShown: false,
        }}
      >
        {!isAuthenticated ? (
          <Stack.Screen name="Auth" component={AuthStack} />
        ) : (
          <>
            <Stack.Screen name="Main" component={MainTabs} />
            <Stack.Screen name="Chart" component={ChartScreen} />
            <Stack.Screen name="Orders" component={OrdersScreen} />
            <Stack.Screen name="History" component={HistoryScreen} />
            <Stack.Screen name="Deposits" component={DepositsScreen} />
            <Stack.Screen name="Withdrawals" component={WithdrawalsScreen} />
            <Stack.Screen name="Settings" component={SettingsScreen} />
            <Stack.Screen name="Notifications" component={NotificationsScreen} />
            <Stack.Screen name="Alerts" component={AlertsScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}

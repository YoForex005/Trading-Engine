import React, { useEffect } from 'react';
import { StatusBar, AppState, Appearance } from 'react-native';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { store, persistor } from './store';
import { AppNavigator } from './navigation';
import { notificationService } from './services/notifications';
import { websocketService } from './services/websocket';
import { SessionManager, secureStorage } from './utils/security';
import { setIsDarkMode } from './store/slices/uiSlice';
import { logout } from './store/slices/authSlice';

function App(): JSX.Element {
  useEffect(() => {
    // Initialize notifications
    notificationService.initialize();

    // Connect WebSocket
    const initializeWebSocket = async () => {
      const token = await secureStorage.getItem('authToken');
      if (token) {
        websocketService.connect(token);
      }
    };
    initializeWebSocket();

    // Session timeout
    SessionManager.start(() => {
      store.dispatch(logout());
    });

    // Listen for app state changes
    const subscription = AppState.addEventListener('change', (nextAppState) => {
      if (nextAppState === 'active') {
        SessionManager.reset();
      }
    });

    // Listen for theme changes
    const themeSubscription = Appearance.addChangeListener(({ colorScheme }) => {
      const state = store.getState();
      if (state.ui.themeMode === 'auto') {
        store.dispatch(setIsDarkMode(colorScheme === 'dark'));
      }
    });

    return () => {
      subscription.remove();
      themeSubscription.remove();
      SessionManager.stop();
      websocketService.disconnect();
    };
  }, []);

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <Provider store={store}>
        <PersistGate loading={null} persistor={persistor}>
          <StatusBar barStyle="light-content" />
          <AppNavigator />
        </PersistGate>
      </Provider>
    </GestureHandlerRootView>
  );
}

export default App;

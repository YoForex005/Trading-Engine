import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { getWebSocketService } from '@/services/websocket';
import { Layout } from '@/components/Layout';
import { SystemOverview } from '@/components/monitoring/SystemOverview';
import { OrderFlowMonitor } from '@/components/monitoring/OrderFlowMonitor';
import { LPHealthMonitor } from '@/components/monitoring/LPHealthMonitor';
import { UserActivityMonitor } from '@/components/monitoring/UserActivityMonitor';
import { ErrorDashboard } from '@/components/monitoring/ErrorDashboard';
import { RoutingControl } from '@/components/controls/RoutingControl';
import { LPManagement } from '@/components/controls/LPManagement';
import { SymbolManagement } from '@/components/controls/SymbolManagement';
import { UserManagement } from '@/components/controls/UserManagement';
import { RiskControl } from '@/components/controls/RiskControl';
import { PerformanceAnalytics } from '@/components/analytics/PerformanceAnalytics';
import { TradingAnalytics } from '@/components/analytics/TradingAnalytics';
import { UserAnalytics } from '@/components/analytics/UserAnalytics';
import { AuditTrail } from '@/components/analytics/AuditTrail';
import { Login } from '@/components/Login';

export const App = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [wsConnected, setWsConnected] = useState(false);
  const [darkMode, setDarkMode] = useState(() => {
    return localStorage.getItem('darkMode') === 'true';
  });

  useEffect(() => {
    // Check authentication
    const token = localStorage.getItem('auth_token');
    setIsAuthenticated(!!token);

    // Apply dark mode
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  useEffect(() => {
    if (!isAuthenticated) return;

    // Connect WebSocket
    const ws = getWebSocketService();

    ws.onConnect(() => {
      console.log('WebSocket connected');
      setWsConnected(true);
    });

    ws.onDisconnect(() => {
      console.log('WebSocket disconnected');
      setWsConnected(false);
    });

    ws.connect().catch(err => {
      console.error('WebSocket connection failed:', err);
    });

    return () => {
      ws.disconnect();
    };
  }, [isAuthenticated]);

  const toggleDarkMode = () => {
    setDarkMode(prev => {
      const newValue = !prev;
      localStorage.setItem('darkMode', String(newValue));
      return newValue;
    });
  };

  const handleLogin = () => {
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    setIsAuthenticated(false);
  };

  if (!isAuthenticated) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <BrowserRouter>
      <Layout
        darkMode={darkMode}
        onToggleDarkMode={toggleDarkMode}
        wsConnected={wsConnected}
        onLogout={handleLogout}
      >
        <Routes>
          {/* Monitoring */}
          <Route path="/" element={<Navigate to="/monitoring/system" replace />} />
          <Route path="/monitoring/system" element={<SystemOverview />} />
          <Route path="/monitoring/orders" element={<OrderFlowMonitor />} />
          <Route path="/monitoring/lp" element={<LPHealthMonitor />} />
          <Route path="/monitoring/users" element={<UserActivityMonitor />} />
          <Route path="/monitoring/errors" element={<ErrorDashboard />} />

          {/* Controls */}
          <Route path="/controls/routing" element={<RoutingControl />} />
          <Route path="/controls/lp" element={<LPManagement />} />
          <Route path="/controls/symbols" element={<SymbolManagement />} />
          <Route path="/controls/users" element={<UserManagement />} />
          <Route path="/controls/risk" element={<RiskControl />} />

          {/* Analytics */}
          <Route path="/analytics/performance" element={<PerformanceAnalytics />} />
          <Route path="/analytics/trading" element={<TradingAnalytics />} />
          <Route path="/analytics/users" element={<UserAnalytics />} />
          <Route path="/analytics/audit" element={<AuditTrail />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
};

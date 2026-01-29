import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { Table } from '@/components/shared/Table';
import { Badge } from '@/components/shared/Badge';
import { Button } from '@/components/shared/Button';
import { Users, Clock, TrendingUp } from 'lucide-react';
import { getWebSocketService } from '@/services/websocket';
import { api } from '@/services/api';
import { format, formatDistanceToNow } from 'date-fns';
import type { UserSession } from '@/types';

export const UserActivityMonitor = () => {
  const [sessions, setSessions] = useState<UserSession[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSessions();

    const ws = getWebSocketService();
    const unsubscribe = ws.on<UserSession[]>('user_activity', (data) => {
      setSessions(data);
    });

    return () => unsubscribe();
  }, []);

  const loadSessions = async () => {
    try {
      setLoading(true);
      const data = await api.getUserSessions();
      setSessions(data);
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTerminate = async (sessionId: string) => {
    if (!confirm('Are you sure you want to terminate this session?')) return;

    try {
      await api.terminateSession(sessionId);
      await loadSessions();
    } catch (error) {
      console.error('Failed to terminate session:', error);
    }
  };

  const columns = [
    {
      key: 'username',
      header: 'User',
      sortable: true,
      render: (session: UserSession) => (
        <div>
          <div className="font-semibold">{session.username}</div>
          <div className="text-xs text-gray-500 font-mono">{session.userId.slice(0, 8)}</div>
        </div>
      ),
    },
    {
      key: 'sessionId',
      header: 'Session ID',
      render: (session: UserSession) => (
        <span className="font-mono text-xs">{session.sessionId.slice(0, 12)}</span>
      ),
    },
    {
      key: 'loginTime',
      header: 'Login Time',
      sortable: true,
      render: (session: UserSession) => format(session.loginTime, 'MMM dd, HH:mm:ss'),
    },
    {
      key: 'lastActivity',
      header: 'Last Activity',
      sortable: true,
      render: (session: UserSession) => formatDistanceToNow(session.lastActivity, { addSuffix: true }),
    },
    {
      key: 'ipAddress',
      header: 'IP Address',
      render: (session: UserSession) => (
        <span className="font-mono text-sm">{session.ipAddress}</span>
      ),
    },
    {
      key: 'activeOrders',
      header: 'Active Orders',
      sortable: true,
      render: (session: UserSession) => (
        <Badge variant={session.activeOrders > 0 ? 'info' : 'default'}>
          {session.activeOrders}
        </Badge>
      ),
    },
    {
      key: 'totalOrders',
      header: 'Total Orders',
      sortable: true,
    },
    {
      key: 'pnl',
      header: 'P&L',
      sortable: true,
      render: (session: UserSession) => (
        <span className={session.pnl >= 0 ? 'text-green-600' : 'text-red-600'}>
          ${session.pnl.toFixed(2)}
        </span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (session: UserSession) => (
        <Button
          size="sm"
          variant="danger"
          onClick={() => handleTerminate(session.sessionId)}
        >
          Terminate
        </Button>
      ),
    },
  ];

  const stats = {
    total: sessions.length,
    avgOrders: sessions.reduce((sum, s) => sum + s.totalOrders, 0) / (sessions.length || 1),
    totalPnL: sessions.reduce((sum, s) => sum + s.pnl, 0),
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Active Users</div>
              <div className="text-2xl font-semibold text-gray-900 dark:text-white mt-1">
                {stats.total}
              </div>
            </div>
            <Users className="w-8 h-8 text-blue-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Avg Orders/User</div>
              <div className="text-2xl font-semibold text-gray-900 dark:text-white mt-1">
                {stats.avgOrders.toFixed(1)}
              </div>
            </div>
            <TrendingUp className="w-8 h-8 text-purple-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Total P&L</div>
              <div className={`text-2xl font-semibold mt-1 ${stats.totalPnL >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                ${stats.totalPnL.toFixed(2)}
              </div>
            </div>
            <Clock className="w-8 h-8 text-orange-600" />
          </div>
        </Card>
      </div>

      <Card title="Active User Sessions">
        <Table
          data={sessions}
          columns={columns}
          loading={loading}
          emptyMessage="No active sessions"
        />
      </Card>
    </div>
  );
};

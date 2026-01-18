import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { Badge } from '@/components/shared/Badge';
import { Button } from '@/components/shared/Button';
import { Table } from '@/components/shared/Table';
import { Activity, TrendingUp, AlertTriangle, CheckCircle } from 'lucide-react';
import { getWebSocketService } from '@/services/websocket';
import { api } from '@/services/api';
import type { LiquidityProvider, LPStatus } from '@/types';

const statusColors: Record<LPStatus, 'success' | 'warning' | 'error'> = {
  connected: 'success',
  degraded: 'warning',
  disconnected: 'error',
};

const statusIcons: Record<LPStatus, React.ReactNode> = {
  connected: <CheckCircle className="w-4 h-4" />,
  degraded: <AlertTriangle className="w-4 h-4" />,
  disconnected: <AlertTriangle className="w-4 h-4" />,
};

export const LPHealthMonitor = () => {
  const [lps, setLps] = useState<LiquidityProvider[]>([]);
  const [loading, setLoading] = useState(true);
  const [testing, setTesting] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadLPs();

    const ws = getWebSocketService();
    const unsubscribe = ws.on<LiquidityProvider>('lp_status', (lp) => {
      setLps(prev => {
        const index = prev.findIndex(l => l.id === lp.id);
        if (index >= 0) {
          const updated = [...prev];
          updated[index] = lp;
          return updated;
        }
        return [...prev, lp];
      });
    });

    return () => unsubscribe();
  }, []);

  const loadLPs = async () => {
    try {
      setLoading(true);
      const data = await api.getLPs();
      setLps(data);
    } catch (error) {
      console.error('Failed to load LPs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async (lpId: string) => {
    setTesting(prev => new Set(prev).add(lpId));
    try {
      const result = await api.testLPConnection(lpId);
      console.log('Test result:', result);
    } catch (error) {
      console.error('Test failed:', error);
    } finally {
      setTesting(prev => {
        const updated = new Set(prev);
        updated.delete(lpId);
        return updated;
      });
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'LP Name',
      sortable: true,
      render: (lp: LiquidityProvider) => (
        <div className="flex items-center gap-2">
          <span className="font-semibold">{lp.name}</span>
          {statusIcons[lp.status]}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (lp: LiquidityProvider) => (
        <Badge variant={statusColors[lp.status]}>
          {lp.status.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'latency',
      header: 'Latency',
      sortable: true,
      render: (lp: LiquidityProvider) => (
        <div>
          <span className={lp.latency > 100 ? 'text-red-600' : 'text-green-600'}>
            {lp.latency}ms
          </span>
          <div className="text-xs text-gray-500">
            Avg: {lp.metrics.averageLatency.toFixed(0)}ms
          </div>
        </div>
      ),
    },
    {
      key: 'uptime',
      header: 'Uptime',
      sortable: true,
      render: (lp: LiquidityProvider) => (
        <div>
          <span>{lp.uptime.toFixed(2)}%</span>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5 mt-1">
            <div
              className="bg-green-600 h-1.5 rounded-full"
              style={{ width: `${lp.uptime}%` }}
            />
          </div>
        </div>
      ),
    },
    {
      key: 'fillRate',
      header: 'Fill Rate',
      sortable: true,
      render: (lp: LiquidityProvider) => (
        <div>
          <span>{(lp.metrics.fillRate * 100).toFixed(1)}%</span>
          <div className="text-xs text-gray-500">
            Rejection: {(lp.metrics.rejectionRate * 100).toFixed(1)}%
          </div>
        </div>
      ),
    },
    {
      key: 'ordersRouted',
      header: 'Orders Routed',
      sortable: true,
      render: (lp: LiquidityProvider) => lp.metrics.ordersRouted.toLocaleString(),
    },
    {
      key: 'activeSymbols',
      header: 'Active Symbols',
      render: (lp: LiquidityProvider) => (
        <span className="text-sm">{lp.activeSymbols.length} symbols</span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (lp: LiquidityProvider) => (
        <Button
          size="sm"
          variant="ghost"
          onClick={() => handleTest(lp.id)}
          loading={testing.has(lp.id)}
        >
          Test Connection
        </Button>
      ),
    },
  ];

  const stats = {
    total: lps.length,
    connected: lps.filter(lp => lp.status === 'connected').length,
    degraded: lps.filter(lp => lp.status === 'degraded').length,
    disconnected: lps.filter(lp => lp.status === 'disconnected').length,
    avgLatency: lps.reduce((sum, lp) => sum + lp.latency, 0) / (lps.length || 1),
    avgFillRate: lps.reduce((sum, lp) => sum + lp.metrics.fillRate, 0) / (lps.length || 1),
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Connected LPs</div>
              <div className="text-2xl font-semibold text-green-600 dark:text-green-400 mt-1">
                {stats.connected}/{stats.total}
              </div>
            </div>
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Avg Latency</div>
              <div className="text-2xl font-semibold text-gray-900 dark:text-white mt-1">
                {stats.avgLatency.toFixed(0)}ms
              </div>
            </div>
            <Activity className="w-8 h-8 text-blue-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Avg Fill Rate</div>
              <div className="text-2xl font-semibold text-gray-900 dark:text-white mt-1">
                {(stats.avgFillRate * 100).toFixed(1)}%
              </div>
            </div>
            <TrendingUp className="w-8 h-8 text-purple-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Issues</div>
              <div className="text-2xl font-semibold text-orange-600 dark:text-orange-400 mt-1">
                {stats.degraded + stats.disconnected}
              </div>
            </div>
            <AlertTriangle className="w-8 h-8 text-orange-600" />
          </div>
        </Card>
      </div>

      <Card title="Liquidity Provider Status">
        <Table
          data={lps}
          columns={columns}
          loading={loading}
          emptyMessage="No liquidity providers configured"
        />
      </Card>
    </div>
  );
};

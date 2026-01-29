import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Cpu, HardDrive, Activity, Users } from 'lucide-react';
import { getWebSocketService } from '@/services/websocket';
import { api } from '@/services/api';
import type { SystemMetrics } from '@/types';

type MetricCardProps = {
  title: string;
  value: string;
  change?: string;
  icon: React.ReactNode;
  trend?: 'up' | 'down' | 'neutral';
};

const MetricCard = ({ title, value, change, icon, trend = 'neutral' }: MetricCardProps) => {
  const trendColors = {
    up: 'text-green-600',
    down: 'text-red-600',
    neutral: 'text-gray-600',
  };

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
          <p className="mt-2 text-3xl font-semibold text-gray-900 dark:text-white">{value}</p>
          {change && (
            <p className={`mt-2 text-sm ${trendColors[trend]} dark:opacity-90`}>
              {change}
            </p>
          )}
        </div>
        <div className="ml-4 p-3 bg-primary-100 dark:bg-primary-900 rounded-lg">
          {icon}
        </div>
      </div>
    </Card>
  );
};

export const SystemOverview = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [history, setHistory] = useState<Array<SystemMetrics & { time: string }>>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const ws = getWebSocketService();

    // Initial load
    api.getSystemMetrics()
      .then(data => {
        setMetrics(data);
        setHistory([{
          ...data,
          time: new Date(data.timestamp).toLocaleTimeString(),
        }]);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });

    // Real-time updates
    const unsubscribe = ws.on<SystemMetrics>('system_metrics', (data) => {
      setMetrics(data);
      setHistory(prev => {
        const newHistory = [
          ...prev,
          {
            ...data,
            time: new Date(data.timestamp).toLocaleTimeString(),
          },
        ];
        // Keep last 20 data points
        return newHistory.slice(-20);
      });
    });

    return () => {
      unsubscribe();
    };
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <p className="text-red-800 dark:text-red-200">Error loading metrics: {error}</p>
      </div>
    );
  }

  if (!metrics) return null;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="CPU Usage"
          value={`${metrics.cpu.usage.toFixed(1)}%`}
          change={`${metrics.cpu.cores} cores`}
          icon={<Cpu className="w-6 h-6 text-primary-600" />}
          trend={metrics.cpu.usage > 80 ? 'down' : 'neutral'}
        />
        <MetricCard
          title="Memory Usage"
          value={`${metrics.memory.percentage.toFixed(1)}%`}
          change={`${(metrics.memory.used / 1024 / 1024 / 1024).toFixed(2)} GB / ${(metrics.memory.total / 1024 / 1024 / 1024).toFixed(2)} GB`}
          icon={<HardDrive className="w-6 h-6 text-primary-600" />}
          trend={metrics.memory.percentage > 85 ? 'down' : 'neutral'}
        />
        <MetricCard
          title="Active Connections"
          value={metrics.connections.active.toLocaleString()}
          change={`${metrics.connections.total.toLocaleString()} total`}
          icon={<Users className="w-6 h-6 text-primary-600" />}
          trend="neutral"
        />
        <MetricCard
          title="Orders/sec"
          value={metrics.throughput.ordersPerSecond.toFixed(1)}
          change={`${metrics.throughput.messagesPerSecond.toFixed(1)} msg/s`}
          icon={<Activity className="w-6 h-6 text-primary-600" />}
          trend="up"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card title="CPU Usage Trend">
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={history}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
              <XAxis
                dataKey="time"
                className="text-gray-600 dark:text-gray-400"
              />
              <YAxis
                domain={[0, 100]}
                className="text-gray-600 dark:text-gray-400"
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(255, 255, 255, 0.95)',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                }}
              />
              <Area
                type="monotone"
                dataKey="cpu.usage"
                stroke="#0ea5e9"
                fill="#0ea5e9"
                fillOpacity={0.3}
              />
            </AreaChart>
          </ResponsiveContainer>
        </Card>

        <Card title="Memory Usage Trend">
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={history}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
              <XAxis
                dataKey="time"
                className="text-gray-600 dark:text-gray-400"
              />
              <YAxis
                domain={[0, 100]}
                className="text-gray-600 dark:text-gray-400"
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(255, 255, 255, 0.95)',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                }}
              />
              <Area
                type="monotone"
                dataKey="memory.percentage"
                stroke="#10b981"
                fill="#10b981"
                fillOpacity={0.3}
              />
            </AreaChart>
          </ResponsiveContainer>
        </Card>

        <Card title="Throughput Trend">
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={history}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
              <XAxis
                dataKey="time"
                className="text-gray-600 dark:text-gray-400"
              />
              <YAxis className="text-gray-600 dark:text-gray-400" />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(255, 255, 255, 0.95)',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                }}
              />
              <Line
                type="monotone"
                dataKey="throughput.ordersPerSecond"
                stroke="#8b5cf6"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </Card>

        <Card title="Active Connections Trend">
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={history}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
              <XAxis
                dataKey="time"
                className="text-gray-600 dark:text-gray-400"
              />
              <YAxis className="text-gray-600 dark:text-gray-400" />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(255, 255, 255, 0.95)',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                }}
              />
              <Line
                type="monotone"
                dataKey="connections.active"
                stroke="#f59e0b"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </Card>
      </div>
    </div>
  );
};

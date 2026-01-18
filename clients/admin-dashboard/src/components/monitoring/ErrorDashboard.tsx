import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { Table } from '@/components/shared/Table';
import { Badge } from '@/components/shared/Badge';
import { AlertTriangle, AlertCircle, Info, XCircle } from 'lucide-react';
import { getWebSocketService } from '@/services/websocket';
import { api } from '@/services/api';
import { format } from 'date-fns';
import type { ErrorLog, ErrorSeverity, Sort } from '@/types';

const severityColors: Record<ErrorSeverity, 'info' | 'warning' | 'error' | 'default'> = {
  info: 'info',
  warning: 'warning',
  error: 'error',
  critical: 'error',
};

const severityIcons: Record<ErrorSeverity, React.ReactNode> = {
  info: <Info className="w-4 h-4" />,
  warning: <AlertCircle className="w-4 h-4" />,
  error: <AlertTriangle className="w-4 h-4" />,
  critical: <XCircle className="w-4 h-4" />,
};

export const ErrorDashboard = () => {
  const [errors, setErrors] = useState<ErrorLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [sort, setSort] = useState<Sort>({ field: 'timestamp', direction: 'desc' });
  const [filterSeverity, setFilterSeverity] = useState<ErrorSeverity | 'all'>('all');

  useEffect(() => {
    loadErrors();

    const ws = getWebSocketService();
    const unsubscribe = ws.on<ErrorLog>('error_log', (error) => {
      setErrors(prev => [error, ...prev].slice(0, 100));
    });

    return () => unsubscribe();
  }, []);

  const loadErrors = async () => {
    try {
      setLoading(true);
      const response = await api.getErrorLogs(
        filterSeverity !== 'all' ? [{ field: 'severity', operator: 'eq', value: filterSeverity }] : undefined,
        { page: 1, pageSize: 100, total: 0 }
      );
      setErrors(response.logs);
    } catch (error) {
      console.error('Failed to load errors:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadErrors();
  }, [filterSeverity]);

  const columns = [
    {
      key: 'timestamp',
      header: 'Time',
      sortable: true,
      render: (error: ErrorLog) => format(error.timestamp, 'MMM dd, HH:mm:ss'),
    },
    {
      key: 'severity',
      header: 'Severity',
      sortable: true,
      render: (error: ErrorLog) => (
        <div className="flex items-center gap-2">
          {severityIcons[error.severity]}
          <Badge variant={severityColors[error.severity]}>
            {error.severity.toUpperCase()}
          </Badge>
        </div>
      ),
    },
    {
      key: 'component',
      header: 'Component',
      sortable: true,
      render: (error: ErrorLog) => (
        <span className="font-mono text-sm">{error.component}</span>
      ),
    },
    {
      key: 'message',
      header: 'Message',
      render: (error: ErrorLog) => (
        <div className="max-w-md truncate" title={error.message}>
          {error.message}
        </div>
      ),
    },
    {
      key: 'userId',
      header: 'User ID',
      render: (error: ErrorLog) => error.userId ? (
        <span className="font-mono text-xs">{error.userId.slice(0, 8)}</span>
      ) : (
        <span className="text-gray-400">-</span>
      ),
    },
    {
      key: 'orderId',
      header: 'Order ID',
      render: (error: ErrorLog) => error.orderId ? (
        <span className="font-mono text-xs">{error.orderId.slice(0, 8)}</span>
      ) : (
        <span className="text-gray-400">-</span>
      ),
    },
  ];

  const stats = {
    total: errors.length,
    critical: errors.filter(e => e.severity === 'critical').length,
    error: errors.filter(e => e.severity === 'error').length,
    warning: errors.filter(e => e.severity === 'warning').length,
    info: errors.filter(e => e.severity === 'info').length,
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Critical</div>
              <div className="text-2xl font-semibold text-red-600 dark:text-red-400 mt-1">
                {stats.critical}
              </div>
            </div>
            <XCircle className="w-8 h-8 text-red-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Errors</div>
              <div className="text-2xl font-semibold text-orange-600 dark:text-orange-400 mt-1">
                {stats.error}
              </div>
            </div>
            <AlertTriangle className="w-8 h-8 text-orange-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Warnings</div>
              <div className="text-2xl font-semibold text-yellow-600 dark:text-yellow-400 mt-1">
                {stats.warning}
              </div>
            </div>
            <AlertCircle className="w-8 h-8 text-yellow-600" />
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Info</div>
              <div className="text-2xl font-semibold text-blue-600 dark:text-blue-400 mt-1">
                {stats.info}
              </div>
            </div>
            <Info className="w-8 h-8 text-blue-600" />
          </div>
        </Card>
      </div>

      <Card
        title="Error Logs"
        actions={
          <select
            value={filterSeverity}
            onChange={(e) => setFilterSeverity(e.target.value as ErrorSeverity | 'all')}
            className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          >
            <option value="all">All Severity</option>
            <option value="critical">Critical</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>
        }
      >
        <Table
          data={errors}
          columns={columns}
          loading={loading}
          sort={sort}
          onSort={(field) => setSort(prev => ({
            field,
            direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
          }))}
          emptyMessage="No errors logged"
        />
      </Card>
    </div>
  );
};

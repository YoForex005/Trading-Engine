import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { Table } from '@/components/shared/Table';
import { Badge } from '@/components/shared/Badge';
import { Button } from '@/components/shared/Button';
import { Download, Filter as FilterIcon } from 'lucide-react';
import { getWebSocketService } from '@/services/websocket';
import { api } from '@/services/api';
import { format } from 'date-fns';
import type { Order, OrderStatus, Sort } from '@/types';

const statusColors: Record<OrderStatus, 'success' | 'warning' | 'error' | 'info' | 'default'> = {
  filled: 'success',
  partial: 'warning',
  pending: 'info',
  rejected: 'error',
  cancelled: 'default',
};

export const OrderFlowMonitor = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [sort, setSort] = useState<Sort>({ field: 'timestamp', direction: 'desc' });
  const [filterStatus, setFilterStatus] = useState<OrderStatus | 'all'>('all');

  useEffect(() => {
    loadOrders();

    const ws = getWebSocketService();
    const unsubscribe = ws.on<Order>('order_update', (order) => {
      setOrders(prev => {
        const index = prev.findIndex(o => o.id === order.id);
        if (index >= 0) {
          const updated = [...prev];
          updated[index] = order;
          return updated;
        }
        return [order, ...prev].slice(0, 100); // Keep last 100 orders
      });
    });

    return () => unsubscribe();
  }, []);

  const loadOrders = async () => {
    try {
      setLoading(true);
      const response = await api.getOrders(
        filterStatus !== 'all' ? [{ field: 'status', operator: 'eq', value: filterStatus }] : undefined,
        sort,
        { page: 1, pageSize: 100, total: 0 }
      );
      setOrders(response.orders);
    } catch (error) {
      console.error('Failed to load orders:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, [sort, filterStatus]);

  const handleSort = (field: string) => {
    setSort(prev => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  };

  const handleExport = async () => {
    try {
      const blob = await api.exportData('orders', 'csv',
        filterStatus !== 'all' ? [{ field: 'status', operator: 'eq', value: filterStatus }] : undefined
      );
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `orders-${Date.now()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
    }
  };

  const columns = [
    {
      key: 'timestamp',
      header: 'Time',
      sortable: true,
      render: (order: Order) => format(order.timestamp, 'HH:mm:ss'),
    },
    {
      key: 'id',
      header: 'Order ID',
      render: (order: Order) => (
        <span className="font-mono text-xs">{order.id.slice(0, 8)}</span>
      ),
    },
    {
      key: 'userId',
      header: 'User',
      render: (order: Order) => (
        <span className="font-mono text-xs">{order.userId.slice(0, 8)}</span>
      ),
    },
    {
      key: 'symbol',
      header: 'Symbol',
      sortable: true,
      render: (order: Order) => (
        <span className="font-semibold">{order.symbol}</span>
      ),
    },
    {
      key: 'side',
      header: 'Side',
      render: (order: Order) => (
        <Badge variant={order.side === 'buy' ? 'success' : 'error'}>
          {order.side.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      render: (order: Order) => (
        <span className="text-sm capitalize">{order.type.replace('_', ' ')}</span>
      ),
    },
    {
      key: 'quantity',
      header: 'Quantity',
      sortable: true,
      render: (order: Order) => order.quantity.toFixed(4),
    },
    {
      key: 'price',
      header: 'Price',
      sortable: true,
      render: (order: Order) => order.price ? `$${order.price.toFixed(2)}` : '-',
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (order: Order) => (
        <Badge variant={statusColors[order.status]}>
          {order.status.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'filledQuantity',
      header: 'Filled',
      render: (order: Order) => {
        const percentage = (order.filledQuantity / order.quantity) * 100;
        return (
          <div>
            <div className="text-sm">{order.filledQuantity.toFixed(4)}</div>
            <div className="text-xs text-gray-500">{percentage.toFixed(1)}%</div>
          </div>
        );
      },
    },
    {
      key: 'lpRoute',
      header: 'LP Route',
      render: (order: Order) => order.lpRoute ? (
        <span className="text-xs font-mono">{order.lpRoute}</span>
      ) : (
        <span className="text-xs text-gray-400">-</span>
      ),
    },
  ];

  const stats = {
    total: orders.length,
    pending: orders.filter(o => o.status === 'pending').length,
    filled: orders.filter(o => o.status === 'filled').length,
    rejected: orders.filter(o => o.status === 'rejected').length,
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400">Total Orders</div>
          <div className="text-2xl font-semibold text-gray-900 dark:text-white mt-1">
            {stats.total}
          </div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400">Pending</div>
          <div className="text-2xl font-semibold text-blue-600 dark:text-blue-400 mt-1">
            {stats.pending}
          </div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400">Filled</div>
          <div className="text-2xl font-semibold text-green-600 dark:text-green-400 mt-1">
            {stats.filled}
          </div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400">Rejected</div>
          <div className="text-2xl font-semibold text-red-600 dark:text-red-400 mt-1">
            {stats.rejected}
          </div>
        </Card>
      </div>

      <Card
        title="Order Flow"
        actions={
          <div className="flex gap-2">
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value as OrderStatus | 'all')}
              className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="filled">Filled</option>
              <option value="partial">Partial</option>
              <option value="rejected">Rejected</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <Button size="sm" variant="secondary" onClick={handleExport}>
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
          </div>
        }
      >
        <Table
          data={orders}
          columns={columns}
          loading={loading}
          sort={sort}
          onSort={handleSort}
          emptyMessage="No orders found"
        />
      </Card>
    </div>
  );
};

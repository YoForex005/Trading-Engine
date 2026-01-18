import { useState, useEffect } from 'react';
import { Card } from '@/components/shared/Card';
import { Table } from '@/components/shared/Table';
import { Badge } from '@/components/shared/Badge';
import { Button } from '@/components/shared/Button';
import { Plus, Edit, Trash2 } from 'lucide-react';
import { api } from '@/services/api';
import type { RoutingRule, RoutingMode } from '@/types';

const modeColors: Record<RoutingMode, 'info' | 'success' | 'warning'> = {
  'a-book': 'info',
  'b-book': 'success',
  'c-book': 'warning',
};

export const RoutingControl = () => {
  const [rules, setRules] = useState<RoutingRule[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadRules();
  }, []);

  const loadRules = async () => {
    try {
      setLoading(true);
      const data = await api.getRoutingRules();
      setRules(data.sort((a, b) => b.priority - a.priority));
    } catch (error) {
      console.error('Failed to load routing rules:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (ruleId: string, enabled: boolean) => {
    try {
      await api.updateRoutingRule(ruleId, { enabled });
      await loadRules();
    } catch (error) {
      console.error('Failed to update rule:', error);
    }
  };

  const handleDelete = async (ruleId: string) => {
    if (!confirm('Are you sure you want to delete this routing rule?')) return;

    try {
      await api.deleteRoutingRule(ruleId);
      await loadRules();
    } catch (error) {
      console.error('Failed to delete rule:', error);
    }
  };

  const columns = [
    {
      key: 'priority',
      header: 'Priority',
      sortable: true,
      render: (rule: RoutingRule) => (
        <Badge variant="default">{rule.priority}</Badge>
      ),
    },
    {
      key: 'name',
      header: 'Rule Name',
      sortable: true,
      render: (rule: RoutingRule) => (
        <span className="font-semibold">{rule.name}</span>
      ),
    },
    {
      key: 'mode',
      header: 'Mode',
      render: (rule: RoutingRule) => (
        <Badge variant={modeColors[rule.mode]}>
          {rule.mode.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'conditions',
      header: 'Conditions',
      render: (rule: RoutingRule) => (
        <div className="text-sm">
          {rule.conditions.userGroups && (
            <div>Groups: {rule.conditions.userGroups.join(', ')}</div>
          )}
          {rule.conditions.symbols && (
            <div>Symbols: {rule.conditions.symbols.join(', ')}</div>
          )}
          {rule.conditions.minVolume && (
            <div>Min Volume: ${rule.conditions.minVolume}</div>
          )}
        </div>
      ),
    },
    {
      key: 'lpTargets',
      header: 'LP Targets',
      render: (rule: RoutingRule) => rule.lpTargets ? (
        <span className="text-sm">{rule.lpTargets.join(', ')}</span>
      ) : (
        <span className="text-gray-400">All</span>
      ),
    },
    {
      key: 'enabled',
      header: 'Status',
      render: (rule: RoutingRule) => (
        <label className="relative inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            checked={rule.enabled}
            onChange={(e) => handleToggle(rule.id, e.target.checked)}
            className="sr-only peer"
          />
          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary-600"></div>
        </label>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (rule: RoutingRule) => (
        <div className="flex gap-2">
          <Button size="sm" variant="ghost">
            <Edit className="w-4 h-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => handleDelete(rule.id)}
          >
            <Trash2 className="w-4 h-4 text-red-600" />
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <Card
        title="Routing Rules"
        subtitle="Configure A-Book, B-Book, and C-Book routing rules"
        actions={
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Add Rule
          </Button>
        }
      >
        <Table
          data={rules}
          columns={columns}
          loading={loading}
          emptyMessage="No routing rules configured"
        />
      </Card>

      <Card title="Routing Statistics">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">A-Book</div>
            <div className="text-2xl font-semibold text-blue-600">0 orders</div>
            <div className="text-sm text-gray-500">Direct to LP</div>
          </div>
          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">B-Book</div>
            <div className="text-2xl font-semibold text-green-600">0 orders</div>
            <div className="text-sm text-gray-500">Internal matching</div>
          </div>
          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">C-Book</div>
            <div className="text-2xl font-semibold text-orange-600">0 orders</div>
            <div className="text-sm text-gray-500">Hybrid routing</div>
          </div>
        </div>
      </Card>
    </div>
  );
};

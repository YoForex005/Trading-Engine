import { ReactNode, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { clsx } from 'clsx';
import {
  LayoutDashboard,
  Activity,
  Settings,
  BarChart3,
  Menu,
  X,
  Moon,
  Sun,
  LogOut,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { Button } from '@/components/shared/Button';

type LayoutProps = {
  children: ReactNode;
  darkMode: boolean;
  onToggleDarkMode: () => void;
  wsConnected: boolean;
  onLogout: () => void;
};

type NavItem = {
  label: string;
  icon: React.ReactNode;
  path?: string;
  children?: NavItem[];
};

const navigation: NavItem[] = [
  {
    label: 'Monitoring',
    icon: <Activity className="w-5 h-5" />,
    children: [
      { label: 'System Overview', icon: <LayoutDashboard className="w-4 h-4" />, path: '/monitoring/system' },
      { label: 'Order Flow', icon: <Activity className="w-4 h-4" />, path: '/monitoring/orders' },
      { label: 'LP Health', icon: <Activity className="w-4 h-4" />, path: '/monitoring/lp' },
      { label: 'User Activity', icon: <Activity className="w-4 h-4" />, path: '/monitoring/users' },
      { label: 'Errors', icon: <Activity className="w-4 h-4" />, path: '/monitoring/errors' },
    ],
  },
  {
    label: 'Controls',
    icon: <Settings className="w-5 h-5" />,
    children: [
      { label: 'Routing', icon: <Settings className="w-4 h-4" />, path: '/controls/routing' },
      { label: 'LP Management', icon: <Settings className="w-4 h-4" />, path: '/controls/lp' },
      { label: 'Symbols', icon: <Settings className="w-4 h-4" />, path: '/controls/symbols' },
      { label: 'Users', icon: <Settings className="w-4 h-4" />, path: '/controls/users' },
      { label: 'Risk Control', icon: <Settings className="w-4 h-4" />, path: '/controls/risk' },
    ],
  },
  {
    label: 'Analytics',
    icon: <BarChart3 className="w-5 h-5" />,
    children: [
      { label: 'Performance', icon: <BarChart3 className="w-4 h-4" />, path: '/analytics/performance' },
      { label: 'Trading', icon: <BarChart3 className="w-4 h-4" />, path: '/analytics/trading' },
      { label: 'Users', icon: <BarChart3 className="w-4 h-4" />, path: '/analytics/users' },
      { label: 'Audit Trail', icon: <BarChart3 className="w-4 h-4" />, path: '/analytics/audit' },
    ],
  },
];

export const Layout = ({ children, darkMode, onToggleDarkMode, wsConnected, onLogout }: LayoutProps) => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['Monitoring']));
  const location = useLocation();

  const toggleSection = (label: string) => {
    setExpandedSections(prev => {
      const updated = new Set(prev);
      if (updated.has(label)) {
        updated.delete(label);
      } else {
        updated.add(label);
      }
      return updated;
    });
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
      {/* Sidebar */}
      <aside
        className={clsx(
          'fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transform transition-transform duration-200',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h1 className="text-xl font-bold text-gray-900 dark:text-white">Admin Dashboard</h1>
          <button
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <nav className="p-4 space-y-1 overflow-y-auto h-[calc(100vh-8rem)]">
          {navigation.map((section) => (
            <div key={section.label}>
              <button
                onClick={() => toggleSection(section.label)}
                className="w-full flex items-center justify-between px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                <div className="flex items-center gap-2">
                  {section.icon}
                  {section.label}
                </div>
                <span className={clsx(
                  'transform transition-transform',
                  expandedSections.has(section.label) && 'rotate-90'
                )}>
                  ▶
                </span>
              </button>
              {expandedSections.has(section.label) && section.children && (
                <div className="ml-4 mt-1 space-y-1">
                  {section.children.map((child) => (
                    <Link
                      key={child.path}
                      to={child.path || '#'}
                      className={clsx(
                        'flex items-center gap-2 px-3 py-2 text-sm rounded-lg transition-colors',
                        location.pathname === child.path
                          ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-200'
                          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                      )}
                    >
                      {child.icon}
                      {child.label}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          ))}
        </nav>

        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {wsConnected ? (
                <Wifi className="w-4 h-4 text-green-600" />
              ) : (
                <WifiOff className="w-4 h-4 text-red-600" />
              )}
              <span className="text-xs text-gray-600 dark:text-gray-400">
                {wsConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
            <button
              onClick={onLogout}
              className="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
              title="Logout"
            >
              <LogOut className="w-4 h-4 text-gray-600 dark:text-gray-400" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div
        className={clsx(
          'transition-all duration-200',
          sidebarOpen ? 'lg:ml-64' : 'ml-0'
        )}
      >
        {/* Header */}
        <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-40">
          <div className="px-4 py-3 flex items-center justify-between">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <Menu className="w-5 h-5" />
            </button>

            <div className="flex items-center gap-2">
              <button
                onClick={onToggleDarkMode}
                className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
                title={darkMode ? 'Light mode' : 'Dark mode'}
              >
                {darkMode ? (
                  <Sun className="w-5 h-5" />
                ) : (
                  <Moon className="w-5 h-5" />
                )}
              </button>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          {children}
        </main>
      </div>
    </div>
  );
};

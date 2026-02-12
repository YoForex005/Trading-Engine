'use client';

import { useState } from 'react';
import { Account } from '../types';
import RtxLayout from '../components/layout/RtxLayout';
import AccountsView from '../components/dashboard/AccountsView';
import OrdersView from '../components/dashboard/OrdersView';
import HistoryView from '../components/dashboard/HistoryView';
import ChartGrid, { ChartLayout } from '../components/dashboard/ChartGrid';
import AuthGuard from '../components/auth/AuthGuard';
import AdminUsersView from '../components/dashboard/AdminUsersView';
import RiskDashboard from '../components/dashboard/RiskDashboard';
import OrderFlowDashboard from '../components/dashboard/OrderFlowDashboard';
import ReportsView from '../components/dashboard/ReportsView';
import DashboardOverview from '../components/dashboard/DashboardOverview';
import SymbolManagement from '../components/dashboard/SymbolManagement';

import AccountWindow from '../components/dashboard/AccountWindow';
import AuditLogViewer from '../components/dashboard/AuditLogViewer';
import WithdrawalQueue from '../components/dashboard/WithdrawalQueue';
import UserManagement from '../components/dashboard/UserManagement';
import GroupManagement from '../components/dashboard/GroupManagement';
import ClientManagement from '../components/dashboard/ClientManagement';
import LPHealthDashboard from '../components/dashboard/LPHealthDashboard';
import ApiKeyManagement from '../components/dashboard/ApiKeyManagement';
import WebhookManagement from '../components/dashboard/WebhookManagement';
import CommissionReport from '../components/dashboard/CommissionReport';
import BrokerPnLReport from '../components/dashboard/BrokerPnlReport';
import IPFilterManagement from '../components/dashboard/IPFilterManagement';
import TradeCopierManagement from '../components/dashboard/TradeCopierManagement';
import KYCManagement from '../components/dashboard/KYCManagement';
import ComplianceDashboard from '../components/dashboard/ComplianceDashboard';
import ComplianceReports from '../components/dashboard/ComplianceReports';
import NotificationCenter from '../components/dashboard/NotificationCenter';
import FinancialOperations from '../components/dashboard/FinancialOperations';
import ServerConfig from '../components/dashboard/ServerConfig';
import DealingDesk from '../components/dashboard/DealingDesk';
import EmailTemplates from '../components/dashboard/EmailTemplateManager';
import SystemHealth from '../components/dashboard/SystemHealth';
import PerformanceMonitor from '../components/dashboard/PerformanceMonitor';
import AffiliateManagement from '../components/dashboard/AffiliateManagement';
import TradeHistory from '../components/dashboard/TradeHistory';
import ClientPortalConfig from '../components/dashboard/ClientPortalConfig';
import RiskRulesEngine from '../components/dashboard/RiskRulesEngine';
import AnnouncementManager from '../components/dashboard/AnnouncementManager';
import LPConfiguration from '../components/dashboard/LPConfiguration';
import MarketHoursConfig from '../components/dashboard/MarketHoursConfig';
import BrokerOverview from '../components/dashboard/BrokerOverview';
import ExecutionReport from '../components/dashboard/ExecutionReport';
import AccountTypeConfig from '../components/dashboard/AccountTypeConfig';
import PaymentGatewayConfig from '../components/dashboard/PaymentGatewayConfig';
import SwapRateConfig from '../components/dashboard/SwapRateConfig';
import TradingCompetitions from '../components/dashboard/TradingCompetitions';
import SupportTickets from '../components/dashboard/SupportTickets';
import LocalizationConfig from '../components/dashboard/LocalizationConfig';
import RegulatoryReporting from '../components/dashboard/RegulatoryReporting';
import MarginCallDashboard from '../components/dashboard/MarginCallDashboard';
import WhiteLabelConfig from '../components/dashboard/WhiteLabelConfig';
import ClientActivityLog from '../components/dashboard/ClientActivityLog';
import CommunicationLog from '../components/dashboard/CommunicationLog';
import PromotionManager from '../components/dashboard/PromotionManager';
import ClientSegmentation from '../components/dashboard/ClientSegmentation';
import AuditTrailDashboard from '../components/dashboard/AuditTrailDashboard';
import TradingSessions from '../components/dashboard/TradingSessions';
import TradeReconciliation from '../components/dashboard/TradeReconciliation';
import ClientNotes from '../components/dashboard/ClientNotes';
import SpreadMonitor from '../components/dashboard/SpreadMonitor';
import MamPammManager from '../components/dashboard/MamPammManager';
import ServerLogViewer from '../components/dashboard/ServerLogViewer';
import PlatformConfig from '../components/dashboard/PlatformConfig';
import DepositWithdrawalQueue from '../components/dashboard/DepositWithdrawalQueue';
import CustomerLifecycle from '../components/dashboard/CustomerLifecycle';
import ScheduledReports from '../components/dashboard/ScheduledReports';
import DemoAccountManager from '../components/dashboard/DemoAccountManager';
import ClientDocuments from '../components/dashboard/ClientDocuments';
import FraudDetection from '../components/dashboard/FraudDetection';
import CommissionTiers from '../components/dashboard/CommissionTiers';
import BackupManager from '../components/dashboard/BackupManager';
import ReferralProgram from '../components/dashboard/ReferralProgram';
import TradingSignals from '../components/dashboard/TradingSignals';
import WalletManagement from '../components/dashboard/WalletManagement';
import AccountStatement from '../components/dashboard/AccountStatement';
import InstrumentGroups from '../components/dashboard/InstrumentGroups';
import ClientRiskScoring from '../components/dashboard/ClientRiskScoring';

export default function AdminDesktop() {
  const [activeTab, setActiveTab] = useState('dashboard-overview');
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [editingOrder, setEditingOrder] = useState<any | null>(null);
  const [chartLayout, setChartLayout] = useState<ChartLayout>('1x1');

  const handleNavigate = (viewId: string) => {
    // Map navigator IDs to views
    if (viewId === 'dashboard-overview') setActiveTab('dashboard-overview');
    else if (viewId === 'accounts' || viewId === 'an-accounts') setActiveTab('accounts');
    else if (viewId === 'orders' || viewId === 'an-orders') setActiveTab('orders');
    else if (viewId === 'rep-trades' || viewId === 'history') setActiveTab('history');
    else if (viewId === 'charts') setActiveTab('charts');
    else if (viewId === 'admin-users') setActiveTab('admin-users');
    else if (viewId === 'risk-dashboard') setActiveTab('risk-dashboard');
    else if (viewId === 'order-flow') setActiveTab('order-flow');
    else if (viewId === 'reports') setActiveTab('reports');
    else if (viewId === 'symbol-management') setActiveTab('symbol-management');
    else if (viewId === 'audit-log') setActiveTab('audit-log');
    else if (viewId === 'withdrawal-queue') setActiveTab('withdrawal-queue');
    else if (viewId === 'user-management') setActiveTab('user-management');
    else if (viewId === 'group-management') setActiveTab('group-management');
    else if (viewId === 'client-management') setActiveTab('client-management');
    else if (viewId === 'lp-health') setActiveTab('lp-health');
    else if (viewId === 'api-keys') setActiveTab('api-keys');
    else if (viewId === 'webhooks') setActiveTab('webhooks');
    else if (viewId === 'commissions') setActiveTab('commissions');
    else if (viewId === 'broker-pnl') setActiveTab('broker-pnl');
    else if (viewId === 'ip-filter') setActiveTab('ip-filter');
    else if (viewId === 'trade-copier') setActiveTab('trade-copier');
    else if (viewId === 'kyc') setActiveTab('kyc');
    else if (viewId === 'compliance') setActiveTab('compliance');
    else if (viewId === 'notifications') setActiveTab('notifications');
    else if (viewId === 'financial-ops') setActiveTab('financial-ops');
    else if (viewId === 'server-config') setActiveTab('server-config');
    else if (viewId === 'platform-config') setActiveTab('platform-config');
    else if (viewId === 'dealing') setActiveTab('dealing');
    else if (viewId === 'email-templates') setActiveTab('email-templates');
    else if (viewId === 'system-health') setActiveTab('system-health');
    else if (viewId === 'performance-monitor') setActiveTab('performance-monitor');
    else if (viewId === 'affiliate-management') setActiveTab('affiliate-management');
    else if (viewId === 'trade-history') setActiveTab('trade-history');
    else if (viewId === 'client-portal') setActiveTab('client-portal');
    else if (viewId === 'risk-rules') setActiveTab('risk-rules');
    else if (viewId === 'announcement-manager') setActiveTab('announcement-manager');
    else if (viewId === 'lp-config') setActiveTab('lp-config');
    else if (viewId === 'swap-config') setActiveTab('swap-config');
    else if (viewId === 'trading-competitions') setActiveTab('trading-competitions');
    else if (viewId === 'market-hours-config') setActiveTab('market-hours-config');
    else if (viewId === 'trading-sessions') setActiveTab('trading-sessions');
    else if (viewId === 'execution-report') setActiveTab('execution-report');
    else if (viewId === 'account-type-config') setActiveTab('account-type-config');
    else if (viewId === 'payment-gateway-config') setActiveTab('payment-gateway-config');
    else if (viewId === 'support-tickets') setActiveTab('support-tickets');
    else if (viewId === 'localization-config') setActiveTab('localization-config');
    else if (viewId === 'regulatory-reporting') setActiveTab('regulatory-reporting');
    else if (viewId === 'margin-call-dashboard') setActiveTab('margin-call-dashboard');
    else if (viewId === 'white-label') setActiveTab('white-label');
    else if (viewId === 'client-activity-log') setActiveTab('client-activity-log');
    else if (viewId === 'communication-log') setActiveTab('communication-log');
    else if (viewId === 'promotion-manager') setActiveTab('promotion-manager');
    else if (viewId === 'client-segmentation') setActiveTab('client-segmentation');
    else if (viewId === 'audit-trail') setActiveTab('audit-trail');
    else if (viewId === 'trade-reconciliation') setActiveTab('trade-reconciliation');
    else if (viewId === 'client-notes') setActiveTab('client-notes');
    else if (viewId === 'spread-monitor') setActiveTab('spread-monitor');
    else if (viewId === 'mam-pamm') setActiveTab('mam-pamm');
    else if (viewId === 'server-logs') setActiveTab('server-logs');
    else if (viewId === 'customer-lifecycle') setActiveTab('customer-lifecycle');
    else if (viewId === 'scheduled-reports') setActiveTab('scheduled-reports');
    else if (viewId === 'demo-account-manager') setActiveTab('demo-account-manager');
    else if (viewId === 'client-documents') setActiveTab('client-documents');
    else if (viewId === 'fraud-detection') setActiveTab('fraud-detection');
    else if (viewId === 'commission-tiers') setActiveTab('commission-tiers');
    else if (viewId === 'backup-manager') setActiveTab('backup-manager');
    else if (viewId === 'referral-program') setActiveTab('referral-program');
    else if (viewId === 'trading-signals') setActiveTab('trading-signals');
    else if (viewId === 'wallet-management') setActiveTab('wallet-management');
    else if (viewId === 'account-statement') setActiveTab('account-statement');
    else if (viewId === 'instrument-groups') setActiveTab('instrument-groups');
    else if (viewId === 'client-risk-scoring') setActiveTab('client-risk-scoring');
    else setActiveTab(viewId);
  };

  const renderView = () => {
    switch (activeTab) {
      case 'dashboard-overview':
        return <BrokerOverview />;
      case 'accounts':
        return <AccountsView />;
      case 'orders':
        return <OrdersView onOrderDoubleClick={(order) => setEditingOrder(order)} />;
      case 'history':
        return <HistoryView />;
      case 'charts':
        return <ChartGrid layout={chartLayout} />;
      case 'admin-users':
        return <AdminUsersView />;
      case 'risk-dashboard':
        return <RiskDashboard />;
      case 'order-flow':
        return <OrderFlowDashboard />;
      case 'reports':
        return <ReportsView />;
      case 'symbol-management':
        return <SymbolManagement />;
      case 'audit-log':
        return <AuditLogViewer />;
      case 'withdrawal-queue':
        return <WithdrawalQueue />;
      case 'user-management':
        return <UserManagement />;
      case 'group-management':
        return <GroupManagement />;
      case 'client-management':
        return <ClientManagement />;
      case 'lp-health':
        return <LPHealthDashboard />;
      case 'api-keys':
        return <ApiKeyManagement />;
      case 'webhooks':
        return <WebhookManagement />;
      case 'commissions':
        return <CommissionReport />;
      case 'broker-pnl':
        return <BrokerPnLReport />;
      case 'ip-filter':
        return <IPFilterManagement />;
      case 'trade-copier':
        return <TradeCopierManagement />;
      case 'kyc':
        return <KYCManagement />;
      case 'compliance':
        return <ComplianceDashboard />;
      case 'notifications':
        return <NotificationCenter onNavigate={handleNavigate} />;
      case 'financial-ops':
        return <FinancialOperations />;
      case 'server-config':
        return <ServerConfig />;
      case 'platform-config':
        return <PlatformConfig />;
      case 'dealing':
        return <DealingDesk />;
      case 'email-templates':
        return <EmailTemplates />;
      case 'system-health':
        return <SystemHealth />;
      case 'performance-monitor':
        return <PerformanceMonitor />;
      case 'affiliate-management':
        return <AffiliateManagement />;
      case 'trade-history':
        return <TradeHistory />;
      case 'client-portal':
        return <ClientPortalConfig />;
      case 'risk-rules':
        return <RiskRulesEngine />;
      case 'announcement-manager':
        return <AnnouncementManager />;
      case 'lp-config':
        return <LPConfiguration />;
      case 'swap-config':
        return <SwapRateConfig />;
      case 'trading-competitions':
        return <TradingCompetitions />;
      case 'market-hours-config':
        return <MarketHoursConfig />;
      case 'trading-sessions':
        return <TradingSessions />;
      case 'execution-report':
        return <ExecutionReport />;
      case 'account-type-config':
        return <AccountTypeConfig />;
      case 'payment-gateway-config':
        return <PaymentGatewayConfig />;
      case 'support-tickets':
        return <SupportTickets />;
      case 'localization-config':
        return <LocalizationConfig />;
      case 'regulatory-reporting':
        return <RegulatoryReporting />;
      case 'margin-call-dashboard':
        return <MarginCallDashboard />;
      case 'white-label':
        return <WhiteLabelConfig />;
      case 'client-activity-log':
        return <ClientActivityLog />;
      case 'communication-log':
        return <CommunicationLog />;
      case 'promotion-manager':
        return <PromotionManager />;
      case 'client-segmentation':
        return <ClientSegmentation />;
      case 'audit-trail':
        return <AuditTrailDashboard />;
      case 'trade-reconciliation':
        return <TradeReconciliation />;
      case 'client-notes':
        return <ClientNotes />;
      case 'spread-monitor':
        return <SpreadMonitor />;
      case 'mam-pamm':
        return <MamPammManager />;
      case 'server-logs':
        return <ServerLogViewer />;
      case 'customer-lifecycle':
        return <CustomerLifecycle />;
      case 'scheduled-reports':
        return <ScheduledReports />;
      case 'demo-account-manager':
        return <DemoAccountManager />;
      case 'client-documents':
        return <ClientDocuments />;
      case 'fraud-detection':
        return <FraudDetection />;
      case 'commission-tiers':
        return <CommissionTiers />;
      case 'backup-manager':
        return <BackupManager />;
      case 'referral-program':
        return <ReferralProgram />;
      case 'trading-signals':
        return <TradingSignals />;
      case 'wallet-management':
        return <WalletManagement />;
      case 'account-statement':
        return <AccountStatement />;
      case 'instrument-groups':
        return <InstrumentGroups />;
      case 'client-risk-scoring':
        return <ClientRiskScoring />;
      default:
        // Fallback for not-yet-implemented views
        return (
          <div className="flex flex-col items-center justify-center h-full text-[#666]">
            <div className="text-4xl mb-4 opacity-20">🚧</div>
            <div className="text-lg">View: {activeTab}</div>
            <div className="text-xs mt-2">This module is under construction</div>
          </div>
        );
    }
  };

  return (
    <AuthGuard>
      <RtxLayout
        onNavigate={handleNavigate}
        chartLayout={chartLayout}
        onLayoutChange={setChartLayout}
      >
        {/* Account Window Modal */}
        {editingOrder && (
          <AccountWindow
            order={editingOrder}
            accountStr="680962851, Trader, USD, 1:100, Hedge"
            onClose={() => setEditingOrder(null)}
          />
        )}

        {/* Main Content Area */}
        <div className="flex flex-col h-full bg-[#121316]">
          {/* Tab Strip for Main Window - Native Styling */}
        <div className="h-7 bg-[#1E2026] border-b border-[#383A42] flex items-end px-1 gap-0.5 flex-shrink-0 select-none">
          <div
            onClick={() => setActiveTab('accounts')}
            className={`
              px-4 h-6 flex items-center text-xs font-bold cursor-pointer transition-none
              ${activeTab === 'accounts'
                ? 'bg-[#121316] text-[#F5C542] border-t-2 border-[#F5C542]'
                : 'bg-[#1E2026] text-[#888] border-t border-transparent hover:bg-[#25272E] hover:text-[#CCC]'}
            `}
          >
            Accounts
          </div>
          <div
            onClick={() => setActiveTab('orders')}
            className={`
              px-4 h-6 flex items-center text-xs font-bold cursor-pointer transition-none
              ${activeTab === 'orders'
                ? 'bg-[#121316] text-[#F5C542] border-t-2 border-[#F5C542]'
                : 'bg-[#1E2026] text-[#888] border-t border-transparent hover:bg-[#25272E] hover:text-[#CCC]'}
            `}
          >
            Orders
          </div>
          <div
            onClick={() => setActiveTab('history')}
            className={`
              px-4 h-6 flex items-center text-xs font-bold cursor-pointer transition-none
              ${activeTab === 'history'
                ? 'bg-[#121316] text-[#F5C542] border-t-2 border-[#F5C542]'
                : 'bg-[#1E2026] text-[#888] border-t border-transparent hover:bg-[#25272E] hover:text-[#CCC]'}
            `}
          >
            History
          </div>
          <div
            onClick={() => setActiveTab('charts')}
            className={`
              px-4 h-6 flex items-center text-xs font-bold cursor-pointer transition-none
              ${activeTab === 'charts'
                ? 'bg-[#121316] text-[#F5C542] border-t-2 border-[#F5C542]'
                : 'bg-[#1E2026] text-[#888] border-t border-transparent hover:bg-[#25272E] hover:text-[#CCC]'}
            `}
          >
            Charts
          </div>
        </div>

          {/* Primary Data Grid */}
          <div className="flex-1 overflow-hidden relative bg-[#121316] border-t border-[#383A42]">
            {renderView()}
          </div>
        </div>
      </RtxLayout>
    </AuthGuard>
  );
}


'use client';

import { useState } from 'react';
import { Account } from '../types';
import RtxLayout from '../components/layout/RtxLayout';
import AccountsView from '../components/dashboard/AccountsView';
import OrdersView from '../components/dashboard/OrdersView';
import HistoryView from '../components/dashboard/HistoryView';

import AccountWindow from '../components/dashboard/AccountWindow';

export default function AdminDesktop() {
  const [activeTab, setActiveTab] = useState('accounts');
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [editingOrder, setEditingOrder] = useState<any | null>(null);

  const handleNavigate = (viewId: string) => {
    // Map navigator IDs to views
    if (viewId === 'accounts' || viewId === 'an-accounts') setActiveTab('accounts');
    else if (viewId === 'orders' || viewId === 'an-orders') setActiveTab('orders');
    else if (viewId === 'rep-trades' || viewId === 'history') setActiveTab('history');
    else setActiveTab(viewId);
  };

  const renderView = () => {
    switch (activeTab) {
      case 'accounts':
        return <AccountsView />;
      case 'orders':
        return <OrdersView onOrderDoubleClick={(order) => setEditingOrder(order)} />;
      case 'history':
        return <HistoryView />;
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
    <RtxLayout onNavigate={handleNavigate}>
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
        </div>

        {/* Primary Data Grid */}
        <div className="flex-1 overflow-hidden relative bg-[#121316] border-t border-[#383A42]">
          {renderView()}
        </div>
      </div>
    </RtxLayout>
  );
}


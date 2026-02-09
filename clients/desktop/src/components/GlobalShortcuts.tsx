/**
 * Global Shortcuts Component
 * Registers all MT5-compatible global shortcuts
 */

import { useEffect, useState } from 'react';
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcut';
import { ShortcutFeedback } from './ShortcutFeedback';
import { ShortcutHelpDialog } from './ShortcutHelpDialog';
import { SHORTCUT_CATEGORIES } from '../services/keyboardShortcutManager';
import { useAppStore } from '../store/useAppStore';

interface GlobalShortcutsProps {
  onOpenOrderPanel: (symbol: string, price: { bid: number; ask: number }) => void;
  onQuickBuy: () => void;
  onQuickSell: () => void;
  onSaveWorkspace: () => void;
  onOpenDataFolder: () => void;
  onPrintChart: () => void;
  onToggleFullScreen: () => void;
  selectedSymbol: string;
}

export function GlobalShortcuts({
  onOpenOrderPanel,
  onQuickBuy,
  onQuickSell,
  onSaveWorkspace,
  onOpenDataFolder,
  onPrintChart,
  onToggleFullScreen,
  selectedSymbol
}: GlobalShortcutsProps) {
  const [feedbackText, setFeedbackText] = useState<string | null>(null);
  const [showHelpDialog, setShowHelpDialog] = useState(false);

  const showFeedback = (text: string) => {
    setFeedbackText(text);
  };

  // Get current tick for selected symbol
  const getCurrentTick = () => {
    const ticks = useAppStore.getState().ticks;
    return ticks[selectedSymbol] || { bid: 0, ask: 0 };
  };

  useKeyboardShortcuts([
    // File menu shortcuts
    {
      id: 'file.save',
      key: 's',
      ctrl: true,
      description: 'Save workspace',
      category: SHORTCUT_CATEGORIES.FILE,
      priority: 100,
      handler: () => {
        showFeedback('Save Workspace (Ctrl+S)');
        onSaveWorkspace();
      }
    },
    {
      id: 'file.open-data-folder',
      key: 'D',
      ctrl: true,
      shift: true,
      description: 'Open data folder',
      category: SHORTCUT_CATEGORIES.FILE,
      priority: 100,
      handler: () => {
        showFeedback('Open Data Folder (Ctrl+Shift+D)');
        onOpenDataFolder();
      }
    },
    {
      id: 'file.print',
      key: 'p',
      ctrl: true,
      description: 'Print chart',
      category: SHORTCUT_CATEGORIES.FILE,
      priority: 100,
      handler: () => {
        showFeedback('Print Chart (Ctrl+P)');
        onPrintChart();
      }
    },
    {
      id: 'file.exit',
      key: 'F4',
      alt: true,
      description: 'Exit application',
      category: SHORTCUT_CATEGORIES.FILE,
      priority: 100,
      allowInInput: true,
      handler: () => {
        if (confirm('Are you sure you want to exit?')) {
          showFeedback('Exit (Alt+F4)');
          window.close();
        }
      }
    },

    // View menu shortcuts
    {
      id: 'view.fullscreen',
      key: 'F11',
      description: 'Full Screen',
      category: SHORTCUT_CATEGORIES.VIEW,
      priority: 100,
      allowInInput: true,
      handler: () => {
        showFeedback('Toggle Fullscreen (F11)');
        onToggleFullScreen();
      }
    },

    // Tools menu shortcuts
    {
      id: 'tools.new-order',
      key: 'F9',
      description: 'New Order',
      category: SHORTCUT_CATEGORIES.TOOLS,
      priority: 100,
      allowInInput: true,
      handler: () => {
        const tick = getCurrentTick();
        showFeedback(`New Order - ${selectedSymbol} (F9)`);
        onOpenOrderPanel(selectedSymbol, { bid: tick.bid, ask: tick.ask });
      }
    },

    // Trading shortcuts
    {
      id: 'trading.quick-buy',
      key: 'b',
      alt: true,
      description: 'Quick Buy',
      category: SHORTCUT_CATEGORIES.TRADING,
      priority: 95,
      handler: () => {
        showFeedback(`Quick Buy - ${selectedSymbol} (Alt+B)`);
        onQuickBuy();
      }
    },
    {
      id: 'trading.quick-sell',
      key: 's',
      alt: true,
      description: 'Quick Sell',
      category: SHORTCUT_CATEGORIES.TRADING,
      priority: 95,
      handler: () => {
        showFeedback(`Quick Sell - ${selectedSymbol} (Alt+S)`);
        onQuickSell();
      }
    },

    // Chart shortcuts
    {
      id: 'charts.zoom-in',
      key: '+',
      description: 'Zoom In',
      category: SHORTCUT_CATEGORIES.CHARTS,
      priority: 80,
      handler: () => {
        showFeedback('Zoom In (+)');
        window.dispatchEvent(new CustomEvent('chart-zoom', { detail: { direction: 'in' } }));
      }
    },
    {
      id: 'charts.zoom-out',
      key: '-',
      description: 'Zoom Out',
      category: SHORTCUT_CATEGORIES.CHARTS,
      priority: 80,
      handler: () => {
        showFeedback('Zoom Out (-)');
        window.dispatchEvent(new CustomEvent('chart-zoom', { detail: { direction: 'out' } }));
      }
    },

    // Navigation shortcuts
    {
      id: 'navigation.close-modal',
      key: 'Escape',
      description: 'Close Modal',
      category: SHORTCUT_CATEGORIES.NAVIGATION,
      priority: 100,
      allowInInput: true,
      handler: () => {
        if (showHelpDialog) {
          setShowHelpDialog(false);
        } else {
          window.dispatchEvent(new CustomEvent('close-modal'));
        }
      }
    },

    // Help shortcuts
    {
      id: 'help.shortcuts',
      key: 'F1',
      description: 'Show Keyboard Shortcuts',
      category: SHORTCUT_CATEGORIES.HELP,
      priority: 100,
      allowInInput: true,
      handler: () => {
        showFeedback('Keyboard Shortcuts (F1)');
        setShowHelpDialog(true);
      }
    }
  ]);

  return (
    <>
      <ShortcutFeedback
        shortcutText={feedbackText}
        onClose={() => setFeedbackText(null)}
      />
      <ShortcutHelpDialog
        isOpen={showHelpDialog}
        onClose={() => setShowHelpDialog(false)}
      />
    </>
  );
}

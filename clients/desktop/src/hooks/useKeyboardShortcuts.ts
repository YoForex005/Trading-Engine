/**
 * Keyboard Shortcuts Hook
 * Global keyboard shortcuts for power users
 */

import { useEffect } from 'react';

export type ShortcutAction =
  | 'NEW_ORDER'
  | 'CLOSE_ALL'
  | 'TOGGLE_CHART'
  | 'NEXT_SYMBOL'
  | 'PREV_SYMBOL'
  | 'CLOSE_MODAL'
  | 'INCREASE_VOLUME'
  | 'DECREASE_VOLUME'
  | 'BUY_MARKET'
  | 'SELL_MARKET'
  | 'SAVE_LAYOUT'
  | 'LOAD_LAYOUT'
  | 'DEPTH_OF_MARKET'
  | 'SYMBOLS_DIALOG'
  | 'POPUP_PRICES';

export interface ShortcutConfig {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  action: ShortcutAction;
  description: string;
}

const DEFAULT_SHORTCUTS: ShortcutConfig[] = [
  { key: 'F9', action: 'NEW_ORDER', description: 'Open new order dialog' },
  { key: 'F10', action: 'POPUP_PRICES', description: 'Open popup prices' },
  { key: 'Escape', action: 'CLOSE_MODAL', description: 'Close modal/dialog' },
  { key: 'F11', action: 'TOGGLE_CHART', description: 'Toggle chart fullscreen' },
  { key: 'ArrowUp', action: 'NEXT_SYMBOL', description: 'Select next symbol' },
  { key: 'ArrowDown', action: 'PREV_SYMBOL', description: 'Select previous symbol' },
  { key: '+', ctrl: true, action: 'INCREASE_VOLUME', description: 'Increase volume' },
  { key: '-', ctrl: true, action: 'DECREASE_VOLUME', description: 'Decrease volume' },
  { key: 'b', ctrl: true, action: 'BUY_MARKET', description: 'Buy market order' },
  { key: 's', ctrl: true, action: 'SELL_MARKET', description: 'Sell market order' },
  { key: 'b', alt: true, action: 'DEPTH_OF_MARKET', description: 'Open depth of market' },
  { key: 'u', ctrl: true, action: 'SYMBOLS_DIALOG', description: 'Open symbols dialog' },
  { key: 'w', ctrl: true, shift: true, action: 'CLOSE_ALL', description: 'Close all positions' },
  { key: 's', ctrl: true, shift: true, action: 'SAVE_LAYOUT', description: 'Save workspace layout' },
  { key: 'l', ctrl: true, shift: true, action: 'LOAD_LAYOUT', description: 'Load workspace layout' },
];

export function useKeyboardShortcuts(
  callbacks: Partial<Record<ShortcutAction, () => void>>,
  enabled = true
) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      // Ignore shortcuts when typing in input fields
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement ||
        event.target instanceof HTMLSelectElement
      ) {
        // Allow ESC to close modals even in input fields
        if (event.key !== 'Escape') {
          return;
        }
      }

      for (const shortcut of DEFAULT_SHORTCUTS) {
        const keyMatch = event.key.toLowerCase() === shortcut.key.toLowerCase();
        const ctrlMatch = shortcut.ctrl ? event.ctrlKey || event.metaKey : !event.ctrlKey && !event.metaKey;
        const shiftMatch = shortcut.shift ? event.shiftKey : !event.shiftKey;
        const altMatch = shortcut.alt ? event.altKey : !event.altKey;

        if (keyMatch && ctrlMatch && shiftMatch && altMatch) {
          const callback = callbacks[shortcut.action];
          if (callback) {
            event.preventDefault();
            callback();
            break;
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [callbacks, enabled]);
}

// Hook to get all available shortcuts for help dialog
export function useShortcutsList(): ShortcutConfig[] {
  return DEFAULT_SHORTCUTS;
}

// Helper to format shortcut for display
export function formatShortcut(shortcut: ShortcutConfig): string {
  const parts: string[] = [];

  if (shortcut.ctrl) parts.push('Ctrl');
  if (shortcut.shift) parts.push('Shift');
  if (shortcut.alt) parts.push('Alt');
  parts.push(shortcut.key);

  return parts.join(' + ');
}

// Type placeholder for command bus - will be replaced when Agent 1 completes
type Command = {
  type: string;
  payload: any;
};

export function registerKeyboardShortcuts(dispatch: (cmd: Command) => void) {
  const handleKeyDown = (e: KeyboardEvent) => {
    // F9: Open Order Panel
    if (e.key === 'F9') {
      e.preventDefault();
      console.log('[KEYBOARD] F9 pressed - Open Order Panel');
      dispatch({
        type: 'OPEN_ORDER_PANEL',
        payload: {
          symbol: 'XAUUSD', // Default symbol, will be overridden by active chart
          price: { bid: 0, ask: 0 } // Will be populated from market data
        }
      });
    }

    // Ctrl+I: Open Indicators Navigator
    if (e.ctrlKey && e.key === 'i') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+I pressed - Open Indicator Navigator');
      dispatch({
        type: 'OPEN_INDICATOR_NAVIGATOR',
        payload: {}
      });
    }

    // Esc: Cancel drawing mode (select cursor tool)
    if (e.key === 'Escape') {
      e.preventDefault();
      console.log('[KEYBOARD] Escape pressed - Cancel drawing mode');
      dispatch({
        type: 'SELECT_TOOL',
        payload: { tool: 'cursor' }
      });
    }

    // +/=: Zoom In
    if (e.key === '+' || e.key === '=') {
      e.preventDefault();
      console.log('[KEYBOARD] + pressed - Zoom In');
      dispatch({
        type: 'ZOOM_IN',
        payload: {}
      });
    }

    // -/_: Zoom Out
    if (e.key === '-' || e.key === '_') {
      e.preventDefault();
      console.log('[KEYBOARD] - pressed - Zoom Out');
      dispatch({
        type: 'ZOOM_OUT',
        payload: {}
      });
    }

    // Ctrl+Z: Undo (future implementation)
    if (e.ctrlKey && e.key === 'z') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+Z pressed - Undo');
      dispatch({
        type: 'UNDO',
        payload: {}
      });
    }

    // Ctrl+Y: Redo (future implementation)
    if (e.ctrlKey && e.key === 'y') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+Y pressed - Redo');
      dispatch({
        type: 'REDO',
        payload: {}
      });
    }

    // Delete: Delete selected drawing
    if (e.key === 'Delete') {
      e.preventDefault();
      console.log('[KEYBOARD] Delete pressed - Delete selected drawing');
      dispatch({
        type: 'DELETE_SELECTED_DRAWING',
        payload: {}
      });
    }

    // Ctrl+A: Select all drawings
    if (e.ctrlKey && e.key === 'a') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+A pressed - Select all drawings');
      dispatch({
        type: 'SELECT_ALL_DRAWINGS',
        payload: {}
      });
    }

    // C: Crosshair tool shortcut
    if (e.key === 'c' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] C pressed - Toggle Crosshair');
      dispatch({
        type: 'TOGGLE_CROSSHAIR',
        payload: {}
      });
    }

    // T: Trendline tool shortcut
    if (e.key === 't' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] T pressed - Select Trendline Tool');
      dispatch({
        type: 'SELECT_TOOL',
        payload: { tool: 'trendline' }
      });
    }

    // H: Horizontal line tool shortcut
    if (e.key === 'h' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] H pressed - Select Horizontal Line Tool');
      dispatch({
        type: 'SELECT_TOOL',
        payload: { tool: 'hline' }
      });
    }

    // V: Vertical line tool shortcut
    if (e.key === 'v' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] V pressed - Select Vertical Line Tool');
      dispatch({
        type: 'SELECT_TOOL',
        payload: { tool: 'vline' }
      });
    }

    // X: Text tool shortcut
    if (e.key === 'x' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] X pressed - Select Text Tool');
      dispatch({
        type: 'SELECT_TOOL',
        payload: { tool: 'text' }
      });
    }
  };

  console.log('[KEYBOARD] Registering keyboard shortcuts');
  document.addEventListener('keydown', handleKeyDown);

  return () => {
    console.log('[KEYBOARD] Unregistering keyboard shortcuts');
    document.removeEventListener('keydown', handleKeyDown);
  };
}

// Helper function to check if an input element is focused
// Prevents shortcuts from firing when typing in text fields
export function isInputFocused(): boolean {
  const activeElement = document.activeElement;
  if (!activeElement) return false;

  const tagName = activeElement.tagName.toLowerCase();
  return (
    tagName === 'input' ||
    tagName === 'textarea' ||
    activeElement.getAttribute('contenteditable') === 'true'
  );
}

// Enhanced version that respects input focus
export function registerKeyboardShortcutsWithInputCheck(dispatch: (cmd: Command) => void) {
  const handleKeyDown = (e: KeyboardEvent) => {
    // Skip shortcuts if user is typing in an input field
    // Exception: F9 and Escape should always work
    if (isInputFocused() && e.key !== 'F9' && e.key !== 'Escape') {
      return;
    }

    // Same logic as registerKeyboardShortcuts
    // (Duplicated for clarity - can be refactored)
    if (e.key === 'F9') {
      e.preventDefault();
      console.log('[KEYBOARD] F9 pressed - Open Order Panel');
      dispatch({ type: 'OPEN_ORDER_PANEL', payload: { symbol: 'XAUUSD', price: { bid: 0, ask: 0 } } });
    }

    if (e.ctrlKey && e.key === 'i') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+I pressed - Open Indicator Navigator');
      dispatch({ type: 'OPEN_INDICATOR_NAVIGATOR', payload: {} });
    }

    if (e.key === 'Escape') {
      e.preventDefault();
      console.log('[KEYBOARD] Escape pressed - Cancel drawing mode');
      dispatch({ type: 'SELECT_TOOL', payload: { tool: 'cursor' } });
    }

    if (e.key === '+' || e.key === '=') {
      e.preventDefault();
      console.log('[KEYBOARD] + pressed - Zoom In');
      dispatch({ type: 'ZOOM_IN', payload: {} });
    }

    if (e.key === '-' || e.key === '_') {
      e.preventDefault();
      console.log('[KEYBOARD] - pressed - Zoom Out');
      dispatch({ type: 'ZOOM_OUT', payload: {} });
    }

    if (e.ctrlKey && e.key === 'z') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+Z pressed - Undo');
      dispatch({ type: 'UNDO', payload: {} });
    }

    if (e.ctrlKey && e.key === 'y') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+Y pressed - Redo');
      dispatch({ type: 'REDO', payload: {} });
    }

    if (e.key === 'Delete') {
      e.preventDefault();
      console.log('[KEYBOARD] Delete pressed - Delete selected drawing');
      dispatch({ type: 'DELETE_SELECTED_DRAWING', payload: {} });
    }

    if (e.ctrlKey && e.key === 'a') {
      e.preventDefault();
      console.log('[KEYBOARD] Ctrl+A pressed - Select all drawings');
      dispatch({ type: 'SELECT_ALL_DRAWINGS', payload: {} });
    }

    if (e.key === 'c' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] C pressed - Toggle Crosshair');
      dispatch({ type: 'TOGGLE_CROSSHAIR', payload: {} });
    }

    if (e.key === 't' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] T pressed - Select Trendline Tool');
      dispatch({ type: 'SELECT_TOOL', payload: { tool: 'trendline' } });
    }

    if (e.key === 'h' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] H pressed - Select Horizontal Line Tool');
      dispatch({ type: 'SELECT_TOOL', payload: { tool: 'hline' } });
    }

    if (e.key === 'v' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] V pressed - Select Vertical Line Tool');
      dispatch({ type: 'SELECT_TOOL', payload: { tool: 'vline' } });
    }

    if (e.key === 'x' && !e.ctrlKey) {
      e.preventDefault();
      console.log('[KEYBOARD] X pressed - Select Text Tool');
      dispatch({ type: 'SELECT_TOOL', payload: { tool: 'text' } });
    }
  };

  console.log('[KEYBOARD] Registering keyboard shortcuts with input check');
  document.addEventListener('keydown', handleKeyDown);

  return () => {
    console.log('[KEYBOARD] Unregistering keyboard shortcuts');
    document.removeEventListener('keydown', handleKeyDown);
  };
}

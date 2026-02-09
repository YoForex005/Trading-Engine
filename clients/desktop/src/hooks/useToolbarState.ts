import { useReducer, useEffect } from 'react';
import type { ToolbarAction } from '../store/toolbarState';
import { toolbarReducer, initialToolbarState } from '../store/toolbarState';
import { useCommandBus } from './useCommandBus';
import type { Command } from '../types/commands';

export function useToolbarState() {
  const [state, dispatchLocal] = useReducer(toolbarReducer, initialToolbarState);
  const { dispatch: dispatchCommand, subscribe } = useCommandBus();

  // Subscribe to command bus for incoming commands that affect toolbar state
  useEffect(() => {
    const unsubscribe = subscribe((cmd: Command) => {
      // Map commands to toolbar actions
      switch (cmd.type) {
        case 'SELECT_TOOL':
          dispatchLocal({ type: 'SELECT_TOOL', tool: cmd.payload.tool, subtype: (cmd.payload as any).subtype });
          break;

        case 'SET_CHART_TYPE':
          dispatchLocal({ type: 'SET_CHART_TYPE', chartType: cmd.payload.chartType });
          break;

        case 'SET_TIMEFRAME':
          dispatchLocal({ type: 'SET_TIMEFRAME', timeframe: cmd.payload.timeframe });
          break;

        case 'TOGGLE_CROSSHAIR':
          dispatchLocal({ type: 'TOGGLE_CROSSHAIR' });
          break;

        case 'ZOOM_IN':
          dispatchLocal({ type: 'ZOOM', delta: 2 });
          break;

        case 'ZOOM_OUT':
          dispatchLocal({ type: 'ZOOM', delta: -2 });
          break;

        case 'DELETE_DRAWING':
          dispatchLocal({ type: 'DELETE_DRAWING', id: cmd.payload.id });
          break;

        case 'SET_AUTO_SCROLL':
          dispatchLocal({ type: 'TOGGLE_AUTO_SCROLL' });
          break;

        case 'SET_CHART_SHIFT':
          dispatchLocal({ type: 'TOGGLE_CHART_SHIFT' });
          break;

        default:
          break;
      }
    });

    return unsubscribe;
  }, [subscribe]);

  // Wrapper to dispatch both locally and to command bus
  const dispatch = (action: ToolbarAction) => {
    // Update local state
    dispatchLocal(action);

    // Emit to command bus for other components (map ToolbarAction to Command)
    // Note: Not all ToolbarActions map to Commands - some are internal only
    switch (action.type) {
      case 'SELECT_TOOL':
        dispatchCommand({
          type: 'SELECT_TOOL',
          payload: { tool: action.tool || 'cursor', subtype: action.subtype }
        } as Command);
        break;

      case 'SET_CHART_TYPE':
        dispatchCommand({
          type: 'SET_CHART_TYPE',
          payload: { chartType: action.chartType }
        } as Command);
        break;

      case 'SET_TIMEFRAME':
        dispatchCommand({
          type: 'SET_TIMEFRAME',
          payload: { timeframe: action.timeframe }
        } as Command);
        break;

      case 'TOGGLE_CROSSHAIR':
        dispatchCommand({
          type: 'TOGGLE_CROSSHAIR',
          payload: {}
        } as Command);
        break;

      case 'ZOOM':
        // Map ZOOM action to ZOOM_IN or ZOOM_OUT command based on delta
        if (action.delta > 0) {
          dispatchCommand({
            type: 'ZOOM_IN',
            payload: {}
          } as Command);
        } else {
          dispatchCommand({
            type: 'ZOOM_OUT',
            payload: {}
          } as Command);
        }
        break;

      case 'DELETE_DRAWING':
        dispatchCommand({
          type: 'DELETE_DRAWING',
          payload: { id: action.id }
        } as Command);
        break;
      case 'TOGGLE_AUTO_SCROLL':
        dispatchCommand({
          type: 'SET_AUTO_SCROLL',
          payload: { enabled: !state.autoScrollEnabled }
        } as Command);
        break;

      case 'TOGGLE_CHART_SHIFT':
        dispatchCommand({
          type: 'SET_CHART_SHIFT',
          payload: { enabled: !state.chartShiftEnabled }
        } as Command);
        break;

      default:
        break;
    }
  };

  return {
    state,
    dispatch
  };
}

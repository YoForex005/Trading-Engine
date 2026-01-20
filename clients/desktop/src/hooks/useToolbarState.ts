import { useReducer, useEffect } from 'react';
import type { ToolbarAction, ToolbarState } from '../store/toolbarState';
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
          dispatchLocal({ type: 'SELECT_TOOL', tool: cmd.payload.tool });
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

        case 'ADD_INDICATOR':
          dispatchLocal({ type: 'ADD_INDICATOR', indicator: cmd.payload.indicator });
          break;

        case 'REMOVE_INDICATOR':
          dispatchLocal({ type: 'REMOVE_INDICATOR', name: cmd.payload.name });
          break;

        case 'ADD_DRAWING':
          dispatchLocal({ type: 'ADD_DRAWING', drawing: cmd.payload.drawing });
          break;

        case 'DELETE_DRAWING':
          dispatchLocal({ type: 'DELETE_DRAWING', id: cmd.payload.id });
          break;

        case 'CLEAR_DRAWINGS':
          dispatchLocal({ type: 'CLEAR_DRAWINGS' });
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
          payload: { tool: action.tool || 'cursor' }
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

      case 'ADD_INDICATOR':
        dispatchCommand({
          type: 'ADD_INDICATOR',
          payload: {
            name: action.indicator.name,
            params: action.indicator.params
          }
        } as Command);
        break;

      case 'DELETE_DRAWING':
        dispatchCommand({
          type: 'DELETE_DRAWING',
          payload: { id: action.id }
        } as Command);
        break;

      // These actions are internal state changes and don't need to broadcast commands:
      // - REMOVE_INDICATOR (internal state management)
      // - ADD_DRAWING (drawings are handled via chart manager)
      // - CLEAR_DRAWINGS (internal state management)
      default:
        break;
    }
  };

  return {
    state,
    dispatch
  };
}

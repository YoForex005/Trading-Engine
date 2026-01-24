export interface ToolbarState {
  activeTool: 'cursor' | 'trendline' | 'hline' | 'vline' | 'text' | null;
  chartType: 'candlestick' | 'bar' | 'line' | 'area';
  timeframe: string;
  crosshairEnabled: boolean;
  candleWidth: number; // for zoom
  indicators: Array<{ name: string; params: any }>;
  drawings: Array<{ id: string; type: string; points: any[] }>;
}

export type ToolbarAction =
  | { type: 'SELECT_TOOL'; tool: ToolbarState['activeTool'] }
  | { type: 'SET_CHART_TYPE'; chartType: ToolbarState['chartType'] }
  | { type: 'SET_TIMEFRAME'; timeframe: string }
  | { type: 'TOGGLE_CROSSHAIR' }
  | { type: 'ZOOM'; delta: number }
  | { type: 'ADD_INDICATOR'; indicator: { name: string; params: any } }
  | { type: 'REMOVE_INDICATOR'; name: string }
  | { type: 'ADD_DRAWING'; drawing: { id: string; type: string; points: any[] } }
  | { type: 'DELETE_DRAWING'; id: string }
  | { type: 'CLEAR_DRAWINGS' };

export const initialToolbarState: ToolbarState = {
  activeTool: null,
  chartType: 'candlestick',
  timeframe: 'm1',
  crosshairEnabled: true,
  candleWidth: 10,
  indicators: [],
  drawings: []
};

export function toolbarReducer(state: ToolbarState, action: ToolbarAction): ToolbarState {
  switch (action.type) {
    case 'SELECT_TOOL':
      // Only one drawing tool can be active at a time
      // Selecting 'cursor' deselects all drawing tools
      return {
        ...state,
        activeTool: action.tool === 'cursor' ? null : action.tool
      };

    case 'SET_CHART_TYPE':
      return {
        ...state,
        chartType: action.chartType
      };

    case 'SET_TIMEFRAME':
      return {
        ...state,
        timeframe: action.timeframe
      };

    case 'TOGGLE_CROSSHAIR':
      return {
        ...state,
        crosshairEnabled: !state.crosshairEnabled
      };

    case 'ZOOM':
      // Zoom by adjusting candle width
      // Minimum: 2px, Maximum: 50px
      const newWidth = Math.max(2, Math.min(50, state.candleWidth + action.delta));
      return {
        ...state,
        candleWidth: newWidth
      };

    case 'ADD_INDICATOR':
      // Prevent duplicate indicators
      if (state.indicators.some(ind => ind.name === action.indicator.name)) {
        return state;
      }
      return {
        ...state,
        indicators: [...state.indicators, action.indicator]
      };

    case 'REMOVE_INDICATOR':
      return {
        ...state,
        indicators: state.indicators.filter(ind => ind.name !== action.name)
      };

    case 'ADD_DRAWING':
      return {
        ...state,
        drawings: [...state.drawings, action.drawing]
      };

    case 'DELETE_DRAWING':
      return {
        ...state,
        drawings: state.drawings.filter(d => d.id !== action.id)
      };

    case 'CLEAR_DRAWINGS':
      return {
        ...state,
        drawings: []
      };

    default:
      return state;
  }
}

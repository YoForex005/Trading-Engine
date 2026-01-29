import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Ticker } from '@/types';

interface MarketState {
  tickers: Record<string, Ticker>;
  watchlist: string[];
}

const initialState: MarketState = {
  tickers: {},
  watchlist: ['BTCUSD', 'ETHUSD', 'EURUSD', 'GBPUSD'],
};

const marketSlice = createSlice({
  name: 'market',
  initialState,
  reducers: {
    updateTicker: (state, action: PayloadAction<Ticker>) => {
      state.tickers[action.payload.symbol] = action.payload;
    },
    updateTickers: (state, action: PayloadAction<Ticker[]>) => {
      action.payload.forEach(ticker => {
        state.tickers[ticker.symbol] = ticker;
      });
    },
    addToWatchlist: (state, action: PayloadAction<string>) => {
      if (!state.watchlist.includes(action.payload)) {
        state.watchlist.push(action.payload);
      }
    },
    removeFromWatchlist: (state, action: PayloadAction<string>) => {
      state.watchlist = state.watchlist.filter(s => s !== action.payload);
    },
  },
});

export const { updateTicker, updateTickers, addToWatchlist, removeFromWatchlist } =
  marketSlice.actions;
export default marketSlice.reducer;

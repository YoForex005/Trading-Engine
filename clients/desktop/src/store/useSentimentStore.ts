/**
 * Sentiment Analysis Store with Zustand
 * Manages market sentiment data and indicators
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export type SentimentTrend = 'up' | 'down' | 'neutral';
export type NewsSentiment = 'positive' | 'negative' | 'neutral';
export type FearGreedLevel = 'Extreme Fear' | 'Fear' | 'Neutral' | 'Greed' | 'Extreme Greed';

export interface SymbolSentiment {
  symbol: string;
  bullishPercent: number; // 0-100
  bearishPercent: number; // 0-100
  trend: SentimentTrend;
  retailLongPercent: number; // 0-100
  retailShortPercent: number; // 0-100
}

export interface NewsItem {
  id: string;
  headline: string;
  source: string;
  timestamp: Date;
  sentiment: NewsSentiment;
  impactScore: number; // 0-100
}

export interface SentimentIndicators {
  fearGreedIndex: number; // 0-100
  fearGreedLevel: FearGreedLevel;
  vixLevel: number;
  putCallRatio: number;
  netSpeculativePositions: number; // -100 to 100 (negative = net short, positive = net long)
}

interface SentimentState {
  overallSentiment: number; // 0-100 (0 = extreme bearish, 50 = neutral, 100 = extreme bullish)
  symbolSentiments: SymbolSentiment[];
  indicators: SentimentIndicators;
  news: NewsItem[];
  autoRefresh: boolean;
  refreshInterval: number; // seconds
  lastUpdate: Date;

  setAutoRefresh: (enabled: boolean) => void;
  setRefreshInterval: (seconds: number) => void;
  refreshData: () => void;
}

// Mock data generators
function generateSymbolSentiments(): SymbolSentiment[] {
  const symbols = [
    'EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD',
    'NZDUSD', 'USDCHF', 'EURGBP', 'EURJPY', 'GBPJPY',
    'XAUUSD', 'BTCUSD', 'ETHUSD', 'SPX500', 'US30'
  ];

  return symbols.map((symbol) => {
    // Generate bullish/bearish split that adds up to ~100%
    const bullishPercent = 30 + Math.random() * 40; // 30-70%
    const bearishPercent = 100 - bullishPercent;

    // Retail positioning (often contrarian to professional traders)
    const retailLongPercent = 20 + Math.random() * 60; // 20-80%
    const retailShortPercent = 100 - retailLongPercent;

    // Determine trend
    let trend: SentimentTrend;
    if (bullishPercent > 60) trend = 'up';
    else if (bearishPercent > 60) trend = 'down';
    else trend = 'neutral';

    return {
      symbol,
      bullishPercent: Math.round(bullishPercent * 10) / 10,
      bearishPercent: Math.round(bearishPercent * 10) / 10,
      trend,
      retailLongPercent: Math.round(retailLongPercent * 10) / 10,
      retailShortPercent: Math.round(retailShortPercent * 10) / 10,
    };
  });
}

function generateNewsItems(): NewsItem[] {
  const headlines = [
    { text: 'Fed signals potential rate cut in Q2', sentiment: 'positive' as NewsSentiment, impact: 85 },
    { text: 'Global trade tensions ease amid new agreements', sentiment: 'positive' as NewsSentiment, impact: 70 },
    { text: 'Tech sector shows signs of weakness', sentiment: 'negative' as NewsSentiment, impact: 60 },
    { text: 'Oil prices stabilize after recent volatility', sentiment: 'neutral' as NewsSentiment, impact: 50 },
    { text: 'Unemployment rate drops to 3-year low', sentiment: 'positive' as NewsSentiment, impact: 75 },
    { text: 'Central banks maintain dovish stance', sentiment: 'positive' as NewsSentiment, impact: 80 },
    { text: 'Manufacturing PMI disappoints expectations', sentiment: 'negative' as NewsSentiment, impact: 65 },
    { text: 'Retail sales exceed forecasts', sentiment: 'positive' as NewsSentiment, impact: 70 },
    { text: 'Geopolitical risks weigh on market sentiment', sentiment: 'negative' as NewsSentiment, impact: 75 },
    { text: 'Corporate earnings season begins with mixed results', sentiment: 'neutral' as NewsSentiment, impact: 55 },
  ];

  const sources = ['Reuters', 'Bloomberg', 'Financial Times', 'Wall Street Journal', 'CNBC'];

  const news: NewsItem[] = [];
  for (let i = 0; i < 5; i++) {
    const headline = headlines[Math.floor(Math.random() * headlines.length)];
    const minutesAgo = Math.floor(Math.random() * 180); // Last 3 hours

    news.push({
      id: `news-${Date.now()}-${i}`,
      headline: headline.text,
      source: sources[Math.floor(Math.random() * sources.length)],
      timestamp: new Date(Date.now() - minutesAgo * 60 * 1000),
      sentiment: headline.sentiment,
      impactScore: headline.impact,
    });
  }

  return news.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}

function generateSentimentIndicators(): SentimentIndicators {
  const fearGreedIndex = 20 + Math.random() * 60; // 20-80

  let fearGreedLevel: FearGreedLevel;
  if (fearGreedIndex < 25) fearGreedLevel = 'Extreme Fear';
  else if (fearGreedIndex < 45) fearGreedLevel = 'Fear';
  else if (fearGreedIndex < 55) fearGreedLevel = 'Neutral';
  else if (fearGreedIndex < 75) fearGreedLevel = 'Greed';
  else fearGreedLevel = 'Extreme Greed';

  const vixLevel = 12 + Math.random() * 25; // 12-37
  const putCallRatio = 0.7 + Math.random() * 0.6; // 0.7-1.3
  const netSpeculativePositions = -50 + Math.random() * 100; // -50 to 50

  return {
    fearGreedIndex: Math.round(fearGreedIndex * 10) / 10,
    fearGreedLevel,
    vixLevel: Math.round(vixLevel * 10) / 10,
    putCallRatio: Math.round(putCallRatio * 100) / 100,
    netSpeculativePositions: Math.round(netSpeculativePositions * 10) / 10,
  };
}

function calculateOverallSentiment(symbolSentiments: SymbolSentiment[]): number {
  if (symbolSentiments.length === 0) return 50;

  const avgBullish = symbolSentiments.reduce((sum, s) => sum + s.bullishPercent, 0) / symbolSentiments.length;
  return Math.round(avgBullish * 10) / 10;
}

export const useSentimentStore = create<SentimentState>()(
  devtools(
    (set, get) => {
      const symbolSentiments = generateSymbolSentiments();
      const overallSentiment = calculateOverallSentiment(symbolSentiments);

      return {
        overallSentiment,
        symbolSentiments,
        indicators: generateSentimentIndicators(),
        news: generateNewsItems(),
        autoRefresh: true,
        refreshInterval: 30,
        lastUpdate: new Date(),

        setAutoRefresh: (enabled) => set({ autoRefresh: enabled }),
        setRefreshInterval: (seconds) => set({ refreshInterval: seconds }),

        refreshData: () => {
          const symbolSentiments = generateSymbolSentiments();
          const overallSentiment = calculateOverallSentiment(symbolSentiments);

          set({
            overallSentiment,
            symbolSentiments,
            indicators: generateSentimentIndicators(),
            news: generateNewsItems(),
            lastUpdate: new Date(),
          });
        },
      };
    },
    { name: 'sentiment-store' }
  )
);

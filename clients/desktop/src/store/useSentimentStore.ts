/**
 * Sentiment Analysis Store with Zustand
 * Manages market sentiment data and indicators
 * Connected to backend: GET /admin/market-news/stats (for sentiment aggregates)
 * and GET /admin/market-news/articles (for news with sentiment scores)
 *
 * NOTE: No dedicated /api/sentiment endpoint exists in the backend.
 * Sentiment data is derived from market-news articles' sentiment scores
 * and the stats endpoint. Mock data is used as fallback.
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { API_ENDPOINTS } from '../config/api';

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
  isLoading: boolean;
  error: string | null;

  setAutoRefresh: (enabled: boolean) => void;
  setRefreshInterval: (seconds: number) => void;
  refreshData: () => void;
  fetchSentimentData: () => Promise<void>;
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

// Fetch sentiment-related data from backend
// Uses /admin/market-news/stats for aggregate sentiment and /admin/market-news/articles for news items
async function fetchSentimentFromAPI(): Promise<{
  news: NewsItem[];
  overallBullish: number;
} | null> {
  // Get auth token
  let authToken: string | null = null;
  try {
    const { useAppStore } = await import('./useAppStore');
    authToken = useAppStore.getState().authToken;
  } catch {
    // fallback
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }

  // Fetch articles with sentiment data
  const articlesRes = await fetch(API_ENDPOINTS.marketNews.articles, { headers });
  if (!articlesRes.ok) {
    throw new Error(`Failed to fetch articles: ${articlesRes.status}`);
  }

  const articlesData = await articlesRes.json();
  const articles = Array.isArray(articlesData) ? articlesData : (articlesData.articles || articlesData.data || []);

  // Transform articles into sentiment NewsItems
  const news: NewsItem[] = articles.slice(0, 10).map((article: any, idx: number) => {
    const sentiment = article.sentiment || { bullish: 0.5, bearish: 0.3, neutral: 0.2, overall: 'neutral' };
    let newsSentiment: NewsSentiment = 'neutral';
    if (sentiment.overall === 'bullish' || sentiment.bullish > 0.5) newsSentiment = 'positive';
    else if (sentiment.overall === 'bearish' || sentiment.bearish > 0.5) newsSentiment = 'negative';

    return {
      id: `news-${article.id || Date.now()}-${idx}`,
      headline: article.title || '',
      source: article.source || 'Unknown',
      timestamp: new Date(article.published_at || Date.now()),
      sentiment: newsSentiment,
      impactScore: Math.round((Math.max(sentiment.bullish, sentiment.bearish) || 0.5) * 100),
    };
  }).sort((a: NewsItem, b: NewsItem) => b.timestamp.getTime() - a.timestamp.getTime());

  // Calculate overall bullish from articles
  let totalBullish = 0;
  let count = 0;
  for (const article of articles) {
    if (article.sentiment) {
      totalBullish += article.sentiment.bullish || 0;
      count++;
    }
  }
  const overallBullish = count > 0 ? (totalBullish / count) * 100 : 50;

  return { news, overallBullish };
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
        isLoading: false,
        error: null,

        setAutoRefresh: (enabled) => set({ autoRefresh: enabled }),
        setRefreshInterval: (seconds) => set({ refreshInterval: seconds }),

        fetchSentimentData: async () => {
          set({ isLoading: true, error: null });
          try {
            const result = await fetchSentimentFromAPI();
            if (result && result.news.length > 0) {
              set({
                news: result.news,
                overallSentiment: Math.round(result.overallBullish * 10) / 10,
                lastUpdate: new Date(),
                isLoading: false,
              });
            } else {
              set({ isLoading: false });
            }
          } catch (err: any) {
            console.warn('[SentimentStore] Failed to fetch from API, using mock data:', err.message);
            set({ isLoading: false, error: err.message });
          }
        },

        refreshData: () => {
          // Try API first, fall back to mock data
          get().fetchSentimentData().catch(() => {
            const symbolSentiments = generateSymbolSentiments();
            const overallSentiment = calculateOverallSentiment(symbolSentiments);

            set({
              overallSentiment,
              symbolSentiments,
              indicators: generateSentimentIndicators(),
              news: generateNewsItems(),
              lastUpdate: new Date(),
            });
          });
        },
      };
    },
    { name: 'sentiment-store' }
  )
);

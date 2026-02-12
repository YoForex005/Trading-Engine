/**
 * News Store with Zustand
 * Manages financial news feed (ephemeral, no persistence)
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export type NewsImpact = 'high' | 'medium' | 'low';
export type NewsCategory = 'All' | 'Forex' | 'Crypto' | 'Economy' | 'Central Banks';

export interface NewsItem {
  id: string;
  headline: string;
  summary: string;
  source: string;
  impact: NewsImpact;
  category: Exclude<NewsCategory, 'All'>;
  symbols: string[];
  timestamp: Date;
  isRead: boolean;
}

interface NewsState {
  news: NewsItem[];
  activeCategory: NewsCategory;
  setNews: (news: NewsItem[]) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  setActiveCategory: (category: NewsCategory) => void;
  refreshNews: () => void;
  getFilteredNews: () => NewsItem[];
}

// Mock news generator
function generateMockNews(): NewsItem[] {
  const newsItems: Omit<NewsItem, 'id' | 'timestamp' | 'isRead'>[] = [
    {
      headline: 'Fed Holds Rates Steady at 5.25-5.50%',
      summary: 'Federal Reserve maintains current interest rate range, citing balanced inflation risks and stable labor market. Officials signal no rate cuts expected before mid-year.',
      source: 'Reuters',
      impact: 'high',
      category: 'Central Banks',
      symbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD'],
    },
    {
      headline: 'ECB Signals Rate Cut in April Meeting',
      summary: 'European Central Bank President Christine Lagarde hints at possible rate reduction in April, as eurozone inflation continues to moderate. Market expects 25bp cut.',
      source: 'Bloomberg',
      impact: 'high',
      category: 'Central Banks',
      symbols: ['EURUSD', 'EURGBP', 'EURJPY'],
    },
    {
      headline: 'Bitcoin Surges Past $105,000 on ETF Inflows',
      summary: 'Bitcoin reaches new all-time high as institutional investors pour billions into spot Bitcoin ETFs. Record trading volumes reported across major exchanges.',
      source: 'CoinDesk',
      impact: 'medium',
      category: 'Crypto',
      symbols: ['BTCUSD', 'ETHUSD'],
    },
    {
      headline: 'UK GDP Grows 0.3% in Q4, Beating Expectations',
      summary: 'British economy shows resilience with stronger-than-forecast growth in fourth quarter. Services sector drives expansion, manufacturing remains weak.',
      source: 'Financial Times',
      impact: 'medium',
      category: 'Economy',
      symbols: ['GBPUSD', 'EURGBP', 'GBPJPY'],
    },
    {
      headline: 'OPEC+ Extends Production Cuts Through Q2',
      summary: 'Oil cartel agrees to maintain output restrictions until June, supporting crude prices. Saudi Arabia and Russia reaffirm commitment to supply discipline.',
      source: 'Reuters',
      impact: 'medium',
      category: 'Economy',
      symbols: ['WTICOUSD', 'BCOUSD'],
    },
    {
      headline: 'Japan Core Inflation Accelerates to 2.8% YoY',
      summary: 'Japanese consumer prices rise faster than expected, pressuring Bank of Japan to reconsider ultra-loose monetary policy. Yen strengthens on speculation.',
      source: 'Nikkei',
      impact: 'high',
      category: 'Economy',
      symbols: ['USDJPY', 'EURJPY', 'GBPJPY'],
    },
    {
      headline: 'Ethereum Upgrade Completed Successfully',
      summary: 'Ethereum network implements Dencun upgrade, reducing transaction costs and improving scalability. ETH price rallies 8% on technical milestone.',
      source: 'CoinTelegraph',
      impact: 'medium',
      category: 'Crypto',
      symbols: ['ETHUSD'],
    },
    {
      headline: 'US Non-Farm Payrolls Surge by 275K in January',
      summary: 'American employers add jobs at robust pace, while unemployment rate ticks up to 3.9%. Wage growth moderates, easing inflation concerns.',
      source: 'Bloomberg',
      impact: 'high',
      category: 'Economy',
      symbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD'],
    },
    {
      headline: 'Australian Dollar Rallies on China Trade Data',
      summary: 'AUD strengthens as Chinese imports surge, boosting demand outlook for Australian commodities. RBA maintains hawkish stance on inflation.',
      source: 'Reuters',
      impact: 'medium',
      category: 'Forex',
      symbols: ['AUDUSD', 'AUDNZD', 'AUDJPY'],
    },
    {
      headline: 'Swiss National Bank Unexpectedly Cuts Rates',
      summary: 'SNB surprises markets with 50bp rate reduction to counter franc strength. CHF falls sharply against euro and dollar on dovish pivot.',
      source: 'Reuters',
      impact: 'high',
      category: 'Central Banks',
      symbols: ['USDCHF', 'EURCHF', 'GBPCHF'],
    },
    {
      headline: 'Gold Hits Record High Above $2,850/oz',
      summary: 'Precious metal rallies to new peak amid geopolitical tensions and central bank buying. Investors seek safe-haven assets as equity volatility rises.',
      source: 'Kitco',
      impact: 'medium',
      category: 'Economy',
      symbols: ['XAUUSD', 'XAGUSD'],
    },
    {
      headline: 'Canadian Dollar Drops on Weak Retail Sales',
      summary: 'Canadian retail sales fall 1.2% in December, missing forecasts. Bank of Canada rate cut expectations rise, pressuring loonie lower.',
      source: 'Bloomberg',
      impact: 'medium',
      category: 'Economy',
      symbols: ['USDCAD', 'AUDCAD', 'EURCAD'],
    },
    {
      headline: 'Ripple Wins SEC Lawsuit, XRP Soars 40%',
      summary: 'Federal court rules in favor of Ripple Labs in securities case. XRP cryptocurrency surges as regulatory uncertainty clears.',
      source: 'CoinDesk',
      impact: 'medium',
      category: 'Crypto',
      symbols: ['XRPUSD'],
    },
    {
      headline: 'German Factory Orders Decline for Third Month',
      summary: 'Industrial orders in Germany drop 3.7% in December, signaling economic weakness. EUR under pressure as recession risks mount.',
      source: 'Reuters',
      impact: 'medium',
      category: 'Economy',
      symbols: ['EURUSD', 'EURGBP', 'EURJPY'],
    },
    {
      headline: 'New Zealand Central Bank Holds Rates at 5.5%',
      summary: 'RBNZ keeps benchmark rate unchanged but signals potential cuts later in year. NZD weakens on dovish commentary.',
      source: 'Bloomberg',
      impact: 'medium',
      category: 'Central Banks',
      symbols: ['NZDUSD', 'AUDNZD', 'NZDJPY'],
    },
    {
      headline: 'Oil Inventories Fall More Than Expected',
      summary: 'US crude stockpiles decline 9.2 million barrels, largest drop in six months. WTI crude rallies above $82 per barrel on supply tightness.',
      source: 'Reuters',
      impact: 'low',
      category: 'Economy',
      symbols: ['WTICOUSD', 'USDCAD'],
    },
    {
      headline: 'Bank of England Cuts Rates by 25bp to 4.75%',
      summary: 'BoE delivers second consecutive rate cut as UK inflation falls to 2.1%. Pound falls against major currencies on dovish outlook.',
      source: 'Financial Times',
      impact: 'high',
      category: 'Central Banks',
      symbols: ['GBPUSD', 'EURGBP', 'GBPJPY'],
    },
    {
      headline: 'Solana Network Outage Triggers 12% Price Drop',
      summary: 'SOL cryptocurrency tumbles after blockchain experiences 5-hour downtime. Developers cite validator consensus issue, network now restored.',
      source: 'CoinTelegraph',
      impact: 'low',
      category: 'Crypto',
      symbols: ['SOLUSD'],
    },
    {
      headline: 'Chinese Yuan Strengthens to 6-Month High',
      summary: 'CNY appreciates as PBOC injects liquidity and signals support. Asian currencies rally alongside yuan strength.',
      source: 'Reuters',
      impact: 'medium',
      category: 'Forex',
      symbols: ['USDCNH', 'AUDUSD', 'NZDUSD'],
    },
    {
      headline: 'US Consumer Confidence Index Jumps to 112.8',
      summary: 'American consumer sentiment improves sharply in February, exceeding analyst estimates. Labor market optimism drives gains.',
      source: 'Conference Board',
      impact: 'low',
      category: 'Economy',
      symbols: ['EURUSD', 'GBPUSD', 'USDJPY'],
    },
  ];

  // Generate news items with randomized timestamps (last 24 hours)
  return newsItems.map((item, index) => {
    const minutesAgo = Math.floor(Math.random() * 1440); // Random time in last 24 hours
    const timestamp = new Date(Date.now() - minutesAgo * 60 * 1000);

    return {
      ...item,
      id: `news-${index + 1}`,
      timestamp,
      isRead: Math.random() > 0.6, // 40% unread
    };
  }).sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}

export const useNewsStore = create<NewsState>()(
  devtools(
    (set, get) => ({
      news: generateMockNews(),
      activeCategory: 'All',

      setNews: (news) => set({ news }),

      markAsRead: (id) =>
        set((state) => ({
          news: state.news.map((item) =>
            item.id === id ? { ...item, isRead: true } : item
          ),
        })),

      markAllAsRead: () =>
        set((state) => ({
          news: state.news.map((item) => ({ ...item, isRead: true })),
        })),

      setActiveCategory: (category) => set({ activeCategory: category }),

      refreshNews: () => {
        // Simulate refresh by updating timestamps
        set((state) => ({
          news: state.news.map((item) => ({
            ...item,
            timestamp: new Date(Date.now() - Math.floor(Math.random() * 1440) * 60 * 1000),
            isRead: Math.random() > 0.6,
          })).sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime()),
        }));
      },

      getFilteredNews: () => {
        const { news, activeCategory } = get();
        if (activeCategory === 'All') return news;
        return news.filter((item) => item.category === activeCategory);
      },
    }),
    { name: 'news-store' }
  )
);

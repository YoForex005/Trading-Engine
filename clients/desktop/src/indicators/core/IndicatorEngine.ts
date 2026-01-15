import * as TI from 'technicalindicators';

// OHLC data structure
export type OHLC = {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
};

// Indicator types - comprehensive list for MT5 parity
export type IndicatorType =
  // Trend Indicators
  | 'SMA' | 'EMA' | 'WMA' | 'VWAP'
  | 'BollingerBands' | 'KeltnerChannels' | 'DonchianChannels'
  // Momentum Indicators
  | 'RSI' | 'Stochastic' | 'MACD' | 'CCI' | 'WilliamsR' | 'ROC' | 'Momentum'
  // Volatility Indicators
  | 'ATR' | 'StandardDeviation'
  // Volume Indicators
  | 'OBV' | 'MFI' | 'ADL' | 'ForceIndex'
  // Oscillators
  | 'AwesomeOscillator' | 'StochasticRSI'
  // Other Popular Indicators
  | 'Ichimoku' | 'ParabolicSAR' | 'ADX' | 'Aroon' | 'TRIX';

// Indicator display mode
export type IndicatorDisplayMode = 'overlay' | 'pane';

// Indicator parameter types
export type IndicatorParams = {
  period?: number;
  fastPeriod?: number;
  slowPeriod?: number;
  signalPeriod?: number;
  stdDev?: number;
  kPeriod?: number;
  dPeriod?: number;
  smooth?: number;
  multiplier?: number;
  acceleration?: number;
  maximum?: number;
  // For multi-output indicators
  [key: string]: number | string | undefined;
};

// Indicator result can be single value or multi-value (e.g., MACD has histogram, signal, MACD)
export type IndicatorValue = number | Record<string, number>;

export type IndicatorResult = {
  time: number;
  value: IndicatorValue;
}[];

// Indicator metadata
export type IndicatorMeta = {
  name: string;
  type: IndicatorType;
  category: 'trend' | 'momentum' | 'volatility' | 'volume' | 'oscillator';
  displayMode: IndicatorDisplayMode;
  defaultParams: IndicatorParams;
  outputs: string[]; // e.g., ['MACD', 'signal', 'histogram'] for MACD
  description: string;
  formula?: string;
};

// Indicator Engine - main calculation class
export class IndicatorEngine {
  // Get default parameters for an indicator
  static getDefaultParams(type: IndicatorType): IndicatorParams {
    const defaults: Record<IndicatorType, IndicatorParams> = {
      // Trend
      SMA: { period: 20 },
      EMA: { period: 20 },
      WMA: { period: 20 },
      VWAP: {},
      BollingerBands: { period: 20, stdDev: 2 },
      KeltnerChannels: { period: 20, multiplier: 2 },
      DonchianChannels: { period: 20 },
      // Momentum
      RSI: { period: 14 },
      Stochastic: { kPeriod: 14, dPeriod: 3, smooth: 3 },
      MACD: { fastPeriod: 12, slowPeriod: 26, signalPeriod: 9 },
      CCI: { period: 20 },
      WilliamsR: { period: 14 },
      ROC: { period: 12 },
      Momentum: { period: 10 },
      // Volatility
      ATR: { period: 14 },
      StandardDeviation: { period: 20 },
      // Volume
      OBV: {},
      MFI: { period: 14 },
      ADL: {},
      ForceIndex: { period: 13 },
      // Oscillators
      AwesomeOscillator: { fastPeriod: 5, slowPeriod: 34 },
      StochasticRSI: { rsiPeriod: 14, kPeriod: 14, dPeriod: 3, smooth: 3 },
      // Others
      Ichimoku: { conversionPeriod: 9, basePeriod: 26, spanPeriod: 52, displacement: 26 },
      ParabolicSAR: { acceleration: 0.02, maximum: 0.2 },
      ADX: { period: 14 },
      Aroon: { period: 25 },
      TRIX: { period: 15 },
    };
    return defaults[type] || {};
  }

  // Get metadata for an indicator
  static getMeta(type: IndicatorType): IndicatorMeta {
    const metadata: Record<IndicatorType, Omit<IndicatorMeta, 'type'>> = {
      // Trend Indicators
      SMA: {
        name: 'Simple Moving Average',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('SMA'),
        outputs: ['SMA'],
        description: 'Average of prices over a specified period',
      },
      EMA: {
        name: 'Exponential Moving Average',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('EMA'),
        outputs: ['EMA'],
        description: 'Weighted average giving more weight to recent prices',
      },
      WMA: {
        name: 'Weighted Moving Average',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('WMA'),
        outputs: ['WMA'],
        description: 'Linear weighted moving average',
      },
      VWAP: {
        name: 'Volume Weighted Average Price',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('VWAP'),
        outputs: ['VWAP'],
        description: 'Average price weighted by volume',
      },
      BollingerBands: {
        name: 'Bollinger Bands',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('BollingerBands'),
        outputs: ['upper', 'middle', 'lower'],
        description: 'Volatility bands around moving average',
      },
      KeltnerChannels: {
        name: 'Keltner Channels',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('KeltnerChannels'),
        outputs: ['upper', 'middle', 'lower'],
        description: 'ATR-based volatility channels',
      },
      DonchianChannels: {
        name: 'Donchian Channels',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('DonchianChannels'),
        outputs: ['upper', 'middle', 'lower'],
        description: 'Highest high and lowest low channels',
      },
      // Momentum Indicators
      RSI: {
        name: 'Relative Strength Index',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('RSI'),
        outputs: ['RSI'],
        description: 'Momentum oscillator (0-100)',
      },
      Stochastic: {
        name: 'Stochastic Oscillator',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('Stochastic'),
        outputs: ['k', 'd'],
        description: 'Momentum indicator comparing close to high-low range',
      },
      MACD: {
        name: 'Moving Average Convergence Divergence',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('MACD'),
        outputs: ['MACD', 'signal', 'histogram'],
        description: 'Trend-following momentum indicator',
      },
      CCI: {
        name: 'Commodity Channel Index',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('CCI'),
        outputs: ['CCI'],
        description: 'Measures deviation from average price',
      },
      WilliamsR: {
        name: 'Williams %R',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('WilliamsR'),
        outputs: ['WilliamsR'],
        description: 'Momentum indicator (0 to -100)',
      },
      ROC: {
        name: 'Rate of Change',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('ROC'),
        outputs: ['ROC'],
        description: 'Percentage change in price',
      },
      Momentum: {
        name: 'Momentum',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('Momentum'),
        outputs: ['Momentum'],
        description: 'Current price minus price n periods ago',
      },
      // Volatility Indicators
      ATR: {
        name: 'Average True Range',
        category: 'volatility',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('ATR'),
        outputs: ['ATR'],
        description: 'Measures market volatility',
      },
      StandardDeviation: {
        name: 'Standard Deviation',
        category: 'volatility',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('StandardDeviation'),
        outputs: ['SD'],
        description: 'Statistical measure of price dispersion',
      },
      // Volume Indicators
      OBV: {
        name: 'On Balance Volume',
        category: 'volume',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('OBV'),
        outputs: ['OBV'],
        description: 'Cumulative volume based on price direction',
      },
      MFI: {
        name: 'Money Flow Index',
        category: 'volume',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('MFI'),
        outputs: ['MFI'],
        description: 'Volume-weighted RSI',
      },
      ADL: {
        name: 'Accumulation/Distribution Line',
        category: 'volume',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('ADL'),
        outputs: ['ADL'],
        description: 'Volume flow indicator',
      },
      ForceIndex: {
        name: 'Force Index',
        category: 'volume',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('ForceIndex'),
        outputs: ['ForceIndex'],
        description: 'Price change multiplied by volume',
      },
      // Oscillators
      AwesomeOscillator: {
        name: 'Awesome Oscillator',
        category: 'oscillator',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('AwesomeOscillator'),
        outputs: ['AO'],
        description: 'Momentum indicator using SMA of median prices',
      },
      StochasticRSI: {
        name: 'Stochastic RSI',
        category: 'oscillator',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('StochasticRSI'),
        outputs: ['k', 'd'],
        description: 'Stochastic applied to RSI values',
      },
      // Others
      Ichimoku: {
        name: 'Ichimoku Cloud',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('Ichimoku'),
        outputs: ['conversionLine', 'baseLine', 'leadingSpanA', 'leadingSpanB', 'laggingSpan'],
        description: 'Comprehensive trend and support/resistance indicator',
      },
      ParabolicSAR: {
        name: 'Parabolic SAR',
        category: 'trend',
        displayMode: 'overlay',
        defaultParams: this.getDefaultParams('ParabolicSAR'),
        outputs: ['PSAR'],
        description: 'Stop and reverse trend indicator',
      },
      ADX: {
        name: 'Average Directional Index',
        category: 'trend',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('ADX'),
        outputs: ['ADX', '+DI', '-DI'],
        description: 'Trend strength indicator',
      },
      Aroon: {
        name: 'Aroon',
        category: 'trend',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('Aroon'),
        outputs: ['aroonUp', 'aroonDown'],
        description: 'Identifies trend changes',
      },
      TRIX: {
        name: 'TRIX',
        category: 'momentum',
        displayMode: 'pane',
        defaultParams: this.getDefaultParams('TRIX'),
        outputs: ['TRIX'],
        description: 'Triple smoothed EMA momentum indicator',
      },
    };

    return { type, ...metadata[type] };
  }

  // Calculate indicator values
  static calculate(
    type: IndicatorType,
    ohlcData: OHLC[],
    params: IndicatorParams = {}
  ): IndicatorResult {
    const finalParams = { ...this.getDefaultParams(type), ...params };

    // Extract price arrays
    const close = ohlcData.map((d) => d.close);
    // const open = ohlcData.map((d) => d.open); // Reserved for future indicators
    const high = ohlcData.map((d) => d.high);
    const low = ohlcData.map((d) => d.low);
    const volume = ohlcData.map((d) => d.volume || 0);
    const times = ohlcData.map((d) => d.time);

    let results: any[] = [];
    // outputs tracked per indicator type for potential future use in metadata
    // let outputs: string[] = [];

    try {
      switch (type) {
        // Trend Indicators
        case 'SMA':
          results = TI.SMA.calculate({ period: finalParams.period!, values: close });
          // outputs = ['SMA'];
          break;
        case 'EMA':
          results = TI.EMA.calculate({ period: finalParams.period!, values: close });
          // outputs = ['EMA'];
          break;
        case 'WMA':
          results = TI.WMA.calculate({ period: finalParams.period!, values: close });
          // outputs = ['WMA'];
          break;
        case 'VWAP':
          results = TI.VWAP.calculate({ high, low, close, volume });
          // outputs = ['VWAP'];
          break;
        case 'BollingerBands':
          results = TI.BollingerBands.calculate({
            period: finalParams.period!,
            stdDev: finalParams.stdDev!,
            values: close,
          });
          // outputs = ['upper', 'middle', 'lower'];
          break;
        // Momentum Indicators
        case 'RSI':
          results = TI.RSI.calculate({ period: finalParams.period!, values: close });
          // outputs = ['RSI'];
          break;
        case 'Stochastic':
          results = TI.Stochastic.calculate({
            high,
            low,
            close,
            period: finalParams.kPeriod!,
            signalPeriod: finalParams.dPeriod!,
          });
          // outputs = ['k', 'd'];
          break;
        case 'MACD':
          results = TI.MACD.calculate({
            values: close,
            fastPeriod: finalParams.fastPeriod!,
            slowPeriod: finalParams.slowPeriod!,
            signalPeriod: finalParams.signalPeriod!,
            SimpleMAOscillator: false,
            SimpleMASignal: false,
          });
          // outputs = ['MACD', 'signal', 'histogram'];
          break;
        case 'CCI':
          results = TI.CCI.calculate({ high, low, close, period: finalParams.period! });
          // outputs = ['CCI'];
          break;
        case 'WilliamsR':
          results = TI.WilliamsR.calculate({ high, low, close, period: finalParams.period! });
          // outputs = ['WilliamsR'];
          break;
        case 'ROC':
          results = TI.ROC.calculate({ period: finalParams.period!, values: close });
          // outputs = ['ROC'];
          break;
        // Volatility
        case 'ATR':
          results = TI.ATR.calculate({ high, low, close, period: finalParams.period! });
          // outputs = ['ATR'];
          break;
        case 'StandardDeviation':
          results = TI.SD.calculate({ period: finalParams.period!, values: close });
          // outputs = ['SD'];
          break;
        // Volume
        case 'OBV':
          results = TI.OBV.calculate({ close, volume });
          // outputs = ['OBV'];
          break;
        case 'MFI':
          results = TI.MFI.calculate({ high, low, close, volume, period: finalParams.period! });
          // outputs = ['MFI'];
          break;
        case 'ADL':
          results = TI.ADL.calculate({ high, low, close, volume });
          // outputs = ['ADL'];
          break;
        case 'ForceIndex':
          results = TI.ForceIndex.calculate({ close, volume, period: finalParams.period! });
          // outputs = ['ForceIndex'];
          break;
        // Oscillators
        case 'AwesomeOscillator':
          results = TI.AwesomeOscillator.calculate({
            high,
            low,
            fastPeriod: finalParams.fastPeriod!,
            slowPeriod: finalParams.slowPeriod!,
          });
          // outputs = ['AO'];
          break;
        case 'StochasticRSI':
          results = TI.StochasticRSI.calculate({
            values: close,
            rsiPeriod: Number(finalParams.rsiPeriod) || 14,
            stochasticPeriod: finalParams.kPeriod!,
            kPeriod: finalParams.kPeriod!,
            dPeriod: finalParams.dPeriod!,
          });
          // outputs = ['k', 'd'];
          break;
        // Others
        case 'ParabolicSAR':
          results = TI.PSAR.calculate({
            high,
            low,
            step: finalParams.acceleration!,
            max: finalParams.maximum!,
          });
          // outputs = ['PSAR'];
          break;
        case 'ADX':
          results = TI.ADX.calculate({ high, low, close, period: finalParams.period! });
          // outputs = ['ADX', 'pdi', 'mdi'];
          break;
        default:
          throw new Error(`Indicator ${type} not yet implemented`);
      }

      // Format results to match IndicatorResult type
      const formattedResults: IndicatorResult = [];
      const offset = ohlcData.length - results.length;

      results.forEach((result, i) => {
        const time = times[i + offset];
        let value: IndicatorValue;

        if (typeof result === 'number') {
          value = result;
        } else if (typeof result === 'object' && result !== null) {
          value = result as Record<string, number>;
        } else {
          return; // Skip invalid results
        }

        formattedResults.push({ time, value });
      });

      return formattedResults;
    } catch (error) {
      console.error(`Error calculating ${type}:`, error);
      return [];
    }
  }

  // Validate parameters
  static validate(type: IndicatorType, params: IndicatorParams): boolean {
    const defaults = this.getDefaultParams(type);

    // Check all required parameters are provided and valid
    for (const [key, defaultValue] of Object.entries(defaults)) {
      const value = params[key] ?? defaultValue;
      if (typeof value === 'number' && (isNaN(value) || value <= 0)) {
        return false;
      }
    }

    return true;
  }

  // Get all available indicators
  static getAllIndicators(): IndicatorType[] {
    return [
      'SMA', 'EMA', 'WMA', 'VWAP',
      'BollingerBands', 'KeltnerChannels', 'DonchianChannels',
      'RSI', 'Stochastic', 'MACD', 'CCI', 'WilliamsR', 'ROC', 'Momentum',
      'ATR', 'StandardDeviation',
      'OBV', 'MFI', 'ADL', 'ForceIndex',
      'AwesomeOscillator', 'StochasticRSI',
      'Ichimoku', 'ParabolicSAR', 'ADX', 'Aroon', 'TRIX',
    ];
  }

  // Get indicators by category
  static getIndicatorsByCategory(category: 'trend' | 'momentum' | 'volatility' | 'volume' | 'oscillator'): IndicatorType[] {
    return this.getAllIndicators().filter(type => this.getMeta(type).category === category);
  }
}

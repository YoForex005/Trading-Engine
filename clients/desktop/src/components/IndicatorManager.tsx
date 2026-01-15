import { useState, useMemo } from 'react';
import { IndicatorEngine } from '../indicators/core/IndicatorEngine';
import type { IndicatorType, IndicatorParams, IndicatorMeta } from '../indicators/core/IndicatorEngine';
import { X, Search, TrendingUp, Activity, BarChart3, Volume2, Zap } from 'lucide-react';

type Category = 'all' | 'trend' | 'momentum' | 'volatility' | 'volume' | 'oscillator';

type IndicatorManagerProps = {
  isOpen: boolean;
  onClose: () => void;
  onAddIndicator: (type: IndicatorType, params?: IndicatorParams) => void;
  currentIndicators?: string[]; // Array of indicator types already on chart
};

const CATEGORY_ICONS: Record<Category, any> = {
  all: BarChart3,
  trend: TrendingUp,
  momentum: Activity,
  volatility: Zap,
  volume: Volume2,
  oscillator: Activity,
};

const CATEGORY_LABELS: Record<Category, string> = {
  all: 'All Indicators',
  trend: 'Trend',
  momentum: 'Momentum',
  volatility: 'Volatility',
  volume: 'Volume',
  oscillator: 'Oscillators',
};

export default function IndicatorManager({
  isOpen,
  onClose,
  onAddIndicator,
  currentIndicators = [],
}: IndicatorManagerProps) {
  const [selectedCategory, setSelectedCategory] = useState<Category>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIndicator, setSelectedIndicator] = useState<IndicatorType | null>(null);
  const [params, setParams] = useState<IndicatorParams>({});

  // Get all indicators
  const allIndicators = useMemo(() => {
    return IndicatorEngine.getAllIndicators().map(type => IndicatorEngine.getMeta(type));
  }, []);

  // Filter indicators by category and search
  const filteredIndicators = useMemo(() => {
    let filtered = allIndicators;

    // Filter by category
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(ind => ind.category === selectedCategory);
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        ind =>
          ind.name.toLowerCase().includes(query) ||
          ind.type.toLowerCase().includes(query) ||
          ind.description.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [allIndicators, selectedCategory, searchQuery]);

  // Handle indicator selection
  const handleSelectIndicator = (indicator: IndicatorMeta) => {
    setSelectedIndicator(indicator.type);
    setParams(indicator.defaultParams);
  };

  // Handle add indicator
  const handleAddIndicator = () => {
    if (selectedIndicator) {
      onAddIndicator(selectedIndicator, params);
      setSelectedIndicator(null);
      setParams({});
      onClose();
    }
  };

  // Handle parameter change
  const handleParamChange = (key: string, value: number) => {
    setParams(prev => ({ ...prev, [key]: value }));
  };

  if (!isOpen) return null;

  const selectedMeta = selectedIndicator ? IndicatorEngine.getMeta(selectedIndicator) : null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900 rounded-lg w-full max-w-4xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
          <h2 className="text-lg font-semibold text-white">Add Indicator</h2>
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-white transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Search Bar */}
        <div className="p-4 border-b border-zinc-800">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={18} />
            <input
              type="text"
              placeholder="Search indicators..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-zinc-800 text-white pl-10 pr-4 py-2 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
        </div>

        {/* Category Tabs */}
        <div className="flex gap-2 px-4 py-3 border-b border-zinc-800 overflow-x-auto">
          {Object.entries(CATEGORY_LABELS).map(([key, label]) => {
            const Icon = CATEGORY_ICONS[key as Category];
            const isActive = selectedCategory === key;
            return (
              <button
                key={key}
                onClick={() => setSelectedCategory(key as Category)}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg whitespace-nowrap transition-colors ${
                  isActive
                    ? 'bg-emerald-600 text-white'
                    : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                }`}
              >
                <Icon size={16} />
                {label}
              </button>
            );
          })}
        </div>

        {/* Content */}
        <div className="flex flex-1 overflow-hidden">
          {/* Indicator List */}
          <div className="w-1/2 border-r border-zinc-800 overflow-y-auto">
            {filteredIndicators.length === 0 ? (
              <div className="p-8 text-center text-zinc-500">
                No indicators found matching your search.
              </div>
            ) : (
              <div className="p-2">
                {filteredIndicators.map((indicator) => {
                  const isSelected = selectedIndicator === indicator.type;
                  const isOnChart = currentIndicators.includes(indicator.type);
                  return (
                    <div
                      key={indicator.type}
                      onClick={() => handleSelectIndicator(indicator)}
                      className={`p-3 rounded-lg cursor-pointer transition-colors mb-1 ${
                        isSelected
                          ? 'bg-emerald-600/20 border border-emerald-600'
                          : 'hover:bg-zinc-800'
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <h3 className="text-white font-medium">{indicator.name}</h3>
                            {isOnChart && (
                              <span className="text-xs bg-zinc-700 text-zinc-300 px-2 py-0.5 rounded">
                                On Chart
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-zinc-400 mt-1">{indicator.description}</p>
                          <div className="flex items-center gap-2 mt-2">
                            <span className="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded">
                              {indicator.category}
                            </span>
                            <span className="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded">
                              {indicator.displayMode}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Indicator Details & Parameters */}
          <div className="w-1/2 overflow-y-auto">
            {selectedMeta ? (
              <div className="p-4">
                <h3 className="text-white font-semibold text-lg mb-2">{selectedMeta.name}</h3>
                <p className="text-zinc-400 text-sm mb-4">{selectedMeta.description}</p>

                {/* Parameters */}
                <div className="bg-zinc-800 rounded-lg p-4 mb-4">
                  <h4 className="text-white font-medium mb-3">Parameters</h4>
                  {Object.keys(selectedMeta.defaultParams).length === 0 ? (
                    <p className="text-zinc-500 text-sm">No parameters to configure</p>
                  ) : (
                    <div className="space-y-3">
                      {Object.entries(selectedMeta.defaultParams).map(([key, defaultValue]) => (
                        <div key={key}>
                          <label className="text-sm text-zinc-400 capitalize block mb-1">
                            {key.replace(/([A-Z])/g, ' $1').trim()}
                          </label>
                          <input
                            type="number"
                            value={params[key] ?? defaultValue}
                            onChange={(e) => handleParamChange(key, Number(e.target.value))}
                            className="w-full bg-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500"
                          />
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                {/* Outputs */}
                <div className="bg-zinc-800 rounded-lg p-4 mb-4">
                  <h4 className="text-white font-medium mb-2">Outputs</h4>
                  <div className="flex flex-wrap gap-2">
                    {selectedMeta.outputs.map((output) => (
                      <span
                        key={output}
                        className="bg-zinc-700 text-zinc-300 px-3 py-1 rounded text-sm"
                      >
                        {output}
                      </span>
                    ))}
                  </div>
                </div>

                {/* Display Mode */}
                <div className="bg-zinc-800 rounded-lg p-4">
                  <h4 className="text-white font-medium mb-2">Display Mode</h4>
                  <div className="flex items-center gap-2">
                    <span className={`px-3 py-1 rounded text-sm ${
                      selectedMeta.displayMode === 'overlay'
                        ? 'bg-blue-600 text-white'
                        : 'bg-purple-600 text-white'
                    }`}>
                      {selectedMeta.displayMode === 'overlay' ? 'Overlay on Chart' : 'Separate Pane'}
                    </span>
                  </div>
                  <p className="text-zinc-500 text-xs mt-2">
                    {selectedMeta.displayMode === 'overlay'
                      ? 'This indicator will be drawn on top of the price chart'
                      : 'This indicator will appear in a separate panel below the chart'}
                  </p>
                </div>
              </div>
            ) : (
              <div className="h-full flex items-center justify-center text-zinc-500">
                <div className="text-center">
                  <BarChart3 size={48} className="mx-auto mb-3 opacity-50" />
                  <p>Select an indicator to view details</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-zinc-800 flex justify-between items-center">
          <div className="text-sm text-zinc-400">
            {filteredIndicators.length} indicator{filteredIndicators.length !== 1 ? 's' : ''} available
          </div>
          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-zinc-800 text-white rounded-lg hover:bg-zinc-700 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleAddIndicator}
              disabled={!selectedIndicator}
              className={`px-4 py-2 rounded-lg transition-colors ${
                selectedIndicator
                  ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
                  : 'bg-zinc-700 text-zinc-500 cursor-not-allowed'
              }`}
            >
              Add Indicator
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

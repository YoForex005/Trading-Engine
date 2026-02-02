import React, { useState, useEffect, useRef } from 'react';
import { X, Download, Image as ImageIcon, FileImage } from 'lucide-react';
import { ChartExporter } from '../../services/chartExporter';
import type { ChartMetadata, ChartExportOptions } from '../../services/chartExporter';

export interface SavePictureDialogProps {
  isOpen: boolean;
  onClose: () => void;
  chartCanvas?: HTMLCanvasElement | null;
  chartContainer?: HTMLElement | null;
  symbol: string;
  timeframe: string;
}

export const SavePictureDialog: React.FC<SavePictureDialogProps> = ({
  isOpen,
  onClose,
  chartCanvas,
  chartContainer,
  symbol,
  timeframe
}) => {
  const [format, setFormat] = useState<'png' | 'jpg'>('png');
  const [quality, setQuality] = useState(92);
  const [scale, setScale] = useState(1);
  const [exportArea, setExportArea] = useState<'visible' | 'full'>('visible');
  const [includeWatermark, setIncludeWatermark] = useState(false);
  const [watermarkText, setWatermarkText] = useState('Trading Engine');
  const [includeTimestamp, setIncludeTimestamp] = useState(true);
  const [includeSymbolInfo, setIncludeSymbolInfo] = useState(true);
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [isExporting, setIsExporting] = useState(false);
  const [error, setError] = useState<string>('');

  const previewCanvasRef = useRef<HTMLCanvasElement>(null);

  // Generate preview when dialog opens or settings change
  useEffect(() => {
    if (isOpen && chartCanvas) {
      try {
        const preview = ChartExporter.getPreview(chartCanvas, 300, 200);
        setPreviewUrl(preview);
      } catch (err) {
        console.error('Failed to generate preview:', err);
      }
    }
  }, [isOpen, chartCanvas]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!isOpen) {
      setError('');
      setIsExporting(false);
    }
  }, [isOpen]);

  const handleExport = async () => {
    if (!chartCanvas && !chartContainer) {
      setError('No chart available to export');
      return;
    }

    setIsExporting(true);
    setError('');

    try {
      const metadata: ChartMetadata = {
        symbol,
        timeframe,
        timestamp: new Date()
      };

      const options: Partial<ChartExportOptions> = {
        format,
        quality: quality / 100,
        scale,
        exportArea,
        includeWatermark,
        watermarkText: includeWatermark ? watermarkText : undefined,
        includeTimestamp,
        includeSymbolInfo,
        backgroundColor: '#000000'
      };

      let blob: Blob;

      if (chartCanvas) {
        // Export from canvas (preferred method)
        blob = await ChartExporter.exportCanvas(chartCanvas, metadata, options);
      } else if (chartContainer) {
        // Export from HTML element (fallback)
        blob = await ChartExporter.exportElement(chartContainer, metadata, options);
      } else {
        throw new Error('No chart source available');
      }

      // Generate filename and download
      const filename = ChartExporter.generateFilename(metadata, format);
      ChartExporter.downloadBlob(blob, filename);

      // Close dialog on success
      setTimeout(() => {
        onClose();
      }, 500);
    } catch (err) {
      console.error('Export failed:', err);
      setError(err instanceof Error ? err.message : 'Export failed');
    } finally {
      setIsExporting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/70">
      <div className="bg-gray-900 border border-gray-700 rounded-lg shadow-2xl w-[600px] max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800">
          <div className="flex items-center gap-2">
            <FileImage size={18} className="text-blue-400" />
            <h2 className="text-sm font-semibold text-white">Save Chart As Picture</h2>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
            disabled={isExporting}
          >
            <X size={18} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Preview */}
          <div className="bg-gray-800 border border-gray-700 rounded p-3">
            <label className="text-xs font-semibold text-gray-400 mb-2 block">Preview</label>
            {previewUrl ? (
              <div className="flex items-center justify-center bg-black rounded p-2">
                <img
                  src={previewUrl}
                  alt="Chart preview"
                  className="max-w-full h-auto rounded"
                  style={{ maxHeight: '200px' }}
                />
              </div>
            ) : (
              <div className="flex items-center justify-center bg-black rounded p-8 h-[200px]">
                <ImageIcon size={48} className="text-gray-600" />
              </div>
            )}
          </div>

          {/* Format Selection */}
          <div className="bg-gray-800 border border-gray-700 rounded p-3 space-y-3">
            <label className="text-xs font-semibold text-gray-400 block">Image Format</label>
            <div className="flex gap-3">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="format"
                  value="png"
                  checked={format === 'png'}
                  onChange={(e) => setFormat(e.target.value as 'png' | 'jpg')}
                  className="text-blue-500"
                />
                <span className="text-sm text-white">PNG (Lossless)</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="format"
                  value="jpg"
                  checked={format === 'jpg'}
                  onChange={(e) => setFormat(e.target.value as 'png' | 'jpg')}
                  className="text-blue-500"
                />
                <span className="text-sm text-white">JPG (Smaller)</span>
              </label>
            </div>

            {/* Quality slider for JPG */}
            {format === 'jpg' && (
              <div>
                <label className="text-xs text-gray-400 mb-1 block">
                  Quality: {quality}%
                </label>
                <input
                  type="range"
                  min="10"
                  max="100"
                  step="5"
                  value={quality}
                  onChange={(e) => setQuality(Number(e.target.value))}
                  className="w-full"
                />
              </div>
            )}
          </div>

          {/* Export Options */}
          <div className="bg-gray-800 border border-gray-700 rounded p-3 space-y-3">
            <label className="text-xs font-semibold text-gray-400 block">Export Options</label>

            {/* Resolution Scale */}
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Resolution Scale</label>
              <select
                value={scale}
                onChange={(e) => setScale(Number(e.target.value))}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white"
              >
                <option value={1}>1x - Standard (Fast)</option>
                <option value={2}>2x - High DPI / Retina</option>
                <option value={3}>3x - Ultra High DPI</option>
              </select>
            </div>

            {/* Export Area */}
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Export Area</label>
              <select
                value={exportArea}
                onChange={(e) => setExportArea(e.target.value as 'visible' | 'full')}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white"
              >
                <option value="visible">Visible Area</option>
                <option value="full">Full Chart (All Data)</option>
              </select>
            </div>
          </div>

          {/* Overlay Options */}
          <div className="bg-gray-800 border border-gray-700 rounded p-3 space-y-3">
            <label className="text-xs font-semibold text-gray-400 block">Chart Overlays</label>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={includeSymbolInfo}
                onChange={(e) => setIncludeSymbolInfo(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm text-white">Include Symbol & Timeframe</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={includeTimestamp}
                onChange={(e) => setIncludeTimestamp(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm text-white">Include Timestamp</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={includeWatermark}
                onChange={(e) => setIncludeWatermark(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm text-white">Include Watermark</span>
            </label>

            {includeWatermark && (
              <div className="pl-6">
                <input
                  type="text"
                  value={watermarkText}
                  onChange={(e) => setWatermarkText(e.target.value)}
                  placeholder="Watermark text"
                  className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white placeholder-gray-500"
                />
              </div>
            )}
          </div>

          {/* Error Message */}
          {error && (
            <div className="bg-red-900/30 border border-red-700 rounded p-3">
              <p className="text-sm text-red-400">{error}</p>
            </div>
          )}

          {/* Info Text */}
          <div className="text-xs text-gray-500 space-y-1">
            <p>• PNG format provides lossless quality, best for charts with text</p>
            <p>• JPG format creates smaller files, adjust quality to balance size vs quality</p>
            <p>• Higher resolution scales improve quality but increase file size</p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-gray-700 bg-gray-800">
          <div className="text-xs text-gray-500">
            {symbol} - {timeframe}
          </div>
          <div className="flex gap-2">
            <button
              onClick={onClose}
              disabled={isExporting}
              className="px-4 py-2 text-sm font-medium text-gray-300 hover:text-white transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              onClick={handleExport}
              disabled={isExporting}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isExporting ? (
                <>
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>Exporting...</span>
                </>
              ) : (
                <>
                  <Download size={16} />
                  <span>Save Picture</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SavePictureDialog;

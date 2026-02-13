import React, { useState, useEffect } from 'react';
import { Save, FolderOpen, Trash2, ChevronDown, X, Download, Upload } from 'lucide-react';
import { drawingManager } from '../../services/drawingManager';
import { drawingTemplateManager, type DrawingTemplate } from '../../services/drawingTemplateManager';

interface DrawingTemplatesDropdownProps {
  symbol?: string;
  accountId?: number;
}

export const DrawingTemplatesDropdown: React.FC<DrawingTemplatesDropdownProps> = ({
  symbol = 'BTCUSD',
  accountId = 1,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);
  const [templateName, setTemplateName] = useState('');
  const [templateDescription, setTemplateDescription] = useState('');
  const [templates, setTemplates] = useState<DrawingTemplate[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  // Load templates on mount and when symbol changes
  useEffect(() => {
    loadTemplates();
  }, [symbol, accountId]);

  // Listen for template changes
  useEffect(() => {
    const handleSaved = () => loadTemplates();
    const handleDeleted = () => loadTemplates();

    window.addEventListener('drawing-template:saved', handleSaved);
    window.addEventListener('drawing-template:deleted', handleDeleted);

    return () => {
      window.removeEventListener('drawing-template:saved', handleSaved);
      window.removeEventListener('drawing-template:deleted', handleDeleted);
    };
  }, [symbol, accountId]);

  const loadTemplates = async () => {
    setIsLoading(true);
    try {
      const loaded = await drawingTemplateManager.loadTemplates(symbol, accountId);
      setTemplates(loaded.filter(t => t.symbol === symbol));
    } catch (error) {
      console.error('Failed to load templates:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveTemplate = async () => {
    if (!templateName.trim()) {
      alert('Please enter a template name');
      return;
    }

    const currentDrawings = drawingManager.getDrawings();
    if (currentDrawings.length === 0) {
      alert('No drawings to save. Please add some drawings first.');
      return;
    }

    try {
      await drawingTemplateManager.saveTemplate(
        templateName.trim(),
        templateDescription.trim(),
        currentDrawings,
        symbol,
        accountId
      );

      setTemplateName('');
      setTemplateDescription('');
      setSaveDialogOpen(false);
      setIsOpen(false);
    } catch (error) {
      console.error('Failed to save template:', error);
      alert('Failed to save template. Please try again.');
    }
  };

  const handleLoadTemplate = (template: DrawingTemplate) => {
    if (drawingManager.getDrawings().length > 0) {
      if (!confirm('Loading this template will replace all current drawings. Continue?')) {
        return;
      }
    }

    // Clear existing drawings
    drawingManager.clearAllDrawings();

    // Load template drawings
    template.drawings.forEach(drawing => {
      const manager = drawingManager as any;
      manager.drawings.push({ ...drawing, selected: false });
    });

    drawingManager.renderAllDrawings();
    setIsOpen(false);

    // Dispatch success event
    window.dispatchEvent(new CustomEvent('notification:show', {
      detail: {
        type: 'success',
        title: 'Template Loaded',
        message: `Loaded ${template.drawings.length} drawing(s) from "${template.name}"`,
      }
    }));
  };

  const handleDeleteTemplate = async (template: DrawingTemplate, e: React.MouseEvent) => {
    e.stopPropagation();

    if (!confirm(`Delete template "${template.name}"?`)) {
      return;
    }

    try {
      await drawingTemplateManager.deleteTemplate(template.id, symbol, accountId);
    } catch (error) {
      console.error('Failed to delete template:', error);
      alert('Failed to delete template. Please try again.');
    }
  };

  const handleExport = () => {
    const data = drawingTemplateManager.exportTemplates();
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `drawing-templates-${symbol}-${Date.now()}.json`;
    link.click();
    URL.revokeObjectURL(url);
    setIsOpen(false);
  };

  const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const data = JSON.parse(event.target?.result as string);
        if (data.version && data.templates && Array.isArray(data.templates)) {
          drawingTemplateManager.importTemplates(data.templates);
          loadTemplates();
          alert(`Imported ${data.templates.length} template(s)`);
        } else {
          alert('Invalid template file format');
        }
      } catch (err) {
        console.error('Import error:', err);
        alert('Failed to import templates');
      }
    };
    reader.readAsText(file);

    // Reset input
    e.target.value = '';
  };

  return (
    <div className="relative">
      <button
        className="flex items-center gap-1 px-1.5 py-1.5 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#333] transition-all"
        onClick={() => setIsOpen(!isOpen)}
        title="Drawing Templates"
      >
        <FolderOpen size={15} />
        <ChevronDown size={10} className="text-zinc-600" />
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-[90]"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute top-full right-0 mt-1 w-64 bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl z-[100] py-1 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-150">
            {/* Save Current Drawings */}
            <DropdownItem
              icon={<Save size={13} className="text-emerald-400" />}
              label="Save Current Drawings"
              onClick={() => {
                setSaveDialogOpen(true);
                setIsOpen(false);
              }}
            />

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            {/* Template List */}
            <div className="max-h-64 overflow-y-auto">
              {isLoading ? (
                <div className="px-3 py-4 text-center text-[11px] text-zinc-500">
                  Loading templates...
                </div>
              ) : templates.length === 0 ? (
                <div className="px-3 py-4 text-center text-[11px] text-zinc-500">
                  No saved templates
                </div>
              ) : (
                templates.map((template) => (
                  <div
                    key={template.id}
                    className="group flex items-center gap-2 px-3 py-2 text-[11px] text-zinc-300 hover:bg-[#2d2d2d] cursor-pointer transition-colors"
                    onClick={() => handleLoadTemplate(template)}
                  >
                    <FolderOpen size={13} className="text-blue-400 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium truncate">{template.name}</div>
                      <div className="text-[10px] text-zinc-500">
                        {template.drawings.length} drawing{template.drawings.length !== 1 ? 's' : ''}
                        {template.description && ` • ${template.description}`}
                      </div>
                    </div>
                    <button
                      onClick={(e) => handleDeleteTemplate(template, e)}
                      className="opacity-0 group-hover:opacity-100 p-1 hover:bg-zinc-700 rounded transition-opacity flex-shrink-0"
                      title="Delete"
                    >
                      <Trash2 size={12} className="text-red-400" />
                    </button>
                  </div>
                ))
              )}
            </div>

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            {/* Import/Export */}
            <DropdownItem
              icon={<Download size={13} />}
              label="Export Templates"
              onClick={handleExport}
            />
            <label>
              <DropdownItem
                icon={<Upload size={13} />}
                label="Import Templates"
                onClick={() => {}}
              />
              <input
                type="file"
                accept=".json"
                onChange={handleImport}
                className="hidden"
              />
            </label>
          </div>
        </>
      )}

      {/* Save Dialog */}
      {saveDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[200]">
          <div className="bg-[#1e1e1e] border border-zinc-800 rounded-lg shadow-xl w-full max-w-md p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-zinc-100">Save Drawing Template</h3>
              <button
                onClick={() => {
                  setSaveDialogOpen(false);
                  setTemplateName('');
                  setTemplateDescription('');
                }}
                className="p-1 hover:bg-zinc-800 rounded transition-colors"
              >
                <X className="w-5 h-5 text-zinc-400" />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Template Name *</label>
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  placeholder="e.g., Support/Resistance Setup"
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                  autoFocus
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && templateName.trim()) {
                      handleSaveTemplate();
                    }
                  }}
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1">Description (optional)</label>
                <textarea
                  value={templateDescription}
                  onChange={(e) => setTemplateDescription(e.target.value)}
                  placeholder="Brief description of this template"
                  rows={2}
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-blue-500 resize-none"
                />
              </div>

              <div className="text-xs text-zinc-500">
                {drawingManager.getDrawings().length} drawing(s) will be saved
              </div>
            </div>

            <div className="flex gap-2 mt-4">
              <button
                onClick={() => {
                  setSaveDialogOpen(false);
                  setTemplateName('');
                  setTemplateDescription('');
                }}
                className="flex-1 px-4 py-2 text-sm font-medium bg-zinc-800 text-zinc-300 hover:bg-zinc-700 rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveTemplate}
                disabled={!templateName.trim()}
                className="flex-1 px-4 py-2 text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Save Template
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const DropdownItem = ({ icon, label, onClick }: { icon: React.ReactNode, label: string, onClick: () => void }) => (
  <div
    className="flex items-center gap-3 px-3 py-2 text-[11px] text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors"
    onClick={onClick}
  >
    <span className="w-4 flex justify-center text-zinc-500">{icon}</span>
    <span>{label}</span>
  </div>
);

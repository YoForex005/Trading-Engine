/**
 * Drawing Template Manager
 * Manages saving/loading drawing configurations as templates
 */

import type { Drawing } from './drawingManager';
import { API_ENDPOINTS } from '../config/api';

export interface DrawingTemplate {
  id: string;
  name: string;
  description?: string;
  symbol: string;
  accountId: number;
  drawings: Drawing[];
  createdAt: number;
  updatedAt: number;
}

export class DrawingTemplateManager {
  private templates: DrawingTemplate[] = [];
  private currentSymbol: string | null = null;
  private currentAccountId: number = 1;

  /**
   * Load templates from backend for a specific symbol and account
   */
  async loadTemplates(symbol: string, accountId: number = 1): Promise<DrawingTemplate[]> {
    this.currentSymbol = symbol;
    this.currentAccountId = accountId;

    try {
      const response = await fetch(
        `${API_ENDPOINTS.workspace.templates}?type=drawing&symbol=${symbol}&accountId=${accountId}`
      );

      if (response.ok) {
        const data = await response.json();
        this.templates = Array.isArray(data) ? data : [];
        return this.templates;
      } else {
        console.warn('Failed to load templates from backend, using localStorage');
        return this.loadFromStorage(symbol, accountId);
      }
    } catch (error) {
      console.error('Failed to load templates from backend:', error);
      return this.loadFromStorage(symbol, accountId);
    }
  }

  /**
   * Save a new template
   */
  async saveTemplate(
    name: string,
    description: string,
    drawings: Drawing[],
    symbol: string,
    accountId: number = 1
  ): Promise<DrawingTemplate> {
    const template: DrawingTemplate = {
      id: `drawing_template_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name,
      description,
      symbol,
      accountId,
      drawings: JSON.parse(JSON.stringify(drawings)), // Deep copy
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };

    try {
      const response = await fetch(API_ENDPOINTS.workspace.templates, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...template,
          type: 'drawing',
          data: { drawings: template.drawings },
        }),
      });

      if (response.ok) {
        const savedTemplate = await response.json();
        // Use backend-generated ID if provided
        if (savedTemplate.id) {
          template.id = savedTemplate.id;
        }
        this.templates.push(template);
        this.saveToStorage(symbol, accountId);

        // Dispatch event
        window.dispatchEvent(new CustomEvent('drawing-template:saved', {
          detail: template
        }));

        return template;
      } else {
        console.warn('Failed to save template to backend, using localStorage');
        this.templates.push(template);
        this.saveToStorage(symbol, accountId);
        return template;
      }
    } catch (error) {
      console.error('Failed to save template to backend:', error);
      this.templates.push(template);
      this.saveToStorage(symbol, accountId);
      return template;
    }
  }

  /**
   * Load a specific template
   */
  loadTemplate(id: string): DrawingTemplate | null {
    const template = this.templates.find(t => t.id === id);
    if (template) {
      // Dispatch event
      window.dispatchEvent(new CustomEvent('drawing-template:loaded', {
        detail: template
      }));
      return template;
    }
    return null;
  }

  /**
   * Delete a template
   */
  async deleteTemplate(id: string, symbol: string, accountId: number = 1): Promise<boolean> {
    try {
      const response = await fetch(`${API_ENDPOINTS.workspace.templates}/${id}?symbol=${symbol}&accountId=${accountId}`, {
        method: 'DELETE',
      });

      if (response.ok || response.status === 404) {
        // Remove from local cache
        const index = this.templates.findIndex(t => t.id === id);
        if (index !== -1) {
          const deleted = this.templates.splice(index, 1)[0];
          this.saveToStorage(symbol, accountId);

          // Dispatch event
          window.dispatchEvent(new CustomEvent('drawing-template:deleted', {
            detail: { id }
          }));
        }
        return true;
      } else {
        console.warn('Failed to delete template from backend');
        return false;
      }
    } catch (error) {
      console.error('Failed to delete template from backend:', error);
      // Still remove from local storage
      const index = this.templates.findIndex(t => t.id === id);
      if (index !== -1) {
        this.templates.splice(index, 1);
        this.saveToStorage(symbol, accountId);
      }
      return false;
    }
  }

  /**
   * Get all templates
   */
  getTemplates(): DrawingTemplate[] {
    return [...this.templates];
  }

  /**
   * Get templates for current symbol
   */
  getTemplatesForSymbol(symbol: string): DrawingTemplate[] {
    return this.templates.filter(t => t.symbol === symbol);
  }

  /**
   * localStorage fallback
   */
  private saveToStorage(symbol: string, accountId: number): void {
    try {
      const key = `drawing-templates-${symbol}-${accountId}`;
      localStorage.setItem(key, JSON.stringify(this.templates));
    } catch (error) {
      console.error('Failed to save templates to localStorage:', error);
    }
  }

  private loadFromStorage(symbol: string, accountId: number): DrawingTemplate[] {
    try {
      const key = `drawing-templates-${symbol}-${accountId}`;
      const data = localStorage.getItem(key);
      if (data) {
        const parsed = JSON.parse(data);
        this.templates = Array.isArray(parsed) ? parsed : [];
        return this.templates;
      }
    } catch (error) {
      console.error('Failed to load templates from localStorage:', error);
    }
    return [];
  }

  /**
   * Export templates as JSON
   */
  exportTemplates(): { version: number; templates: DrawingTemplate[] } {
    return {
      version: 1,
      templates: this.templates,
    };
  }

  /**
   * Import templates from JSON
   */
  importTemplates(templates: DrawingTemplate[]): void {
    if (!Array.isArray(templates)) return;

    const existingIds = new Set(this.templates.map(t => t.id));
    const newTemplates = templates.map(t => {
      if (existingIds.has(t.id)) {
        // Regenerate ID if conflict
        return {
          ...t,
          id: `drawing_template_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          updatedAt: Date.now(),
        };
      }
      return t;
    });

    this.templates.push(...newTemplates);
    if (this.currentSymbol) {
      this.saveToStorage(this.currentSymbol, this.currentAccountId);
    }
  }
}

// Export singleton instance
export const drawingTemplateManager = new DrawingTemplateManager();

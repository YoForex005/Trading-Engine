import { useState, useEffect } from 'react';
import type { Drawing } from '../types';

export function useDrawings(symbol: string, accountId: number) {
  const [drawings, setDrawings] = useState<Drawing[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Fetch drawings on symbol change
  useEffect(() => {
    let cancelled = false;

    async function fetchDrawings() {
      try {
        setLoading(true);
        setError(null);

        const res = await fetch(
          `http://localhost:8080/api/drawings?accountId=${accountId}&symbol=${symbol}`
        );

        if (!res.ok) {
          throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }

        const data = await res.json();

        if (!cancelled) {
          setDrawings(data || []);
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          console.error('Failed to fetch drawings:', err);
          setError(err as Error);
          setLoading(false);
        }
      }
    }

    fetchDrawings();

    return () => {
      cancelled = true;
    };
  }, [symbol, accountId]);

  const updateDrawing = async (drawing: Drawing) => {
    // Optimistic update
    setDrawings(prev => {
      const exists = prev.find(d => d.id === drawing.id);
      if (exists) {
        return prev.map(d => (d.id === drawing.id ? drawing : d));
      }
      return [...prev, drawing];
    });

    // Save to backend
    try {
      await fetch('http://localhost:8080/api/drawings/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(drawing),
      });
    } catch (err) {
      console.error('Failed to save drawing:', err);
      throw err;
    }
  };

  const deleteDrawing = async (id: string) => {
    // Optimistic update
    setDrawings(prev => prev.filter(d => d.id !== id));

    // Delete from backend
    try {
      await fetch(
        `http://localhost:8080/api/drawings/delete?id=${id}&accountId=${accountId}`,
        { method: 'POST' }
      );
    } catch (err) {
      console.error('Failed to delete drawing:', err);
      throw err;
    }
  };

  return {
    drawings,
    loading,
    error,
    updateDrawing,
    deleteDrawing,
  };
}

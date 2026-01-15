import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '../../test/utils'
import { TradingChart } from '../TradingChart'

// Mock lightweight-charts to avoid rendering issues in tests
vi.mock('lightweight-charts', () => ({
  createChart: vi.fn(() => ({
    addSeries: vi.fn(() => ({
      setData: vi.fn(),
      update: vi.fn(),
      priceToCoordinate: vi.fn(() => 100),
      coordinateToPrice: vi.fn(() => 1.0850),
    })),
    removeSeries: vi.fn(),
    applyOptions: vi.fn(),
    remove: vi.fn(),
    timeScale: vi.fn(() => ({
      subscribeVisibleTimeRangeChange: vi.fn(),
      unsubscribeVisibleTimeRangeChange: vi.fn(),
    })),
    subscribeClick: vi.fn(),
    unsubscribeClick: vi.fn(),
  })),
  ColorType: { Solid: 0 },
  CrosshairMode: { Normal: 0 },
  CandlestickSeries: {},
  LineSeries: {},
  BarSeries: {},
  AreaSeries: {},
}))

describe('TradingChart', () => {
  it('renders chart container with accessibility label', () => {
    renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    // Use accessibility-focused query (not class names or IDs)
    const chartContainer = screen.getByRole('region', { name: /trading chart/i })
    expect(chartContainer).toBeInTheDocument()
  })

  it('displays symbol name in header', () => {
    renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    expect(screen.getByText(/EURUSD/i)).toBeInTheDocument()
  })

  it('renders with empty positions array without crashing', () => {
    renderWithProviders(
      <TradingChart symbol="EURUSD" positions={[]} />
    )

    // Chart should handle empty positions gracefully
    expect(screen.getByText(/EURUSD/i)).toBeInTheDocument()
  })

  it('updates when symbol changes', () => {
    const { rerender } = renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    expect(screen.getByText(/EURUSD/i)).toBeInTheDocument()

    // Change symbol
    rerender(<TradingChart symbol="GBPUSD" />)

    expect(screen.getByText(/GBPUSD/i)).toBeInTheDocument()
  })

  it('handles price updates without errors', () => {
    const { rerender } = renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    // Simulate new price arriving
    const newPrice = { bid: 1.0850, ask: 1.0852 }
    rerender(<TradingChart symbol="EURUSD" currentPrice={newPrice} />)

    // Chart should re-render without crashing
    expect(screen.getByText(/EURUSD/i)).toBeInTheDocument()
  })

  it('displays indicators button', () => {
    renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    const indicatorButton = screen.getByRole('button', { name: /indicators/i })
    expect(indicatorButton).toBeInTheDocument()
  })

  it('displays timeframe', () => {
    renderWithProviders(
      <TradingChart symbol="EURUSD" timeframe="1h" />
    )

    expect(screen.getByText(/1h/i)).toBeInTheDocument()
  })
})

describe('TradingChart - Error Handling', () => {
  it('renders without crashing when initialized', () => {
    // This tests basic initialization without errors
    renderWithProviders(
      <TradingChart symbol="EURUSD" />
    )

    // Should show the symbol, meaning chart initialized successfully
    expect(screen.getByText(/EURUSD/i)).toBeInTheDocument()
  })
})

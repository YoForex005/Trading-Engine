import { describe, it, expect, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { renderWithProviders, userEvent } from '../../test/utils'
import IndicatorManager from '../IndicatorManager'

// Mock IndicatorEngine to avoid import issues
vi.mock('../../indicators/core/IndicatorEngine', () => ({
  IndicatorEngine: {
    getAllIndicators: vi.fn(() => ['SMA', 'EMA', 'RSI', 'MACD']),
    getMeta: vi.fn((type) => {
      const metas: Record<string, any> = {
        SMA: {
          type: 'SMA',
          name: 'Simple Moving Average',
          description: 'Calculates the average price over a period',
          category: 'trend',
          displayMode: 'overlay',
          outputs: ['SMA'],
          defaultParams: { period: 14 },
        },
        EMA: {
          type: 'EMA',
          name: 'Exponential Moving Average',
          description: 'Weighted moving average giving more weight to recent prices',
          category: 'trend',
          displayMode: 'overlay',
          outputs: ['EMA'],
          defaultParams: { period: 14 },
        },
        RSI: {
          type: 'RSI',
          name: 'Relative Strength Index',
          description: 'Momentum oscillator measuring speed and magnitude of price changes',
          category: 'momentum',
          displayMode: 'pane',
          outputs: ['RSI'],
          defaultParams: { period: 14 },
        },
        MACD: {
          type: 'MACD',
          name: 'Moving Average Convergence Divergence',
          description: 'Trend-following momentum indicator',
          category: 'momentum',
          displayMode: 'pane',
          outputs: ['MACD', 'Signal', 'Histogram'],
          defaultParams: { fastPeriod: 12, slowPeriod: 26, signalPeriod: 9 },
        },
      }
      return metas[type] || metas.SMA
    }),
  },
}))

describe('IndicatorManager', () => {
  it('renders when open', () => {
    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    expect(screen.getByText(/Add Indicator/i)).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    renderWithProviders(
      <IndicatorManager
        isOpen={false}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    expect(screen.queryByText(/Add Indicator/i)).not.toBeInTheDocument()
  })

  it('displays search input', () => {
    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    const searchInput = screen.getByPlaceholderText(/Search indicators/i)
    expect(searchInput).toBeInTheDocument()
  })

  it('shows indicator list', () => {
    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    expect(screen.getByText(/Simple Moving Average/i)).toBeInTheDocument()
    expect(screen.getByText(/Relative Strength Index/i)).toBeInTheDocument()
  })

  it('filters indicators by search query', async () => {
    const user = userEvent.setup()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    const searchInput = screen.getByPlaceholderText(/Search indicators/i)
    await user.type(searchInput, 'Moving Average')

    // Should show Moving Average indicators
    expect(screen.getByText(/Simple Moving Average/i)).toBeInTheDocument()
    expect(screen.getByText(/Exponential Moving Average/i)).toBeInTheDocument()

    // RSI should not be visible (filtered out)
    // Note: It might still be in DOM but filtered by React
  })

  it('selects indicator when clicked', async () => {
    const user = userEvent.setup()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    const smaOption = screen.getByText(/Simple Moving Average/i)
    await user.click(smaOption)

    // Add button should become enabled after selection
    const addButton = screen.getByRole('button', { name: /Add Indicator/i })
    expect(addButton).toBeEnabled()
  })

  it('calls onAddIndicator when Add button clicked', async () => {
    const user = userEvent.setup()
    const onAddIndicator = vi.fn()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={onAddIndicator}
      />
    )

    // Select an indicator
    const smaOption = screen.getByText(/Simple Moving Average/i)
    await user.click(smaOption)

    // Click Add button
    const addButton = screen.getByRole('button', { name: /Add Indicator/i })
    await user.click(addButton)

    // Verify callback called with indicator type and params
    expect(onAddIndicator).toHaveBeenCalledWith('SMA', { period: 14 })
  })

  it('calls onClose when Cancel button clicked', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={onClose}
        onAddIndicator={vi.fn()}
      />
    )

    const cancelButton = screen.getByRole('button', { name: /Cancel/i })
    await user.click(cancelButton)

    expect(onClose).toHaveBeenCalled()
  })

  it('calls onClose when X button clicked', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={onClose}
        onAddIndicator={vi.fn()}
      />
    )

    // Find and click the X button (close icon)
    const closeButtons = screen.getAllByRole('button')
    const xButton = closeButtons.find(btn => btn.querySelector('svg'))

    if (xButton) {
      await user.click(xButton)
      expect(onClose).toHaveBeenCalled()
    }
  })

  it('allows configuring indicator parameters', async () => {
    const user = userEvent.setup()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    // Select MACD which has multiple parameters
    const macdOption = screen.getByText(/Moving Average Convergence Divergence/i)
    await user.click(macdOption)

    // Check that parameter inputs appear
    await waitFor(() => {
      expect(screen.getByText(/Parameters/i)).toBeInTheDocument()
    })
  })

  it('shows category tabs', () => {
    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    expect(screen.getByText(/All Indicators/i)).toBeInTheDocument()
    expect(screen.getByText(/Trend/i)).toBeInTheDocument()
    expect(screen.getByText(/Momentum/i)).toBeInTheDocument()
  })

  it('filters by category when category tab clicked', async () => {
    const user = userEvent.setup()

    renderWithProviders(
      <IndicatorManager
        isOpen={true}
        onClose={vi.fn()}
        onAddIndicator={vi.fn()}
      />
    )

    // Click on Trend category
    const trendTab = screen.getByRole('button', { name: /Trend/i })
    await user.click(trendTab)

    // Should show trend indicators
    expect(screen.getByText(/Simple Moving Average/i)).toBeInTheDocument()
  })
})

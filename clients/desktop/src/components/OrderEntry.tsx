import { useState, useEffect } from 'react';

type OrderType = 'MARKET' | 'BUY_LIMIT' | 'SELL_LIMIT' | 'BUY_STOP' | 'SELL_STOP';

interface OrderEntryProps {
    symbol: string;
    currentPrice?: { bid: number; ask: number };
    onOrderPlaced?: () => void;
}

export function OrderEntry({ symbol, currentPrice, onOrderPlaced }: OrderEntryProps) {
    const [orderType, setOrderType] = useState<OrderType>('MARKET');
    const [volume, setVolume] = useState('0.01');
    const [triggerPrice, setTriggerPrice] = useState('');
    const [sl, setSL] = useState('');
    const [tp, setTP] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    // Reset trigger price when order type changes
    useEffect(() => {
        if (orderType === 'MARKET') {
            setTriggerPrice('');
        } else if (currentPrice && !triggerPrice) {
            // Set default trigger price based on order type
            const defaultPrice = orderType.startsWith('BUY') ? currentPrice.ask : currentPrice.bid;
            setTriggerPrice(defaultPrice.toFixed(5));
        }
    }, [orderType, currentPrice]);

    const validateTriggerPrice = (): string | null => {
        if (!currentPrice || !triggerPrice) return null;

        const trigger = parseFloat(triggerPrice);
        const { bid, ask } = currentPrice;

        switch (orderType) {
            case 'BUY_LIMIT':
                if (trigger >= ask) {
                    return `BUY LIMIT trigger price must be below current ask (${ask.toFixed(5)})`;
                }
                break;
            case 'SELL_LIMIT':
                if (trigger <= bid) {
                    return `SELL LIMIT trigger price must be above current bid (${bid.toFixed(5)})`;
                }
                break;
            case 'BUY_STOP':
                if (trigger <= ask) {
                    return `BUY STOP trigger price must be above current ask (${ask.toFixed(5)})`;
                }
                break;
            case 'SELL_STOP':
                if (trigger >= bid) {
                    return `SELL STOP trigger price must be below current bid (${bid.toFixed(5)})`;
                }
                break;
        }
        return null;
    };

    const handleSubmit = async () => {
        setError('');

        // Validate
        const validationError = validateTriggerPrice();
        if (validationError) {
            setError(validationError);
            return;
        }

        if (parseFloat(volume) <= 0) {
            setError('Volume must be greater than 0');
            return;
        }

        setLoading(true);

        try {
            if (orderType === 'MARKET') {
                // Market order
                const side = 'BUY'; // Default to BUY for market orders, should have separate buttons
                const response = await fetch('http://localhost:8080/api/orders/market', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        symbol,
                        side,
                        volume: parseFloat(volume),
                        sl: sl ? parseFloat(sl) : 0,
                        tp: tp ? parseFloat(tp) : 0,
                    }),
                });

                if (!response.ok) {
                    const data = await response.text();
                    throw new Error(data || 'Order failed');
                }

                onOrderPlaced?.();
            } else {
                // Pending order
                const response = await fetch('http://localhost:8080/api/orders/pending', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        type: orderType,
                        symbol,
                        volume: parseFloat(volume),
                        triggerPrice: parseFloat(triggerPrice),
                        sl: sl ? parseFloat(sl) : 0,
                        tp: tp ? parseFloat(tp) : 0,
                    }),
                });

                if (!response.ok) {
                    const data = await response.text();
                    throw new Error(data || 'Pending order failed');
                }

                const result = await response.json();
                console.log('Pending order created:', result);
                onOrderPlaced?.();
            }

            // Reset form
            setVolume('0.01');
            setTriggerPrice('');
            setSL('');
            setTP('');
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        } finally {
            setLoading(false);
        }
    };

    const currentAsk = currentPrice?.ask.toFixed(5) || '---';
    const currentBid = currentPrice?.bid.toFixed(5) || '---';

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold text-zinc-200">{symbol}</h3>
                <div className="text-xs text-zinc-400">
                    <span className="text-blue-400">Bid: {currentBid}</span>
                    <span className="mx-2">|</span>
                    <span className="text-red-400">Ask: {currentAsk}</span>
                </div>
            </div>

            {/* Order Type Selector */}
            <div className="mb-4">
                <label className="text-xs text-zinc-500 mb-2 block">Order Type</label>
                <div className="grid grid-cols-3 gap-1">
                    {(['MARKET', 'BUY_LIMIT', 'SELL_LIMIT', 'BUY_STOP', 'SELL_STOP'] as OrderType[]).map((type) => (
                        <button
                            key={type}
                            onClick={() => setOrderType(type)}
                            className={`py-2 px-2 text-xs font-medium rounded transition-colors ${
                                orderType === type
                                    ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                                    : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700 border border-zinc-700'
                            }`}
                        >
                            {type.replace('_', ' ')}
                        </button>
                    ))}
                </div>
            </div>

            {/* Volume */}
            <div className="mb-3">
                <label className="text-xs text-zinc-500 mb-1 block">Volume (Lots)</label>
                <input
                    type="number"
                    value={volume}
                    onChange={(e) => setVolume(e.target.value)}
                    step="0.01"
                    min="0.01"
                    className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200"
                />
            </div>

            {/* Trigger Price (for pending orders only) */}
            {orderType !== 'MARKET' && (
                <div className="mb-3">
                    <label className="text-xs text-zinc-500 mb-1 block">Trigger Price</label>
                    <input
                        type="number"
                        value={triggerPrice}
                        onChange={(e) => setTriggerPrice(e.target.value)}
                        step="0.00001"
                        className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200"
                        placeholder={orderType.startsWith('BUY') ? currentAsk : currentBid}
                    />
                    {validateTriggerPrice() && (
                        <p className="text-xs text-orange-400 mt-1">{validateTriggerPrice()}</p>
                    )}
                </div>
            )}

            {/* Stop Loss */}
            <div className="mb-3">
                <label className="text-xs text-zinc-500 mb-1 block">Stop Loss (optional)</label>
                <input
                    type="number"
                    value={sl}
                    onChange={(e) => setSL(e.target.value)}
                    step="0.00001"
                    className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200"
                    placeholder="0.00000"
                />
            </div>

            {/* Take Profit */}
            <div className="mb-4">
                <label className="text-xs text-zinc-500 mb-1 block">Take Profit (optional)</label>
                <input
                    type="number"
                    value={tp}
                    onChange={(e) => setTP(e.target.value)}
                    step="0.00001"
                    className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200"
                    placeholder="0.00000"
                />
            </div>

            {/* Error Display */}
            {error && (
                <div className="mb-3 p-2 bg-red-500/10 border border-red-500/30 rounded text-xs text-red-400">
                    {error}
                </div>
            )}

            {/* Submit Button */}
            <button
                onClick={handleSubmit}
                disabled={loading || (orderType !== 'MARKET' && !triggerPrice)}
                className={`w-full py-2.5 rounded font-medium text-sm transition-colors ${
                    loading || (orderType !== 'MARKET' && !triggerPrice)
                        ? 'bg-zinc-700 text-zinc-500 cursor-not-allowed'
                        : orderType.startsWith('BUY') || orderType === 'MARKET'
                        ? 'bg-blue-600 hover:bg-blue-500 text-white'
                        : 'bg-red-600 hover:bg-red-500 text-white'
                }`}
            >
                {loading ? 'Processing...' : orderType === 'MARKET' ? 'Execute Market Order' : 'Place Pending Order'}
            </button>
        </div>
    );
}

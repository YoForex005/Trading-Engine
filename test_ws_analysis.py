#!/usr/bin/env python3
"""
WebSocket Quote Analysis Tool
Analyzes ticks from ws://localhost:7999/ws to determine if they are real or simulated
"""

import asyncio
import json
import websockets
import sys
from datetime import datetime
from collections import defaultdict
from statistics import mean, stdev

# Configuration
WS_URL = "ws://localhost:7999/ws"
CAPTURE_COUNT = 20
TIMEOUT = 30

# Data collection
captured_ticks = defaultdict(list)
total_tick_count = 0
start_time = datetime.now()

# Analysis results
analysis = {
    'lp_sources': set(),
    'symbols': set(),
    'evidence': []
}

def analyze_lp(ticks):
    """Analyze LP sources"""
    print('\n📊 LP SOURCE ANALYSIS:')
    print('═' * 60)

    lp_counts = {}
    for tick in ticks:
        lp = tick.get('lp', 'UNKNOWN')
        lp_counts[lp] = lp_counts.get(lp, 0) + 1
        analysis['lp_sources'].add(lp)

    for lp, count in lp_counts.items():
        percentage = (count / len(ticks)) * 100
        print(f"  {lp}: {count} ticks ({percentage:.1f}%)")

        if lp == 'SIMULATED':
            analysis['evidence'].append('❌ LP field shows "SIMULATED" - clear simulation marker')
        elif lp == 'YOFX':
            analysis['evidence'].append('✅ LP field shows "YOFX" - indicates real FIX gateway data')

def analyze_price_patterns(ticks):
    """Analyze price movement patterns"""
    print('\n📈 PRICE PATTERN ANALYSIS:')
    print('═' * 60)

    by_symbol = defaultdict(list)
    for tick in ticks:
        by_symbol[tick['symbol']].append(tick)

    for symbol, symbol_ticks in by_symbol.items():
        if len(symbol_ticks) < 3:
            continue

        # Calculate price changes
        price_changes = []
        for i in range(1, len(symbol_ticks)):
            change = symbol_ticks[i]['bid'] - symbol_ticks[i-1]['bid']
            price_changes.append(change)

        # Check for regular patterns
        unique_changes = len(set(round(c, 6) for c in price_changes))
        regularity_ratio = unique_changes / len(price_changes) if price_changes else 0

        print(f"  {symbol}:")
        print(f"    - Price changes: {len(price_changes)}")
        print(f"    - Unique changes: {unique_changes}")
        print(f"    - Regularity ratio: {regularity_ratio:.3f} (lower = more regular)")

        if regularity_ratio < 0.5:
            analysis['evidence'].append(f"⚠️  {symbol}: Low regularity ratio ({regularity_ratio:.3f}) suggests simulated data")
        else:
            analysis['evidence'].append(f"✅ {symbol}: High price variation ({regularity_ratio:.3f}) suggests real market data")

        # Sample changes
        sample_changes = [f"{c:.6f}" for c in price_changes[:5]]
        print(f"    - Sample changes: [{', '.join(sample_changes)}]")

def analyze_timestamps(ticks):
    """Analyze timestamp patterns"""
    print('\n⏰ TIMESTAMP ANALYSIS:')
    print('═' * 60)

    timestamps = sorted([t['timestamp'] for t in ticks])
    gaps = [timestamps[i] - timestamps[i-1] for i in range(1, len(timestamps))]

    if not gaps:
        print("  Not enough data for timestamp analysis")
        return

    avg_gap = mean(gaps)
    unique_gaps = len(set(gaps))
    gap_regularity = unique_gaps / len(gaps)

    print(f"  Average gap: {avg_gap:.2f}s")
    print(f"  Unique gaps: {unique_gaps} / {len(gaps)}")
    print(f"  Gap regularity: {gap_regularity:.3f}")

    # Check for regular intervals
    if gap_regularity < 0.3 and avg_gap < 1:
        analysis['evidence'].append(f"⚠️  Very regular timestamp intervals ({avg_gap:.2f}s) suggests simulation")
    else:
        analysis['evidence'].append("✅ Irregular timestamp intervals suggest real market data")

    # Sample gaps
    sample_gaps = [f"{g:.2f}" for g in gaps[:10]]
    print(f"  Sample gaps (s): [{', '.join(sample_gaps)}]")

def analyze_spread(ticks):
    """Analyze spread consistency"""
    print('\n💰 SPREAD ANALYSIS:')
    print('═' * 60)

    by_symbol = defaultdict(list)
    for tick in ticks:
        by_symbol[tick['symbol']].append(tick)

    for symbol, symbol_ticks in by_symbol.items():
        spreads = [round(t['ask'] - t['bid'], 6) for t in symbol_ticks]
        unique_spreads = len(set(spreads))
        spread_consistency = 1 - (unique_spreads / len(spreads))

        print(f"  {symbol}:")
        print(f"    - Total ticks: {len(spreads)}")
        print(f"    - Unique spreads: {unique_spreads}")
        print(f"    - Consistency: {spread_consistency * 100:.1f}%")

        if spread_consistency > 0.95:
            analysis['evidence'].append(f"⚠️  {symbol}: Perfectly consistent spread ({spread_consistency * 100:.1f}%) suggests simulation")
        else:
            analysis['evidence'].append(f"✅ {symbol}: Variable spread suggests real market conditions")

        # Sample spreads
        sample_spreads = [f"{s:.6f}" for s in list(set(spreads))[:5]]
        print(f"    - Sample spreads: [{', '.join(sample_spreads)}]")

def generate_verdict():
    """Generate final verdict"""
    print('\n\n' + '═' * 60)
    print('🔍 FINAL VERDICT')
    print('═' * 60)

    # Count indicators
    simulated_indicators = sum(1 for e in analysis['evidence'] if '⚠️' in e or '❌' in e)
    real_indicators = sum(1 for e in analysis['evidence'] if '✅' in e)

    print('\nEVIDENCE SUMMARY:')
    for evidence in analysis['evidence']:
        print(f"  {evidence}")

    print('\n' + '─' * 60)
    print(f"Simulated indicators: {simulated_indicators}")
    print(f"Real data indicators: {real_indicators}")
    print('─' * 60)

    # Determine verdict
    if 'SIMULATED' in analysis['lp_sources']:
        verdict = 'SIMULATED'
        print('\n🎭 VERDICT: SIMULATED DATA')
        print('   The LP field explicitly shows "SIMULATED"')
    elif 'YOFX' in analysis['lp_sources'] and real_indicators > simulated_indicators:
        verdict = 'REAL'
        print('\n✅ VERDICT: REAL MARKET DATA')
        print('   LP field shows "YOFX" (FIX gateway)')
        print('   Price patterns and timestamps show real market characteristics')
    elif simulated_indicators > real_indicators:
        verdict = 'LIKELY_SIMULATED'
        print('\n⚠️  VERDICT: LIKELY SIMULATED')
        print('   Patterns suggest simulation despite LP field')
    else:
        verdict = 'INCONCLUSIVE'
        print('\n❓ VERDICT: INCONCLUSIVE')
        print('   Mixed indicators - need more data')

    duration = (datetime.now() - start_time).total_seconds()
    print('\n' + '═' * 60)
    print(f"Total ticks analyzed: {total_tick_count}")
    print(f"Symbols detected: {', '.join(sorted(analysis['symbols']))}")
    print(f"Duration: {duration:.1f}s")
    print('═' * 60)

async def capture_and_analyze():
    """Main capture and analysis function"""
    global total_tick_count

    print('═' * 60)
    print('🔌 WebSocket Quote Analysis Tool')
    print('═' * 60)
    print(f"Connecting to: {WS_URL}")
    print(f"Target: Capture {CAPTURE_COUNT} ticks per symbol")
    print(f"Timeout: {TIMEOUT}s")
    print('═' * 60)

    try:
        async with websockets.connect(WS_URL, ping_interval=None) as websocket:
            print('✅ Connected to WebSocket')
            print('📊 Capturing ticks...\n')

            start = datetime.now()

            while True:
                # Check timeout
                if (datetime.now() - start).total_seconds() > TIMEOUT:
                    print('\n⏰ Analysis timeout reached')
                    break

                try:
                    # Receive message with timeout
                    message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                    tick = json.loads(message)

                    if tick.get('type') == 'tick':
                        total_tick_count += 1
                        symbol = tick['symbol']
                        analysis['symbols'].add(symbol)

                        # Store tick
                        if len(captured_ticks[symbol]) < CAPTURE_COUNT:
                            captured_ticks[symbol].append(tick)

                        # Log first few ticks
                        if total_tick_count <= 5:
                            timestamp = datetime.fromtimestamp(tick['timestamp']).isoformat()
                            print(f"[{total_tick_count}] {symbol} | Bid: {tick['bid']:.5f} | Ask: {tick['ask']:.5f} | LP: {tick.get('lp', 'N/A')} | Time: {timestamp}")

                        # Check if we have enough data
                        symbols_with_enough = sum(1 for ticks in captured_ticks.values() if len(ticks) >= CAPTURE_COUNT)
                        if symbols_with_enough >= min(3, len(analysis['symbols'])) and total_tick_count >= CAPTURE_COUNT:
                            break

                except asyncio.TimeoutError:
                    continue
                except json.JSONDecodeError as e:
                    print(f"Error parsing tick: {e}")
                    continue

    except Exception as e:
        print(f"❌ Connection error: {e}")
        sys.exit(1)

    print('\n🔌 WebSocket disconnected')

    # Perform analysis
    all_ticks = [tick for ticks in captured_ticks.values() for tick in ticks]

    if not all_ticks:
        print('\n❌ No ticks captured. Server may not be sending data.')
        sys.exit(1)

    analyze_lp(all_ticks)
    analyze_price_patterns(all_ticks)
    analyze_timestamps(all_ticks)
    analyze_spread(all_ticks)
    generate_verdict()

if __name__ == '__main__':
    try:
        asyncio.run(capture_and_analyze())
    except KeyboardInterrupt:
        print('\n\n⚠️  Analysis interrupted by user')
        sys.exit(0)

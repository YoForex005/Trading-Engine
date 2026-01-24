#!/usr/bin/env node
/**
 * WebSocket Quote Analysis Tool
 * Connects to ws://localhost:7999/ws and analyzes ticks to determine if they are real or simulated
 */

const WebSocket = require('ws');

// Configuration
const WS_URL = 'ws://localhost:7999/ws';
const CAPTURE_COUNT = 20; // Number of ticks to capture per symbol
const ANALYSIS_TIMEOUT = 30000; // 30 seconds timeout

// Data collection
const capturedTicks = {};
let totalTickCount = 0;
let startTime = Date.now();

// Analysis results
const analysis = {
    lpSources: new Set(),
    pricePatterns: [],
    timestampGaps: [],
    spreadVariations: [],
    symbols: new Set(),
    verdict: null,
    evidence: []
};

function analyzeLP(ticks) {
    const lpCounts = {};
    ticks.forEach(tick => {
        const lp = tick.lp || 'UNKNOWN';
        lpCounts[lp] = (lpCounts[lp] || 0) + 1;
        analysis.lpSources.add(lp);
    });

    console.log('\n📊 LP SOURCE ANALYSIS:');
    console.log('═'.repeat(60));
    Object.entries(lpCounts).forEach(([lp, count]) => {
        const percentage = ((count / ticks.length) * 100).toFixed(1);
        console.log(`  ${lp}: ${count} ticks (${percentage}%)`);

        if (lp === 'SIMULATED') {
            analysis.evidence.push(`❌ LP field shows "SIMULATED" - clear simulation marker`);
        } else if (lp === 'YOFX') {
            analysis.evidence.push(`✅ LP field shows "YOFX" - indicates real FIX gateway data`);
        }
    });
}

function analyzePricePatterns(ticks) {
    console.log('\n📈 PRICE PATTERN ANALYSIS:');
    console.log('═'.repeat(60));

    // Group by symbol
    const bySymbol = {};
    ticks.forEach(tick => {
        if (!bySymbol[tick.symbol]) bySymbol[tick.symbol] = [];
        bySymbol[tick.symbol].push(tick);
    });

    Object.entries(bySymbol).forEach(([symbol, symbolTicks]) => {
        if (symbolTicks.length < 3) return;

        // Calculate price changes
        const priceChanges = [];
        for (let i = 1; i < symbolTicks.length; i++) {
            const change = symbolTicks[i].bid - symbolTicks[i-1].bid;
            priceChanges.push(change);
        }

        // Check for regular patterns (simulated data often has regular patterns)
        const uniqueChanges = new Set(priceChanges.map(c => c.toFixed(6)));
        const regularityRatio = uniqueChanges.size / priceChanges.length;

        console.log(`  ${symbol}:`);
        console.log(`    - Price changes: ${priceChanges.length}`);
        console.log(`    - Unique changes: ${uniqueChanges.size}`);
        console.log(`    - Regularity ratio: ${regularityRatio.toFixed(3)} (lower = more regular)`);

        if (regularityRatio < 0.5) {
            analysis.evidence.push(`⚠️  ${symbol}: Low regularity ratio (${regularityRatio.toFixed(3)}) suggests simulated data`);
        } else {
            analysis.evidence.push(`✅ ${symbol}: High price variation (${regularityRatio.toFixed(3)}) suggests real market data`);
        }

        // Sample of actual changes
        const sampleChanges = priceChanges.slice(0, 5).map(c => c.toFixed(6));
        console.log(`    - Sample changes: [${sampleChanges.join(', ')}]`);
    });
}

function analyzeTimestamps(ticks) {
    console.log('\n⏰ TIMESTAMP ANALYSIS:');
    console.log('═'.repeat(60));

    const timestamps = ticks.map(t => t.timestamp).sort((a, b) => a - b);
    const gaps = [];

    for (let i = 1; i < timestamps.length; i++) {
        const gap = timestamps[i] - timestamps[i-1];
        gaps.push(gap);
    }

    // Statistical analysis
    const avgGap = gaps.reduce((a, b) => a + b, 0) / gaps.length;
    const uniqueGaps = new Set(gaps);
    const gapRegularity = uniqueGaps.size / gaps.length;

    console.log(`  Average gap: ${avgGap.toFixed(2)}s`);
    console.log(`  Unique gaps: ${uniqueGaps.size} / ${gaps.length}`);
    console.log(`  Gap regularity: ${gapRegularity.toFixed(3)}`);

    // Check for perfectly regular intervals (sign of simulation)
    if (gapRegularity < 0.3 && avgGap < 1) {
        analysis.evidence.push(`⚠️  Very regular timestamp intervals (${avgGap.toFixed(2)}s) suggests simulation`);
    } else {
        analysis.evidence.push(`✅ Irregular timestamp intervals suggest real market data`);
    }

    // Sample gaps
    const sampleGaps = gaps.slice(0, 10).map(g => g.toFixed(2));
    console.log(`  Sample gaps (s): [${sampleGaps.join(', ')}]`);
}

function analyzeSpread(ticks) {
    console.log('\n💰 SPREAD ANALYSIS:');
    console.log('═'.repeat(60));

    const bySymbol = {};
    ticks.forEach(tick => {
        if (!bySymbol[tick.symbol]) bySymbol[tick.symbol] = [];
        bySymbol[tick.symbol].push(tick);
    });

    Object.entries(bySymbol).forEach(([symbol, symbolTicks]) => {
        const spreads = symbolTicks.map(t => (t.ask - t.bid).toFixed(6));
        const uniqueSpreads = new Set(spreads);
        const spreadConsistency = 1 - (uniqueSpreads.size / spreads.length);

        console.log(`  ${symbol}:`);
        console.log(`    - Total ticks: ${spreads.length}`);
        console.log(`    - Unique spreads: ${uniqueSpreads.size}`);
        console.log(`    - Consistency: ${(spreadConsistency * 100).toFixed(1)}%`);

        if (spreadConsistency > 0.95) {
            analysis.evidence.push(`⚠️  ${symbol}: Perfectly consistent spread (${(spreadConsistency * 100).toFixed(1)}%) suggests simulation`);
        } else {
            analysis.evidence.push(`✅ ${symbol}: Variable spread suggests real market conditions`);
        }

        // Sample spreads
        const sampleSpreads = [...uniqueSpreads].slice(0, 5);
        console.log(`    - Sample spreads: [${sampleSpreads.join(', ')}]`);
    });
}

function generateVerdict() {
    console.log('\n\n' + '═'.repeat(60));
    console.log('🔍 FINAL VERDICT');
    console.log('═'.repeat(60));

    // Count positive and negative indicators
    const simulatedIndicators = analysis.evidence.filter(e => e.includes('⚠️') || e.includes('❌')).length;
    const realIndicators = analysis.evidence.filter(e => e.includes('✅')).length;

    console.log('\nEVIDENCE SUMMARY:');
    analysis.evidence.forEach(evidence => console.log(`  ${evidence}`));

    console.log('\n' + '─'.repeat(60));
    console.log(`Simulated indicators: ${simulatedIndicators}`);
    console.log(`Real data indicators: ${realIndicators}`);
    console.log('─'.repeat(60));

    // Determine verdict
    if (analysis.lpSources.has('SIMULATED')) {
        analysis.verdict = 'SIMULATED';
        console.log('\n🎭 VERDICT: SIMULATED DATA');
        console.log('   The LP field explicitly shows "SIMULATED"');
    } else if (analysis.lpSources.has('YOFX') && realIndicators > simulatedIndicators) {
        analysis.verdict = 'REAL';
        console.log('\n✅ VERDICT: REAL MARKET DATA');
        console.log('   LP field shows "YOFX" (FIX gateway)');
        console.log('   Price patterns and timestamps show real market characteristics');
    } else if (simulatedIndicators > realIndicators) {
        analysis.verdict = 'LIKELY_SIMULATED';
        console.log('\n⚠️  VERDICT: LIKELY SIMULATED');
        console.log('   Patterns suggest simulation despite LP field');
    } else {
        analysis.verdict = 'INCONCLUSIVE';
        console.log('\n❓ VERDICT: INCONCLUSIVE');
        console.log('   Mixed indicators - need more data');
    }

    console.log('\n' + '═'.repeat(60));
    console.log(`Total ticks analyzed: ${totalTickCount}`);
    console.log(`Symbols detected: ${[...analysis.symbols].join(', ')}`);
    console.log(`Duration: ${((Date.now() - startTime) / 1000).toFixed(1)}s`);
    console.log('═'.repeat(60));
}

function main() {
    console.log('═'.repeat(60));
    console.log('🔌 WebSocket Quote Analysis Tool');
    console.log('═'.repeat(60));
    console.log(`Connecting to: ${WS_URL}`);
    console.log(`Target: Capture ${CAPTURE_COUNT} ticks per symbol`);
    console.log(`Timeout: ${ANALYSIS_TIMEOUT/1000}s`);
    console.log('═'.repeat(60));

    const ws = new WebSocket(WS_URL);
    let analysisTimeout;

    ws.on('open', () => {
        console.log('✅ Connected to WebSocket');
        console.log('📊 Capturing ticks...\n');

        // Set timeout for analysis
        analysisTimeout = setTimeout(() => {
            console.log('\n⏰ Analysis timeout reached');
            ws.close();
        }, ANALYSIS_TIMEOUT);
    });

    ws.on('message', (data) => {
        try {
            const tick = JSON.parse(data.toString());

            if (tick.type === 'tick') {
                totalTickCount++;
                analysis.symbols.add(tick.symbol);

                // Store tick
                if (!capturedTicks[tick.symbol]) {
                    capturedTicks[tick.symbol] = [];
                }

                if (capturedTicks[tick.symbol].length < CAPTURE_COUNT) {
                    capturedTicks[tick.symbol].push(tick);

                    // Log first few ticks
                    if (totalTickCount <= 5) {
                        console.log(`[${totalTickCount}] ${tick.symbol} | Bid: ${tick.bid.toFixed(5)} | Ask: ${tick.ask.toFixed(5)} | LP: ${tick.lp || 'N/A'} | Time: ${new Date(tick.timestamp * 1000).toISOString()}`);
                    }
                }

                // Check if we have enough data
                const symbolsWithEnoughData = Object.values(capturedTicks).filter(arr => arr.length >= CAPTURE_COUNT).length;
                if (symbolsWithEnoughData >= Math.min(3, analysis.symbols.size) && totalTickCount >= CAPTURE_COUNT) {
                    clearTimeout(analysisTimeout);
                    ws.close();
                }
            }
        } catch (error) {
            console.error('Error parsing tick:', error.message);
        }
    });

    ws.on('error', (error) => {
        console.error('❌ WebSocket error:', error.message);
        process.exit(1);
    });

    ws.on('close', () => {
        console.log('\n🔌 WebSocket disconnected');

        // Perform analysis
        const allTicks = Object.values(capturedTicks).flat();

        if (allTicks.length === 0) {
            console.log('\n❌ No ticks captured. Server may not be running or not sending data.');
            process.exit(1);
        }

        analyzeLP(allTicks);
        analyzePricePatterns(allTicks);
        analyzeTimestamps(allTicks);
        analyzeSpread(allTicks);
        generateVerdict();

        process.exit(0);
    });
}

// Run the analysis
main();

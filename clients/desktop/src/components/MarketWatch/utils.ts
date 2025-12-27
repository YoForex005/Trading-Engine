export const getSymbolCategory = (symbol: string): 'Forex' | 'Crypto' | 'Metals' | 'Indices' | 'Other' => {
    const s = symbol.toUpperCase();

    if (s.includes('BTC') || s.includes('ETH') || s.includes('SOL') || s.includes('XRP') || s.includes('LTC') || s.includes('DOGE')) return 'Crypto';
    if (s.includes('XAU') || s.includes('XAG') || s.includes('GOLD') || s.includes('SILVER')) return 'Metals';
    if (s.includes('US30') || s.includes('NAS100') || s.includes('SPX') || s.includes('GER30') || s.includes('UK100') || s.includes('JP225')) return 'Indices';

    // Default to Forex for standard pairs
    // Assuming 6 char pairs like EURUSD, GBPJPY are forex
    if (s.length === 6 && !s.includes('USD')) {
        // Logic for non-USD pairs might be tricky, but generally correct
        return 'Forex';
    }
    if (s.includes('USD') || s.includes('EUR') || s.includes('GBP') || s.includes('JPY') || s.includes('AUD') || s.includes('CAD') || s.includes('CHF') || s.includes('NZD')) return 'Forex';

    return 'Other';
};

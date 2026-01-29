import React from 'react';
import {
    PlusSquare,
    Save,
    FileText,
    FolderOpen,
    Printer,
    UserPlus,
    LogIn,
    Globe,
    LogOut,
    Box,
    Layout,
    Layers,
    Monitor,
    Settings,
    CreditCard,
    HelpCircle,
    ChevronRight,
    TrendingUp,
    Activity,
    BarChart2,
    Zap,
    Code,
    Cpu,
    Terminal
} from 'lucide-react';

export type MenuItem = {
    label: string;
    shortcut?: string;
    icon?: React.ReactNode;
    children?: MenuItem[];
    divider?: boolean;
    disabled?: boolean;
    action?: () => void;
    scrollableChildren?: boolean;
    category?: 'trend' | 'oscillator' | 'volume' | 'bill-williams' | 'custom' | 'expert' | 'script';
};

// Helper to create simple label items
const item = (label: string, shortcut?: string, icon?: React.ReactNode): MenuItem => ({ label, shortcut, icon });
const divider = (): MenuItem => ({ label: '', divider: true });

// --- INDICATORS STRUCTURE ---

const trendIndicators: MenuItem[] = [
    item('Adaptive Moving Average'),
    item('Average Directional Movement Index'),
    item('Average Directional Movement Index Wilder'),
    item('Bollinger Bands'),
    item('Double Exponential Moving Average'),
    item('Envelopes'),
    item('Fractal Adaptive Moving Average'),
    item('Ichimoku Kinko Hyo'),
    item('Moving Average'),
    item('Parabolic SAR'),
    item('Standard Deviation'),
    item('Triple Exponential Moving Average'),
    item('Variable Index Dynamic Average'),
];

const oscillatorIndicators: MenuItem[] = [
    item('Average True Range'),
    item('Bears Power'),
    item('Bulls Power'),
    item('Chaikin Oscillator'),
    item('Commodity Channel Index'),
    item('DeMarker'),
    item('Force Index'),
    item('MACD'),
    item('Momentum'),
    item('Moving Average of Oscillator'),
    item('Relative Strength Index'),
    item('Relative Vigor Index'),
    item('Stochastic Oscillator'),
    item('Triple Exponential Average'),
    item('Williams\' Percent Range'),
];

const volumeIndicators: MenuItem[] = [
    item('Accumulation/Distribution'),
    item('Money Flow Index'),
    item('On Balance Volume'),
    item('Volumes'),
];

const billWilliamsIndicators: MenuItem[] = [
    item('Accelerator Oscillator'),
    item('Alligator'),
    item('Awesome Oscillator'),
    item('Fractals'),
    item('Gator Oscillator'),
    item('Market Facilitation Index'),
];

// --- OBJECTS STRUCTURE ---

const lineObjects: MenuItem[] = [
    item('Vertical Line'),
    item('Horizontal Line'),
    item('Trendline'),
    item('Trendline by Angle'),
    item('Cycle Lines'),
    item('Arrow Line'),
];

const channelObjects: MenuItem[] = [
    item('Equidistant Channel'),
    item('Standard Deviation Channel'),
    item('Regression Channel'),
    item('Andrews\' Pitchfork'),
];

const gannObjects: MenuItem[] = [
    item('Gann Line'),
    item('Gann Fan'),
    item('Gann Grid'),
];

const fibonacciObjects: MenuItem[] = [
    item('Fibonacci Retracement'),
    item('Fibonacci Time Zones'),
    item('Fibonacci Fan'),
    item('Fibonacci Arcs'),
    item('Fibonacci Channel'),
    item('Fibonacci Expansion'),
];

const shapeObjects: MenuItem[] = [
    item('Rectangle'),
    item('Triangle'),
    item('Ellipse'),
];

const arrowObjects: MenuItem[] = [
    item('Thumbs Up'),
    item('Thumbs Down'),
    item('Arrow Up'),
    item('Arrow Down'),
    item('Stop Sign'),
    item('Check Sign'),
];

// --- EXPERTS & SCRIPTS ---

const expertsList: MenuItem[] = [
    item('Macd Sample', undefined, <Cpu size={ 14} className = "text-zinc-500" />),
    item('Moving Average', undefined, <Cpu size={ 14} className = "text-zinc-500" />),
    item('MQL5 \ Experts \ ...', undefined, <FolderOpen size={ 14} className = "text-zinc-600" />),
];

const scriptsList: MenuItem[] = [
    item('PeriodConverter', undefined, <Terminal size={ 14} className = "text-zinc-500" />),
    item('MQL5 \ Scripts \ ...', undefined, <FolderOpen size={ 14} className = "text-zinc-600" />),
];

// --- MAIN MENU DATA ---

export const getMenuItems = (
    languageList: MenuItem[]
): { [key: string]: MenuItem[] } => ({
    File: [
        item('New Chart', undefined, <PlusSquare size={ 14} className = "text-emerald-500" />),
        item('Open Deleted', undefined), // Disabled handled in component if needed, or update type
        // ... Reusing logic from existing MenuBar, but cleaner
        { label: 'Profiles', children: [item('Default'), item('Euro'), item('Market')] },
        item('Close', 'Ctrl+F4'),
        divider(),
        item('Save', 'Ctrl+S', <Save size={ 14} />),
        item('Save As Picture', undefined, <FileText size={ 14} />),
        divider(),
        item('Open Data Folder', 'Ctrl+Shift+D', <FolderOpen size={ 14} />),
        divider(),
        item('Print', 'Ctrl+P', <Printer size={ 14} />),
        divider(),
        item('Open an Account', undefined, <UserPlus size={ 14} className = "text-blue-400" />),
        item('Login to Trade Account', undefined, <LogIn size={ 14} className = "text-blue-400" />),
        item('Login to Web Trader', undefined, <Globe size={ 14} />),
        divider(),
        item('Exit', undefined, <LogOut size={ 14} className = "text-rose-400" />)
    ],
    View: [
        { label: 'Languages', children: languageList, scrollableChildren: true },
        { label: 'Color Themes', children: [item('Dark'), item('Light'), item('Color')] },
        divider(),
        { label: 'Toolbars', children: [item('Standard', undefined, <Box size={ 12} />), item('Line Studies'), item('Timeframes')] },
        item('Status Bar', undefined, <Layout size={ 14} />),
        item('Charts Bar', undefined, <Layers size={ 14} />),
        divider(),
        item('Symbols', 'Ctrl+U'),
        item('Depth of Market', 'Alt+B'),
        item('Market Watch', 'Ctrl+M', <Monitor size={ 14} />),
        item('Data Window', 'Ctrl+D'),
        item('Navigator', 'Ctrl+N'),
        item('Toolbox', 'Ctrl+T'),
        item('Strategy Tester', 'Ctrl+R'),
        divider(),
        item('Full Screen', 'F11')
    ],
    Insert: [
        {
            label: 'Indicators',
            icon: <Activity size={ 14} />,
        children: [
            { label: 'Trend', children: trendIndicators, icon: <TrendingUp size={ 12} /> },
    { label: 'Oscillators', children: oscillatorIndicators, icon: <Activity size={ 12}/> },
{ label: 'Volumes', children: volumeIndicators, icon: <BarChart2 size={ 12 }/> },
{ label: 'Bill Williams', children: billWilliamsIndicators },
{ label: 'Custom', children: [item('Custom Indicator...')] },
            ]
        },
divider(),
{
    label: 'Objects',
    children: [
        { label: 'Lines', children: lineObjects },
        { label: 'Channels', children: channelObjects },
        { label: 'Gann', children: gannObjects },
        { label: 'Fibonacci', children: fibonacciObjects },
        { label: 'Shapes', children: shapeObjects },
        { label: 'Arrows', children: arrowObjects },
    ]
},
    divider(),
{
    label: 'Experts',
    icon: <Cpu size={ 14 } />,
children: expertsList
        },
{
    label: 'Scripts',
        icon: <Terminal size={ 14 } />,
    children: scriptsList
}
    ],
Charts: [
    item('Depth of Market', 'Alt+B'),
    item('Indicator List', 'Ctrl+I'),
    { label: 'Objects', children: [item('Objects List', 'Ctrl+B')] },
    divider(),
    item('Bar Chart', 'Alt+1'),
    item('Candlesticks', 'Alt+2'),
    item('Line Chart', 'Alt+3'),
    divider(),
    item('Grid', 'Ctrl+G'),
    item('Auto Scroll'),
    item('Chart Shift'),
    item('Volumes', 'Ctrl+L'),
    item('Tick Volumes'),
    divider(),
    item('Zoom In', '+'),
    item('Zoom Out', '-'),
    divider(),
    item('Properties', 'F8', <Settings size={ 14} />)
],
    Tools: [
        item('New Order', 'F9', <PlusSquare size={ 14} className = "text-emerald-500" />),
        item('History Center', 'F2'),
        item('Global Variables', 'F3'),
        item('MetaQuotes Language Editor', 'F4', <CreditCard size={ 14} />),
        divider(),
        item('Options', 'Ctrl+O', <Settings size={ 14} />)
    ],
        Window: [
            item('New Window'),
            divider(),
            item('Tile Windows', 'Alt+R'),
            item('Cascade'),
            divider(),
            { label: 'Resolution', children: [item('1920x1080'), item('1280x720')] }
        ],
            Help: [
                item('Help Topics', 'F1', <HelpCircle size={ 14} />),
                item('Web Terminal'),
                item('MQL5.community'),
                divider(),
                item('About', undefined, <Box size={ 14} />)
            ]
});

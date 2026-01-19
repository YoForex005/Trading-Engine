import React, { useState, useEffect, useRef } from 'react';
import {
    ChevronRight,
    PlusSquare,
    Save,
    FileText,
    FolderOpen,
    Printer,
    UserPlus,
    LogIn,
    LogOut,
    Layout,
    Box,
    Layers,
    Monitor,
    CreditCard,
    Globe,
    List,
    Delete,
    Trash2,
    Trash,
    MoveHorizontal,
    MousePointer2,
    Undo2,
    Grid,
    MoveDown,
    MoveRight,
    BarChart,
    CandlestickChart,
    BarChart3,
    Settings,
    HelpCircle,
    LineChart as LucideLineChart, // Alias if needed, or just LineChart
    Cpu,
    Activity,
    Server,
    Shield,
    LayoutGrid,
    AlignHorizontalJustifyCenter,
    AlignVerticalJustifyCenter,
    Zap,
    Send,
    PlayCircle,
    Book,
    BookOpen,
    Brain,
    Code,
    Briefcase,
    ShoppingBag,
    Radio,
    ArrowUpRight,
    MessageSquare,
    Cloud,
    Smartphone,
    Laptop,
    Terminal,
    Download,
    CheckCircle2,
    AlertTriangle
} from 'lucide-react';
import { OptionsDialog } from '../settings/OptionsDialog';

// Local interface definition to avoid export issues
interface MenuItem {
    label: string;
    shortcut?: string;
    icon?: React.ReactNode;
    children?: MenuItem[];
    divider?: boolean;
    disabled?: boolean;
    action?: () => void;
    scrollableChildren?: boolean;
}

export const MenuBar = () => {
    const [activeMenu, setActiveMenu] = useState<string | null>(null);
    const [isOptionsOpen, setIsOptionsOpen] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setActiveMenu(null);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const languageList: MenuItem[] = [
        // Western / Latin
        { label: 'English', icon: <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div> },
        { label: 'German' },
        { label: 'French' },
        { label: 'Spanish' },
        { label: 'Italian' },
        { label: 'Portuguese (Brazil)' },
        { label: 'Portuguese (Portugal)' },
        { label: 'Dutch' },
        { label: 'Danish' },
        { label: 'Swedish' },
        { label: 'Finnish' },
        { label: 'Lithuanian' },
        { label: 'Latvian' },
        { label: 'Estonian' },
        { label: 'Polish' },
        { label: 'Czech' },
        { label: 'Slovak' },
        { label: 'Hungarian' },
        { label: 'Romanian' },
        { label: 'Slovenian' },
        { label: 'Croatian' },
        { divider: true, label: '' },

        // Cyrillic
        { label: 'Russian' },
        { label: 'Bulgarian' },
        { label: 'Serbian' },
        { label: 'Mongolian' },
        { divider: true, label: '' },

        // Asian
        { label: 'Chinese (Simplified)' },
        { label: 'Chinese (Traditional)' },
        { label: 'Japanese' },
        { label: 'Korean' },
        { label: 'Indonesian' },
        { label: 'Malay' },
        { label: 'Javanese' },
        { label: 'Vietnamese' },
        { label: 'Thai' },
        { divider: true, label: '' },

        // Middle Eastern / Indic
        { label: 'Arabic' },
        { label: 'Persian' },
        { label: 'Hebrew' },
        { label: 'Turkish' },
        { label: 'Hindi' },
        { label: 'Punjabi (India)' },
        { label: 'Punjabi (Pakistan)' },
        { label: 'Bengali' },
        { label: 'Marathi' },
        { label: 'Swahili' },
        { label: 'Hausa' }
    ];

    const menuItems: { [key: string]: MenuItem[] } = {
        File: [
            { label: 'New Chart', icon: <PlusSquare size={14} className="text-emerald-500" /> },
            { label: 'Open Deleted', disabled: true },
            { label: 'Profiles', children: [{ label: 'Default' }, { label: 'Euro' }, { label: 'Market' }] },
            { label: 'Close', shortcut: 'Ctrl+F4' },
            { divider: true, label: '' },
            { label: 'Save', icon: <Save size={14} />, shortcut: 'Ctrl+S' },
            { label: 'Save As Picture', icon: <FileText size={14} /> },
            { divider: true, label: '' },
            { label: 'Open Data Folder', icon: <FolderOpen size={14} />, shortcut: 'Ctrl+Shift+D' },
            { divider: true, label: '' },
            { label: 'Print', icon: <Printer size={14} />, shortcut: 'Ctrl+P' },
            { label: 'Print Preview' },
            { divider: true, label: '' },
            { label: 'Open an Account', icon: <UserPlus size={14} className="text-blue-400" /> },
            { label: 'Login to Trade Account', icon: <LogIn size={14} className="text-blue-400" /> },
            { label: 'Login to Web Trader', icon: <Globe size={14} /> },
            { divider: true, label: '' },
            { label: 'Exit', icon: <LogOut size={14} className="text-rose-400" />, action: () => alert('Exit clicked') }
        ],
        View: [
            { label: 'Languages', children: languageList, scrollableChildren: true },
            { label: 'Color Themes', children: [{ label: 'Dark' }, { label: 'Light' }, { label: 'Color' }] },
            { divider: true, label: '' },
            { label: 'Toolbars', children: [{ label: 'Standard', icon: <Box size={12} /> }, { label: 'Line Studies' }, { label: 'Timeframes' }] },
            { label: 'Status Bar', icon: <Layout size={14} /> },
            { label: 'Charts Bar', icon: <Layers size={14} /> },
            { divider: true, label: '' },
            { label: 'Symbols', shortcut: 'Ctrl+U' },
            { label: 'Depth of Market', shortcut: 'Alt+B' },
            { label: 'Market Watch', shortcut: 'Ctrl+M', icon: <Monitor size={14} /> },
            { label: 'Data Window', shortcut: 'Ctrl+D' },
            { label: 'Navigator', shortcut: 'Ctrl+N' },
            { label: 'Toolbox', shortcut: 'Ctrl+T' },
            { label: 'Strategy Tester', shortcut: 'Ctrl+R' },
            { divider: true, label: '' },
            { label: 'Full Screen', shortcut: 'F11' }
        ],
        Insert: [
            { label: 'Indicators', children: [{ label: 'Trend' }, { label: 'Oscillators' }, { label: 'Volumes' }] },
            { label: 'Objects', children: [{ label: 'Lines' }, { label: 'Channels' }, { label: 'Gann' }, { label: 'Fibonacci' }] },
            { label: 'Experts' },
            { label: 'Scripts' }
        ],
        Charts: [
            { label: 'Depth of Market', shortcut: 'Alt+B', icon: <Box size={14} /> },
            { label: 'Indicator List', shortcut: 'Ctrl+I', icon: <LucideLineChart size={14} /> },
            {
                label: 'Objects',
                children: [
                    // Management
                    { label: 'Object List', shortcut: 'Ctrl+B', icon: <List size={14} /> },
                    { divider: true, label: '' },

                    // Destructive Actions (Muted Red)
                    { label: 'Delete Last', shortcut: 'Backspace', icon: <Delete size={14} className="text-rose-400" /> },
                    { label: 'Delete All Selected', shortcut: 'Del', icon: <Trash2 size={14} className="text-rose-400" /> },
                    { label: 'Delete All Arrows', icon: <MoveHorizontal size={14} className="text-rose-400" /> },
                    { label: 'Delete All', icon: <Trash size={14} className="text-rose-500" /> },
                    { divider: true, label: '' },

                    // Selection / Undo
                    { label: 'Unselect All', icon: <MousePointer2 size={14} /> },
                    { label: 'Undo Delete', shortcut: 'Ctrl+Z', icon: <Undo2 size={14} /> }
                ]
            },
            { divider: true, label: '' },
            { label: 'Bar Chart', shortcut: 'Alt+1', icon: <BarChart size={14} /> },
            { label: 'Candlesticks', shortcut: 'Alt+2', icon: <CandlestickChart size={14} /> },
            { label: 'Line Chart', shortcut: 'Alt+3', icon: <LucideLineChart size={14} /> },
            { divider: true, label: '' },
            { label: 'Grid', shortcut: 'Ctrl+G', icon: <Grid size={14} /> },
            { label: 'Auto Scroll', icon: <MoveDown size={14} /> },
            { label: 'Chart Shift', icon: <MoveRight size={14} /> },
            { label: 'Volumes', shortcut: 'Ctrl+L', icon: <BarChart3 size={14} /> },
            { label: 'Tick Volumes' },
            { divider: true, label: '' },
            { label: 'Zoom In', shortcut: '+' },
            { label: 'Zoom Out', shortcut: '-' },
            { divider: true, label: '' },
            { label: 'Properties', shortcut: 'F8', icon: <Settings size={14} /> }
        ],
        Tools: [
            // Primary Trading
            { label: 'New Order', shortcut: 'F9', icon: <PlusSquare size={14} className="text-emerald-500" /> },
            { divider: true, label: '' },

            // Development & Automation (RTX5 Rebranded)
            { label: 'Strategy Tester', shortcut: 'Ctrl+R', icon: <Activity size={14} /> },
            { label: 'RTX5 Script Editor', shortcut: 'F4', icon: <CreditCard size={14} /> },
            { label: 'Agents Manager', shortcut: 'F6', icon: <Cpu size={14} /> },
            { divider: true, label: '' },

            // System & Monitoring
            { label: 'Task Manager', shortcut: 'F2', icon: <Server size={14} /> },
            { label: 'Global Variables', shortcut: 'F3', icon: <Globe size={14} /> },
            { divider: true, label: '' },

            // RTX5 Services
            { label: 'RTX5 Marketplace', icon: <Box size={14} className="text-blue-400" /> },
            { label: 'RTX5 Signals Hub', icon: <Activity size={14} className="text-emerald-400" /> },
            { label: 'RTX5 Cloud Hosting', icon: <Shield size={14} className="text-purple-400" /> },
            { divider: true, label: '' },

            // Settings
            { label: 'Options', shortcut: 'Ctrl+O', icon: <Settings size={14} />, action: () => setIsOptionsOpen(true) }
        ],
        Window: [
            { label: 'Tile Windows', shortcut: 'Alt+R', icon: <LayoutGrid size={14} /> },
            { label: 'Cascade', icon: <Layers size={14} /> },
            { label: 'Tile Horizontally', icon: <AlignHorizontalJustifyCenter size={14} /> },
            { label: 'Tile Vertically', icon: <AlignVerticalJustifyCenter size={14} /> },
            { label: 'Arrange Icons', icon: <Grid size={14} /> },
            { divider: true, label: '' },
            {
                label: 'Resolution',
                icon: <Monitor size={14} />,
                children: [
                    { label: '2160p', shortcut: '3840x2160' },
                    { label: '1440p', shortcut: '2560x1440' },
                    { label: '1080p', shortcut: '1920x1080' },
                    { label: '720p', shortcut: '1280x720' },
                    { label: '480p', shortcut: '854x480' },
                    { label: '360p', shortcut: '640x360' }
                ]
            }
        ],
        Help: [
            { label: 'Help Topics', shortcut: 'F1', icon: <HelpCircle size={14} /> },
            { label: "What's New", icon: <Zap size={14} className="text-yellow-400" /> },
            { label: 'RTX5 Telegram Channel', icon: <Send size={14} className="text-blue-400" /> },
            { label: 'Video Guides', icon: <PlayCircle size={14} />, children: [{ label: 'Getting Started' }, { label: 'Trading' }, { label: 'Analysis' }] },
            { divider: true, label: '' },
            { label: 'RTX5 Web Trader', icon: <Globe size={14} /> },
            { label: 'RTX5 Documentation', icon: <Book size={14} /> },
            { label: 'RTX5 AlgoBook', icon: <BookOpen size={14} className="text-amber-400" /> },
            { label: 'RTX5 NeuroBook', icon: <Brain size={14} className="text-purple-400" /> },
            { label: 'RTX5 Articles', icon: <FileText size={14} /> },
            { label: 'RTX5 Code Base', icon: <Code size={14} /> },
            { label: 'RTX5 Jobs', icon: <Briefcase size={14} /> },
            { label: 'RTX5 Marketplace', icon: <ShoppingBag size={14} className="text-blue-400" /> },
            { label: 'RTX5 Signals', icon: <Radio size={14} className="text-emerald-400" /> },
            { label: 'RTX5 Quotes', icon: <ArrowUpRight size={14} /> },
            { label: 'RTX5 Forum', icon: <MessageSquare size={14} /> },
            { label: 'RTX5 Cloud Hosting', icon: <Cloud size={14} className="text-sky-400" /> },
            { divider: true, label: '' },
            {
                label: 'Mobile',
                icon: <Smartphone size={14} />,
                children: [
                    { label: 'Economic Calendar' },
                    { label: 'RTX5 Messenger' },
                    { label: 'RTX5 for iOS' },
                    { label: 'RTX5 for Android' }
                ]
            },
            { label: 'RTX5 for Mac', icon: <Laptop size={14} /> },
            { label: 'RTX5 for Linux', icon: <Terminal size={14} /> },
            { divider: true, label: '' },
            {
                label: 'Check For Updates',
                icon: <Download size={14} className="text-blue-500" />,
                children: [
                    { label: 'Latest Release Version', icon: <CheckCircle2 size={12} className="text-emerald-500" /> },
                    { label: 'Latest Beta Version', icon: <AlertTriangle size={12} className="text-amber-500" /> }
                ]
            },
            { divider: true, label: '' },
            { label: 'About', icon: <Box size={14} /> }
        ]
    };

    return (
        <>
            <div className="h-7 bg-[#252528] border-b border-zinc-800 flex items-center px-1 select-none z-50 relative" ref={menuRef}>
                {Object.keys(menuItems).map((key) => (
                    <div key={key} className="relative">
                        <button
                            onClick={() => setActiveMenu(activeMenu === key ? null : key)}
                            onMouseEnter={() => activeMenu && setActiveMenu(key)}
                            className={`px-3 py-1 text-[12px] rounded transition-colors cursor-default outline-none flex items-center
                                ${activeMenu === key
                                    ? 'bg-[#3b82f6] text-white shadow-sm font-medium'
                                    : 'text-zinc-300 hover:bg-zinc-700/50 hover:text-zinc-100'}`}
                        >
                            {key}
                        </button>

                        {/* Dropdown */}
                        {activeMenu === key && (
                            <div className="absolute top-full left-0 mt-1 min-w-[240px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[100] animate-in fade-in zoom-in-95 duration-100 origin-top-left">
                                {menuItems[key].map((item, index) => (
                                    <MenuItem key={index} item={item} />
                                ))}
                            </div>
                        )}
                    </div>
                ))}
            </div>

            {/* Options Dialog Modal */}
            <OptionsDialog isOpen={isOptionsOpen} onClose={() => setIsOptionsOpen(false)} />
        </>
    );
};

const MenuItem = ({ item }: { item: MenuItem }) => {
    if (item.divider) {
        return <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>;
    }

    return (
        <div className="relative group px-1">
            <button
                disabled={item.disabled}
                onClick={() => {
                    if (item.action) {
                        item.action();
                        // We rely on parent to close, but since we don't have access to setActiveMenu here easily without context, 
                        // we assume the action might involve closing or the blur will handle it.
                        // Actually, for "Options" we want it to close the menu.
                        // A simple hack is to simulate a click on body or just let the blur handler do it if we click away.
                        // But since we are opening a modal, the focus moves? 
                        // Let's just execute action.
                    }
                }}
                className={`w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left
                    ${item.disabled ? 'text-zinc-600 cursor-not-allowed' : 'text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group'}`}
            >
                {/* Icon Area */}
                <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                    {item.icon}
                </div>

                {/* Label */}
                <span className="flex-1">{item.label}</span>

                {/* Shortcut or Submenu Arrow */}
                {item.children ? (
                    <ChevronRight size={12} className="text-zinc-500" />
                ) : (
                    item.shortcut && <span className="text-[10px] text-zinc-500 font-mono tracking-tighter">{item.shortcut}</span>
                )}
            </button>

            {/* Submenu Logic (Hover-based simplified) */}
            {item.children && (
                <div className={`absolute left-full top-0 ml-[-4px] mt-0 hidden group-hover:block min-w-[180px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[101]
                    ${item.scrollableChildren ? 'max-h-[400px] overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-600 scrollbar-track-transparent' : ''}
                 `}>
                    {item.children.map((subItem, idx) => (
                        <MenuItem key={idx} item={subItem} />
                    ))}
                </div>
            )}
        </div>
    );
};

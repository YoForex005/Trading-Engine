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
    AlertTriangle,
    GraduationCap,
    Moon,
    Sun
} from 'lucide-react';
import { OptionsDialog } from '../settings/OptionsDialog';
import { FileMenu } from './FileMenu';
import { useTranslation } from 'react-i18next';
import { useTheme } from '../../contexts/ThemeContext';

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
    const { t, i18n } = useTranslation();
    const { theme, setTheme } = useTheme();
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

    // Dynamic Symbol Loading
    const [availableSymbols, setAvailableSymbols] = useState<{ symbol: string; name: string; category: string }[]>([]);
    const [isLoadingSymbols, setIsLoadingSymbols] = useState(false);

    useEffect(() => {
        const fetchSymbols = async () => {
            try {
                // Dynamically import API to avoid circular deps if any, or just use global
                // Assuming api is exported from services/api
                const { default: api } = await import('../../services/api');
                setIsLoadingSymbols(true);
                const symbols = await api.market.getAvailableSymbols();
                if (symbols && symbols.length > 0) {
                    setAvailableSymbols(symbols);
                }
            } catch (error) {
                console.error('Failed to load menu symbols:', error);
            } finally {
                setIsLoadingSymbols(false);
            }
        };

        fetchSymbols();
    }, []);

    const languageList: MenuItem[] = [
        // Western / Latin - Fully Translated
        { label: 'English', action: () => i18n.changeLanguage('en'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'en' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Deutsch', action: () => i18n.changeLanguage('de'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'de' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Français', action: () => i18n.changeLanguage('fr'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'fr' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Español', action: () => i18n.changeLanguage('es'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'es' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Italiano', action: () => i18n.changeLanguage('it'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'it' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Português (BR)', action: () => i18n.changeLanguage('pt-BR'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'pt-BR' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Nederlands', action: () => i18n.changeLanguage('nl'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'nl' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Polski', action: () => i18n.changeLanguage('pl'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'pl' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Türkçe', action: () => i18n.changeLanguage('tr'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'tr' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { divider: true, label: '' },

        // Cyrillic - Fully Translated
        { label: 'Русский', action: () => i18n.changeLanguage('ru'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'ru' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { divider: true, label: '' },

        // Asian - Fully Translated
        { label: '简体中文', action: () => i18n.changeLanguage('zh-CN'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'zh-CN' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: '繁體中文', action: () => i18n.changeLanguage('zh-TW'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'zh-TW' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: '日本語', action: () => i18n.changeLanguage('ja'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'ja' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: '한국어', action: () => i18n.changeLanguage('ko'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'ko' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Bahasa Indonesia', action: () => i18n.changeLanguage('id'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'id' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'Tiếng Việt', action: () => i18n.changeLanguage('vi'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'vi' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'ภาษาไทย', action: () => i18n.changeLanguage('th'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'th' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { divider: true, label: '' },

        // Middle Eastern / Indic - Fully Translated
        { label: 'العربية', action: () => i18n.changeLanguage('ar'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'ar' ? 'bg-blue-500' : 'bg-transparent'}`}></div> },
        { label: 'हिंदी', action: () => i18n.changeLanguage('hi'), icon: <div className={`w-1.5 h-1.5 rounded-full ${i18n.language === 'hi' ? 'bg-blue-500' : 'bg-transparent'}`}></div> }
    ];

    const menuItems: { [key: string]: MenuItem[] } = {
        File: [
            {
                label: t('menu.file.newChart'),
                icon: <PlusSquare size={14} className="text-emerald-500" />,
                children: isLoadingSymbols
                    ? [{ label: 'Loading...' }]
                    : availableSymbols.length > 0
                        ? (() => {
                            const buildNestedMenu = (symbols: typeof availableSymbols): MenuItem[] => {
                                const root: MenuItem[] = [];
                                const groups: Record<string, typeof availableSymbols> = {};

                                symbols.forEach(sym => {
                                    // If category is empty, it's a root item
                                    if (!sym.category || sym.category.trim() === '') {
                                        root.push({
                                            label: sym.name,
                                            action: () => window.dispatchEvent(new CustomEvent('open-chart', { detail: { symbol: sym.symbol } }))
                                        });
                                        return;
                                    }

                                    const parts = sym.category.split('.');
                                    const topLevel = parts[0];
                                    if (!groups[topLevel]) groups[topLevel] = [];

                                    if (parts.length > 1) {
                                        groups[topLevel].push({ ...sym, category: parts.slice(1).join('.') });
                                    } else {
                                        groups[topLevel].push({ ...sym, category: '' });
                                    }
                                });

                                // Add grouped items (folders) after root items
                                Object.keys(groups).sort().forEach(key => {
                                    const groupSymbols = groups[key];
                                    // Recurse
                                    root.push({
                                        label: key,
                                        children: buildNestedMenu(groupSymbols)
                                    });
                                });
                                return root;
                            };
                            return buildNestedMenu(availableSymbols);
                        })()
                        : [
                            // Fallback if no API symbols
                            { label: 'EURUSD' }, { label: 'GBPUSD' }, { label: 'USDJPY' }
                        ]
            },
            { label: t('menu.file.openDeleted'), disabled: true },
            { label: t('menu.file.profiles'), children: [{ label: 'Default' }, { label: 'Euro' }, { label: 'Market' }] },
            { label: t('menu.file.close'), shortcut: 'Ctrl+F4', action: () => window.dispatchEvent(new CustomEvent('close-active-chart')) },
            { divider: true, label: '' },
            { label: t('menu.file.save'), icon: <Save size={14} />, shortcut: 'Ctrl+S' },
            { label: t('menu.file.saveAsPicture'), icon: <FileText size={14} /> },
            { divider: true, label: '' },
            { label: t('menu.file.openDataFolder'), icon: <FolderOpen size={14} />, shortcut: 'Ctrl+Shift+D' },
            { divider: true, label: '' },
            { label: t('menu.file.print'), icon: <Printer size={14} />, shortcut: 'Ctrl+P' },
            { label: t('menu.file.printPreview') },
            { label: t('menu.file.printSetup'), icon: <Settings size={14} /> },
            { divider: true, label: '' },
            { label: t('menu.file.openAccount'), icon: <UserPlus size={14} className="text-blue-400" /> },
            { label: t('menu.file.loginTrade'), icon: <LogIn size={14} className="text-blue-400" /> },
            { label: t('menu.file.loginWeb'), icon: <Globe size={14} /> },
            { label: t('menu.file.loginMql5'), icon: <GraduationCap size={14} className="text-blue-500" /> },
            { divider: true, label: '' },
            { label: t('menu.file.exit'), icon: <LogOut size={14} className="text-rose-400" />, action: () => alert('Exit clicked') }
        ],
        View: [
            { label: t('menu.view.languages'), children: languageList, scrollableChildren: true },
            {
                label: t('menu.view.colorThemes'), children: [
                    {
                        label: 'Dark',
                        icon: theme === 'dark' ? <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div> : <Moon size={14} className="text-zinc-400" />,
                        action: () => setTheme('dark')
                    },
                    {
                        label: 'Light',
                        icon: theme === 'light' ? <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div> : <Sun size={14} className="text-yellow-400" />,
                        action: () => setTheme('light')
                    }
                ]
            },
            { divider: true, label: '' },
            { label: t('menu.view.toolbars'), children: [{ label: 'Standard', icon: <Box size={12} /> }, { label: 'Line Studies' }, { label: 'Timeframes' }] },
            { label: t('menu.view.statusBar'), icon: <Layout size={14} /> },
            { label: t('menu.view.chartsBar'), icon: <Layers size={14} /> },
            { divider: true, label: '' },
            { label: t('menu.view.symbols'), shortcut: 'Ctrl+U' },
            { label: t('menu.view.depthOfMarket'), shortcut: 'Alt+B' },
            { label: t('menu.view.marketWatch'), shortcut: 'Ctrl+M', icon: <Monitor size={14} /> },
            { label: t('menu.view.dataWindow'), shortcut: 'Ctrl+D' },
            { label: t('menu.view.navigator'), shortcut: 'Ctrl+N' },
            { label: t('menu.view.toolbox'), shortcut: 'Ctrl+T' },
            { label: t('menu.view.strategyTester'), shortcut: 'Ctrl+R' },
            { divider: true, label: '' },
            { label: t('menu.view.fullScreen'), shortcut: 'F11' }
        ],
        Insert: [
            { label: t('menu.insert.indicators'), children: [{ label: 'Trend' }, { label: 'Oscillators' }, { label: 'Volumes' }] },
            { label: t('menu.insert.objects'), children: [{ label: 'Lines' }, { label: 'Channels' }, { label: 'Gann' }, { label: 'Fibonacci' }] },
            { label: t('menu.insert.experts') },
            { label: t('menu.insert.scripts') }
        ],
        Charts: [
            { label: t('menu.charts.depthOfMarket'), shortcut: 'Alt+B', icon: <Box size={14} /> },
            { label: t('menu.charts.indicatorList'), shortcut: 'Ctrl+I', icon: <LucideLineChart size={14} /> },
            {
                label: t('menu.charts.objects'),
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
            { label: t('menu.charts.barChart'), shortcut: 'Alt+1', icon: <BarChart size={14} /> },
            { label: t('menu.charts.candlesticks'), shortcut: 'Alt+2', icon: <CandlestickChart size={14} /> },
            { label: t('menu.charts.lineChart'), shortcut: 'Alt+3', icon: <LucideLineChart size={14} /> },
            { divider: true, label: '' },
            { label: t('menu.charts.grid'), shortcut: 'Ctrl+G', icon: <Grid size={14} /> },
            { label: t('menu.charts.autoScroll'), icon: <MoveDown size={14} /> },
            { label: t('menu.charts.chartShift'), icon: <MoveRight size={14} /> },
            { label: t('menu.charts.volumes'), shortcut: 'Ctrl+L', icon: <BarChart3 size={14} /> },
            { label: t('menu.charts.tickVolumes') },
            { divider: true, label: '' },
            { label: t('menu.charts.zoomIn'), shortcut: '+' },
            { label: t('menu.charts.zoomOut'), shortcut: '-' },
            { divider: true, label: '' },
            { label: t('menu.charts.properties'), shortcut: 'F8', icon: <Settings size={14} /> }
        ],
        Tools: [
            // Primary Trading
            { label: t('menu.tools.newOrder'), shortcut: 'F9', icon: <PlusSquare size={14} className="text-emerald-500" /> },
            { divider: true, label: '' },

            // Development & Automation (RTX5 Rebranded)
            { label: t('menu.tools.strategyTester'), shortcut: 'Ctrl+R', icon: <Activity size={14} /> },
            { label: t('menu.tools.scriptEditor'), shortcut: 'F4', icon: <CreditCard size={14} /> },
            { label: t('menu.tools.agentsManager'), shortcut: 'F6', icon: <Cpu size={14} /> },
            { divider: true, label: '' },

            // System & Monitoring
            { label: t('menu.tools.taskManager'), shortcut: 'F2', icon: <Server size={14} /> },
            { label: t('menu.tools.globalVariables'), shortcut: 'F3', icon: <Globe size={14} /> },
            { divider: true, label: '' },

            // RTX5 Services
            { label: t('menu.tools.marketplace'), icon: <Box size={14} className="text-blue-400" /> },
            { label: t('menu.tools.signalsHub'), icon: <Activity size={14} className="text-emerald-400" /> },
            { label: t('menu.tools.cloudHosting'), icon: <Shield size={14} className="text-purple-400" /> },
            { divider: true, label: '' },

            // Settings
            { label: t('menu.tools.options'), shortcut: 'Ctrl+O', icon: <Settings size={14} />, action: () => setIsOptionsOpen(true) }
        ],
        Window: [
            { label: t('menu.window.tileWindows'), shortcut: 'Alt+R', icon: <LayoutGrid size={14} /> },
            { label: t('menu.window.cascade'), icon: <Layers size={14} /> },
            { label: t('menu.window.tileHorizontally'), icon: <AlignHorizontalJustifyCenter size={14} /> },
            { label: t('menu.window.tileVertically'), icon: <AlignVerticalJustifyCenter size={14} /> },
            { label: t('menu.window.arrangeIcons'), icon: <Grid size={14} /> },
            { divider: true, label: '' },
            {
                label: t('menu.window.resolution'),
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
            { label: t('menu.help.helpTopics'), shortcut: 'F1', icon: <HelpCircle size={14} /> },
            { label: t('menu.help.whatsNew'), icon: <Zap size={14} className="text-yellow-400" /> },
            { label: t('menu.help.telegramChannel'), icon: <Send size={14} className="text-blue-400" /> },
            { label: t('menu.help.videoGuides'), icon: <PlayCircle size={14} />, children: [{ label: 'Getting Started' }, { label: 'Trading' }, { label: 'Analysis' }] },
            { divider: true, label: '' },
            { label: t('menu.help.webTrader'), icon: <Globe size={14} /> },
            { label: t('menu.help.documentation'), icon: <Book size={14} /> },
            { label: t('menu.help.algoBook'), icon: <BookOpen size={14} className="text-amber-400" /> },
            { label: t('menu.help.neuroBook'), icon: <Brain size={14} className="text-purple-400" /> },
            { label: t('menu.help.articles'), icon: <FileText size={14} /> },
            { label: t('menu.help.codeBase'), icon: <Code size={14} /> },
            { label: t('menu.help.jobs'), icon: <Briefcase size={14} /> },
            { label: t('menu.help.marketplace'), icon: <ShoppingBag size={14} className="text-blue-400" /> },
            { label: t('menu.help.signals'), icon: <Radio size={14} className="text-emerald-400" /> },
            { label: t('menu.help.quotes'), icon: <ArrowUpRight size={14} /> },
            { label: t('menu.help.forum'), icon: <MessageSquare size={14} /> },
            { label: t('menu.help.cloudHosting'), icon: <Cloud size={14} className="text-sky-400" /> },
            { divider: true, label: '' },
            {
                label: t('menu.help.mobile'),
                icon: <Smartphone size={14} />,
                children: [
                    { label: 'Economic Calendar' },
                    { label: 'RTX5 Messenger' },
                    { label: 'RTX5 for iOS' },
                    { label: 'RTX5 for Android' }
                ]
            },
            { label: t('menu.help.mac'), icon: <Laptop size={14} /> },
            { label: t('menu.help.linux'), icon: <Terminal size={14} /> },
            { divider: true, label: '' },
            {
                label: t('menu.help.checkForUpdates'),
                icon: <Download size={14} className="text-blue-500" />,
                children: [
                    { label: 'Latest Release Version', icon: <CheckCircle2 size={12} className="text-emerald-500" /> },
                    { label: 'Latest Beta Version', icon: <AlertTriangle size={12} className="text-amber-500" /> }
                ]
            },
            { divider: true, label: '' },
            { label: t('menu.help.about'), icon: <Box size={14} /> }
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
                            <div className="absolute top-full left-0 mt-1 z-[100]">
                                {key === 'File' ? (
                                    <FileMenu
                                        onSave={() => {
                                            console.log('[MenuBar] Save workspace triggered');
                                            // Dispatch custom event for save workspace
                                            window.dispatchEvent(new CustomEvent('saveWorkspace'));
                                            setActiveMenu(null);
                                        }}
                                        onSaveAsPicture={() => {
                                            console.log('Save as picture');
                                            setActiveMenu(null);
                                        }}
                                        onOpenDataFolder={() => {
                                            console.log('[MenuBar] Open data folder triggered');
                                            window.dispatchEvent(new CustomEvent('openDataFolder'));
                                            setActiveMenu(null);
                                        }}
                                        onPrint={() => {
                                            console.log('[MenuBar] Print triggered');
                                            window.dispatchEvent(new CustomEvent('printChart'));
                                            setActiveMenu(null);
                                        }}
                                        onPrintPreview={() => {
                                            console.log('Print preview');
                                            setActiveMenu(null);
                                        }}
                                        onPrintSetup={() => {
                                            console.log('Print setup');
                                            setActiveMenu(null);
                                        }}
                                        onExit={() => {
                                            console.log('Exit');
                                            setActiveMenu(null);
                                        }}
                                        hasUnsavedChanges={false}
                                    />
                                ) : (
                                    <div className="min-w-[240px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 animate-in fade-in zoom-in-95 duration-100 origin-top-left">
                                        {menuItems[key].map((item, index) => (
                                            <MenuItem
                                                key={index}
                                                item={item}
                                                onClose={() => setActiveMenu(null)}
                                            />
                                        ))}
                                    </div>
                                )}
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

const MenuItem = ({ item, onClose }: { item: MenuItem; onClose?: () => void }) => {
    if (item.divider) {
        return <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>;
    }

    return (
        <div className="relative group px-1 [&:hover>div]:block">
            <button
                disabled={item.disabled}
                onClick={() => {
                    if (item.action) {
                        item.action();
                        onClose?.();
                    }
                }}
                className={`w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left
                    ${item.disabled ? 'text-zinc-600 cursor-not-allowed' : 'text-zinc-300 hover:bg-[#2a2e39] hover:text-white group-hover:bg-[#2a2e39] group-hover:text-white cursor-default group'}`}
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

            {/* Submenu Logic (Hover-based strict) */}
            {item.children && (
                <div className={`absolute left-full top-0 ml-[-4px] mt-0 hidden min-w-[180px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[101]
                    ${item.scrollableChildren ? 'max-h-[400px] overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-600 scrollbar-track-transparent' : ''}
                 `}>
                    {item.children.map((subItem, idx) => (
                        <MenuItem
                            key={idx}
                            item={subItem}
                            onClose={onClose}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

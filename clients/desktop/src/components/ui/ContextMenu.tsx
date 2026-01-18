import { useState, useEffect, useRef, useCallback } from 'react';

type MenuItem = {
  label: string;
  icon?: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  divider?: boolean;
  shortcut?: string;
};

type ContextMenuProps = {
  items: MenuItem[];
  children: React.ReactNode;
  disabled?: boolean;
};

export function ContextMenu({ items, children, disabled = false }: ContextMenuProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const menuRef = useRef<HTMLDivElement>(null);

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    if (disabled) return;

    e.preventDefault();
    e.stopPropagation();

    setPosition({ x: e.clientX, y: e.clientY });
    setIsOpen(true);
  }, [disabled]);

  const handleClick = useCallback(() => {
    setIsOpen(false);
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen]);

  // Adjust menu position to keep it on screen
  useEffect(() => {
    if (isOpen && menuRef.current) {
      const menu = menuRef.current;
      const rect = menu.getBoundingClientRect();
      const { innerWidth, innerHeight } = window;

      let { x, y } = position;

      if (x + rect.width > innerWidth) {
        x = innerWidth - rect.width - 8;
      }

      if (y + rect.height > innerHeight) {
        y = innerHeight - rect.height - 8;
      }

      if (x !== position.x || y !== position.y) {
        setPosition({ x, y });
      }
    }
  }, [isOpen, position]);

  return (
    <>
      <div onContextMenu={handleContextMenu} onClick={handleClick}>
        {children}
      </div>

      {isOpen && (
        <div
          ref={menuRef}
          className="context-menu"
          style={{
            left: `${position.x}px`,
            top: `${position.y}px`
          }}
        >
          {items.map((item, index) => (
            <div key={index}>
              {item.divider ? (
                <div className="context-menu-divider" />
              ) : (
                <button
                  className={`context-menu-item w-full text-left flex items-center justify-between gap-4 ${
                    item.disabled ? 'opacity-50 cursor-not-allowed' : ''
                  }`}
                  onClick={() => {
                    if (!item.disabled) {
                      item.onClick();
                      setIsOpen(false);
                    }
                  }}
                  disabled={item.disabled}
                >
                  <div className="flex items-center gap-2">
                    {item.icon && <span className="w-4 h-4">{item.icon}</span>}
                    <span>{item.label}</span>
                  </div>
                  {item.shortcut && (
                    <span className="text-2xs text-zinc-500 font-mono">
                      {item.shortcut}
                    </span>
                  )}
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </>
  );
}

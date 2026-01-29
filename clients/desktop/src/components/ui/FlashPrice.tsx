import { useState, useEffect, useRef } from 'react';

type FlashPriceProps = {
  value: number | string;
  format?: (value: number) => string;
  direction?: 'up' | 'down' | 'none';
  flashDuration?: number;
  className?: string;
  colorize?: boolean;
};

export function FlashPrice({
  value,
  format = (v) => v.toFixed(5),
  direction = 'none',
  flashDuration = 200,
  className = '',
  colorize = true
}: FlashPriceProps) {
  const [isFlashing, setIsFlashing] = useState(false);
  const prevValueRef = useRef(value);
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (prevValueRef.current !== value) {
      setIsFlashing(true);

      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        setIsFlashing(false);
      }, flashDuration);

      prevValueRef.current = value;
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [value, flashDuration]);

  const flashClass = isFlashing
    ? direction === 'up'
      ? 'flash-profit'
      : direction === 'down'
      ? 'flash-loss'
      : ''
    : '';

  const textColor = colorize
    ? direction === 'up'
      ? 'text-profit'
      : direction === 'down'
      ? 'text-loss'
      : ''
    : '';

  const displayValue = typeof value === 'number' ? format(value) : value;

  return (
    <span className={`font-mono transition-colors ${flashClass} ${textColor} ${className}`}>
      {displayValue}
    </span>
  );
}

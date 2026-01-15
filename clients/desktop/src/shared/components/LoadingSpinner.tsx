/**
 * LoadingSpinner - Shared loading state component
 * Eliminates duplicated loading UI patterns
 */

type LoadingSpinnerProps = {
  message?: string;
  size?: 'small' | 'medium' | 'large';
};

export function LoadingSpinner({ message = 'Loading...', size = 'medium' }: LoadingSpinnerProps) {
  const sizeClasses = {
    small: 'w-4 h-4',
    medium: 'w-8 h-8',
    large: 'w-12 h-12',
  };

  return (
    <div className="flex flex-col items-center justify-center p-4">
      <div className={`animate-spin rounded-full border-b-2 border-blue-500 ${sizeClasses[size]}`} />
      {message && <span className="mt-2 text-sm text-gray-600">{message}</span>}
    </div>
  );
}

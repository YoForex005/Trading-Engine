import { ReactNode } from 'react';
import { clsx } from 'clsx';
import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';
import type { Sort, SortDirection } from '@/types';

type Column<T> = {
  key: string;
  header: string;
  render?: (row: T) => ReactNode;
  sortable?: boolean;
  className?: string;
};

type TableProps<T> = {
  data: T[];
  columns: Column<T>[];
  loading?: boolean;
  sort?: Sort;
  onSort?: (field: string) => void;
  onRowClick?: (row: T) => void;
  emptyMessage?: string;
  className?: string;
};

export const Table = <T extends Record<string, unknown>>({
  data,
  columns,
  loading = false,
  sort,
  onSort,
  onRowClick,
  emptyMessage = 'No data available',
  className,
}: TableProps<T>) => {
  const getSortIcon = (columnKey: string): ReactNode => {
    if (!sort || sort.field !== columnKey) {
      return <ArrowUpDown className="w-4 h-4 text-gray-400" />;
    }
    return sort.direction === 'asc' ? (
      <ArrowUp className="w-4 h-4 text-primary-600" />
    ) : (
      <ArrowDown className="w-4 h-4 text-primary-600" />
    );
  };

  return (
    <div className={clsx('overflow-x-auto', className)}>
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-800">
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                className={clsx(
                  'px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider',
                  column.sortable && onSort && 'cursor-pointer select-none hover:bg-gray-100 dark:hover:bg-gray-700',
                  column.className
                )}
                onClick={() => column.sortable && onSort?.(column.key)}
              >
                <div className="flex items-center gap-2">
                  {column.header}
                  {column.sortable && getSortIcon(column.key)}
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {loading ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-6 py-12 text-center text-gray-500 dark:text-gray-400"
              >
                <div className="flex items-center justify-center">
                  <svg
                    className="animate-spin h-8 w-8 text-primary-600"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                </div>
              </td>
            </tr>
          ) : data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-6 py-12 text-center text-gray-500 dark:text-gray-400"
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row, index) => (
              <tr
                key={index}
                className={clsx(
                  onRowClick && 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800'
                )}
                onClick={() => onRowClick?.(row)}
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={clsx(
                      'px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100',
                      column.className
                    )}
                  >
                    {column.render
                      ? column.render(row)
                      : String(row[column.key] ?? '')}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};

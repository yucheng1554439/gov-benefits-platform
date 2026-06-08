import { cn } from '@/lib/utils';
import type { ReactNode } from 'react';

interface Column<T> {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
  className?: string;
}

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T | ((row: T) => string);
  emptyMessage?: string;
  className?: string;
}

export function Table<T>({ columns, data, keyField, emptyMessage = 'No records found.', className }: TableProps<T>) {
  const getKey = (row: T) =>
    typeof keyField === 'function' ? keyField(row) : String(row[keyField]);

  return (
    <div className={cn('overflow-x-auto rounded-lg border border-gov-border', className)}>
      <table className="min-w-full divide-y divide-gov-border">
        <thead className="bg-gov-surface">
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                scope="col"
                className={cn('px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gov-navy', col.className)}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gov-border bg-white">
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-8 text-center text-gov-slate">
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row) => (
              <tr key={getKey(row)} className="hover:bg-gov-surface/50">
                {columns.map((col) => (
                  <td key={col.key} className={cn('px-4 py-3 text-sm text-gov-slate', col.className)}>
                    {col.render(row)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}

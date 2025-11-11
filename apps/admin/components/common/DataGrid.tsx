'use client';

import React, { useMemo, useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ArrowUpDown } from 'lucide-react';

/**
 * Column definition for DataGrid
 * T is the type of data being displayed
 */
export interface ColumnDef<T> {
  /** Unique key for this column */
  key: string;
  /** Header text or React element */
  header: string | React.ReactNode;
  /** Function to access the value from a row */
  accessor: (row: T) => any;
  /** Optional custom renderer for cell content */
  render?: (value: any, row: T) => React.ReactNode;
  /** Whether this column is sortable */
  sortable?: boolean;
  /** Optional column width (e.g., "200px" or "20%") */
  width?: string;
  /** Optional CSS classes for the column */
  className?: string;
}

/**
 * Props for DataGrid component
 * T is the type of data being displayed
 */
export interface DataGridProps<T> {
  /** Array of data items to display */
  data: T[];
  /** Column definitions */
  columns: ColumnDef<T>[];
  /** Optional callback when a row is clicked */
  onRowClick?: (row: T) => void;
  /** Whether data is currently loading */
  loading?: boolean;
  /** Message to display when there's no data */
  emptyMessage?: string;
  /** Optional function to get a unique ID for each row */
  getRowId?: (row: T, index: number) => string;
  /** Optional CSS classes for the container */
  className?: string;
}

/**
 * Reusable DataGrid component for displaying tabular data
 *
 * Features:
 * - Type-safe with TypeScript generics
 * - Client-side sorting
 * - Custom cell rendering
 * - Loading skeletons
 * - Empty state handling
 * - Responsive design
 * - Accessible (ARIA labels, keyboard navigation)
 *
 * @example
 * ```tsx
 * interface User {
 *   id: string;
 *   name: string;
 *   email: string;
 * }
 *
 * const columns: ColumnDef<User>[] = [
 *   {
 *     key: 'name',
 *     header: 'Name',
 *     accessor: (user) => user.name,
 *     sortable: true,
 *   },
 *   {
 *     key: 'email',
 *     header: 'Email',
 *     accessor: (user) => user.email,
 *   },
 * ];
 *
 * <DataGrid
 *   data={users}
 *   columns={columns}
 *   onRowClick={(user) => console.log(user)}
 * />
 * ```
 */
export function DataGrid<T>({
  data,
  columns,
  onRowClick,
  loading = false,
  emptyMessage = 'No data available',
  getRowId = (_, index) => index.toString(),
  className = '',
}: DataGridProps<T>) {
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  // Handle column header click for sorting
  const handleSort = (columnKey: string, sortable?: boolean) => {
    if (!sortable) return;

    if (sortColumn === columnKey) {
      // Toggle direction if same column
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // New column, default to ascending
      setSortColumn(columnKey);
      setSortDirection('asc');
    }
  };

  // Sort data based on current sort state
  const sortedData = useMemo(() => {
    if (!sortColumn) return data;

    const column = columns.find((col) => col.key === sortColumn);
    if (!column) return data;

    return [...data].sort((a, b) => {
      const aValue = column.accessor(a);
      const bValue = column.accessor(b);

      // Handle null/undefined values
      if (aValue == null && bValue == null) return 0;
      if (aValue == null) return 1;
      if (bValue == null) return -1;

      // Compare values
      let comparison = 0;
      if (typeof aValue === 'string' && typeof bValue === 'string') {
        comparison = aValue.localeCompare(bValue);
      } else if (typeof aValue === 'number' && typeof bValue === 'number') {
        comparison = aValue - bValue;
      } else {
        // Fallback to string comparison
        comparison = String(aValue).localeCompare(String(bValue));
      }

      return sortDirection === 'asc' ? comparison : -comparison;
    });
  }, [data, columns, sortColumn, sortDirection]);

  // Render loading skeleton
  if (loading) {
    return (
      <div className={`rounded-md border ${className}`}>
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((column) => (
                <TableHead key={column.key} style={{ width: column.width }}>
                  {column.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.from({ length: 5 }).map((_, i) => (
              <TableRow key={i}>
                {columns.map((column) => (
                  <TableCell key={column.key}>
                    <div className="h-4 w-full bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  // Render empty state
  if (data.length === 0) {
    return (
      <div className={`rounded-md border ${className}`}>
        <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
          <p className="text-neutral-500 dark:text-neutral-400">
            {emptyMessage}
          </p>
        </div>
      </div>
    );
  }

  // Render data grid
  return (
    <div className={`rounded-md border overflow-auto ${className}`}>
      <Table>
        <TableHeader>
          <TableRow>
            {columns.map((column) => (
              <TableHead
                key={column.key}
                style={{ width: column.width }}
                className={column.className}
              >
                {column.sortable ? (
                  <button
                    className="flex items-center gap-2 hover:text-neutral-900 dark:hover:text-neutral-100 transition-colors"
                    onClick={() => handleSort(column.key, column.sortable)}
                    aria-label={`Sort by ${column.header}`}
                  >
                    <span>{column.header}</span>
                    <ArrowUpDown
                      className={`h-4 w-4 ${
                        sortColumn === column.key
                          ? 'text-neutral-900 dark:text-neutral-100'
                          : 'text-neutral-400'
                      }`}
                    />
                  </button>
                ) : (
                  column.header
                )}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedData.map((row, index) => {
            const rowId = getRowId(row, index);
            return (
              <TableRow
                key={rowId}
                className={
                  onRowClick
                    ? 'cursor-pointer hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors'
                    : ''
                }
                onClick={() => onRowClick?.(row)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    onRowClick?.(row);
                  }
                }}
                tabIndex={onRowClick ? 0 : undefined}
                role={onRowClick ? 'button' : undefined}
              >
                {columns.map((column) => {
                  const value = column.accessor(row);
                  const content = column.render ? column.render(value, row) : value;

                  return (
                    <TableCell
                      key={column.key}
                      className={column.className}
                    >
                      {content}
                    </TableCell>
                  );
                })}
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}

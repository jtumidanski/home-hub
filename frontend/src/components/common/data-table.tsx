import { type RowData, flexRender, useTable } from "@tanstack/react-table";
import {
  type DataTableColumnDef,
  dataTableFeatures,
} from "@/components/common/data-table-features";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export type { DataTableColumnDef };

interface DataTableProps<TData extends RowData> {
  columns: DataTableColumnDef<TData>[];
  data: TData[];
  isLoading?: boolean;
  emptyMessage?: string;
  onRowClick?: (row: TData) => void;
  skeletonRows?: number;
}

export function DataTable<TData extends RowData>({
  columns,
  data,
  isLoading,
  emptyMessage = "No results.",
  onRowClick,
  skeletonRows = 4,
}: DataTableProps<TData>) {
  const table = useTable({
    features: dataTableFeatures,
    data,
    columns,
  });

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: skeletonRows }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <p className="text-muted-foreground">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <table className="w-full caption-bottom text-sm">
        <thead className="border-b bg-muted/50">
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th
                  key={header.id}
                  className="h-10 px-4 text-left align-middle font-medium text-muted-foreground"
                >
                  {header.isPlaceholder
                    ? null
                    : flexRender(header.column.columnDef.header, header.getContext())}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr
              key={row.id}
              className={cn(
                "border-b transition-colors hover:bg-muted/50",
                onRowClick && "cursor-pointer"
              )}
              onClick={() => onRowClick?.(row.original)}
            >
              {row.getAllCells().map((cell) => (
                <td key={cell.id} className="px-4 py-3 align-middle">
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

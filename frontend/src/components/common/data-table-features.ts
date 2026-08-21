import {
  type ColumnDef,
  type RowData,
  columnSizingFeature,
  tableFeatures,
} from "@tanstack/react-table";

/**
 * The feature set every DataTable is built with. Only column sizing is
 * registered — the table is presentational, with no sorting, filtering, or
 * pagination.
 */
export const dataTableFeatures = tableFeatures({
  columnSizingFeature,
});

/** Column definition accepted by `DataTable`. */
export type DataTableColumnDef<TData extends RowData> = ColumnDef<
  typeof dataTableFeatures,
  TData,
  unknown
>;

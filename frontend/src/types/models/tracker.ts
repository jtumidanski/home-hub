export interface TrackerAttributes {
  name: string;
  scale_type: ScaleType;
  scale_config: RangeConfig | null;
  schedule: number[];
  color: TrackerColor;
  sort_order: number;
  schedule_history?: ScheduleHistoryEntry[];
  created_at: string;
  updated_at: string;
}

export interface Tracker {
  id: string;
  type: "trackers";
  attributes: TrackerAttributes;
}

export interface ScheduleHistoryEntry {
  schedule: number[];
  effective_date: string;
}

export interface RangeConfig {
  min: number;
  max: number;
}

export type ScaleType = "sentiment" | "numeric" | "range";

export type TrackerColor =
  | "red" | "orange" | "amber" | "yellow"
  | "lime" | "green" | "emerald" | "teal"
  | "cyan" | "blue" | "indigo" | "violet"
  | "purple" | "fuchsia" | "pink" | "rose";

export const SCALE_TYPE_LABELS: Record<ScaleType, string> = {
  sentiment: "Sentiment",
  numeric: "Numeric",
  range: "Range",
};

export const COLOR_PALETTE: TrackerColor[] = [
  "red", "orange", "amber", "yellow",
  "lime", "green", "emerald", "teal",
  "cyan", "blue", "indigo", "violet",
  "purple", "fuchsia", "pink", "rose",
];

export const DAY_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

export interface TrackerEntryAttributes {
  tracking_item_id: string;
  date: string;
  value: SentimentValue | NumericValue | RangeValue | null;
  skipped: boolean;
  note?: string | null;
  scheduled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface TrackerEntry {
  id: string;
  type: "tracker-entries";
  attributes: TrackerEntryAttributes;
}

export interface SentimentValue {
  rating: "positive" | "neutral" | "negative";
}

export interface NumericValue {
  count: number;
}

export interface RangeValue {
  value: number;
}

export interface MonthCompletion {
  expected: number;
  filled: number;
  skipped: number;
  remaining: number;
}

export interface MonthSummaryAttributes {
  month: string;
  complete: boolean;
  completion: MonthCompletion;
}

export interface MonthItemInfo {
  id: string;
  name: string;
  scale_type: ScaleType;
  scale_config: RangeConfig | null;
  color: TrackerColor;
  sort_order: number;
  active_from: string;
  active_until: string | null;
  schedule_snapshots: ScheduleHistoryEntry[];
}

export interface MonthSummaryResponse {
  data: {
    type: "tracker-months";
    attributes: MonthSummaryAttributes;
    relationships: {
      items: { data: MonthItemInfo[] };
      entries: { data: TrackerEntry[] };
    };
  };
}

export interface ReportSummary {
  total_items: number;
  completion_rate: number;
  skip_rate: number;
  total_expected: number;
  total_filled: number;
  total_skipped: number;
}

export interface SentimentStats {
  expected_days: number;
  filled_days: number;
  skipped_days: number;
  positive: number;
  neutral: number;
  negative: number;
  positive_ratio: number;
  daily_values: { date: string; rating: string }[];
}

export interface NumericStats {
  expected_days: number;
  filled_days: number;
  skipped_days: number;
  total: number;
  daily_average: number;
  days_with_entries_above_zero: number;
  days_with_entries_above_zero_pct: number;
  min: { date: string; count: number } | null;
  max: { date: string; count: number } | null;
  daily_values: { date: string; count: number }[];
}

export interface RangeStats {
  expected_days: number;
  filled_days: number;
  skipped_days: number;
  average: number;
  min: { date: string; value: number } | null;
  max: { date: string; value: number } | null;
  std_dev: number;
  daily_values: { date: string; value: number }[];
}

export interface ReportItem {
  tracking_item_id: string;
  name: string;
  scale_type: ScaleType;
  stats: SentimentStats | NumericStats | RangeStats;
}

export interface MonthReportAttributes {
  month: string;
  summary: ReportSummary;
  items: ReportItem[];
}

export interface MonthReportResponse {
  data: {
    type: "tracker-reports";
    attributes: MonthReportAttributes;
  };
}

export interface TodayResponse {
  data: {
    type: "tracker-today";
    attributes: { date: string };
    relationships: {
      items: { data: Array<{ id: string; type: string; name: string; scale_type: ScaleType; scale_config: RangeConfig | null; color: TrackerColor; sort_order: number }> };
      entries: { data: TrackerEntry[] };
    };
  };
}

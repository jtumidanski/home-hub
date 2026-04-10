import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useMonthReport } from "@/lib/hooks/api/use-trackers";
import type { SentimentStats, NumericStats, RangeStats, ReportItem } from "@/types/models/tracker";

interface Props {
  month: string;
  onBackToCalendar: () => void;
}

function formatMonth(month: string) {
  const parts = month.split("-");
  const date = new Date(parseInt(parts[0] ?? "2026"), parseInt(parts[1] ?? "1") - 1);
  return date.toLocaleDateString("en-US", { month: "long", year: "numeric" });
}

export function MonthReport({ month, onBackToCalendar }: Props) {
  const { data, isLoading, error } = useMonthReport(month, true);

  if (isLoading) return <Skeleton className="h-96 w-full" />;
  if (error) return <p className="text-sm text-muted-foreground">Report not available for this month.</p>;

  const report = data?.data?.attributes;
  if (!report) return null;

  const sentimentItems = report.items.filter((i) => i.scale_type === "sentiment");
  const numericItems = report.items.filter((i) => i.scale_type === "numeric");
  const rangeItems = report.items.filter((i) => i.scale_type === "range");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">{formatMonth(month)} — Report</h2>
        <Button variant="outline" size="sm" onClick={onBackToCalendar}>Calendar</Button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card>
          <CardContent className="pt-4 text-center">
            <p className="text-2xl font-bold">{Math.round(report.summary.completion_rate * 100)}%</p>
            <p className="text-xs text-muted-foreground">Completion ({report.summary.total_filled}/{report.summary.total_expected})</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 text-center">
            <p className="text-2xl font-bold">{Math.round(report.summary.skip_rate * 100)}%</p>
            <p className="text-xs text-muted-foreground">Skip Rate ({report.summary.total_skipped}/{report.summary.total_expected})</p>
          </CardContent>
        </Card>
      </div>

      {sentimentItems.length > 0 && (
        <div className="space-y-3">
          <h3 className="font-semibold text-sm">Sentiment Items</h3>
          {sentimentItems.map((item) => <SentimentCard key={item.tracking_item_id} item={item} />)}
        </div>
      )}

      {numericItems.length > 0 && (
        <div className="space-y-3">
          <h3 className="font-semibold text-sm">Numeric Items</h3>
          {numericItems.map((item) => <NumericCard key={item.tracking_item_id} item={item} />)}
        </div>
      )}

      {rangeItems.length > 0 && (
        <div className="space-y-3">
          <h3 className="font-semibold text-sm">Range Items</h3>
          {rangeItems.map((item) => <RangeCard key={item.tracking_item_id} item={item} />)}
        </div>
      )}
    </div>
  );
}

function SentimentCard({ item }: { item: ReportItem }) {
  const stats = item.stats as SentimentStats;
  const total = stats.positive + stats.neutral + stats.negative;
  return (
    <Card>
      <CardHeader className="py-3 px-4">
        <CardTitle className="text-sm">{item.name}</CardTitle>
      </CardHeader>
      <CardContent className="py-2 px-4 space-y-2">
        <p className="text-sm">{stats.positive}/{total} positive ({Math.round(stats.positive_ratio * 100)}%)</p>
        <div className="flex gap-1 h-3 rounded overflow-hidden">
          {total > 0 && <>
            <div className="bg-green-500 transition-all" style={{ width: `${(stats.positive / total) * 100}%` }} />
            <div className="bg-yellow-500 transition-all" style={{ width: `${(stats.neutral / total) * 100}%` }} />
            <div className="bg-red-500 transition-all" style={{ width: `${(stats.negative / total) * 100}%` }} />
          </>}
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
          <span>😊 {stats.positive}</span>
          <span>😐 {stats.neutral}</span>
          <span>😞 {stats.negative}</span>
          {stats.skipped_days > 0 && <span>skipped {stats.skipped_days}</span>}
        </div>
      </CardContent>
    </Card>
  );
}

function NumericCard({ item }: { item: ReportItem }) {
  const stats = item.stats as NumericStats;
  const maxVal = stats.max?.count ?? 1;
  return (
    <Card>
      <CardHeader className="py-3 px-4">
        <CardTitle className="text-sm">{item.name}</CardTitle>
      </CardHeader>
      <CardContent className="py-2 px-4 space-y-2">
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
          <span>Total: <strong>{stats.total}</strong></span>
          <span>Avg: <strong>{stats.daily_average}</strong>/day</span>
          <span>{Math.round(stats.days_with_entries_above_zero_pct * 100)}% days &gt;0</span>
        </div>
        <div className="flex items-end gap-px h-8">
          {stats.daily_values.map((dv, i) => (
            <div key={i} className="flex-1 bg-primary/60 rounded-t" style={{ height: `${maxVal > 0 ? (dv.count / maxVal) * 100 : 0}%`, minHeight: dv.count > 0 ? "2px" : "0" }} />
          ))}
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
          {stats.min && <span>Min: {stats.min.count} ({new Date(stats.min.date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })})</span>}
          {stats.max && <span>Max: {stats.max.count} ({new Date(stats.max.date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })})</span>}
        </div>
      </CardContent>
    </Card>
  );
}

function RangeCard({ item }: { item: ReportItem }) {
  const stats = item.stats as RangeStats;
  return (
    <Card>
      <CardHeader className="py-3 px-4">
        <CardTitle className="text-sm">{item.name}</CardTitle>
      </CardHeader>
      <CardContent className="py-2 px-4 space-y-2">
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
          <span>Avg: <strong>{stats.average}</strong></span>
          <span>Std Dev: <strong>{stats.std_dev}</strong></span>
        </div>
        <div className="flex items-end gap-px h-8">
          {stats.daily_values.map((dv, i) => {
            const min = stats.min?.value ?? 0;
            const max = stats.max?.value ?? 100;
            const range = max - min || 1;
            return (
              <div key={i} className="flex-1 bg-primary/60 rounded-t" style={{ height: `${((dv.value - min) / range) * 100}%`, minHeight: "2px" }} />
            );
          })}
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
          {stats.min && <span>Min: {stats.min.value} ({new Date(stats.min.date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })})</span>}
          {stats.max && <span>Max: {stats.max.value} ({new Date(stats.max.date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })})</span>}
        </div>
      </CardContent>
    </Card>
  );
}

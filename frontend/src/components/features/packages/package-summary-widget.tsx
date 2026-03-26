import { usePackageSummary } from "@/lib/hooks/api/use-packages";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Package, Truck, AlertTriangle } from "lucide-react";

export function PackageSummaryWidget() {
  const { data, isLoading } = usePackageSummary();
  const summary = data?.data?.attributes;

  if (isLoading || !summary) return null;

  const hasPackages = summary.arrivingTodayCount > 0 || summary.inTransitCount > 0 || summary.exceptionCount > 0;
  if (!hasPackages) return null;

  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Arriving Today</CardTitle>
          <Truck className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{summary.arrivingTodayCount}</div>
          <p className="text-xs text-muted-foreground">
            package{summary.arrivingTodayCount !== 1 ? "s" : ""} expected
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">In Transit</CardTitle>
          <Package className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{summary.inTransitCount}</div>
          <p className="text-xs text-muted-foreground">
            being tracked
          </p>
        </CardContent>
      </Card>

      {summary.exceptionCount > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Exceptions</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{summary.exceptionCount}</div>
            <p className="text-xs text-muted-foreground">
              need attention
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

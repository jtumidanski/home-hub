import { Link } from "react-router-dom";
import { usePackageSummary } from "@/lib/hooks/api/use-packages";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Package, Truck, AlertTriangle, ChevronRight } from "lucide-react";

export function PackageSummaryWidget() {
  const { data, isLoading, isError } = usePackageSummary();
  const summary = data?.data?.attributes;

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-4 w-48" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="border-destructive">
        <CardContent className="py-3">
          <p className="text-sm text-destructive">Failed to load package summary</p>
        </CardContent>
      </Card>
    );
  }

  const hasPackages = summary && (summary.arrivingTodayCount > 0 || summary.inTransitCount > 0 || summary.exceptionCount > 0);

  if (!hasPackages) {
    return (
      <Link to="/app/packages" className="transition-opacity hover:opacity-80">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Packages</CardTitle>
            <CardAction>
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            </CardAction>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 text-muted-foreground">
              <Package className="h-5 w-5" />
              <p className="text-sm">No packages being tracked</p>
            </div>
          </CardContent>
        </Card>
      </Link>
    );
  }

  return (
    <Link to="/app/packages" className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader>
          <CardTitle className="text-sm font-medium">Packages</CardTitle>
          <CardAction>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <div className="flex items-center gap-2">
              <Truck className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm">{summary.arrivingTodayCount} arriving today</span>
            </div>
            <div className="flex items-center gap-2">
              <Package className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm">{summary.inTransitCount} in transit</span>
            </div>
            {summary.exceptionCount > 0 && (
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-destructive" />
                <span className="text-sm text-destructive">{summary.exceptionCount} exception{summary.exceptionCount !== 1 ? "s" : ""}</span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

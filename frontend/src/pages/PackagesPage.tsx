import { useState, useCallback } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { PackageCard } from "@/components/features/packages/package-card";
import { CreatePackageDialog } from "@/components/features/packages/create-package-dialog";
import { usePackages } from "@/lib/hooks/api/use-packages";
import type { Package } from "@/types/models/package";

function PackagesSkeleton() {
  return (
    <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-9 w-32" />
      </div>
      {Array.from({ length: 3 }).map((_, i) => (
        <Skeleton key={i} className="h-24" />
      ))}
    </div>
  );
}

export function PackagesPage() {
  const [showArchived, setShowArchived] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  const params = showArchived ? "filter[archived]=true" : undefined;
  const { data, isLoading, isError, refetch } = usePackages(params);

  const packages: Package[] = (data?.data ?? []) as Package[];

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  if (isLoading) return <PackagesSkeleton />;

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl md:text-2xl font-semibold">Packages</h1>
            <p className="text-sm text-muted-foreground">
              {packages.length} package{packages.length !== 1 ? "s" : ""}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant={showArchived ? "secondary" : "outline"}
              size="sm"
              onClick={() => setShowArchived(!showArchived)}
            >
              {showArchived ? "Hide Archived" : "Show Archived"}
            </Button>
            <Button size="sm" onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4 mr-1" />
              Add Package
            </Button>
          </div>
        </div>

        {isError && (
          <ErrorCard message="Failed to load packages. Try refreshing the page." />
        )}

        {packages.length === 0 && !isError && (
          <div className="text-center py-12 text-muted-foreground">
            <p className="text-lg">No packages yet</p>
            <p className="text-sm mt-1">Add a tracking number to get started.</p>
          </div>
        )}

        <div className="space-y-3">
          {packages.map((pkg) => (
            <PackageCard key={pkg.id} pkg={pkg} />
          ))}
        </div>

        <CreatePackageDialog
          open={showCreate}
          onClose={() => setShowCreate(false)}
        />
      </div>
    </PullToRefresh>
  );
}

import { useState, useCallback, useMemo } from "react";
import { Package as PackageIcon, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
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

  const packages: Package[] = useMemo(() => {
    const list = (data?.data ?? []) as Package[];
    return [...list].sort((a, b) => {
      const etaA = a.attributes.estimatedDelivery;
      const etaB = b.attributes.estimatedDelivery;
      if (!etaA && !etaB) return 0;
      if (!etaA) return 1;
      if (!etaB) return -1;
      return new Date(etaA).getTime() - new Date(etaB).getTime();
    });
  }, [data]);

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  if (isLoading) return <PackagesSkeleton />;

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <PageHeader
          title="Packages"
          description={`${packages.length} package${packages.length !== 1 ? "s" : ""}`}
          actions={
            <>
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
            </>
          }
        />

        {isError && (
          <ErrorCard message="Failed to load packages. Try refreshing the page." />
        )}

        {packages.length === 0 && !isError && (
          <EmptyState
            icon={PackageIcon}
            title="No packages yet"
            description="Add a tracking number to get started."
            action={
              <Button size="sm" onClick={() => setShowCreate(true)}>
                <Plus className="h-4 w-4 mr-1" />
                Add Package
              </Button>
            }
          />
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

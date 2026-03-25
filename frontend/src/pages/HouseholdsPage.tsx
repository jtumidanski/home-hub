import { useState } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { type Household } from "@/types/models/household";
import { HouseholdCard } from "@/components/features/households/household-card";
import { CreateHouseholdDialog } from "@/components/features/households/create-household-dialog";
import { ErrorCard } from "@/components/common/error-card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { data, isLoading, isError } = useHouseholds();
  const [open, setOpen] = useState(false);

  const households = (data?.data ?? []) as Household[];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load households. Try refreshing the page." />
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl md:text-2xl font-semibold">Households</h1>
        {canCreate && (
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />New Household
          </Button>
        )}
      </div>

      <CreateHouseholdDialog open={open} onOpenChange={setOpen} />

      {households.length === 0 && !isLoading ? (
        <div className="flex flex-col items-center justify-center py-12 text-center space-y-4">
          <p className="text-muted-foreground">No households yet.</p>
          {canCreate && (
            <Button variant="outline" onClick={() => setOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />Create First Household
            </Button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {households.map((household) => (
            <HouseholdCard
              key={household.id}
              household={household}
              isActive={household.id === activeId}
            />
          ))}
        </div>
      )}
    </div>
  );
}

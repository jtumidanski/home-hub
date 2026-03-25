import { useState } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { CreateHouseholdDialog } from "@/components/features/households/create-household-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, Home } from "lucide-react";

function HouseholdsPageSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-9 w-36" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    </div>
  );
}

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { data, isLoading, isError } = useHouseholds();
  const [open, setOpen] = useState(false);

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;

  if (isLoading) {
    return <HouseholdsPageSkeleton />;
  }

  if (isError) {
    return (
      <div className="p-6">
        <Card className="border-destructive">
          <CardContent className="py-3">
            <p className="text-sm text-destructive">
              Failed to load households. Try refreshing the page.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Households</h1>
        {canCreate && (
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />New Household
          </Button>
        )}
      </div>

      <CreateHouseholdDialog open={open} onOpenChange={setOpen} />

      {households.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground">No households yet.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {households.map((hh) => (
            <Card key={hh.id}>
              <CardContent className="flex items-center justify-between py-3">
                <div className="flex items-center gap-3">
                  <Home className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="font-medium">{hh.attributes.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {hh.attributes.timezone} &middot; {hh.attributes.units}
                    </p>
                  </div>
                </div>
                {hh.id === activeId && <Badge>Active</Badge>}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

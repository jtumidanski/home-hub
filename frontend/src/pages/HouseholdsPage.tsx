import { useState } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { type Household } from "@/types/models/household";
import { CreateHouseholdDialog } from "@/components/features/households/create-household-dialog";
import { DataTable } from "@/components/common/data-table";
import { ErrorCard } from "@/components/common/error-card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Home } from "lucide-react";

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { data, isLoading, isError } = useHouseholds();
  const [open, setOpen] = useState(false);

  const households = (data?.data ?? []) as Household[];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;

  const columns: ColumnDef<Household, unknown>[] = [
    {
      id: "icon",
      header: "",
      size: 40,
      cell: () => <Home className="h-5 w-5 text-muted-foreground" />,
    },
    {
      accessorKey: "attributes.name",
      header: "Name",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.attributes.name}</p>
          <p className="text-xs text-muted-foreground">
            {row.original.attributes.timezone} &middot; {row.original.attributes.units}
          </p>
        </div>
      ),
    },
    {
      id: "status",
      header: "",
      cell: ({ row }) =>
        row.original.id === activeId ? <Badge>Active</Badge> : null,
    },
  ];

  if (isLoading) {
    return (
      <div className="p-6 space-y-4" role="status" aria-label="Loading">
        <DataTable columns={columns} data={[]} isLoading skeletonRows={3} />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-6">
        <ErrorCard message="Failed to load households. Try refreshing the page." />
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
        <DataTable
          columns={columns}
          data={households}
          emptyMessage="No households yet."
        />
      )}
    </div>
  );
}

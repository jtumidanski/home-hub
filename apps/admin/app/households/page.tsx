"use client";

import { useEffect, useState, useMemo, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { listHouseholds, Household } from "@/lib/api/households";
import { DataGrid, ColumnDef } from "@/components/common/DataGrid";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { HouseholdFormModal } from "@/components/households/HouseholdFormModal";
import { HouseholdDetailModal } from "@/components/households/HouseholdDetailModal";
import { HouseholdDeleteDialog } from "@/components/households/HouseholdDeleteDialog";
import { AlertCircle, MoreVertical } from "lucide-react";

function HouseholdsContent() {
  const [households, setHouseholds] = useState<Household[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const searchParams = useSearchParams();
  const householdId = searchParams.get("householdId");

  // Modal state
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedHousehold, setSelectedHousehold] = useState<Household | null>(
    null
  );

  // Fetch households on mount
  useEffect(() => {
    fetchHouseholds();
  }, []);

  const fetchHouseholds = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await listHouseholds();
      setHouseholds(data);
    } catch (err) {
      console.error("Failed to fetch households:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  };

  // Filter households if householdId query param is present
  const filteredHouseholds = useMemo(() => {
    if (!householdId) return households;
    return households.filter((h) => h.id === householdId);
  }, [households, householdId]);

  // Action handlers
  const handleView = (household: Household) => {
    setSelectedHousehold(household);
    setDetailModalOpen(true);
  };

  const handleEdit = (household: Household) => {
    setSelectedHousehold(household);
    setEditModalOpen(true);
  };

  const handleDelete = (household: Household) => {
    setSelectedHousehold(household);
    setDeleteDialogOpen(true);
  };

  const handleRowClick = (household: Household) => {
    handleView(household);
  };

  const handleModalClose = () => {
    setCreateModalOpen(false);
    setEditModalOpen(false);
    setDetailModalOpen(false);
    setDeleteDialogOpen(false);
    setSelectedHousehold(null);
  };

  const handleSave = () => {
    fetchHouseholds();
    handleModalClose();
  };

  // Format date helper (consistent with users page)
  const formatDate = (dateString: string): string => {
    if (!dateString) return "—";

    const date = new Date(dateString);

    if (isNaN(date.getTime())) {
      console.error("Invalid date string:", dateString);
      return "Invalid Date";
    }

    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  // Column definitions
  const columns: ColumnDef<Household>[] = [
    {
      key: "name",
      header: "Name",
      accessor: (h) => h.name,
      sortable: true,
    },
    {
      key: "createdAt",
      header: "Created",
      accessor: (h) => h.createdAt,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "updatedAt",
      header: "Updated",
      accessor: (h) => h.updatedAt,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "actions",
      header: "",
      accessor: () => null,
      width: "80px",
      render: (_, household) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleView(household);
              }}
            >
              View
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleEdit(household);
              }}
            >
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleDelete(household);
              }}
              className="text-red-600 focus:text-red-600"
            >
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Header actions
  const headerActions = (
    <Button onClick={() => setCreateModalOpen(true)}>Create Household</Button>
  );

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Households</h1>
        <p className="text-neutral-600 dark:text-neutral-400 mt-2">
          View and manage households in the system
        </p>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Households Grid */}
      <DataGrid
        data={filteredHouseholds}
        columns={columns}
        onRowClick={handleRowClick}
        loading={loading}
        emptyMessage="No households found"
        getRowId={(h) => h.id}
        actions={headerActions}
      />

      {/* Modals */}
      <HouseholdFormModal
        open={createModalOpen}
        mode="create"
        onClose={handleModalClose}
        onSave={handleSave}
      />

      <HouseholdFormModal
        open={editModalOpen}
        mode="edit"
        household={selectedHousehold || undefined}
        onClose={handleModalClose}
        onSave={handleSave}
      />

      <HouseholdDetailModal
        household={selectedHousehold}
        open={detailModalOpen}
        onClose={handleModalClose}
        onUpdate={fetchHouseholds}
      />

      <HouseholdDeleteDialog
        household={selectedHousehold}
        open={deleteDialogOpen}
        onClose={handleModalClose}
        onDeleted={handleSave}
      />
    </div>
  );
}

export default function HouseholdsPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Households</h1>
            <p className="text-neutral-600 dark:text-neutral-400 mt-2">
              Loading households...
            </p>
          </div>
        </div>
      }
    >
      <HouseholdsContent />
    </Suspense>
  );
}

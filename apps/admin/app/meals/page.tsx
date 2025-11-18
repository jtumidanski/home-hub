"use client";

import { useEffect, useState } from "react";
import { listMeals, deleteMeal, Meal } from "@/lib/api/meals";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AlertCircle, MoreVertical, Plus, Trash2, Eye, Edit } from "lucide-react";
import { CreateMealDialog } from "@/components/meals/CreateMealDialog";
import { ViewMealDialog } from "@/components/meals/ViewMealDialog";
import { EditMealDialog } from "@/components/meals/EditMealDialog";

export default function MealsPage() {
  const [meals, setMeals] = useState<Meal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [mealToDelete, setMealToDelete] = useState<Meal | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Dialog state
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedMealId, setSelectedMealId] = useState<string | null>(null);

  // Fetch meals on mount
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const mealsData = await listMeals();
      setMeals(mealsData);
    } catch (err) {
      console.error("Failed to fetch meals:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch meals");
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteClick = (meal: Meal) => {
    setMealToDelete(meal);
    setDeleteConfirmOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!mealToDelete) return;

    try {
      setDeleting(true);
      await deleteMeal(mealToDelete.id);
      await fetchData(); // Refresh list
      setDeleteConfirmOpen(false);
      setMealToDelete(null);
    } catch (err) {
      console.error("Failed to delete meal:", err);
      setError(err instanceof Error ? err.message : "Failed to delete meal");
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteConfirmOpen(false);
    setMealToDelete(null);
  };

  // Dialog handlers
  const handleCreateClick = () => {
    setCreateDialogOpen(true);
  };

  const handleViewClick = (meal: Meal) => {
    setSelectedMealId(meal.id);
    setViewDialogOpen(true);
  };

  const handleEditClick = (meal: Meal) => {
    setSelectedMealId(meal.id);
    setEditDialogOpen(true);
  };

  const handleViewEdit = (mealId: string) => {
    setViewDialogOpen(false);
    setSelectedMealId(mealId);
    setEditDialogOpen(true);
  };

  const handleCreateSuccess = () => {
    fetchData(); // Refresh list
  };

  const handleEditSuccess = () => {
    fetchData(); // Refresh list
  };

  // Format date helper
  const formatDate = (dateString: string): string => {
    if (!dateString) return "—";

    const date = new Date(dateString);

    // Check if date is invalid
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

  // Truncate text helper
  const truncate = (text: string, maxLength: number): string => {
    if (!text) return "—";
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + "...";
  };

  // Column definitions
  const columns: ColumnDef<Meal>[] = [
    {
      key: "title",
      header: "Title",
      accessor: (meal) => meal.title,
      sortable: true,
    },
    {
      key: "description",
      header: "Description",
      accessor: (meal) => meal.description,
      render: (value) => truncate(value, 60),
      sortable: true,
    },
    {
      key: "createdAt",
      header: "Created",
      accessor: (meal) => meal.createdAt,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "actions",
      header: "",
      accessor: () => null,
      width: "80px",
      render: (_, meal) => (
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
                handleViewClick(meal);
              }}
            >
              <Eye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleEditClick(meal);
              }}
            >
              <Edit className="h-4 w-4 mr-2" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleDeleteClick(meal);
              }}
              className="text-red-600 dark:text-red-400"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Meals</h1>
          <p className="text-neutral-600 dark:text-neutral-400 mt-2">
            Manage meals and recipes with AI-powered ingredient parsing
          </p>
        </div>
        <Button onClick={handleCreateClick}>
          <Plus className="h-4 w-4 mr-2" />
          Add Meal
        </Button>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Meals Grid */}
      <DataGrid
        data={meals}
        columns={columns}
        loading={loading}
        emptyMessage="No meals found. Add a new meal to get started."
        getRowId={(meal) => meal.id}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Meal</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{mealToDelete?.title}</strong>?
              This action cannot be undone and will also delete all associated ingredients.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={handleDeleteCancel}
              disabled={deleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
              disabled={deleting}
            >
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Meal Dialog */}
      <CreateMealDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={handleCreateSuccess}
      />

      {/* View Meal Dialog */}
      <ViewMealDialog
        mealId={selectedMealId}
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        onEdit={handleViewEdit}
      />

      {/* Edit Meal Dialog */}
      <EditMealDialog
        mealId={selectedMealId}
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        onSuccess={handleEditSuccess}
      />
    </div>
  );
}

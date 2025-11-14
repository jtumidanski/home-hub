"use client";

import { useEffect, useState } from "react";
import {
  getDevices,
  deleteDevice,
  Device,
  formatDeviceType,
} from "@/lib/api/devices";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AlertCircle, MoreVertical, Plus, Trash2, Eye, Edit } from "lucide-react";
import Link from "next/link";
import { RegisterDeviceDialog } from "@/components/devices/RegisterDeviceDialog";
import { ViewDeviceDialog } from "@/components/devices/ViewDeviceDialog";
import { EditDeviceDialog } from "@/components/devices/EditDeviceDialog";

/** Device with resolved household name */
interface DeviceWithHousehold extends Device {
  householdName?: string;
}

export default function DevicesPage() {
  const [devices, setDevices] = useState<DeviceWithHousehold[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deviceToDelete, setDeviceToDelete] = useState<DeviceWithHousehold | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Dialog state for register, view, and edit
  const [registerDialogOpen, setRegisterDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);

  // Fetch devices and households on mount
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch devices and households in parallel
      const [devicesData, householdsData] = await Promise.all([
        getDevices(),
        listHouseholds(),
      ]);

      // Create household lookup map
      const householdMap = new Map<string, string>();
      householdsData.forEach((h: Household) => {
        householdMap.set(h.id, h.name);
      });

      // Enrich devices with household names
      const enrichedDevices: DeviceWithHousehold[] = devicesData.map((device) => ({
        ...device,
        householdName: householdMap.get(device.householdId),
      }));

      setDevices(enrichedDevices);
    } catch (err) {
      console.error("Failed to fetch data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteClick = (device: DeviceWithHousehold) => {
    setDeviceToDelete(device);
    setDeleteConfirmOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!deviceToDelete) return;

    try {
      setDeleting(true);
      await deleteDevice(deviceToDelete.id);
      await fetchData(); // Refresh list
      setDeleteConfirmOpen(false);
      setDeviceToDelete(null);
    } catch (err) {
      console.error("Failed to delete device:", err);
      setError(err instanceof Error ? err.message : "Failed to delete device");
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteConfirmOpen(false);
    setDeviceToDelete(null);
  };

  // Dialog handlers
  const handleRegisterClick = () => {
    setRegisterDialogOpen(true);
  };

  const handleViewClick = (device: DeviceWithHousehold) => {
    setSelectedDeviceId(device.id);
    setViewDialogOpen(true);
  };

  const handleEditClick = (device: DeviceWithHousehold) => {
    setSelectedDeviceId(device.id);
    setEditDialogOpen(true);
  };

  const handleViewEdit = (deviceId: string) => {
    setViewDialogOpen(false);
    setSelectedDeviceId(deviceId);
    setEditDialogOpen(true);
  };

  const handleRegisterSuccess = () => {
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

  // Column definitions
  const columns: ColumnDef<DeviceWithHousehold>[] = [
    {
      key: "name",
      header: "Name",
      accessor: (device) => device.name,
      sortable: true,
    },
    {
      key: "type",
      header: "Type",
      accessor: (device) => device.type,
      render: (value) => formatDeviceType(value),
      sortable: true,
    },
    {
      key: "householdName",
      header: "Household",
      accessor: (device) => device.householdName,
      render: (value, device) => {
        if (!value || !device.householdId) return "—";

        return (
          <Link
            href={`/households?householdId=${device.householdId}`}
            className="text-blue-600 hover:underline dark:text-blue-400"
            onClick={(e) => e.stopPropagation()}
          >
            {value}
          </Link>
        );
      },
      sortable: true,
    },
    {
      key: "createdAt",
      header: "Created",
      accessor: (device) => device.created_at,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "actions",
      header: "",
      accessor: () => null,
      width: "80px",
      render: (_, device) => (
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
                handleViewClick(device);
              }}
            >
              <Eye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleEditClick(device);
              }}
            >
              <Edit className="h-4 w-4 mr-2" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleDeleteClick(device);
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
          <h1 className="text-3xl font-bold tracking-tight">Devices</h1>
          <p className="text-neutral-600 dark:text-neutral-400 mt-2">
            Manage kiosks and other devices in your household
          </p>
        </div>
        <Button onClick={handleRegisterClick}>
          <Plus className="h-4 w-4 mr-2" />
          Register Device
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

      {/* Devices Grid */}
      <DataGrid
        data={devices}
        columns={columns}
        loading={loading}
        emptyMessage="No devices found. Register a new device to get started."
        getRowId={(device) => device.id}
        onRowClick={handleViewClick}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Device</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{deviceToDelete?.name}</strong>?
              This action cannot be undone and will also delete all device preferences.
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

      {/* Register Device Dialog */}
      <RegisterDeviceDialog
        open={registerDialogOpen}
        onOpenChange={setRegisterDialogOpen}
        onSuccess={handleRegisterSuccess}
      />

      {/* View Device Dialog */}
      <ViewDeviceDialog
        deviceId={selectedDeviceId}
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        onEdit={handleViewEdit}
      />

      {/* Edit Device Dialog */}
      <EditDeviceDialog
        deviceId={selectedDeviceId}
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        onSuccess={handleEditSuccess}
      />
    </div>
  );
}

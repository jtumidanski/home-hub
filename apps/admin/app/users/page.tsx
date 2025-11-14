"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { listUsers, User } from "@/lib/api/users";
import { listHouseholds, Household } from "@/lib/api/households";
import { DataGrid, ColumnDef } from "@/components/common/DataGrid";
import { UserViewModal } from "@/components/users/UserViewModal";
import { UserEditModal } from "@/components/users/UserEditModal";
import { UserHouseholdModal } from "@/components/users/UserHouseholdModal";
import { UserTasksModal } from "@/components/users/UserTasksModal";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AlertCircle, MoreVertical } from "lucide-react";

/** User with resolved household name */
interface UserWithHousehold extends User {
  householdName?: string;
}

export default function UsersPage() {
  const [users, setUsers] = useState<UserWithHousehold[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<UserWithHousehold | null>(
    null,
  );
  const [viewModalOpen, setViewModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [householdModalOpen, setHouseholdModalOpen] = useState(false);
  const [tasksModalOpen, setTasksModalOpen] = useState(false);

  // Fetch users and households on mount
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch users and households in parallel
      const [usersData, householdsData] = await Promise.all([
        listUsers(),
        listHouseholds(),
      ]);

      // Create household lookup map
      const householdMap = new Map<string, string>();
      householdsData.forEach((h: Household) => {
        householdMap.set(h.id, h.name);
      });

      // Enrich users with household names
      const enrichedUsers: UserWithHousehold[] = usersData.map((user) => ({
        ...user,
        householdName: user.householdId
          ? householdMap.get(user.householdId)
          : undefined,
      }));

      setUsers(enrichedUsers);
    } catch (err) {
      console.error("Failed to fetch data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  };

  const handleView = (user: UserWithHousehold) => {
    setSelectedUser(user);
    setViewModalOpen(true);
  };

  const handleEdit = (user: UserWithHousehold) => {
    setSelectedUser(user);
    setEditModalOpen(true);
  };

  const handleManageHousehold = (user: UserWithHousehold) => {
    setSelectedUser(user);
    setHouseholdModalOpen(true);
  };

  const handleManageTasks = (user: UserWithHousehold) => {
    setSelectedUser(user);
    setTasksModalOpen(true);
  };

  const handleModalClose = () => {
    setViewModalOpen(false);
    setEditModalOpen(false);
    setHouseholdModalOpen(false);
    setTasksModalOpen(false);
    setSelectedUser(null);
  };

  const handleModalSave = () => {
    // Refresh users after save
    fetchData();
    handleModalClose();
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
  const columns: ColumnDef<UserWithHousehold>[] = [
    {
      key: "displayName",
      header: "Display Name",
      accessor: (user) => user.displayName,
      sortable: true,
    },
    {
      key: "email",
      header: "Email",
      accessor: (user) => user.email,
      sortable: true,
    },
    {
      key: "householdName",
      header: "Household",
      accessor: (user) => user.householdName,
      render: (value, user) => {
        if (!value || !user.householdId) return "—";

        return (
          <Link
            href={`/households?householdId=${user.householdId}`}
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
      accessor: (user) => user.createdAt,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "updatedAt",
      header: "Updated",
      accessor: (user) => user.updatedAt,
      render: (value) => formatDate(value),
      sortable: true,
    },
    {
      key: "actions",
      header: "",
      accessor: () => null,
      width: "80px",
      render: (_, user) => (
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
                handleView(user);
              }}
            >
              View
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleEdit(user);
              }}
            >
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleManageHousehold(user);
              }}
            >
              Manage Household
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                handleManageTasks(user);
              }}
            >
              Manage Tasks
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Users</h1>
        <p className="text-neutral-600 dark:text-neutral-400 mt-2">
          View and manage users in the system
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

      {/* Users Grid */}
      <DataGrid
        data={users}
        columns={columns}
        loading={loading}
        emptyMessage="No users found"
        getRowId={(user) => user.id}
      />

      {/* User Modals */}
      {selectedUser && (
        <>
          <UserViewModal
            user={selectedUser}
            householdName={selectedUser.householdName}
            open={viewModalOpen}
            onClose={handleModalClose}
          />
          <UserEditModal
            user={selectedUser}
            householdName={selectedUser.householdName}
            open={editModalOpen}
            onClose={handleModalClose}
            onSave={handleModalSave}
          />
          <UserHouseholdModal
            user={selectedUser}
            currentHouseholdName={selectedUser.householdName}
            open={householdModalOpen}
            onClose={handleModalClose}
            onSave={handleModalSave}
          />
          <UserTasksModal
            user={selectedUser}
            open={tasksModalOpen}
            onClose={handleModalClose}
            onSave={handleModalSave}
          />
        </>
      )}
    </div>
  );
}

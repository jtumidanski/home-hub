"use client";

import { useEffect, useState } from "react";
import { listUsers, User } from "@/lib/api/users";
import { listHouseholds, Household } from "@/lib/api/households";
import { DataGrid, ColumnDef } from "@/components/common/DataGrid";
import { UserDetailModal } from "@/components/users/UserDetailModal";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";

/** User with resolved household name */
interface UserWithHousehold extends User {
  householdName?: string;
}

export default function UsersPage() {
  const [users, setUsers] = useState<UserWithHousehold[]>([]);
  const [households, setHouseholds] = useState<Map<string, string>>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<UserWithHousehold | null>(
    null,
  );
  const [modalOpen, setModalOpen] = useState(false);

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
      setHouseholds(householdMap);

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

  const handleRowClick = (user: UserWithHousehold) => {
    setSelectedUser(user);
    setModalOpen(true);
  };

  const handleModalClose = () => {
    setModalOpen(false);
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
      render: (value) => value || "—",
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
        onRowClick={handleRowClick}
        loading={loading}
        emptyMessage="No users found"
        getRowId={(user) => user.id}
      />

      {/* User Detail Modal */}
      {selectedUser && (
        <UserDetailModal
          user={selectedUser}
          householdName={selectedUser.householdName}
          open={modalOpen}
          onClose={handleModalClose}
          onSave={handleModalSave}
        />
      )}
    </div>
  );
}

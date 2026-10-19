import { useState } from "react";
import { ChevronDown, ChevronUp, RefreshCw, Archive, ArchiveRestore, Trash2, Lock, Unlock, Pencil, MoreHorizontal } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { StatusBadge } from "./status-badge";
import { CarrierIcon } from "./carrier-icon";
import { PackageDetail } from "./package-detail";
import { EditPackageDialog } from "./edit-package-dialog";
import {
  useRefreshPackage,
  useArchivePackage,
  useUnarchivePackage,
  useDeletePackage,
  useUpdatePackage,
} from "@/lib/hooks/api/use-packages";
import type { Package } from "@/types/models/package";

interface PackageCardProps {
  pkg: Package;
}

export function PackageCard({ pkg }: PackageCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const refreshMutation = useRefreshPackage();
  const archiveMutation = useArchivePackage();
  const unarchiveMutation = useUnarchivePackage();
  const deleteMutation = useDeletePackage();
  const updateMutation = useUpdatePackage();

  const { attributes: a } = pkg;
  const isRedacted = a.private && !a.isOwner;

  const truncatedTracking = a.trackingNumber
    ? a.trackingNumber.length > 16
      ? a.trackingNumber.slice(0, 8) + "..." + a.trackingNumber.slice(-4)
      : a.trackingNumber
    : null;

  const isArchived = a.status === "archived";
  const archiveToggleMutation = isArchived ? unarchiveMutation : archiveMutation;
  const archiveLabel = isArchived ? "Unarchive" : "Archive";
  const privateLabel = a.private ? "Make visible" : "Make private";
  const handleRefresh = () => refreshMutation.mutate(pkg.id);
  const handleTogglePrivate = () =>
    updateMutation.mutate({ id: pkg.id, attrs: { private: !a.private } });
  const handleArchiveToggle = () => archiveToggleMutation.mutate(pkg.id);

  return (
    <>
      <Card className="overflow-hidden">
        <CardContent className="p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-start gap-3 min-w-0 flex-1">
              <CarrierIcon carrier={a.carrier} />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium truncate">
                    {a.label ?? "Package"}
                  </span>
                  {a.status && <StatusBadge status={a.status} />}
                  {a.private && (
                    <Lock className="h-3 w-3 text-muted-foreground" aria-label="Private" role="img" />
                  )}
                </div>
                {!isRedacted && truncatedTracking && (
                  <p className="text-xs text-muted-foreground font-mono mt-0.5">
                    {truncatedTracking}
                  </p>
                )}
                <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                  {a.estimatedDelivery && (
                    <span>ETA: {a.estimatedDelivery}</span>
                  )}
                  {a.actualDelivery && (
                    <span>Delivered: {new Date(a.actualDelivery).toLocaleDateString()}</span>
                  )}
                </div>
                {!isRedacted && a.notes && (
                  <p className="text-xs text-muted-foreground mt-1 truncate">
                    {a.notes}
                  </p>
                )}
              </div>
            </div>

            <div className="flex items-center gap-1 shrink-0 justify-end">
              {a.isOwner && !isRedacted && (
                <>
                  {/* Desktop: inline icon buttons */}
                  <div className="hidden sm:flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleRefresh}
                      disabled={refreshMutation.isPending}
                      title="Refresh tracking"
                      aria-label="Refresh tracking"
                    >
                      <RefreshCw className={cn("h-3.5 w-3.5", refreshMutation.isPending && "animate-spin")} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setEditing(true)}
                      title="Edit"
                      aria-label="Edit"
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleTogglePrivate}
                      title={privateLabel}
                      aria-label={privateLabel}
                    >
                      {a.private ? <Unlock className="h-3.5 w-3.5" /> : <Lock className="h-3.5 w-3.5" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleArchiveToggle}
                      disabled={archiveToggleMutation.isPending}
                      title={archiveLabel}
                      aria-label={archiveLabel}
                    >
                      {isArchived ? <ArchiveRestore className="h-3.5 w-3.5" /> : <Archive className="h-3.5 w-3.5" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="ml-2 text-destructive hover:text-destructive"
                      onClick={() => setConfirmingDelete(true)}
                      disabled={deleteMutation.isPending}
                      title="Delete"
                      aria-label="Delete"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                  {/* Mobile: overflow menu */}
                  <div className="sm:hidden">
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={<Button variant="ghost" size="icon" aria-label="More actions" />}
                      >
                        <MoreHorizontal className="h-4 w-4" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuItem onClick={handleRefresh} disabled={refreshMutation.isPending}>
                          <RefreshCw /> Refresh tracking
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setEditing(true)}>
                          <Pencil /> Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={handleTogglePrivate}>
                          {a.private ? <Unlock /> : <Lock />} {privateLabel}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={handleArchiveToggle} disabled={archiveToggleMutation.isPending}>
                          {isArchived ? <ArchiveRestore /> : <Archive />} {archiveLabel}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          variant="destructive"
                          onClick={() => setConfirmingDelete(true)}
                          disabled={deleteMutation.isPending}
                        >
                          <Trash2 /> Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </>
              )}
              {!isRedacted && (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setExpanded(!expanded)}
                  title={expanded ? "Collapse" : "Expand"}
                  aria-label={expanded ? "Collapse" : "Expand"}
                >
                  {expanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                </Button>
              )}
            </div>
          </div>

          {expanded && !isRedacted && (
            <PackageDetail packageId={pkg.id} carrier={a.carrier} trackingNumber={a.trackingNumber} />
          )}
        </CardContent>
      </Card>

      {editing && (
        <EditPackageDialog
          pkg={pkg}
          open={editing}
          onClose={() => setEditing(false)}
        />
      )}

      <Dialog open={confirmingDelete} onOpenChange={setConfirmingDelete}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete package</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Delete &ldquo;{a.label ?? "this package"}&rdquo;? This cannot be undone.
          </p>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" onClick={() => setConfirmingDelete(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              disabled={deleteMutation.isPending}
              onClick={() => {
                deleteMutation.mutate(pkg.id, {
                  onSuccess: () => setConfirmingDelete(false),
                });
              }}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholdMembers, useUpdateMemberRole, useRemoveMember, useLeaveHousehold } from "@/lib/hooks/api/use-memberships";
import { useHouseholdInvitations, useRevokeInvitation } from "@/lib/hooks/api/use-invitations";
import { useUsersByIds } from "@/lib/hooks/api/use-users";
import { InviteMemberDialog } from "@/components/features/households/invite-member-dialog";
import { ErrorCard } from "@/components/common/error-card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { UserAvatar } from "@/components/ui/user-avatar";
import { ArrowLeft, Plus, UserMinus, LogOut, Mail, AlertTriangle, Clock, X } from "lucide-react";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { Membership } from "@/types/models/membership";
import { membershipRoleLabelMap } from "@/types/models/membership";
import { invitationRoleLabelMap } from "@/types/models/invitation";
import type { Invitation } from "@/types/models/invitation";
import type { User } from "@/types/models/user";

export function HouseholdMembersPage() {
  const { id: householdId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { appContext, user } = useAuth();
  const [inviteOpen, setInviteOpen] = useState(false);
  const [confirmRemove, setConfirmRemove] = useState<Membership | null>(null);
  const [confirmLeave, setConfirmLeave] = useState<Membership | null>(null);
  const [confirmRevoke, setConfirmRevoke] = useState<Invitation | null>(null);

  const { data: membersData, isLoading: membersLoading, isError: membersError } = useHouseholdMembers(householdId);
  const { data: invitationsData, isLoading: invitationsLoading } = useHouseholdInvitations(householdId);

  const members = (membersData?.data ?? []) as Membership[];
  const invitations = (invitationsData?.data ?? []) as Invitation[];

  const userIds = members.map((m) => m.relationships.user.data.id);
  const { data: usersData } = useUsersByIds(userIds);
  const users = (usersData?.data ?? []) as User[];

  const updateRole = useUpdateMemberRole();
  const removeMember = useRemoveMember();
  const leaveHousehold = useLeaveHousehold();
  const revokeInvitation = useRevokeInvitation();

  const resolvedRole = appContext?.attributes.resolvedRole;
  const isPrivileged = resolvedRole === "owner" || resolvedRole === "admin";
  const currentUserId = user?.id;

  const getUserForMember = (m: Membership): User | undefined =>
    users.find((u) => u.id === m.relationships.user.data.id);

  const getMyMembership = (): Membership | undefined =>
    members.find((m) => m.relationships.user.data.id === currentUserId);

  const handleRoleChange = async (membershipId: string, role: string) => {
    try {
      await updateRole.mutateAsync({ membershipId, role });
      toast.success("Role updated");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to update role").message);
    }
  };

  const handleRemove = async () => {
    if (!confirmRemove) return;
    try {
      await removeMember.mutateAsync(confirmRemove.id);
      toast.success("Member removed");
      setConfirmRemove(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to remove member").message);
    }
  };

  const handleLeave = async () => {
    if (!confirmLeave) return;
    try {
      await leaveHousehold.mutateAsync(confirmLeave.id);
      toast.success("You have left the household");
      setConfirmLeave(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to leave household").message);
    }
  };

  const handleRevoke = async () => {
    if (!confirmRevoke) return;
    try {
      await revokeInvitation.mutateAsync(confirmRevoke.id);
      toast.success("Invitation revoked");
      setConfirmRevoke(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to revoke invitation").message);
    }
  };

  if (membersLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    );
  }

  if (membersError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load household members." />
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div className="space-y-3">
        <Button variant="ghost" size="sm" onClick={() => navigate("/app/households")}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Households
        </Button>
        <h1 className="text-xl md:text-2xl font-semibold">Household Members</h1>
      </div>

      {/* Members */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle className="text-lg">Members</CardTitle>
          <div className="flex gap-2">
            {(() => {
              const myMembership = getMyMembership();
              if (myMembership) {
                const isLastOwner = myMembership.attributes.isLastOwner;
                return (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => isLastOwner ? toast.error("You are the sole owner. Transfer ownership before leaving.") : setConfirmLeave(myMembership)}
                  >
                    <LogOut className="mr-2 h-4 w-4" />Leave
                  </Button>
                );
              }
              return null;
            })()}
            {isPrivileged && householdId && (
              <Button size="sm" onClick={() => setInviteOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />Invite
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {members.map((member) => {
            const u = getUserForMember(member);
            const isSelf = member.relationships.user.data.id === currentUserId;
            const isTargetOwner = member.attributes.role === "owner";
            const canModify = isPrivileged && !isSelf && !(resolvedRole === "admin" && isTargetOwner);

            return (
              <div key={member.id} className="flex items-center justify-between gap-3 rounded-md border p-3">
                {u && (
                  <UserAvatar
                    avatarUrl={u.attributes.avatarUrl}
                    providerAvatarUrl={u.attributes.providerAvatarUrl}
                    displayName={u.attributes.displayName}
                    userId={u.id}
                    size="md"
                  />
                )}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="font-medium truncate">
                      {u?.attributes.displayName ?? "Unknown User"}
                    </p>
                    {isSelf && <Badge variant="outline">You</Badge>}
                    {member.attributes.isLastOwner && (
                      <Badge variant="destructive" className="gap-1">
                        <AlertTriangle className="h-3 w-3" />Sole Owner
                      </Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground truncate">{u?.attributes.email}</p>
                  <p className="text-xs text-muted-foreground">
                    Joined {new Date(member.attributes.createdAt).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {canModify ? (
                    <Select
                      value={member.attributes.role}
                      onValueChange={(role) => { if (role) handleRoleChange(member.id, role); }}
                    >
                      <SelectTrigger className="w-28">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {Object.entries(membershipRoleLabelMap).map(([value, label]) => (
                          <SelectItem key={value} value={value}>{label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  ) : (
                    <Badge variant="secondary">{membershipRoleLabelMap[member.attributes.role]}</Badge>
                  )}
                  {canModify && (
                    <Button variant="ghost" size="icon" onClick={() => setConfirmRemove(member)}>
                      <UserMinus className="h-4 w-4 text-destructive" />
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </CardContent>
      </Card>

      {/* Pending Invitations */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Pending Invitations</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {invitationsLoading ? (
            <Skeleton className="h-12" />
          ) : invitations.length === 0 ? (
            <p className="text-sm text-muted-foreground">No pending invitations.</p>
          ) : (
            invitations.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between gap-3 rounded-md border p-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <Mail className="h-4 w-4 text-muted-foreground shrink-0" />
                    <p className="font-medium truncate">{inv.attributes.email}</p>
                  </div>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant="secondary">{invitationRoleLabelMap[inv.attributes.role]}</Badge>
                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      Expires {new Date(inv.attributes.expiresAt).toLocaleDateString()}
                    </span>
                  </div>
                </div>
                {isPrivileged && (
                  <Button variant="ghost" size="icon" onClick={() => setConfirmRevoke(inv)}>
                    <X className="h-4 w-4 text-destructive" />
                  </Button>
                )}
              </div>
            ))
          )}
        </CardContent>
      </Card>

      {/* Dialogs */}
      {householdId && (
        <InviteMemberDialog open={inviteOpen} onOpenChange={setInviteOpen} householdId={householdId} />
      )}

      <Dialog open={!!confirmRemove} onOpenChange={() => setConfirmRemove(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Member</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove this member from the household? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmRemove(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleRemove} disabled={removeMember.isPending}>
              Remove
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!confirmLeave} onOpenChange={() => setConfirmLeave(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Leave Household</DialogTitle>
            <DialogDescription>
              Are you sure you want to leave this household? You will lose access to all shared data.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmLeave(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleLeave} disabled={leaveHousehold.isPending}>
              Leave
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!confirmRevoke} onOpenChange={() => setConfirmRevoke(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke Invitation</DialogTitle>
            <DialogDescription>
              Are you sure you want to revoke this invitation? The user will no longer be able to accept it.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmRevoke(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleRevoke} disabled={revokeInvitation.isPending}>
              Revoke
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

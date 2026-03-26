import { useState } from "react";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { useMyInvitations, useAcceptInvitation, useDeclineInvitation } from "@/lib/hooks/api/use-invitations";
import { type Household } from "@/types/models/household";
import { type Invitation } from "@/types/models/invitation";
import { invitationRoleLabelMap } from "@/types/models/invitation";
import { HouseholdCard } from "@/components/features/households/household-card";
import { CreateHouseholdDialog } from "@/components/features/households/create-household-dialog";
import { ErrorCard } from "@/components/common/error-card";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Check, X, Clock } from "lucide-react";

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { data, isLoading, isError } = useHouseholds();
  const [open, setOpen] = useState(false);

  const { data: myInvitationsData } = useMyInvitations();
  const acceptInvitation = useAcceptInvitation();
  const declineInvitation = useDeclineInvitation();

  const households = (data?.data ?? []) as Household[];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;
  const myInvitations = (myInvitationsData?.data ?? []) as Invitation[];
  const includedHouseholds = ((myInvitationsData as { included?: Household[] })?.included ?? []) as Household[];

  const handleAccept = async (id: string) => {
    try {
      await acceptInvitation.mutateAsync(id);
      toast.success("Invitation accepted");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to accept invitation").message);
    }
  };

  const handleDecline = async (id: string) => {
    try {
      await declineInvitation.mutateAsync(id);
      toast.success("Invitation declined");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to decline invitation").message);
    }
  };

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

      {myInvitations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Pending Invitations</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {myInvitations.map((inv) => {
              const hh = includedHouseholds.find(
                (h) => h.id === inv.relationships.household.data.id,
              );
              return (
                <div key={inv.id} className="flex items-center justify-between gap-3 rounded-md border p-3">
                  <div className="min-w-0 flex-1">
                    <p className="font-medium">{hh?.attributes.name ?? "Unknown Household"}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="secondary">{invitationRoleLabelMap[inv.attributes.role]}</Badge>
                      <span className="text-xs text-muted-foreground flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        Expires {new Date(inv.attributes.expiresAt).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={() => handleAccept(inv.id)}
                      disabled={acceptInvitation.isPending}
                    >
                      <Check className="mr-1 h-4 w-4" />Accept
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleDecline(inv.id)}
                      disabled={declineInvitation.isPending}
                    >
                      <X className="mr-1 h-4 w-4" />Decline
                    </Button>
                  </div>
                </div>
              );
            })}
          </CardContent>
        </Card>
      )}

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

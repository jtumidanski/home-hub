import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useCreateTenant, useOnboardingCreateHousehold } from "@/lib/hooks/api/use-context";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { useMyInvitations, useAcceptInvitation, useDeclineInvitation } from "@/lib/hooks/api/use-invitations";
import { createTenantSchema, type CreateTenantFormData } from "@/lib/schemas/tenant.schema";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";
import type { Invitation } from "@/types/models/invitation";
import { invitationRoleLabelMap } from "@/types/models/invitation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Loader2, Check, X, Clock } from "lucide-react";

export function OnboardingPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [step, setStep] = useState<"invitations" | "tenant" | "household">("invitations");
  const [createdTenant, setCreatedTenant] = useState<Tenant | null>(null);
  const createTenant = useCreateTenant();
  const createHousehold = useOnboardingCreateHousehold();
  const logout = useLogout();

  const { data: myInvitationsData, isLoading: invitationsLoading } = useMyInvitations();
  const acceptInvitation = useAcceptInvitation();
  const declineInvitation = useDeclineInvitation();

  const myInvitations = (myInvitationsData?.data ?? []) as Invitation[];
  const includedHouseholds = ((myInvitationsData as { included?: Household[] })?.included ?? []) as Household[];

  // Auto-advance past invitations step if no pending invitations
  const effectiveStep = step === "invitations" && !invitationsLoading && myInvitations.length === 0
    ? "tenant"
    : step;

  const handleAcceptInvitation = async (id: string) => {
    try {
      await acceptInvitation.mutateAsync(id);
      toast.success("Welcome! You've joined the household.");
      navigate("/app");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to accept invitation").message);
    }
  };

  const handleDeclineAll = async () => {
    try {
      for (const inv of myInvitations) {
        await declineInvitation.mutateAsync(inv.id);
      }
      setStep("tenant");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to decline invitations").message);
    }
  };

  const tenantForm = useForm<CreateTenantFormData>({
    resolver: zodResolver(createTenantSchema),
    defaultValues: { name: user?.attributes.displayName ? `${user.attributes.displayName}'s Home` : "" },
  });

  const householdForm = useForm<CreateHouseholdFormData>({
    resolver: zodResolver(createHouseholdSchema),
    defaultValues: {
      ...createHouseholdDefaults,
      name: "Main Home",
    },
  });

  const onTenantSubmit = async (data: CreateTenantFormData) => {
    try {
      const result = await createTenant.mutateAsync(data.name);
      setCreatedTenant(result.data);
      toast.success("Account created");
      setStep("household");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create account").message);
    }
  };

  const onHouseholdSubmit = async (data: CreateHouseholdFormData) => {
    if (!createdTenant) return;
    try {
      await createHousehold.mutateAsync({
        tenant: createdTenant,
        name: data.name,
        timezone: data.timezone,
        units: data.units,
      });
      toast.success("Household created");
      navigate("/app");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create household").message);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome to Home Hub</CardTitle>
          <CardDescription>
            {effectiveStep === "invitations"
              ? "You have pending invitations"
              : effectiveStep === "tenant"
                ? "Let's set up your account"
                : "Now create your first household"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {effectiveStep === "invitations" && !invitationsLoading && myInvitations.length > 0 && (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                You've been invited to join a household. Accept an invitation to get started, or create your own.
              </p>
              <div className="space-y-3">
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
                      <Button
                        size="sm"
                        onClick={() => handleAcceptInvitation(inv.id)}
                        disabled={acceptInvitation.isPending}
                      >
                        <Check className="mr-1 h-4 w-4" />Accept
                      </Button>
                    </div>
                  );
                })}
              </div>
              <div className="flex gap-2 pt-2">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={handleDeclineAll}
                  disabled={declineInvitation.isPending}
                >
                  <X className="mr-1 h-4 w-4" />Decline All
                </Button>
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => setStep("tenant")}
                >
                  Create My Own
                </Button>
              </div>
            </div>
          )}

          {effectiveStep === "tenant" && (
            <Form {...tenantForm}>
              <form onSubmit={tenantForm.handleSubmit(onTenantSubmit)} className="space-y-4">
                <FormField
                  control={tenantForm.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Account Name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g., The Smith Family" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button type="submit" className="w-full" disabled={tenantForm.formState.isSubmitting}>
                  {tenantForm.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Continue
                </Button>
              </form>
            </Form>
          )}

          {effectiveStep === "household" && (
            <Form {...householdForm}>
              <form onSubmit={householdForm.handleSubmit(onHouseholdSubmit)} className="space-y-4">
                <FormField
                  control={householdForm.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Household Name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g., Main Home" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={householdForm.control}
                  name="timezone"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Timezone</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={householdForm.control}
                  name="units"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Units</FormLabel>
                      <FormControl>
                        <RadioGroup value={field.value} onValueChange={field.onChange}>
                          <RadioGroupItem value="imperial">Imperial</RadioGroupItem>
                          <RadioGroupItem value="metric">Metric</RadioGroupItem>
                        </RadioGroup>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button type="submit" className="w-full" disabled={householdForm.formState.isSubmitting}>
                  {householdForm.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Get Started
                </Button>
              </form>
            </Form>
          )}
        </CardContent>
        <CardFooter className="justify-center">
          <Button
            variant="link"
            size="sm"
            className="text-muted-foreground"
            onClick={() => logout.mutate()}
          >
            Sign out
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

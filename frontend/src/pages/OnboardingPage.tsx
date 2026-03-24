import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useAuth } from "@/components/providers/auth-provider";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const tenantSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

const householdSchema = z.object({
  name: z.string().min(1, "Name is required"),
  timezone: z.string().min(1, "Timezone is required"),
  units: z.enum(["imperial", "metric"]),
});

type TenantForm = z.infer<typeof tenantSchema>;
type HouseholdForm = z.infer<typeof householdSchema>;

export function OnboardingPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [step, setStep] = useState<"tenant" | "household">("tenant");
  const [error, setError] = useState<string | null>(null);

  const tenantForm = useForm<TenantForm>({
    resolver: zodResolver(tenantSchema),
    defaultValues: { name: user?.attributes.displayName ? `${user.attributes.displayName}'s Home` : "" },
  });

  const householdForm = useForm<HouseholdForm>({
    resolver: zodResolver(householdSchema),
    defaultValues: {
      name: "Main Home",
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      units: "imperial",
    },
  });

  const onTenantSubmit = async (data: TenantForm) => {
    try {
      setError(null);
      await accountService.createTenant(data.name);
      setStep("household");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to create tenant");
    }
  };

  const onHouseholdSubmit = async (data: HouseholdForm) => {
    try {
      setError(null);
      await accountService.createHousehold(data.name, data.timezone, data.units);
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      navigate("/app");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to create household");
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome to Home Hub</CardTitle>
          <CardDescription>
            {step === "tenant" ? "Let's set up your account" : "Now create your first household"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && (
            <div className="mb-4 rounded border border-destructive p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {step === "tenant" && (
            <form onSubmit={tenantForm.handleSubmit(onTenantSubmit)} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="tenant-name" className="text-sm font-medium">
                  Account Name
                </label>
                <Input
                  id="tenant-name"
                  placeholder="e.g., The Smith Family"
                  {...tenantForm.register("name")}
                />
                {tenantForm.formState.errors.name && (
                  <p className="text-sm text-destructive">{tenantForm.formState.errors.name.message}</p>
                )}
              </div>
              <Button type="submit" className="w-full" disabled={tenantForm.formState.isSubmitting}>
                {tenantForm.formState.isSubmitting ? "Creating..." : "Continue"}
              </Button>
            </form>
          )}

          {step === "household" && (
            <form onSubmit={householdForm.handleSubmit(onHouseholdSubmit)} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="hh-name" className="text-sm font-medium">
                  Household Name
                </label>
                <Input
                  id="hh-name"
                  placeholder="e.g., Main Home"
                  {...householdForm.register("name")}
                />
                {householdForm.formState.errors.name && (
                  <p className="text-sm text-destructive">{householdForm.formState.errors.name.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <label htmlFor="hh-timezone" className="text-sm font-medium">
                  Timezone
                </label>
                <Input
                  id="hh-timezone"
                  {...householdForm.register("timezone")}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Units</label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <input type="radio" value="imperial" {...householdForm.register("units")} />
                    <span className="text-sm">Imperial</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input type="radio" value="metric" {...householdForm.register("units")} />
                    <span className="text-sm">Metric</span>
                  </label>
                </div>
              </div>
              <Button type="submit" className="w-full" disabled={householdForm.formState.isSubmitting}>
                {householdForm.formState.isSubmitting ? "Creating..." : "Get Started"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

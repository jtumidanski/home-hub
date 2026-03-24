import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { createTenantSchema, type CreateTenantFormData } from "@/lib/schemas/tenant.schema";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { getErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Loader2 } from "lucide-react";

export function OnboardingPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [step, setStep] = useState<"tenant" | "household">("tenant");

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
      await accountService.createTenant(data.name);
      toast.success("Account created");
      setStep("household");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create account"));
    }
  };

  const onHouseholdSubmit = async (data: CreateHouseholdFormData) => {
    try {
      await accountService.createHousehold(data.name, data.timezone, data.units);
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      toast.success("Household created");
      navigate("/app");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create household"));
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
          {step === "tenant" && (
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

          {step === "household" && (
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
                      <div className="flex gap-4">
                        <label className="flex items-center gap-2">
                          <input
                            type="radio"
                            value="imperial"
                            checked={field.value === "imperial"}
                            onChange={() => field.onChange("imperial")}
                          />
                          <span className="text-sm">Imperial</span>
                        </label>
                        <label className="flex items-center gap-2">
                          <input
                            type="radio"
                            value="metric"
                            checked={field.value === "metric"}
                            onChange={() => field.onChange("metric")}
                          />
                          <span className="text-sm">Metric</span>
                        </label>
                      </div>
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
      </Card>
    </div>
  );
}

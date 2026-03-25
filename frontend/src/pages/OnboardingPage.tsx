import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useCreateTenant, useOnboardingCreateHousehold } from "@/lib/hooks/api/use-context";
import { createTenantSchema, type CreateTenantFormData } from "@/lib/schemas/tenant.schema";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Loader2 } from "lucide-react";

export function OnboardingPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [step, setStep] = useState<"tenant" | "household">("tenant");
  const [createdTenant, setCreatedTenant] = useState<Tenant | null>(null);
  const createTenant = useCreateTenant();
  const createHousehold = useOnboardingCreateHousehold();

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
      </Card>
    </div>
  );
}

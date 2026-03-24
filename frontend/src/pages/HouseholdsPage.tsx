import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useHouseholds, householdKeys } from "@/lib/hooks/api/use-households";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { getErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, Home, Loader2 } from "lucide-react";

function HouseholdsPageSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-9 w-36" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    </div>
  );
}

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { tenantId } = useTenant();
  const { data, isLoading } = useHouseholds();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const form = useForm<CreateHouseholdFormData>({
    resolver: zodResolver(createHouseholdSchema),
    defaultValues: createHouseholdDefaults,
  });

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;

  const onSubmit = async (values: CreateHouseholdFormData) => {
    if (!tenantId) return;
    try {
      await accountService.createHousehold(tenantId, values.name, values.timezone, values.units);
      await queryClient.invalidateQueries({ queryKey: householdKeys.list(tenantId) });
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      toast.success("Household created");
      form.reset(createHouseholdDefaults);
      setOpen(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create household"));
    }
  };

  if (isLoading) {
    return <HouseholdsPageSkeleton />;
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Households</h1>
        {canCreate && (
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger>
              <Button size="sm"><Plus className="mr-2 h-4 w-4" />New Household</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Household</DialogTitle>
              </DialogHeader>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter household name" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
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
                    control={form.control}
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
                  <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
                    {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    Create Household
                  </Button>
                </form>
              </Form>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {households.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground">No households yet.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {households.map((hh) => (
            <Card key={hh.id}>
              <CardContent className="flex items-center justify-between py-3">
                <div className="flex items-center gap-3">
                  <Home className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="font-medium">{hh.attributes.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {hh.attributes.timezone} &middot; {hh.attributes.units}
                    </p>
                  </div>
                </div>
                {hh.id === activeId && <Badge>Active</Badge>}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

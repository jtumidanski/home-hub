import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds, householdKeys } from "@/lib/hooks/api/use-households";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Plus, Home } from "lucide-react";

const householdSchema = z.object({
  name: z.string().min(1, "Name is required"),
  timezone: z.string().min(1, "Timezone is required"),
  units: z.enum(["imperial", "metric"]),
});

type HouseholdForm = z.infer<typeof householdSchema>;

export function HouseholdsPage() {
  const { appContext } = useAuth();
  const { data, isLoading } = useHouseholds();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const form = useForm<HouseholdForm>({
    resolver: zodResolver(householdSchema),
    defaultValues: {
      name: "",
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      units: "imperial",
    },
  });

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const canCreate = appContext?.attributes.canCreateHousehold;

  const onSubmit = async (values: HouseholdForm) => {
    await accountService.createHousehold(values.name, values.timezone, values.units);
    await queryClient.invalidateQueries({ queryKey: householdKeys.list });
    await queryClient.invalidateQueries({ queryKey: contextKeys.current });
    form.reset();
    setOpen(false);
  };

  if (isLoading) {
    return <div className="p-6"><div className="h-8 w-48 animate-pulse rounded bg-muted" /></div>;
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
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input id="name" {...form.register("name")} />
                  {form.formState.errors.name && (
                    <p className="text-sm text-destructive">{form.formState.errors.name.message}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="timezone">Timezone</Label>
                  <Input id="timezone" {...form.register("timezone")} />
                </div>
                <div className="space-y-2">
                  <Label>Units</Label>
                  <div className="flex gap-4">
                    <label className="flex items-center gap-2">
                      <input type="radio" value="imperial" {...form.register("units")} />
                      <span className="text-sm">Imperial</span>
                    </label>
                    <label className="flex items-center gap-2">
                      <input type="radio" value="metric" {...form.register("units")} />
                      <span className="text-sm">Metric</span>
                    </label>
                  </div>
                </div>
                <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
                  {form.formState.isSubmitting ? "Creating..." : "Create Household"}
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>

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
    </div>
  );
}

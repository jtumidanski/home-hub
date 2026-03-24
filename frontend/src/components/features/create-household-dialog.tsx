import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { useTenant } from "@/context/tenant-context";
import { accountService } from "@/services/api/account";
import { useInvalidateHouseholds } from "@/lib/hooks/api/use-households";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { getErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Loader2 } from "lucide-react";

interface CreateHouseholdDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateHouseholdDialog({ open, onOpenChange }: CreateHouseholdDialogProps) {
  const { tenantId } = useTenant();
  const queryClient = useQueryClient();
  const invalidateHouseholds = useInvalidateHouseholds();

  const form = useForm<CreateHouseholdFormData>({
    resolver: zodResolver(createHouseholdSchema),
    defaultValues: createHouseholdDefaults,
  });

  const handleOpenChange = (next: boolean) => {
    if (form.formState.isSubmitting) return;
    if (!next) form.reset(createHouseholdDefaults);
    onOpenChange(next);
  };

  const onSubmit = async (values: CreateHouseholdFormData) => {
    if (!tenantId) return;
    try {
      await accountService.createHousehold(tenantId, values.name, values.timezone, values.units);
      invalidateHouseholds.invalidateLists();
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      toast.success("Household created");
      form.reset(createHouseholdDefaults);
      onOpenChange(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create household"));
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
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
  );
}

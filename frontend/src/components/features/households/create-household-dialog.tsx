import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useCreateHousehold } from "@/lib/hooks/api/use-households";
import { createHouseholdSchema, type CreateHouseholdFormData, createHouseholdDefaults } from "@/lib/schemas/household.schema";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Loader2 } from "lucide-react";

interface CreateHouseholdDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateHouseholdDialog({ open, onOpenChange }: CreateHouseholdDialogProps) {
  const createHousehold = useCreateHousehold();

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
    try {
      await createHousehold.mutateAsync(values);
      toast.success("Household created");
      form.reset(createHouseholdDefaults);
      onOpenChange(false);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create household").message);
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

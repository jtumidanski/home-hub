import { z } from "zod";

export const packageFormSchema = z.object({
  trackingNumber: z.string().min(1, "Tracking number is required").max(64, "Tracking number must be 64 characters or fewer"),
  carrier: z.enum(["usps", "ups", "fedex"], { message: "Carrier is required" }),
  label: z.string().max(255, "Label must be 255 characters or fewer").optional(),
  notes: z.string().optional(),
  private: z.boolean().optional(),
});

export type PackageFormData = z.infer<typeof packageFormSchema>;

export const packageFormDefaults: PackageFormData = {
  trackingNumber: "",
  carrier: "usps",
  label: "",
  notes: "",
  private: false,
};

export const packageEditSchema = z.object({
  label: z.string().max(255, "Label must be 255 characters or fewer").optional(),
  notes: z.string().optional(),
  carrier: z.enum(["usps", "ups", "fedex"]).optional(),
  private: z.boolean().optional(),
});

export type PackageEditData = z.infer<typeof packageEditSchema>;

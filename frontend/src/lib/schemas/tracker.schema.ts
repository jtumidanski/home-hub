import { z } from "zod";

export const trackerFormSchema = z.object({
  name: z.string().min(1, "Name is required").max(100, "Name must be 100 characters or fewer"),
  scale_type: z.enum(["sentiment", "numeric", "range"], { message: "Scale type is required" }),
  color: z.string().min(1, "Color is required"),
  schedule: z.array(z.number().min(0).max(6)),
  sort_order: z.number().min(0).optional(),
  range_min: z.number().optional(),
  range_max: z.number().optional(),
}).refine(
  (data) => {
    if (data.scale_type === "range") {
      return data.range_min !== undefined && data.range_max !== undefined && data.range_min < data.range_max;
    }
    return true;
  },
  { message: "Range min must be less than max", path: ["range_max"] }
);

export type TrackerFormData = z.infer<typeof trackerFormSchema>;

export const trackerFormDefaults: TrackerFormData = {
  name: "",
  scale_type: "sentiment",
  color: "blue",
  schedule: [],
  sort_order: 0,
  range_min: 0,
  range_max: 100,
};

export const trackerEditSchema = z.object({
  name: z.string().min(1, "Name is required").max(100).optional(),
  color: z.string().optional(),
  schedule: z.array(z.number().min(0).max(6)).optional(),
  sort_order: z.number().min(0).optional(),
  range_min: z.number().optional(),
  range_max: z.number().optional(),
});

export type TrackerEditData = z.infer<typeof trackerEditSchema>;

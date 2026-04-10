// Workout-tracker domain types. The wire format is JSON:API but a few of the
// composite endpoints (week, today, summary) hand-roll their own document
// shapes — those types live further down in this file rather than going
// through the JSON:API resource wrapper.

export type WorkoutKind = "strength" | "isometric" | "cardio";
export type WeightType = "free" | "bodyweight";
export type WeightUnit = "lb" | "kg";
export type DistanceUnit = "mi" | "km" | "m";
export type PerformanceStatus = "pending" | "done" | "skipped" | "partial";
export type PerformanceMode = "summary" | "per_set";

export interface ThemeAttributes {
  name: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface Theme {
  id: string;
  type: "themes";
  attributes: ThemeAttributes;
}

export interface RegionAttributes {
  name: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface Region {
  id: string;
  type: "regions";
  attributes: RegionAttributes;
}

export interface ExerciseDefaults {
  sets?: number | null;
  reps?: number | null;
  weight?: number | null;
  weightUnit?: WeightUnit | null;
  durationSeconds?: number | null;
  distance?: number | null;
  distanceUnit?: DistanceUnit | null;
}

export interface ExerciseAttributes {
  name: string;
  kind: WorkoutKind;
  weightType: WeightType;
  themeId: string;
  regionId: string;
  secondaryRegionIds: string[];
  defaults: ExerciseDefaults;
  notes?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Exercise {
  id: string;
  type: "exercises";
  attributes: ExerciseAttributes;
}

// --- Composite week document (matches `weekview` package projection) -----

export interface PlannedShape {
  sets?: number | null;
  reps?: number | null;
  weight?: number | null;
  weightUnit?: WeightUnit | null;
  durationSeconds?: number | null;
  distance?: number | null;
  distanceUnit?: DistanceUnit | null;
}

export interface ActualsShape {
  sets?: number | null;
  reps?: number | null;
  weight?: number | null;
  durationSeconds?: number | null;
  distance?: number | null;
  distanceUnit?: DistanceUnit | null;
}

export interface PerformanceSet {
  setNumber: number;
  reps: number;
  weight: number;
}

export interface PerformanceProjection {
  status: PerformanceStatus;
  mode: PerformanceMode;
  weightUnit?: WeightUnit | null;
  actuals?: ActualsShape | null;
  sets?: PerformanceSet[] | null;
  notes?: string | null;
  updatedAt?: string | null;
}

export interface WeekItem {
  id: string;
  dayOfWeek: number;
  position: number;
  exerciseId: string;
  exerciseName: string;
  exerciseDeleted: boolean;
  kind: WorkoutKind;
  weightType: WeightType;
  planned: PlannedShape;
  performance?: PerformanceProjection | null;
  notes?: string | null;
}

export interface WeekDocument {
  data: {
    type: "weeks";
    id: string;
    attributes: {
      weekStartDate: string;
      restDayFlags: number[];
      items: WeekItem[];
    };
  };
}

export interface TodayDocument {
  data: {
    type: "today";
    id: string;
    attributes: {
      date: string;
      weekStartDate: string;
      dayOfWeek: number;
      isRestDay: boolean;
      items: WeekItem[];
    };
  };
}

export interface SummaryDocument {
  data: {
    type: "week-summaries";
    id: string;
    attributes: {
      weekStartDate: string;
      restDayFlags: number[];
      totalPlannedItems: number;
      totalPerformedItems: number;
      totalSkippedItems: number;
      byDay: Array<{
        dayOfWeek: number;
        isRestDay: boolean;
        items: Array<{
          itemId: string;
          exerciseName: string;
          status: PerformanceStatus;
          planned: Record<string, unknown>;
          actualSummary: Record<string, unknown> | null;
        }>;
      }>;
      byTheme: Array<{
        themeId: string;
        themeName: string;
        itemCount: number;
        strengthVolume: { value: number; unit: WeightUnit } | null;
        cardio: { totalDurationSeconds: number; totalDistance: { value: number; unit: DistanceUnit } } | null;
      }>;
      byRegion: Array<{
        regionId: string;
        regionName: string;
        itemCount: number;
        strengthVolume: { value: number; unit: WeightUnit } | null;
        cardio: { totalDurationSeconds: number; totalDistance: { value: number; unit: DistanceUnit } } | null;
      }>;
    };
  };
}

export const DAYS_OF_WEEK_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] as const;

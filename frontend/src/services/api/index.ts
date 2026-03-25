// Base
export { BaseService } from "./base";
export type { ValidationError } from "./base";

// Singleton instances
export { authService } from "./auth";
export { accountService } from "./account";
export { productivityService } from "./productivity";

// Types re-exported per service
export type { AuthProvider } from "./auth";
export type { Tenant, TenantAttributes } from "@/types/models/tenant";
export type { Household, HouseholdAttributes, HouseholdUpdateAttributes } from "@/types/models/household";
export type { Task, TaskAttributes, TaskUpdateAttributes } from "@/types/models/task";
export type { Reminder, ReminderAttributes, ReminderUpdateAttributes } from "@/types/models/reminder";
export type { Preference, PreferenceAttributes } from "@/types/models/preference";
export type { AppContext, ContextAttributes } from "@/types/models/context";
export type { User, UserAttributes } from "@/types/models/user";

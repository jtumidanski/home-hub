import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const ingredientKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["ingredients", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  lists: (tenant: Tenant | null, household: Household | null) =>
    [...ingredientKeys.all(tenant, household), "list"] as const,
  details: (tenant: Tenant | null, household: Household | null) =>
    [...ingredientKeys.all(tenant, household), "detail"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...ingredientKeys.details(tenant, household), id] as const,
  recipes: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...ingredientKeys.detail(tenant, household, id), "recipes"] as const,
};

export const categoryKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["ingredient-categories", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  lists: (tenant: Tenant | null, household: Household | null) =>
    [...categoryKeys.all(tenant, household), "list"] as const,
};

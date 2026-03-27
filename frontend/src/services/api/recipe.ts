import { BaseService } from "./base";
import type {
  RecipeListItem,
  RecipeDetail,
  RecipeCreateAttributes,
  RecipeUpdateAttributes,
  RecipeTag,
  RecipeParseResult,
  RecipeIngredient,
} from "@/types/models/recipe";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse } from "@/types/api/responses";
import { api } from "@/lib/api/client";

interface RecipeListParams {
  search?: string | undefined;
  tags?: string[] | undefined;
  page?: number | undefined;
  pageSize?: number | undefined;
  plannerReady?: boolean | undefined;
  classification?: string | undefined;
  normalizationStatus?: string | undefined;
}

export interface RecipeListResponse extends ApiListResponse<RecipeListItem> {
  meta?: {
    total: number;
    page: number;
    pageSize: number;
  };
}

class RecipeService extends BaseService {
  constructor() {
    super("/recipes");
  }

  listRecipes(tenant: Tenant, params?: RecipeListParams): Promise<RecipeListResponse> {
    this.setTenant(tenant);
    const query = new URLSearchParams();
    if (params?.search) query.set("search", params.search);
    if (params?.tags) {
      for (const tag of params.tags) {
        query.append("tag", tag);
      }
    }
    if (params?.plannerReady !== undefined) query.set("plannerReady", String(params.plannerReady));
    if (params?.classification) query.set("classification", params.classification);
    if (params?.normalizationStatus) query.set("normalizationStatus", params.normalizationStatus);
    if (params?.page) query.set("page[number]", String(params.page));
    if (params?.pageSize) query.set("page[size]", String(params.pageSize));
    const qs = query.toString();
    return api.get<RecipeListResponse>(`/recipes${qs ? `?${qs}` : ""}`);
  }

  getRecipe(tenant: Tenant, id: string) {
    return this.getOne<RecipeDetail>(tenant, `/recipes/${id}`);
  }

  createRecipe(tenant: Tenant, attrs: RecipeCreateAttributes) {
    return this.create<RecipeDetail>(tenant, "/recipes", {
      data: { type: "recipes", attributes: attrs },
    });
  }

  async updateRecipe(tenant: Tenant, id: string, attrs: RecipeUpdateAttributes) {
    return this.update<RecipeDetail>(tenant, `/recipes/${id}`, {
      data: { type: "recipes", id, attributes: attrs },
    });
  }

  deleteRecipe(tenant: Tenant, id: string) {
    return this.remove(tenant, `/recipes/${id}`);
  }

  restoreRecipe(tenant: Tenant, recipeId: string) {
    return this.create<RecipeDetail>(tenant, "/recipes/restorations", {
      data: {
        type: "recipe-restorations",
        attributes: { recipeId },
      },
    });
  }

  listTags(tenant: Tenant) {
    return this.getList<RecipeTag>(tenant, "/recipes/tags");
  }

  parseSource(tenant: Tenant, source: string) {
    return this.create<RecipeParseResult>(tenant, "/recipes/parse", {
      data: {
        type: "recipe-parse",
        attributes: { source },
      },
    });
  }

  resolveIngredient(tenant: Tenant, recipeId: string, ingredientId: string, canonicalIngredientId: string, saveAsAlias: boolean) {
    return this.create<RecipeIngredient>(tenant, `/recipes/${recipeId}/ingredients/${ingredientId}/resolve`, {
      data: {
        type: "ingredient-resolutions",
        attributes: { canonicalIngredientId, saveAsAlias },
      },
    });
  }

  renormalize(tenant: Tenant, recipeId: string) {
    this.setTenant(tenant);
    return api.post(`/recipes/${recipeId}/renormalize`, {});
  }
}

export const recipeService = new RecipeService();

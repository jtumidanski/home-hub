import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";

export interface ValidationError {
  field: string;
  message: string;
}

export class BaseService {
  protected basePath: string;

  constructor(basePath: string) {
    this.basePath = basePath;
  }

  protected setTenant(tenant: { id: string }) {
    api.setTenant(tenant);
  }

  protected validate<T>(_data: T): ValidationError[] {
    return [];
  }

  protected transformResponse<T>(data: T): T {
    return data;
  }

  protected getList<T>(tenant: { id: string }, path?: string): Promise<JsonApiListResponse<T>> {
    this.setTenant(tenant);
    return api.get<JsonApiListResponse<T>>(path ?? this.basePath);
  }

  protected getOne<T>(tenant: { id: string }, path: string): Promise<JsonApiResponse<T>> {
    this.setTenant(tenant);
    return api.get<JsonApiResponse<T>>(path);
  }

  protected create<T>(tenant: { id: string }, path: string, body: unknown): Promise<JsonApiResponse<T>> {
    this.setTenant(tenant);
    return api.post<JsonApiResponse<T>>(path, body);
  }

  protected update<T>(tenant: { id: string }, path: string, body: unknown): Promise<JsonApiResponse<T>> {
    this.setTenant(tenant);
    return api.patch<JsonApiResponse<T>>(path, body);
  }

  protected remove(tenant: { id: string }, path: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(path);
  }
}
